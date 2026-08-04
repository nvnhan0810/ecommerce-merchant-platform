package queries

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type GetProductQuery struct {
	ID                   domain.ProductID
	OwnerMerchantID      string
	IncludeOrderFlags    bool
	IncludeAllCategories bool
}

type GetProductHandler struct {
	repo       domain.ProductRepository
	categories domain.CategoryRepository
	merchants  identitydomain.AccountRepository
	geo        identitydomain.GeoRepository
	publicBase string
}

func NewGetProductHandler(
	repo domain.ProductRepository,
	categories domain.CategoryRepository,
	merchants identitydomain.AccountRepository,
	geo identitydomain.GeoRepository,
	publicBase string,
) *GetProductHandler {
	return &GetProductHandler{repo: repo, categories: categories, merchants: merchants, geo: geo, publicBase: publicBase}
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
	if h.categories != nil {
		linked, err := h.categories.ListByProductIDs([]domain.ProductID{p.ID})
		if err != nil {
			return ProductDTO{}, err
		}
		dto.Categories = filterCategories(linked[p.ID], q.IncludeAllCategories)
	}
	if h.merchants != nil {
		list := &ListProductsHandler{merchants: h.merchants, geo: h.geo, publicBase: h.publicBase}
		list.enrichMerchant(&dto, p.MerchantID, map[string]merchantPublicInfo{})
	}
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
