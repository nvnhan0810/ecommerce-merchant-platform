package queries

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type GetOrderQuery struct {
	ID              domain.OrderID
	OwnerMerchantID string // when set, order must belong to this merchant
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
	owner := strings.TrimSpace(q.OwnerMerchantID)
	if owner != "" && order.MerchantID != owner {
		return OrderDTO{}, domain.ErrOrderNotFound
	}
	return ToDTOWithHistory(order, true), nil
}
