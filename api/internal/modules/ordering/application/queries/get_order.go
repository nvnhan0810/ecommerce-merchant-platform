package queries

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type GetOrderQuery struct {
	ID              domain.OrderID
	OwnerMerchantID string // when set, order must belong to this merchant
	OwnerUserID     string // when set, order must belong to this user
}

type GetOrderHandler struct {
	repo domain.OrderRepository
}

func NewGetOrderHandler(repo domain.OrderRepository) *GetOrderHandler {
	return &GetOrderHandler{repo: repo}
}

func (h *GetOrderHandler) Handle(_ context.Context, q GetOrderQuery) (OrderDTO, error) {
	order, err := h.repo.FindByID(q.ID)
	if err != nil {
		return OrderDTO{}, err
	}
	if owner := strings.TrimSpace(q.OwnerMerchantID); owner != "" && order.MerchantID != owner {
		return OrderDTO{}, domain.ErrOrderNotFound
	}
	if owner := strings.TrimSpace(q.OwnerUserID); owner != "" && order.UserID != owner {
		return OrderDTO{}, domain.ErrOrderNotFound
	}
	return ToDTOWithHistory(order, true), nil
}
