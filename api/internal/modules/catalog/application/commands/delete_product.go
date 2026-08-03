package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type DeleteProductCommand struct {
	ID domain.ProductID
}

type DeleteProductHandler struct {
	repo domain.ProductRepository
}

func NewDeleteProductHandler(repo domain.ProductRepository) *DeleteProductHandler {
	return &DeleteProductHandler{repo: repo}
}

func (h *DeleteProductHandler) Handle(_ context.Context, cmd DeleteProductCommand) error {
	if _, err := h.repo.FindByID(cmd.ID); err != nil {
		return err
	}
	return h.repo.Delete(cmd.ID)
}
