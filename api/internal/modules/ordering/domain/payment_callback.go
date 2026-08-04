package domain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Payment callback providers (extensible for future gateways).
const (
	PaymentProviderOnePay = "onepay"
)

type PaymentCallbackChannel string

const (
	PaymentCallbackChannelIPN    PaymentCallbackChannel = "ipn"
	PaymentCallbackChannelReturn PaymentCallbackChannel = "return"
)

func ParsePaymentCallbackChannel(raw string) (PaymentCallbackChannel, bool) {
	c := PaymentCallbackChannel(strings.ToLower(strings.TrimSpace(raw)))
	switch c {
	case PaymentCallbackChannelIPN, PaymentCallbackChannelReturn:
		return c, true
	default:
		return "", false
	}
}

func (c PaymentCallbackChannel) LabelVI() string {
	switch c {
	case PaymentCallbackChannelIPN:
		return "IPN"
	case PaymentCallbackChannelReturn:
		return "Return"
	default:
		return string(c)
	}
}

type PaymentCallbackEventID string

func NewPaymentCallbackEventID() PaymentCallbackEventID {
	return PaymentCallbackEventID(uuid.NewString())
}

type PaymentCallbackEvent struct {
	ID            PaymentCallbackEventID
	Provider      string
	Channel       PaymentCallbackChannel
	HTTPMethod    string
	PaymentID     PaymentID
	PaymentMethod PaymentMethod
	MerchTxnRef   string
	ResponseCode  string
	Message       string
	Paid          bool
	Success       bool
	ErrorMessage  string
	RawPayload    json.RawMessage
	CreatedAt     time.Time
}

type NewPaymentCallbackEventInput struct {
	Provider      string
	Channel       PaymentCallbackChannel
	HTTPMethod    string
	PaymentID     PaymentID
	PaymentMethod PaymentMethod
	MerchTxnRef   string
	ResponseCode  string
	Message       string
	Paid          bool
	Success       bool
	ErrorMessage  string
	RawPayload    map[string]string
}

func NewPaymentCallbackEvent(in NewPaymentCallbackEventInput) PaymentCallbackEvent {
	provider := strings.ToLower(strings.TrimSpace(in.Provider))
	if provider == "" {
		provider = PaymentProviderOnePay
	}
	raw := in.RawPayload
	if raw == nil {
		raw = map[string]string{}
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		payload = []byte("{}")
	}
	return PaymentCallbackEvent{
		ID:            NewPaymentCallbackEventID(),
		Provider:      provider,
		Channel:       in.Channel,
		HTTPMethod:    strings.ToUpper(strings.TrimSpace(in.HTTPMethod)),
		PaymentID:     in.PaymentID,
		PaymentMethod: in.PaymentMethod,
		MerchTxnRef:   strings.TrimSpace(in.MerchTxnRef),
		ResponseCode:  strings.TrimSpace(in.ResponseCode),
		Message:       strings.TrimSpace(in.Message),
		Paid:          in.Paid,
		Success:       in.Success,
		ErrorMessage:  strings.TrimSpace(in.ErrorMessage),
		RawPayload:    payload,
		CreatedAt:     time.Now().UTC(),
	}
}

type PaymentCallbackListFilter struct {
	Provider    string
	Channel     string
	MerchTxnRef string
	Limit       int
	Offset      int
}

func (f *PaymentCallbackListFilter) Normalize() {
	f.Provider = strings.ToLower(strings.TrimSpace(f.Provider))
	f.Channel = strings.ToLower(strings.TrimSpace(f.Channel))
	f.MerchTxnRef = strings.TrimSpace(f.MerchTxnRef)
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
}
