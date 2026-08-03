package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type GetProductQuery struct {
	ID domain.ProductID
}

type GetProductHandler struct {
	repo       domain.ProductRepository
	publicBase string
}

func NewGetProductHandler(repo domain.ProductRepository, publicBase string) *GetProductHandler {
	return &GetProductHandler{repo: repo, publicBase: publicBase}
}

func (h *GetProductHandler) Handle(_ context.Context, q GetProductQuery) (ProductDTO, error) {
	p, err := h.repo.FindByID(q.ID)
	if err != nil {
		return ProductDTO{}, err
	}
	return toDTO(p, h.publicBase), nil
}
