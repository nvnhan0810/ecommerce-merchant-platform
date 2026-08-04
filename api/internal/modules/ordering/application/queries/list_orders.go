package queries

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type OrderItemDTO struct {
	ID             string    `json:"id"`
	ProductID      string    `json:"product_id"`
	ProductName    string    `json:"product_name"`
	UnitPriceCents int64     `json:"unit_price_cents"`
	Quantity       int       `json:"quantity"`
	LineTotalCents int64     `json:"line_total_cents"`
	CreatedAt      time.Time `json:"created_at"`
}

type OrderEventDTO struct {
	ID              string    `json:"id"`
	EventType       string    `json:"event_type"`
	EventLabel      string    `json:"event_label"`
	FromStatus      string    `json:"from_status,omitempty"`
	FromStatusLabel string    `json:"from_status_label,omitempty"`
	ToStatus        string    `json:"to_status,omitempty"`
	ToStatusLabel   string    `json:"to_status_label,omitempty"`
	Message         string    `json:"message"`
	ActorID         string    `json:"actor_id"`
	ActorEmail      string    `json:"actor_email"`
	ActorRole       string    `json:"actor_role"`
	ActorName       string    `json:"actor_name"`
	CreatedAt       time.Time `json:"created_at"`
}

type DeliveryEventDTO struct {
	ID                   string          `json:"id"`
	EventID              string          `json:"event_id,omitempty"`
	DeliveryTrackingCode string          `json:"delivery_tracking_code"`
	StatusCode           string          `json:"status_code"`
	StatusLabel          string          `json:"status_label"`
	Message              string          `json:"message"`
	Reason               string          `json:"reason,omitempty"`
	OccurredAt           time.Time       `json:"occurred_at"`
	Source               string          `json:"source"`
	RawPayload           json.RawMessage `json:"raw_payload,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
}

type OrderDTO struct {
	ID                   string             `json:"id"`
	Code                 string             `json:"code"`
	UserID               string             `json:"user_id"`
	MerchantID           string             `json:"merchant_id"`
	Status               string             `json:"status"`
	StatusLabel          string             `json:"status_label"`
	Currency             string             `json:"currency"`
	TotalCents           int64              `json:"total_cents"`
	Note                 string             `json:"note"`
	DeliveryTrackingCode string             `json:"deliveryTrackingCode,omitempty"`
	DeliveryCarrier      string             `json:"deliveryCarrier"`
	ShippingName         string             `json:"shipping_name"`
	ShippingPhone        string             `json:"shipping_phone"`
	ShippingAddress      string             `json:"shipping_address"`
	Items                []OrderItemDTO     `json:"items"`
	History              []OrderEventDTO    `json:"history"`
	DeliveryEvents       []DeliveryEventDTO `json:"delivery_events"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

func ToDTO(o domain.Order) OrderDTO {
	return ToDTOWithHistory(o, false)
}

func ToDTOWithHistory(o domain.Order, includeHistory bool) OrderDTO {
	items := make([]OrderItemDTO, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, OrderItemDTO{
			ID:             string(item.ID),
			ProductID:      item.ProductID,
			ProductName:    item.ProductName,
			UnitPriceCents: item.UnitPriceCents,
			Quantity:       item.Quantity,
			LineTotalCents: item.LineTotalCents,
			CreatedAt:      item.CreatedAt,
		})
	}
	history := make([]OrderEventDTO, 0)
	if includeHistory {
		history = make([]OrderEventDTO, 0, len(o.History))
		for _, ev := range o.History {
			history = append(history, toEventDTO(ev))
		}
	}
	deliveryEvents := make([]DeliveryEventDTO, 0, len(o.DeliveryEvents))
	if includeHistory {
		for _, ev := range o.DeliveryEvents {
			deliveryEvents = append(deliveryEvents, toDeliveryEventDTO(ev))
		}
	}
	carrier := o.DeliveryCarrier
	if carrier == "" {
		carrier = domain.DefaultDeliveryCarrier
	}
	return OrderDTO{
		ID:                   string(o.ID),
		Code:                 o.Code,
		UserID:               o.UserID,
		MerchantID:           o.MerchantID,
		Status:               string(o.Status),
		StatusLabel:          o.Status.LabelVI(),
		Currency:             o.Currency,
		TotalCents:           o.TotalCents,
		Note:                 o.Note,
		DeliveryTrackingCode: o.DeliveryTrackingCode,
		DeliveryCarrier:      carrier,
		ShippingName:         o.ShippingName,
		ShippingPhone:        o.ShippingPhone,
		ShippingAddress:      o.ShippingAddress,
		Items:                items,
		History:              history,
		DeliveryEvents:       deliveryEvents,
		CreatedAt:            o.CreatedAt,
		UpdatedAt:            o.UpdatedAt,
	}
}

func toEventDTO(ev domain.OrderEvent) OrderEventDTO {
	dto := OrderEventDTO{
		ID:         string(ev.ID),
		EventType:  string(ev.Type),
		EventLabel: ev.Type.LabelVI(),
		Message:    ev.Message,
		ActorID:    ev.ActorID,
		ActorEmail: ev.ActorEmail,
		ActorRole:  ev.ActorRole,
		ActorName:  ev.ActorName,
		CreatedAt:  ev.CreatedAt,
	}
	if ev.FromStatus != "" {
		dto.FromStatus = string(ev.FromStatus)
		dto.FromStatusLabel = ev.FromStatus.LabelVI()
	}
	if ev.ToStatus != "" {
		dto.ToStatus = string(ev.ToStatus)
		dto.ToStatusLabel = ev.ToStatus.LabelVI()
	}
	return dto
}

func toDeliveryEventDTO(ev domain.DeliveryEvent) DeliveryEventDTO {
	return DeliveryEventDTO{
		ID:                   string(ev.ID),
		EventID:              ev.EventID,
		DeliveryTrackingCode: ev.DeliveryTrackingCode,
		StatusCode:           string(ev.StatusCode),
		StatusLabel:          ev.StatusLabel,
		Message:              ev.Message,
		Reason:               ev.Reason,
		OccurredAt:           ev.OccurredAt,
		Source:               ev.Source,
		RawPayload:           ev.RawPayload,
		CreatedAt:            ev.CreatedAt,
	}
}

type ListOrdersQuery struct {
	Limit      int
	Offset     int
	Code       string
	Status     string
	MerchantID string // when set, only return orders for this merchant
	UserID     string // when set, only return orders for this user
}

type ListOrdersHandler struct {
	repo domain.OrderRepository
}

func NewListOrdersHandler(repo domain.OrderRepository) *ListOrdersHandler {
	return &ListOrdersHandler{repo: repo}
}

func (h *ListOrdersHandler) Handle(_ context.Context, q ListOrdersQuery) ([]OrderDTO, error) {
	merchantID := strings.TrimSpace(q.MerchantID)
	userID := strings.TrimSpace(q.UserID)
	code := strings.TrimSpace(q.Code)
	if code != "" {
		order, err := h.repo.FindByCode(code)
		if err != nil {
			return nil, err
		}
		if merchantID != "" && order.MerchantID != merchantID {
			return []OrderDTO{}, nil
		}
		if userID != "" && order.UserID != userID {
			return []OrderDTO{}, nil
		}
		statusFilter := strings.TrimSpace(q.Status)
		if statusFilter != "" && string(order.Status) != statusFilter {
			return []OrderDTO{}, nil
		}
		return []OrderDTO{ToDTO(order)}, nil
	}

	limit := q.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	var (
		items []domain.Order
		err   error
	)
	switch {
	case merchantID != "":
		items, err = h.repo.ListByMerchant(merchantID, limit, offset)
	case userID != "":
		items, err = h.repo.ListByUser(userID, limit, offset)
	default:
		items, err = h.repo.List(limit, offset)
	}
	if err != nil {
		return nil, err
	}

	statusFilter := strings.TrimSpace(q.Status)
	out := make([]OrderDTO, 0, len(items))
	for _, o := range items {
		if statusFilter != "" && string(o.Status) != statusFilter {
			continue
		}
		out = append(out, ToDTO(o))
	}
	return out, nil
}
