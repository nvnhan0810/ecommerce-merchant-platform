package queries

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type PaymentCallbackEventDTO struct {
	ID                  string          `json:"id"`
	Provider            string          `json:"provider"`
	ProviderLabel       string          `json:"provider_label"`
	Channel             string          `json:"channel"`
	ChannelLabel        string          `json:"channel_label"`
	HTTPMethod          string          `json:"http_method"`
	PaymentID           string          `json:"payment_id,omitempty"`
	PaymentMethod       string          `json:"payment_method,omitempty"`
	PaymentMethodLabel  string          `json:"payment_method_label,omitempty"`
	MerchTxnRef         string          `json:"merch_txn_ref"`
	ResponseCode        string          `json:"response_code"`
	Message             string          `json:"message"`
	Paid                bool            `json:"paid"`
	Success             bool            `json:"success"`
	ErrorMessage        string          `json:"error_message,omitempty"`
	RawPayload          json.RawMessage `json:"raw_payload"`
	CreatedAt           time.Time       `json:"created_at"`
}

func providerLabel(provider string) string {
	switch provider {
	case domain.PaymentProviderOnePay:
		return "OnePay"
	default:
		return provider
	}
}

func ToPaymentCallbackEventDTO(ev domain.PaymentCallbackEvent) PaymentCallbackEventDTO {
	raw := ev.RawPayload
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}
	dto := PaymentCallbackEventDTO{
		ID:            string(ev.ID),
		Provider:      ev.Provider,
		ProviderLabel: providerLabel(ev.Provider),
		Channel:       string(ev.Channel),
		ChannelLabel:  ev.Channel.LabelVI(),
		HTTPMethod:    ev.HTTPMethod,
		PaymentID:     string(ev.PaymentID),
		MerchTxnRef:   ev.MerchTxnRef,
		ResponseCode:  ev.ResponseCode,
		Message:       ev.Message,
		Paid:          ev.Paid,
		Success:       ev.Success,
		ErrorMessage:  ev.ErrorMessage,
		RawPayload:    raw,
		CreatedAt:     ev.CreatedAt,
	}
	if ev.PaymentMethod != "" {
		dto.PaymentMethod = string(ev.PaymentMethod)
		dto.PaymentMethodLabel = ev.PaymentMethod.LabelVI()
	}
	return dto
}

type ListPaymentCallbacksQuery struct {
	Provider    string
	Channel     string
	MerchTxnRef string
	Limit       int
	Offset      int
}

type ListPaymentCallbacksHandler struct {
	payments domain.PaymentRepository
}

func NewListPaymentCallbacksHandler(payments domain.PaymentRepository) *ListPaymentCallbacksHandler {
	return &ListPaymentCallbacksHandler{payments: payments}
}

func (h *ListPaymentCallbacksHandler) Handle(_ context.Context, q ListPaymentCallbacksQuery) ([]PaymentCallbackEventDTO, error) {
	items, err := h.payments.ListCallbackEvents(domain.PaymentCallbackListFilter{
		Provider:    q.Provider,
		Channel:     q.Channel,
		MerchTxnRef: q.MerchTxnRef,
		Limit:       q.Limit,
		Offset:      q.Offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]PaymentCallbackEventDTO, 0, len(items))
	for _, ev := range items {
		out = append(out, ToPaymentCallbackEventDTO(ev))
	}
	return out, nil
}

type GetPaymentCallbackHandler struct {
	payments domain.PaymentRepository
}

func NewGetPaymentCallbackHandler(payments domain.PaymentRepository) *GetPaymentCallbackHandler {
	return &GetPaymentCallbackHandler{payments: payments}
}

func (h *GetPaymentCallbackHandler) Handle(_ context.Context, id string) (PaymentCallbackEventDTO, error) {
	ev, err := h.payments.FindCallbackEventByID(domain.PaymentCallbackEventID(id))
	if err != nil {
		return PaymentCallbackEventDTO{}, err
	}
	return ToPaymentCallbackEventDTO(ev), nil
}
