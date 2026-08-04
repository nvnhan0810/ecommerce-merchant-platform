package commands

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type ApplyDeliveryEventCommand struct {
	OrderID              domain.OrderID // optional; used by admin simulate
	OrderCode            string
	DeliveryTrackingCode string
	DeliveryCarrier      string
	Status               string
	Message              string
	Reason               string
	OccurredAt           time.Time
	EventID              string
	Source               string
	RawPayload           json.RawMessage
}

type ApplyDeliveryEventHandler struct {
	repo domain.OrderRepository
}

func NewApplyDeliveryEventHandler(repo domain.OrderRepository) *ApplyDeliveryEventHandler {
	return &ApplyDeliveryEventHandler{repo: repo}
}

func (h *ApplyDeliveryEventHandler) Handle(_ context.Context, cmd ApplyDeliveryEventCommand) (queries.OrderDTO, error) {
	statusCode, err := domain.ParseDeliveryStatusCode(cmd.Status)
	if err != nil {
		return queries.OrderDTO{}, err
	}

	eventID := strings.TrimSpace(cmd.EventID)
	if eventID != "" {
		exists, err := h.repo.HasDeliveryEventID(eventID)
		if err != nil {
			return queries.OrderDTO{}, err
		}
		if exists {
			return queries.OrderDTO{}, domain.ErrDeliveryEventDuplicate
		}
	}

	order, err := h.resolveOrder(cmd)
	if err != nil {
		return queries.OrderDTO{}, err
	}

	raw := cmd.RawPayload
	if len(raw) == 0 {
		raw = json.RawMessage(`{}`)
	}

	if _, err := order.ApplyDeliveryEvent(domain.ApplyDeliveryInput{
		EventID:              eventID,
		DeliveryTrackingCode: strings.TrimSpace(cmd.DeliveryTrackingCode),
		DeliveryCarrier:      strings.TrimSpace(cmd.DeliveryCarrier),
		StatusCode:           statusCode,
		Message:              cmd.Message,
		Reason:               cmd.Reason,
		OccurredAt:           cmd.OccurredAt,
		Source:               strings.TrimSpace(cmd.Source),
		RawPayload:           raw,
	}); err != nil {
		return queries.OrderDTO{}, err
	}

	if err := h.repo.Save(order); err != nil {
		return queries.OrderDTO{}, err
	}
	saved, err := h.repo.FindByID(order.ID)
	if err != nil {
		return queries.OrderDTO{}, err
	}
	return queries.ToDTOWithHistory(saved, true), nil
}

func (h *ApplyDeliveryEventHandler) resolveOrder(cmd ApplyDeliveryEventCommand) (domain.Order, error) {
	if cmd.OrderID != "" {
		return h.repo.FindByID(cmd.OrderID)
	}
	orderCode := strings.TrimSpace(cmd.OrderCode)
	tracking := strings.TrimSpace(cmd.DeliveryTrackingCode)
	if orderCode != "" {
		return h.repo.FindByCode(orderCode)
	}
	if tracking != "" {
		return h.repo.FindByDeliveryTrackingCode(tracking)
	}
	return domain.Order{}, domain.ErrDeliveryLookupRequired
}
