package commands

import (
	"context"
	"net/url"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/infrastructure"
)

type HandleOnePayCallbackCommand struct {
	Params     map[string]string
	Channel    domain.PaymentCallbackChannel
	HTTPMethod string
}

type HandleOnePayCallbackResult struct {
	Payment queries.PaymentDTO
	Paid    bool
	OrderID string
}

type HandleOnePayCallbackHandler struct {
	payments domain.PaymentRepository
	orders   domain.OrderRepository
}

func NewHandleOnePayCallbackHandler(payments domain.PaymentRepository, orders domain.OrderRepository) *HandleOnePayCallbackHandler {
	return &HandleOnePayCallbackHandler{payments: payments, orders: orders}
}

func (h *HandleOnePayCallbackHandler) Handle(_ context.Context, cmd HandleOnePayCallbackCommand) (result HandleOnePayCallbackResult, err error) {
	params := cmd.Params
	if params == nil {
		params = map[string]string{}
	}
	channel := cmd.Channel
	if channel == "" {
		channel = domain.PaymentCallbackChannelIPN
	}

	ref := strings.TrimSpace(params["vpc_MerchTxnRef"])
	respCode := strings.TrimSpace(params["vpc_TxnResponseCode"])
	message := strings.TrimSpace(params["vpc_Message"])
	var knownPayment domain.Payment
	var hasPayment bool

	defer func() {
		paymentID := domain.PaymentID("")
		method := domain.PaymentMethod("")
		paid := result.Paid
		success := err == nil
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		if result.Payment.ID != "" {
			paymentID = domain.PaymentID(result.Payment.ID)
			method, _ = domain.ParsePaymentMethod(result.Payment.Method)
		} else if hasPayment {
			paymentID = knownPayment.ID
			method = knownPayment.Method
		}
		_ = h.payments.SaveCallbackEvent(domain.NewPaymentCallbackEvent(domain.NewPaymentCallbackEventInput{
			Provider:      domain.PaymentProviderOnePay,
			Channel:       channel,
			HTTPMethod:    cmd.HTTPMethod,
			PaymentID:     paymentID,
			PaymentMethod: method,
			MerchTxnRef:   ref,
			ResponseCode:  respCode,
			Message:       message,
			Paid:          paid,
			Success:       success,
			ErrorMessage:  errMsg,
			RawPayload:    params,
		}))
	}()

	payment, err := h.payments.FindByMerchTxnRef(ref)
	if err != nil {
		return HandleOnePayCallbackResult{}, err
	}
	knownPayment = payment
	hasPayment = true

	settings, err := h.payments.GetSettings()
	if err != nil {
		return HandleOnePayCallbackResult{}, err
	}
	gateway, err := settings.GatewayFor(payment.Method)
	if err != nil {
		return HandleOnePayCallbackResult{}, err
	}
	if err := infrastructure.VerifyOnePaySecureHash(gateway.HashSecret, params); err != nil {
		return HandleOnePayCallbackResult{}, err
	}

	orderID := ""
	if len(payment.OrderIDs) > 0 {
		orderID = string(payment.OrderIDs[0])
	}
	actor := domain.SystemActor()

	// Idempotent: already paid — still sync order status if needed.
	if payment.Status == domain.PaymentStatusPaid {
		if err := h.syncOrders(payment, true, actor); err != nil {
			return HandleOnePayCallbackResult{}, err
		}
		return HandleOnePayCallbackResult{
			Payment: queries.ToPaymentDTO(payment, ""),
			Paid:    true,
			OrderID: orderID,
		}, nil
	}

	txnNo := strings.TrimSpace(params["vpc_TransactionNo"])
	paid := infrastructure.OnePayResponseSuccess(respCode)

	// Already failed with same outcome — repair order cancellation if needed.
	if payment.Status == domain.PaymentStatusFailed && !paid {
		if err := h.syncOrders(payment, false, actor); err != nil {
			return HandleOnePayCallbackResult{}, err
		}
		return HandleOnePayCallbackResult{
			Payment: queries.ToPaymentDTO(payment, ""),
			Paid:    false,
			OrderID: orderID,
		}, nil
	}

	if paid {
		payment.MarkPaid(txnNo, respCode, message)
	} else {
		payment.MarkFailed(respCode, message)
	}
	if err := h.payments.Save(payment); err != nil {
		return HandleOnePayCallbackResult{}, err
	}

	if err := h.syncOrders(payment, paid, actor); err != nil {
		return HandleOnePayCallbackResult{}, err
	}

	return HandleOnePayCallbackResult{
		Payment: queries.ToPaymentDTO(payment, ""),
		Paid:    paid,
		OrderID: orderID,
	}, nil
}

func (h *HandleOnePayCallbackHandler) syncOrders(payment domain.Payment, paid bool, actor domain.Actor) error {
	for _, oid := range payment.OrderIDs {
		order, err := h.orders.FindByID(oid)
		if err != nil {
			return err
		}
		if paid {
			if err := order.ApplyPaymentSuccess(actor); err != nil {
				return err
			}
		} else {
			if err := order.ApplyPaymentFailure(actor); err != nil {
				return err
			}
		}
		if strings.TrimSpace(string(order.PaymentID)) == "" {
			order.AttachPayment(payment.ID, order.PaymentStatus)
		} else {
			order.PaymentID = payment.ID
		}
		if err := h.orders.Save(order); err != nil {
			return err
		}
	}
	return nil
}

func ParamsFromURLValues(values url.Values) map[string]string {
	out := make(map[string]string, len(values))
	for k, vs := range values {
		if len(vs) == 0 {
			continue
		}
		out[k] = vs[0]
	}
	return out
}

type OnePayGatewayInput struct {
	Enabled    bool
	MerchantID string
	AccessCode string
	HashSecret string
	PaymentURL string
}

type UpdatePaymentSettingsCommand struct {
	OnePayReturnURL     string
	OnePayIPNURL        string
	OnePayDomestic      OnePayGatewayInput
	OnePayInternational OnePayGatewayInput
}

type UpdatePaymentSettingsHandler struct {
	payments      domain.PaymentRepository
	publicAPIBase string
}

func NewUpdatePaymentSettingsHandler(payments domain.PaymentRepository, publicAPIBase string) *UpdatePaymentSettingsHandler {
	return &UpdatePaymentSettingsHandler{
		payments:      payments,
		publicAPIBase: strings.TrimRight(strings.TrimSpace(publicAPIBase), "/"),
	}
}

func mergeGateway(input OnePayGatewayInput, current domain.OnePayGatewaySettings, defaultURL string) domain.OnePayGatewaySettings {
	hashSecret := strings.TrimSpace(input.HashSecret)
	if hashSecret == "" {
		hashSecret = current.HashSecret
	}
	paymentURL := strings.TrimSpace(input.PaymentURL)
	if paymentURL == "" {
		paymentURL = defaultURL
	}
	return domain.OnePayGatewaySettings{
		Enabled:    input.Enabled,
		MerchantID: strings.TrimSpace(input.MerchantID),
		AccessCode: strings.TrimSpace(input.AccessCode),
		HashSecret: hashSecret,
		PaymentURL: paymentURL,
	}
}

func (h *UpdatePaymentSettingsHandler) Handle(_ context.Context, cmd UpdatePaymentSettingsCommand) (queries.PaymentSettingsDTO, error) {
	current, err := h.payments.GetSettings()
	if err != nil {
		return queries.PaymentSettingsDTO{}, err
	}

	returnURL := strings.TrimSpace(cmd.OnePayReturnURL)
	if returnURL == "" && h.publicAPIBase != "" {
		returnURL = h.publicAPIBase + "/api/v1/payments/onepay/return"
	}
	ipnURL := strings.TrimSpace(cmd.OnePayIPNURL)
	if ipnURL == "" && h.publicAPIBase != "" {
		ipnURL = h.publicAPIBase + "/api/v1/payments/onepay/ipn"
	}

	settings := domain.PaymentSettings{
		OnePayReturnURL:     returnURL,
		OnePayIPNURL:        ipnURL,
		OnePayDomestic:      mergeGateway(cmd.OnePayDomestic, current.OnePayDomestic, infrastructure.DefaultOnePayDomesticPaymentURL),
		OnePayInternational: mergeGateway(cmd.OnePayInternational, current.OnePayInternational, infrastructure.DefaultOnePayInternationalPaymentURL),
	}
	if err := h.payments.SaveSettings(settings); err != nil {
		return queries.PaymentSettingsDTO{}, err
	}
	saved, err := h.payments.GetSettings()
	if err != nil {
		return queries.PaymentSettingsDTO{}, err
	}
	return queries.ToPaymentSettingsDTO(saved, h.publicAPIBase), nil
}
