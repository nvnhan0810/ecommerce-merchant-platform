package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidPaymentMethod  = errors.New("invalid payment method")
	ErrInvalidPaymentStatus  = errors.New("invalid payment status")
	ErrPaymentNotFound       = errors.New("payment not found")
	ErrOnePayNotConfigured   = errors.New("onepay is not configured")
	ErrOnePayDisabled        = errors.New("onepay is disabled")
	ErrInvalidOnePayHash     = errors.New("invalid onepay secure hash")
	ErrPaymentAlreadyHandled = errors.New("payment already handled")
	ErrPaymentCallbackNotFound = errors.New("payment callback not found")
)

type PaymentMethod string

const (
	PaymentMethodCOD                 PaymentMethod = "cod"
	PaymentMethodOnePayDomestic      PaymentMethod = "onepay_domestic"
	PaymentMethodOnePayInternational PaymentMethod = "onepay_international"
)

func ParsePaymentMethod(raw string) (PaymentMethod, error) {
	m := PaymentMethod(strings.ToLower(strings.TrimSpace(raw)))
	switch m {
	case PaymentMethodCOD, PaymentMethodOnePayDomestic, PaymentMethodOnePayInternational:
		return m, nil
	case "onepay":
		// Legacy alias → domestic.
		return PaymentMethodOnePayDomestic, nil
	case "":
		return PaymentMethodCOD, nil
	default:
		return "", ErrInvalidPaymentMethod
	}
}

func (m PaymentMethod) IsOnePay() bool {
	return m == PaymentMethodOnePayDomestic || m == PaymentMethodOnePayInternational
}

func (m PaymentMethod) LabelVI() string {
	switch m {
	case PaymentMethodCOD:
		return "Thanh toán khi giao hàng"
	case PaymentMethodOnePayDomestic:
		return "Thanh toán thẻ nội địa"
	case PaymentMethodOnePayInternational:
		return "Thanh toán thẻ quốc tế"
	default:
		return string(m)
	}
}

type PaymentStatus string

const (
	PaymentStatusUnpaid    PaymentStatus = "unpaid"
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

func ParsePaymentStatus(raw string) (PaymentStatus, error) {
	s := PaymentStatus(strings.ToLower(strings.TrimSpace(raw)))
	switch s {
	case PaymentStatusUnpaid, PaymentStatusPending, PaymentStatusPaid, PaymentStatusFailed, PaymentStatusCancelled:
		return s, nil
	default:
		return "", ErrInvalidPaymentStatus
	}
}

func (s PaymentStatus) LabelVI() string {
	switch s {
	case PaymentStatusUnpaid:
		return "Chưa thanh toán"
	case PaymentStatusPending:
		return "Đang chờ thanh toán"
	case PaymentStatusPaid:
		return "Đã thanh toán"
	case PaymentStatusFailed:
		return "Thanh toán thất bại"
	case PaymentStatusCancelled:
		return "Đã hủy thanh toán"
	default:
		return string(s)
	}
}

type PaymentID string

func NewPaymentID() PaymentID { return PaymentID(uuid.NewString()) }

func ParsePaymentID(raw string) (PaymentID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrPaymentNotFound
	}
	return PaymentID(raw), nil
}

type Payment struct {
	ID           PaymentID
	UserID       string
	Method       PaymentMethod
	Status       PaymentStatus
	AmountCents  int64
	Currency     string
	MerchTxnRef  string
	GatewayTxnNo string
	ResponseCode string
	Message      string
	OrderIDs     []OrderID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewPayment(userID string, method PaymentMethod, amountCents int64, currency string, orderIDs []OrderID) (Payment, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Payment{}, ErrUserRequired
	}
	if amountCents <= 0 {
		return Payment{}, ErrInvalidOrderPrice
	}
	if len(orderIDs) == 0 {
		return Payment{}, ErrEmptyOrderItems
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "VND"
	}
	now := time.Now().UTC()
	id := NewPaymentID()
	ref := strings.ReplaceAll(string(id), "-", "")
	if len(ref) > 32 {
		ref = ref[:32]
	}
	status := PaymentStatusPending
	if method == PaymentMethodCOD {
		status = PaymentStatusUnpaid
	}
	return Payment{
		ID:          id,
		UserID:      userID,
		Method:      method,
		Status:      status,
		AmountCents: amountCents,
		Currency:    currency,
		MerchTxnRef: ref,
		OrderIDs:    append([]OrderID(nil), orderIDs...),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (p *Payment) MarkPaid(gatewayTxnNo, responseCode, message string) {
	p.Status = PaymentStatusPaid
	p.GatewayTxnNo = strings.TrimSpace(gatewayTxnNo)
	p.ResponseCode = strings.TrimSpace(responseCode)
	p.Message = strings.TrimSpace(message)
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) MarkFailed(responseCode, message string) {
	p.Status = PaymentStatusFailed
	p.ResponseCode = strings.TrimSpace(responseCode)
	p.Message = strings.TrimSpace(message)
	p.UpdatedAt = time.Now().UTC()
}

type OnePayGatewaySettings struct {
	Enabled    bool
	MerchantID string
	AccessCode string
	HashSecret string
	PaymentURL string
}

func (g OnePayGatewaySettings) Ready(returnURL string) bool {
	return g.Enabled &&
		strings.TrimSpace(g.MerchantID) != "" &&
		strings.TrimSpace(g.AccessCode) != "" &&
		strings.TrimSpace(g.HashSecret) != "" &&
		strings.TrimSpace(g.PaymentURL) != "" &&
		strings.TrimSpace(returnURL) != ""
}

type PaymentSettings struct {
	OnePayReturnURL      string
	OnePayIPNURL         string
	OnePayDomestic       OnePayGatewaySettings
	OnePayInternational  OnePayGatewaySettings
	UpdatedAt            time.Time
}

func (s PaymentSettings) GatewayFor(method PaymentMethod) (OnePayGatewaySettings, error) {
	switch method {
	case PaymentMethodOnePayDomestic:
		return s.OnePayDomestic, nil
	case PaymentMethodOnePayInternational:
		return s.OnePayInternational, nil
	default:
		return OnePayGatewaySettings{}, ErrInvalidPaymentMethod
	}
}

func (s PaymentSettings) OnePayReady(method PaymentMethod) bool {
	gw, err := s.GatewayFor(method)
	if err != nil {
		return false
	}
	return gw.Ready(s.OnePayReturnURL)
}

type PaymentRepository interface {
	Save(payment Payment) error
	FindByID(id PaymentID) (Payment, error)
	FindByMerchTxnRef(ref string) (Payment, error)
	GetSettings() (PaymentSettings, error)
	SaveSettings(settings PaymentSettings) error
	SaveCallbackEvent(event PaymentCallbackEvent) error
	FindCallbackEventByID(id PaymentCallbackEventID) (PaymentCallbackEvent, error)
	ListCallbackEvents(filter PaymentCallbackListFilter) ([]PaymentCallbackEvent, error)
}
