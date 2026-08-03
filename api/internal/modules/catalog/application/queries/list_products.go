package queries

import (
	"context"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/mediaurl"
)

type ListProductsQuery struct {
	Limit  int
	Offset int
}

type ProductDTO struct {
	ID          string `json:"id"`
	MerchantID  string `json:"merchant_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Currency    string `json:"currency"`
	Stock       int    `json:"stock"`
	ImageKey    string `json:"image_key"`
	ImageURL    string `json:"image_url"`
}

type ListProductsHandler struct {
	repo       domain.ProductRepository
	publicBase string
}

func NewListProductsHandler(repo domain.ProductRepository, publicBase string) *ListProductsHandler {
	return &ListProductsHandler{repo: repo, publicBase: publicBase}
}

func (h *ListProductsHandler) Handle(_ context.Context, q ListProductsQuery) ([]ProductDTO, error) {
	limit := q.Limit
	if limit <= 0 || limit > 500 {
		limit = 20
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}
	items, err := h.repo.List(limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]ProductDTO, 0, len(items))
	for _, p := range items {
		out = append(out, toDTO(p, h.publicBase))
	}
	return out, nil
}

func toDTO(p domain.Product, publicBase string) ProductDTO {
	return ProductDTO{
		ID:          string(p.ID),
		MerchantID:  p.MerchantID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.Price.AmountCents,
		Currency:    p.Price.Currency,
		Stock:       p.Stock,
		ImageKey:    p.ImageKey,
		ImageURL:    mediaurl.Absolute(publicBase, p.ImageKey),
	}
}
