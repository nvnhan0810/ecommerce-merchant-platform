package domain

import (
	"crypto/rand"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidOrderID          = errors.New("order id is required")
	ErrInvalidOrderCode        = errors.New("invalid order code")
	ErrUserRequired            = errors.New("user is required")
	ErrMerchantRequired        = errors.New("merchant is required")
	ErrInvalidOrderStatus      = errors.New("invalid order status")
	ErrEmptyOrderItems         = errors.New("order must have at least one item")
	ErrInvalidOrderQuantity    = errors.New("order item quantity must be greater than zero")
	ErrInvalidOrderPrice       = errors.New("order item price must be greater than zero")
	ErrProductMerchantMismatch = errors.New("product does not belong to the order merchant")
	ErrOrderNotFound           = errors.New("order not found")
)

// Order tracking code: 10 uppercase alphanumeric characters (A-Z, 0-9).
const OrderCodeLength = 10

var orderCodeAlphabet = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// OrderStatus values stored in DB / API.
type OrderStatus string

const (
	StatusNew       OrderStatus = "new"       // Mới
	StatusPaid      OrderStatus = "paid"      // Đã thanh toán
	StatusConfirmed OrderStatus = "confirmed" // Đã xác nhận
	StatusShipping  OrderStatus = "shipping"  // Đang vận chuyển
	StatusSucceeded OrderStatus = "succeeded" // Thành công
	StatusFailed    OrderStatus = "failed"    // Thất bại
	StatusCancelled OrderStatus = "cancelled" // Huỷ
)

func ParseOrderStatus(raw string) (OrderStatus, error) {
	s := OrderStatus(strings.ToLower(strings.TrimSpace(raw)))
	switch s {
	case StatusNew, StatusPaid, StatusConfirmed, StatusShipping, StatusSucceeded, StatusFailed, StatusCancelled:
		return s, nil
	default:
		return "", ErrInvalidOrderStatus
	}
}

func (s OrderStatus) LabelVI() string {
	switch s {
	case StatusNew:
		return "Mới"
	case StatusPaid:
		return "Đã thanh toán"
	case StatusConfirmed:
		return "Đã xác nhận"
	case StatusShipping:
		return "Đang vận chuyển"
	case StatusSucceeded:
		return "Thành công"
	case StatusFailed:
		return "Thất bại"
	case StatusCancelled:
		return "Huỷ"
	default:
		return string(s)
	}
}

type OrderID string

func NewOrderID() OrderID { return OrderID(uuid.NewString()) }

func ParseOrderID(raw string) (OrderID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidOrderID
	}
	return OrderID(raw), nil
}

type OrderItemID string

func NewOrderItemID() OrderItemID { return OrderItemID(uuid.NewString()) }

type OrderItem struct {
	ID             OrderItemID
	ProductID      string
	ProductName    string
	UnitPriceCents int64
	Quantity       int
	LineTotalCents int64
	CreatedAt      time.Time
}

type OrderLineInput struct {
	ProductID      string
	ProductName    string
	MerchantID     string
	UnitPriceCents int64
	Quantity       int
}

type Order struct {
	ID         OrderID
	Code       string
	UserID     string
	MerchantID string
	Status     OrderStatus
	Currency   string
	TotalCents int64
	Note       string
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewOrderCode returns a random tracking code (letters + digits), e.g. "K7M2P9QX4A".
func NewOrderCode() (string, error) {
	buf := make([]byte, OrderCodeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = orderCodeAlphabet[int(buf[i])%len(orderCodeAlphabet)]
	}
	return string(buf), nil
}

func ParseOrderCode(raw string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if len(code) != OrderCodeLength {
		return "", ErrInvalidOrderCode
	}
	for _, c := range code {
		if (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return "", ErrInvalidOrderCode
		}
	}
	return code, nil
}

func NewOrder(userID, merchantID, currency, note string, lines []OrderLineInput) (Order, error) {
	userID = strings.TrimSpace(userID)
	merchantID = strings.TrimSpace(merchantID)
	if userID == "" {
		return Order{}, ErrUserRequired
	}
	if merchantID == "" {
		return Order{}, ErrMerchantRequired
	}
	if len(lines) == 0 {
		return Order{}, ErrEmptyOrderItems
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "VND"
	}

	code, err := NewOrderCode()
	if err != nil {
		return Order{}, err
	}

	now := time.Now().UTC()
	items := make([]OrderItem, 0, len(lines))
	var total int64
	for _, line := range lines {
		if strings.TrimSpace(line.MerchantID) != merchantID {
			return Order{}, ErrProductMerchantMismatch
		}
		if line.Quantity <= 0 {
			return Order{}, ErrInvalidOrderQuantity
		}
		if line.UnitPriceCents <= 0 {
			return Order{}, ErrInvalidOrderPrice
		}
		name := strings.TrimSpace(line.ProductName)
		if name == "" {
			name = "Product"
		}
		lineTotal := line.UnitPriceCents * int64(line.Quantity)
		items = append(items, OrderItem{
			ID:             NewOrderItemID(),
			ProductID:      strings.TrimSpace(line.ProductID),
			ProductName:    name,
			UnitPriceCents: line.UnitPriceCents,
			Quantity:       line.Quantity,
			LineTotalCents: lineTotal,
			CreatedAt:      now,
		})
		total += lineTotal
	}

	return Order{
		ID:         NewOrderID(),
		Code:       code,
		UserID:     userID,
		MerchantID: merchantID,
		Status:     StatusNew,
		Currency:   currency,
		TotalCents: total,
		Note:       strings.TrimSpace(note),
		Items:      items,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (o *Order) ChangeStatus(status OrderStatus) error {
	if _, err := ParseOrderStatus(string(status)); err != nil {
		return err
	}
	o.Status = status
	o.UpdatedAt = time.Now().UTC()
	return nil
}

type OrderRepository interface {
	Save(order Order) error
	FindByID(id OrderID) (Order, error)
	List(limit, offset int) ([]Order, error)
	Count() (int, error)
}
