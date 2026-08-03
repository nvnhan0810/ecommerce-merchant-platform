package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type GetOrderQuery struct {
	ID domain.OrderID
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
	return ToDTOWithHistory(order, true), nil
}
