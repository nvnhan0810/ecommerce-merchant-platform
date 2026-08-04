package queries

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type GetProductQuery struct {
	ID                domain.ProductID
	OwnerMerchantID   string
	IncludeOrderFlags bool
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
	owner := strings.TrimSpace(q.OwnerMerchantID)
	if owner != "" && p.MerchantID != owner {
		return ProductDTO{}, domain.ErrProductNotFound
	}
	dto := toDTO(p, h.publicBase)
	if q.IncludeOrderFlags {
		hasOrders, err := h.repo.HasOrderItems(p.ID)
		if err != nil {
			return ProductDTO{}, err
		}
		canDelete := !hasOrders
		dto.HasOrders = &hasOrders
		dto.CanDelete = &canDelete
	}
	return dto, nil
}
