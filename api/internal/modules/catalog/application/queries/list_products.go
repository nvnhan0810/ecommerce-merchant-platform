package queries

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/mediaurl"
)

type ListProductsQuery struct {
	Limit             int
	Offset            int
	MerchantID        string
	IncludeOrderFlags bool
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
	HasOrders   *bool  `json:"has_orders,omitempty"`
	CanDelete   *bool  `json:"can_delete,omitempty"`
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

	var (
		items []domain.Product
		err   error
	)
	merchantID := strings.TrimSpace(q.MerchantID)
	if merchantID != "" {
		items, err = h.repo.ListByMerchant(merchantID, limit, offset)
	} else {
		items, err = h.repo.List(limit, offset)
	}
	if err != nil {
		return nil, err
	}

	out := make([]ProductDTO, 0, len(items))
	for _, p := range items {
		dto := toDTO(p, h.publicBase)
		if q.IncludeOrderFlags {
			hasOrders, err := h.repo.HasOrderItems(p.ID)
			if err != nil {
				return nil, err
			}
			canDelete := !hasOrders
			dto.HasOrders = &hasOrders
			dto.CanDelete = &canDelete
		}
		out = append(out, dto)
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
