package commands

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
)

type RepayOrderCommand struct {
	OrderID  string
	UserID   string
	ClientIP string
	AgainLink string
	Actor    domain.Actor
}

type RepayOrderResult struct {
	Order   queries.OrderDTO
	Payment queries.PaymentDTO
}

type RepayOrderHandler struct {
	orders        domain.OrderRepository
	payments      domain.PaymentRepository
	publicAPIBase string
}

func NewRepayOrderHandler(
	orders domain.OrderRepository,
	payments domain.PaymentRepository,
	publicAPIBase string,
) *RepayOrderHandler {
	return &RepayOrderHandler{
		orders:        orders,
		payments:      payments,
		publicAPIBase: strings.TrimRight(strings.TrimSpace(publicAPIBase), "/"),
	}
}

func (h *RepayOrderHandler) Handle(_ context.Context, cmd RepayOrderCommand) (RepayOrderResult, error) {
	oid, err := domain.ParseOrderID(cmd.OrderID)
	if err != nil {
		return RepayOrderResult{}, err
	}
	order, err := h.orders.FindByID(oid)
	if err != nil {
		return RepayOrderResult{}, err
	}
	if strings.TrimSpace(order.UserID) != strings.TrimSpace(cmd.UserID) {
		return RepayOrderResult{}, domain.ErrOrderNotFound
	}
	if !order.CanRepay() {
		return RepayOrderResult{}, domain.ErrOrderNotRepayable
	}

	method := order.PaymentMethod
	settings, err := h.payments.GetSettings()
	if err != nil {
		return RepayOrderResult{}, err
	}
	gateway, err := settings.GatewayFor(method)
	if err != nil {
		return RepayOrderResult{}, err
	}
	if !gateway.Enabled {
		return RepayOrderResult{}, domain.ErrOnePayDisabled
	}
	returnURL := strings.TrimSpace(settings.OnePayReturnURL)
	if returnURL == "" && h.publicAPIBase != "" {
		returnURL = h.publicAPIBase + "/api/v1/payments/onepay/return"
	}
	if !gateway.Ready(returnURL) {
		return RepayOrderResult{}, domain.ErrOnePayNotConfigured
	}

	actor := cmd.Actor
	if strings.TrimSpace(actor.Role) == "" {
		actor.Role = "user"
	}
	if strings.TrimSpace(actor.DisplayName) == "" {
		actor.DisplayName = actor.Email
	}

	// Include sibling orders that share the same failed/pending payment and are still repayable.
	targets := []domain.Order{order}
	seen := map[domain.OrderID]bool{order.ID: true}
	if order.PaymentID != "" {
		if prev, err := h.payments.FindByID(order.PaymentID); err == nil {
			for _, sid := range prev.OrderIDs {
				if seen[sid] {
					continue
				}
				sibling, err := h.orders.FindByID(sid)
				if err != nil {
					continue
				}
				if sibling.UserID != order.UserID || !sibling.CanRepay() {
					continue
				}
				targets = append(targets, sibling)
				seen[sid] = true
			}
		}
	}

	var totalCents int64
	orderIDs := make([]domain.OrderID, 0, len(targets))
	codes := make([]string, 0, len(targets))
	currency := order.Currency
	for i := range targets {
		if err := targets[i].PrepareRepayment(actor); err != nil {
			return RepayOrderResult{}, err
		}
		totalCents += targets[i].TotalCents
		orderIDs = append(orderIDs, targets[i].ID)
		codes = append(codes, targets[i].Code)
		if targets[i].Currency != "" {
			currency = targets[i].Currency
		}
	}

	payment, err := domain.NewPayment(order.UserID, method, totalCents, currency, orderIDs)
	if err != nil {
		return RepayOrderResult{}, err
	}
	if err := h.payments.Save(payment); err != nil {
		return RepayOrderResult{}, err
	}

	for i := range targets {
		targets[i].AttachPayment(payment.ID, domain.PaymentStatusPending)
		if err := h.orders.Save(targets[i]); err != nil {
			return RepayOrderResult{}, err
		}
	}

	redirectURL, err := infrastructure.BuildOnePayRedirectURL(infrastructure.OnePayCheckoutInput{
		Gateway:     gateway,
		ReturnURL:   returnURL,
		AgainLink:   strings.TrimSpace(cmd.AgainLink),
		Title:       "VPC 3-Party",
		MerchTxnRef: payment.MerchTxnRef,
		AmountCents: payment.AmountCents,
		OrderInfo:   "DH " + strings.Join(codes, " "),
		ClientIP:    cmd.ClientIP,
		Method:      method,
	})
	if err != nil {
		return RepayOrderResult{}, err
	}

	saved, err := h.orders.FindByID(order.ID)
	if err != nil {
		return RepayOrderResult{}, err
	}
	return RepayOrderResult{
		Order: queries.ToDTOWithHistory(saved, true),
		Payment: queries.PaymentDTO{
			ID:          string(payment.ID),
			Method:      string(payment.Method),
			MethodLabel: payment.Method.LabelVI(),
			Status:      string(payment.Status),
			StatusLabel: payment.Status.LabelVI(),
			AmountCents: payment.AmountCents,
			Currency:    payment.Currency,
			MerchTxnRef: payment.MerchTxnRef,
			PaymentURL:  redirectURL,
		},
	}, nil
}
