package domain

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidOrderID            = errors.New("order id is required")
	ErrInvalidOrderCode          = errors.New("invalid order code")
	ErrUserRequired              = errors.New("user is required")
	ErrMerchantRequired          = errors.New("merchant is required")
	ErrInvalidOrderStatus          = errors.New("invalid order status")
	ErrInvalidStatusTransition     = errors.New("invalid order status transition")
	ErrMerchantConfirmOnly         = errors.New("merchant may only confirm or cancel new/confirmed orders")
	ErrMerchantCancelReasonRequired = errors.New("cancel reason is required")
	ErrEmptyOrderItems             = errors.New("order must have at least one item")
	ErrMissingShippingInfo         = errors.New("shipping name, phone, and address are required")
	ErrInvalidOrderQuantity      = errors.New("order item quantity must be greater than zero")
	ErrInvalidOrderPrice         = errors.New("order item price must be greater than zero")
	ErrProductMerchantMismatch   = errors.New("product does not belong to the order merchant")
	ErrOrderNotFound             = errors.New("order not found")
	ErrInvalidDeliveryStatus     = errors.New("invalid delivery status")
	ErrDeliveryLookupRequired    = errors.New("order_code or delivery_tracking_code is required")
	ErrDeliveryEventDuplicate    = errors.New("delivery event already processed")
)

// Order tracking code: 10 uppercase alphanumeric characters (A-Z, 0-9).
const OrderCodeLength = 10

var orderCodeAlphabet = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// OrderStatus values stored in DB / API.
type OrderStatus string

const (
	StatusNew       OrderStatus = "new"       // Mới
	StatusPaid      OrderStatus = "paid"      // Đã thanh toán (compat)
	StatusConfirmed OrderStatus = "confirmed" // Đã xác nhận
	StatusShipping  OrderStatus = "shipping"  // Đang vận chuyển
	StatusSucceeded OrderStatus = "succeeded" // Thành công
	StatusReturning OrderStatus = "returning" // Đang hoàn hàng
	StatusReturned  OrderStatus = "returned"  // Đã hoàn hàng
	StatusFailed    OrderStatus = "failed"    // Thất bại
	StatusCancelled OrderStatus = "cancelled" // Huỷ
)

const DefaultDeliveryCarrier = "internal"

func ParseOrderStatus(raw string) (OrderStatus, error) {
	s := OrderStatus(strings.ToLower(strings.TrimSpace(raw)))
	switch s {
	case StatusNew, StatusPaid, StatusConfirmed, StatusShipping, StatusSucceeded,
		StatusReturning, StatusReturned, StatusFailed, StatusCancelled:
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
	case StatusReturning:
		return "Đang hoàn hàng"
	case StatusReturned:
		return "Đã hoàn hàng"
	case StatusFailed:
		return "Thất bại"
	case StatusCancelled:
		return "Huỷ"
	default:
		return string(s)
	}
}

// DeliveryStatusCode values from TMS / webhook.
type DeliveryStatusCode string

const (
	DeliveryAccepted    DeliveryStatusCode = "accepted"
	DeliveryDelivering  DeliveryStatusCode = "delivering"
	DeliveryInTransit   DeliveryStatusCode = "in_transit"
	DeliveryFail        DeliveryStatusCode = "delivery_fail"
	DeliveryDelivered   DeliveryStatusCode = "delivered"
	DeliveryReturning   DeliveryStatusCode = "returning"
	DeliveryReturned    DeliveryStatusCode = "returned"
	DeliveryReturnFail  DeliveryStatusCode = "return_fail"
	DeliveryLost        DeliveryStatusCode = "lost"
	DeliveryDamage      DeliveryStatusCode = "damage"
)

func ParseDeliveryStatusCode(raw string) (DeliveryStatusCode, error) {
	s := DeliveryStatusCode(strings.ToLower(strings.TrimSpace(raw)))
	switch s {
	case DeliveryAccepted, DeliveryDelivering, DeliveryInTransit, DeliveryFail,
		DeliveryDelivered, DeliveryReturning, DeliveryReturned,
		DeliveryReturnFail, DeliveryLost, DeliveryDamage:
		return s, nil
	default:
		return "", ErrInvalidDeliveryStatus
	}
}

func (c DeliveryStatusCode) LabelVI() string {
	switch c {
	case DeliveryAccepted:
		return "Đã tiếp nhận"
	case DeliveryDelivering:
		return "Đang giao"
	case DeliveryInTransit:
		return "Đang trung chuyển"
	case DeliveryFail:
		return "Giao thất bại"
	case DeliveryDelivered:
		return "Đã giao"
	case DeliveryReturning:
		return "Đang hoàn hàng"
	case DeliveryReturned:
		return "Đã hoàn hàng"
	case DeliveryReturnFail:
		return "Hoàn thất bại"
	case DeliveryLost:
		return "Thất lạc"
	case DeliveryDamage:
		return "Hư hỏng"
	default:
		return string(c)
	}
}

// MapDeliveryToOrderStatus returns the coarse order status for a TMS code.
// change=false means keep current business status (log-only / stay shipping).
func MapDeliveryToOrderStatus(code DeliveryStatusCode) (target OrderStatus, change bool) {
	switch code {
	case DeliveryAccepted:
		return StatusShipping, true
	case DeliveryDelivering, DeliveryInTransit, DeliveryFail:
		return StatusShipping, false
	case DeliveryDelivered:
		return StatusSucceeded, true
	case DeliveryReturning:
		return StatusReturning, true
	case DeliveryReturned:
		return StatusReturned, true
	case DeliveryReturnFail, DeliveryLost, DeliveryDamage:
		return StatusFailed, true
	default:
		return "", false
	}
}

// AllDeliveryStatusCodes returns TMS codes in display order.
func AllDeliveryStatusCodes() []DeliveryStatusCode {
	return []DeliveryStatusCode{
		DeliveryAccepted, DeliveryDelivering, DeliveryInTransit, DeliveryFail,
		DeliveryDelivered, DeliveryReturning, DeliveryReturned,
		DeliveryReturnFail, DeliveryLost, DeliveryDamage,
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
	ID                   OrderID
	Code                 string
	UserID               string
	MerchantID           string
	Status               OrderStatus
	Currency             string
	TotalCents           int64
	Note                 string
	DeliveryTrackingCode string
	DeliveryCarrier      string
	ShippingName         string
	ShippingPhone        string
	ShippingAddress      string
	Items                []OrderItem
	History              []OrderEvent
	DeliveryEvents       []DeliveryEvent
	CreatedAt            time.Time
	UpdatedAt            time.Time

	pendingEvents         []OrderEvent
	pendingDeliveryEvents []DeliveryEvent
}

type Actor struct {
	ID          string
	Email       string
	Role        string
	DisplayName string
}

func SystemActor() Actor {
	return Actor{Role: "system", DisplayName: "Hệ thống"}
}

type OrderEventType string

const (
	EventCreated       OrderEventType = "created"
	EventStatusChanged OrderEventType = "status_changed"
	EventCancelled     OrderEventType = "cancelled"
)

func (t OrderEventType) LabelVI() string {
	switch t {
	case EventCreated:
		return "Tạo đơn"
	case EventStatusChanged:
		return "Đổi trạng thái"
	case EventCancelled:
		return "Huỷ đơn"
	default:
		return string(t)
	}
}

type OrderEventID string

func NewOrderEventID() OrderEventID { return OrderEventID(uuid.NewString()) }

type OrderEvent struct {
	ID         OrderEventID
	OrderID    OrderID
	Type       OrderEventType
	FromStatus OrderStatus
	ToStatus   OrderStatus
	Message    string
	ActorID    string
	ActorEmail string
	ActorRole  string
	ActorName  string
	CreatedAt  time.Time
}

type DeliveryEventID string

func NewDeliveryEventID() DeliveryEventID { return DeliveryEventID(uuid.NewString()) }

type DeliveryEvent struct {
	ID                   DeliveryEventID
	OrderID              OrderID
	EventID              string
	DeliveryTrackingCode string
	StatusCode           DeliveryStatusCode
	StatusLabel          string
	Message              string
	Reason               string
	OccurredAt           time.Time
	Source               string
	RawPayload           json.RawMessage
	CreatedAt            time.Time
}

func (o *Order) PendingEvents() []OrderEvent {
	return append([]OrderEvent(nil), o.pendingEvents...)
}

func (o *Order) ClearPendingEvents() {
	o.pendingEvents = nil
}

func (o *Order) PendingDeliveryEvents() []DeliveryEvent {
	return append([]DeliveryEvent(nil), o.pendingDeliveryEvents...)
}

func (o *Order) ClearPendingDeliveryEvents() {
	o.pendingDeliveryEvents = nil
}

func (o *Order) RecordCreated(actor Actor) {
	now := time.Now().UTC()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}
	ev := OrderEvent{
		ID:         NewOrderEventID(),
		OrderID:    o.ID,
		Type:       EventCreated,
		ToStatus:   StatusNew,
		Message:    "Tạo đơn hàng",
		ActorID:    strings.TrimSpace(actor.ID),
		ActorEmail: strings.TrimSpace(actor.Email),
		ActorRole:  strings.TrimSpace(actor.Role),
		ActorName:  strings.TrimSpace(actor.DisplayName),
		CreatedAt:  o.CreatedAt,
	}
	if ev.ActorName == "" {
		ev.ActorName = ev.ActorEmail
	}
	if ev.ActorRole == "" {
		ev.ActorRole = "user"
	}
	o.pendingEvents = append(o.pendingEvents, ev)
	o.History = append(o.History, ev)
}

func (o *Order) ChangeStatus(status OrderStatus, actor Actor) error {
	if _, err := ParseOrderStatus(string(status)); err != nil {
		return err
	}
	if status == o.Status {
		return nil
	}
	from := o.Status
	now := time.Now().UTC()
	o.Status = status
	o.UpdatedAt = now

	eventType := EventStatusChanged
	message := "Đổi trạng thái từ " + from.LabelVI() + " sang " + status.LabelVI()
	if status == StatusCancelled {
		eventType = EventCancelled
		message = "Huỷ đơn hàng"
		if from != StatusNew {
			message = "Huỷ đơn hàng (từ " + from.LabelVI() + ")"
		}
	}

	ev := OrderEvent{
		ID:         NewOrderEventID(),
		OrderID:    o.ID,
		Type:       eventType,
		FromStatus: from,
		ToStatus:   status,
		Message:    message,
		ActorID:    strings.TrimSpace(actor.ID),
		ActorEmail: strings.TrimSpace(actor.Email),
		ActorRole:  strings.TrimSpace(actor.Role),
		ActorName:  strings.TrimSpace(actor.DisplayName),
		CreatedAt:  now,
	}
	if ev.ActorName == "" {
		ev.ActorName = ev.ActorEmail
	}
	if ev.ActorName == "" {
		ev.ActorName = "Hệ thống"
	}
	if ev.ActorRole == "" {
		ev.ActorRole = "system"
	}
	o.pendingEvents = append(o.pendingEvents, ev)
	o.History = append(o.History, ev)
	return nil
}

// MerchantConfirm allows only new → confirmed, then auto-dispatches internal shipment.
func (o *Order) MerchantConfirm(actor Actor) error {
	if o.Status != StatusNew {
		return ErrMerchantConfirmOnly
	}
	if err := o.ChangeStatus(StatusConfirmed, actor); err != nil {
		return err
	}
	return o.DispatchInternalShipment()
}

// DispatchInternalShipment creates an internal TMS accepted event (→ shipping).
func (o *Order) DispatchInternalShipment() error {
	tracking := strings.TrimSpace(o.DeliveryTrackingCode)
	if tracking == "" {
		code, err := NewDeliveryTrackingCode()
		if err != nil {
			return err
		}
		tracking = code
	}
	_, err := o.ApplyDeliveryEvent(ApplyDeliveryInput{
		DeliveryTrackingCode: tracking,
		DeliveryCarrier:      DefaultDeliveryCarrier,
		StatusCode:           DeliveryAccepted,
		Message:              "Đã tạo vận đơn nội bộ",
		Source:               "confirm",
	})
	return err
}

// NewDeliveryTrackingCode returns a carrier-facing tracking id, e.g. "DELI-K7M2P9QX4A".
func NewDeliveryTrackingCode() (string, error) {
	code, err := NewOrderCode()
	if err != nil {
		return "", err
	}
	return "DELI-" + code, nil
}

// MerchantCancel allows new|confirmed → cancelled; reason is required.
// After auto-dispatch, confirmed orders are usually already shipping — cancel only pre-ship.
func (o *Order) MerchantCancel(actor Actor, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ErrMerchantCancelReasonRequired
	}
	switch o.Status {
	case StatusNew, StatusConfirmed:
		// ok
	default:
		return ErrMerchantConfirmOnly
	}
	if err := o.ChangeStatus(StatusCancelled, actor); err != nil {
		return err
	}
	if n := len(o.pendingEvents); n > 0 {
		ev := &o.pendingEvents[n-1]
		ev.Message = "Huỷ đơn hàng: " + reason
		if len(o.History) > 0 {
			o.History[len(o.History)-1].Message = ev.Message
		}
	}
	return nil
}

type ApplyDeliveryInput struct {
	EventID              string
	DeliveryTrackingCode string
	DeliveryCarrier      string
	StatusCode           DeliveryStatusCode
	Message              string
	Reason               string
	OccurredAt           time.Time
	Source               string
	RawPayload           json.RawMessage
}

// ApplyDeliveryEvent appends a TMS log row and optionally updates coarse order status.
func (o *Order) ApplyDeliveryEvent(in ApplyDeliveryInput) (DeliveryEvent, error) {
	code, err := ParseDeliveryStatusCode(string(in.StatusCode))
	if err != nil {
		return DeliveryEvent{}, err
	}

	tracking := strings.TrimSpace(in.DeliveryTrackingCode)
	if tracking != "" && strings.TrimSpace(o.DeliveryTrackingCode) == "" {
		o.DeliveryTrackingCode = tracking
	}
	if tracking == "" {
		tracking = strings.TrimSpace(o.DeliveryTrackingCode)
	}

	carrier := strings.TrimSpace(in.DeliveryCarrier)
	if carrier != "" {
		o.DeliveryCarrier = carrier
	}
	if strings.TrimSpace(o.DeliveryCarrier) == "" {
		o.DeliveryCarrier = DefaultDeliveryCarrier
	}

	occurred := in.OccurredAt.UTC()
	if occurred.IsZero() {
		occurred = time.Now().UTC()
	}
	now := time.Now().UTC()
	label := code.LabelVI()
	msg := strings.TrimSpace(in.Message)
	if msg == "" {
		msg = label
	}
	raw := in.RawPayload
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}

	ev := DeliveryEvent{
		ID:                   NewDeliveryEventID(),
		OrderID:              o.ID,
		EventID:              strings.TrimSpace(in.EventID),
		DeliveryTrackingCode: tracking,
		StatusCode:           code,
		StatusLabel:          label,
		Message:              msg,
		Reason:               strings.TrimSpace(in.Reason),
		OccurredAt:           occurred,
		Source:               strings.TrimSpace(in.Source),
		RawPayload:           raw,
		CreatedAt:            now,
	}
	o.pendingDeliveryEvents = append(o.pendingDeliveryEvents, ev)
	o.DeliveryEvents = append(o.DeliveryEvents, ev)
	o.UpdatedAt = now

	target, change := MapDeliveryToOrderStatus(code)
	if change && target != o.Status {
		actor := SystemActor()
		actor.DisplayName = "TMS"
		if err := o.ChangeStatus(target, actor); err != nil {
			return DeliveryEvent{}, err
		}
	}
	return ev, nil
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

func NewOrder(userID, merchantID, currency, note, shippingName, shippingPhone, shippingAddress string, lines []OrderLineInput) (Order, error) {
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
	shippingName = strings.TrimSpace(shippingName)
	shippingPhone = strings.TrimSpace(shippingPhone)
	shippingAddress = strings.TrimSpace(shippingAddress)
	if shippingName == "" || shippingPhone == "" || shippingAddress == "" {
		return Order{}, ErrMissingShippingInfo
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
			ID:              NewOrderID(),
			Code:            code,
			UserID:          userID,
			MerchantID:      merchantID,
			Status:          StatusNew,
			Currency:        currency,
			TotalCents:      total,
			Note:            strings.TrimSpace(note),
			DeliveryCarrier: DefaultDeliveryCarrier,
			ShippingName:    shippingName,
			ShippingPhone:   shippingPhone,
			ShippingAddress: shippingAddress,
			Items:           items,
			CreatedAt:       now,
			UpdatedAt:       now,
		}, nil
}

type OrderRepository interface {
	Save(order Order) error
	FindByID(id OrderID) (Order, error)
	FindByCode(code string) (Order, error)
	FindByDeliveryTrackingCode(code string) (Order, error)
	HasDeliveryEventID(eventID string) (bool, error)
	List(limit, offset int) ([]Order, error)
	ListByMerchant(merchantID string, limit, offset int) ([]Order, error)
	ListByUser(userID string, limit, offset int) ([]Order, error)
	Count() (int, error)
}

// AllOrderStatuses returns statuses in display order.
func AllOrderStatuses() []OrderStatus {
	return []OrderStatus{
		StatusNew, StatusPaid, StatusConfirmed, StatusShipping,
		StatusSucceeded, StatusReturning, StatusReturned, StatusFailed, StatusCancelled,
	}
}
