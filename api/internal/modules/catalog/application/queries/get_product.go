package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type GetProductQuery struct {
	ID domain.ProductID
}

type GetProductHandler struct {
	repo domain.ProductRepository
}

func NewGetProductHandler(repo domain.ProductRepository) *GetProductHandler {
	return &GetProductHandler{repo: repo}
}

func (h *GetProductHandler) Handle(_ context.Context, q GetProductQuery) (ProductDTO, error) {
	p, err := h.repo.FindByID(q.ID)
	if err != nil {
		return ProductDTO{}, err
	}
	return ProductDTO{
		ID:          string(p.ID),
		MerchantID:  p.MerchantID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.Price.AmountCents,
		Currency:    p.Price.Currency,
		Stock:       p.Stock,
	}, nil
}
