package commands

import (
	"context"
	"fmt"
	"strings"

	catalogdomain "github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
)

type CreateOrderItemInput struct {
	ProductID string
	Quantity  int
}

type CreateOrderCommand struct {
	UserID          string
	Note            string
	ShippingName    string
	ShippingPhone   string
	ShippingAddress string
	PaymentMethod   string
	ClientIP        string
	AgainLink       string
	Items           []CreateOrderItemInput
	Actor           domain.Actor
}

type CreateOrderResult struct {
	Orders  []queries.OrderDTO
	Payment *queries.PaymentDTO
}

type CreateOrderHandler struct {
	orders        domain.OrderRepository
	products      catalogdomain.ProductRepository
	payments      domain.PaymentRepository
	publicAPIBase string
}

func NewCreateOrderHandler(
	orders domain.OrderRepository,
	products catalogdomain.ProductRepository,
	payments domain.PaymentRepository,
	publicAPIBase string,
) *CreateOrderHandler {
	return &CreateOrderHandler{
		orders:        orders,
		products:      products,
		payments:      payments,
		publicAPIBase: strings.TrimRight(strings.TrimSpace(publicAPIBase), "/"),
	}
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) (CreateOrderResult, error) {
	userID := strings.TrimSpace(cmd.UserID)
	if userID == "" {
		return CreateOrderResult{}, domain.ErrUserRequired
	}
	if len(cmd.Items) == 0 {
		return CreateOrderResult{}, domain.ErrEmptyOrderItems
	}
	method, err := domain.ParsePaymentMethod(cmd.PaymentMethod)
	if err != nil {
		return CreateOrderResult{}, err
	}

	type lineAgg struct {
		product  catalogdomain.Product
		quantity int
	}
	byProduct := map[catalogdomain.ProductID]*lineAgg{}
	for _, item := range cmd.Items {
		pid, err := catalogdomain.ParseProductID(item.ProductID)
		if err != nil {
			return CreateOrderResult{}, err
		}
		if item.Quantity <= 0 {
			return CreateOrderResult{}, domain.ErrInvalidOrderQuantity
		}
		if existing, ok := byProduct[pid]; ok {
			existing.quantity += item.Quantity
			continue
		}
		product, err := h.products.FindByID(pid)
		if err != nil {
			return CreateOrderResult{}, err
		}
		byProduct[pid] = &lineAgg{product: product, quantity: item.Quantity}
	}

	byMerchant := map[string][]domain.OrderLineInput{}
	merchantCurrency := map[string]string{}
	for _, agg := range byProduct {
		if err := agg.product.Reserve(agg.quantity); err != nil {
			return CreateOrderResult{}, err
		}
		mid := agg.product.MerchantID
		byMerchant[mid] = append(byMerchant[mid], domain.OrderLineInput{
			ProductID:      string(agg.product.ID),
			ProductName:    agg.product.Name,
			MerchantID:     mid,
			UnitPriceCents: agg.product.Price.AmountCents,
			Quantity:       agg.quantity,
		})
		merchantCurrency[mid] = agg.product.Price.Currency
	}

	for _, agg := range byProduct {
		if err := h.products.Save(agg.product); err != nil {
			return CreateOrderResult{}, fmt.Errorf("save product stock: %w", err)
		}
	}

	actor := cmd.Actor
	if strings.TrimSpace(actor.Role) == "" {
		actor.Role = "user"
	}
	if strings.TrimSpace(actor.DisplayName) == "" {
		actor.DisplayName = actor.Email
	}

	createdOrders := make([]domain.Order, 0, len(byMerchant))
	var totalCents int64
	currency := "VND"
	for merchantID, lines := range byMerchant {
		order, err := domain.NewOrder(
			userID, merchantID, merchantCurrency[merchantID], cmd.Note,
			cmd.ShippingName, cmd.ShippingPhone, cmd.ShippingAddress, method, lines,
		)
		if err != nil {
			return CreateOrderResult{}, err
		}
		order.RecordCreated(actor)
		if err := h.orders.Save(order); err != nil {
			return CreateOrderResult{}, err
		}
		if marker, ok := h.products.(interface {
			MarkOrdered(id catalogdomain.ProductID)
		}); ok {
			for _, line := range lines {
				marker.MarkOrdered(catalogdomain.ProductID(line.ProductID))
			}
		}
		saved, err := h.orders.FindByID(order.ID)
		if err != nil {
			return CreateOrderResult{}, err
		}
		createdOrders = append(createdOrders, saved)
		totalCents += saved.TotalCents
		if saved.Currency != "" {
			currency = saved.Currency
		}
	}

	result := CreateOrderResult{
		Orders: make([]queries.OrderDTO, 0, len(createdOrders)),
	}
	for _, o := range createdOrders {
		result.Orders = append(result.Orders, queries.ToDTOWithHistory(o, true))
	}

	if method == domain.PaymentMethodCOD {
		result.Payment = &queries.PaymentDTO{
			Method:      string(domain.PaymentMethodCOD),
			MethodLabel: domain.PaymentMethodCOD.LabelVI(),
			Status:      string(domain.PaymentStatusUnpaid),
			StatusLabel: domain.PaymentStatusUnpaid.LabelVI(),
			AmountCents: totalCents,
			Currency:    currency,
		}
		return result, nil
	}

	if !method.IsOnePay() {
		return CreateOrderResult{}, domain.ErrInvalidPaymentMethod
	}

	settings, err := h.payments.GetSettings()
	if err != nil {
		return CreateOrderResult{}, err
	}
	gateway, err := settings.GatewayFor(method)
	if err != nil {
		return CreateOrderResult{}, err
	}
	if !gateway.Enabled {
		return CreateOrderResult{}, domain.ErrOnePayDisabled
	}
	returnURL := strings.TrimSpace(settings.OnePayReturnURL)
	if returnURL == "" && h.publicAPIBase != "" {
		returnURL = h.publicAPIBase + "/api/v1/payments/onepay/return"
	}
	if !gateway.Ready(returnURL) {
		return CreateOrderResult{}, domain.ErrOnePayNotConfigured
	}

	orderIDs := make([]domain.OrderID, 0, len(createdOrders))
	for _, o := range createdOrders {
		orderIDs = append(orderIDs, o.ID)
	}
	payment, err := domain.NewPayment(userID, method, totalCents, currency, orderIDs)
	if err != nil {
		return CreateOrderResult{}, err
	}
	if err := h.payments.Save(payment); err != nil {
		return CreateOrderResult{}, err
	}

	for _, o := range createdOrders {
		o.AttachPayment(payment.ID, domain.PaymentStatusPending)
		if err := h.orders.Save(o); err != nil {
			return CreateOrderResult{}, err
		}
	}

	codes := make([]string, 0, len(createdOrders))
	for _, o := range createdOrders {
		codes = append(codes, o.Code)
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
		return CreateOrderResult{}, err
	}

	result.Orders = result.Orders[:0]
	for _, oid := range orderIDs {
		saved, err := h.orders.FindByID(oid)
		if err != nil {
			return CreateOrderResult{}, err
		}
		result.Orders = append(result.Orders, queries.ToDTOWithHistory(saved, true))
	}
	result.Payment = &queries.PaymentDTO{
		ID:          string(payment.ID),
		Method:      string(payment.Method),
		MethodLabel: payment.Method.LabelVI(),
		Status:      string(payment.Status),
		StatusLabel: payment.Status.LabelVI(),
		AmountCents: payment.AmountCents,
		Currency:    payment.Currency,
		MerchTxnRef: payment.MerchTxnRef,
		PaymentURL:  redirectURL,
	}
	return result, nil
}
