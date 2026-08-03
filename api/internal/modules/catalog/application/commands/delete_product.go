package commands

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/storage"
)

type DeleteProductCommand struct {
	ID domain.ProductID
}

type DeleteProductHandler struct {
	repo  domain.ProductRepository
	store storage.ObjectStore
}

func NewDeleteProductHandler(repo domain.ProductRepository, store storage.ObjectStore) *DeleteProductHandler {
	return &DeleteProductHandler{repo: repo, store: store}
}

func (h *DeleteProductHandler) Handle(ctx context.Context, cmd DeleteProductCommand) error {
	product, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return err
	}
	if err := h.repo.Delete(cmd.ID); err != nil {
		return err
	}
	if product.ImageKey != "" && h.store != nil && h.store.Enabled() {
		_ = h.store.Delete(ctx, product.ImageKey)
	}
	return nil
}
