package commands

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/application/queries"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type UpdateOrderStatusCommand struct {
	ID              domain.OrderID
	Status          string
	Reason          string
	Actor           domain.Actor
	OwnerMerchantID string // when set, order must belong to this merchant
}

type UpdateOrderStatusHandler struct {
	repo domain.OrderRepository
}

func NewUpdateOrderStatusHandler(repo domain.OrderRepository) *UpdateOrderStatusHandler {
	return &UpdateOrderStatusHandler{repo: repo}
}

func (h *UpdateOrderStatusHandler) Handle(_ context.Context, cmd UpdateOrderStatusCommand) (queries.OrderDTO, error) {
	order, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return queries.OrderDTO{}, err
	}
	owner := strings.TrimSpace(cmd.OwnerMerchantID)
	if owner != "" && order.MerchantID != owner {
		return queries.OrderDTO{}, domain.ErrOrderNotFound
	}
	status, err := domain.ParseOrderStatus(cmd.Status)
	if err != nil {
		return queries.OrderDTO{}, err
	}
	actor := cmd.Actor
	if strings.TrimSpace(actor.DisplayName) == "" {
		actor.DisplayName = actor.Email
	}
	if strings.TrimSpace(actor.Role) == "" {
		actor.Role = "admin"
	}
	if owner != "" {
		switch status {
		case domain.StatusConfirmed:
			if err := order.MerchantConfirm(actor); err != nil {
				return queries.OrderDTO{}, err
			}
		case domain.StatusCancelled:
			if err := order.MerchantCancel(actor, cmd.Reason); err != nil {
				return queries.OrderDTO{}, err
			}
		default:
			return queries.OrderDTO{}, domain.ErrMerchantConfirmOnly
		}
	} else {
		from := order.Status
		if err := order.ChangeStatus(status, actor); err != nil {
			return queries.OrderDTO{}, err
		}
		// Admin confirm also auto-dispatches internal shipment.
		if from == domain.StatusNew && status == domain.StatusConfirmed {
			if err := order.DispatchInternalShipment(); err != nil {
				return queries.OrderDTO{}, err
			}
		}
	}
	if err := h.repo.Save(order); err != nil {
		return queries.OrderDTO{}, err
	}
	// Reload so history includes persisted events in order.
	saved, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return queries.OrderDTO{}, err
	}
	return queries.ToDTOWithHistory(saved, true), nil
}
