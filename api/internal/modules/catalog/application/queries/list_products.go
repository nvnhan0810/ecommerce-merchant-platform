package queries

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
	identitydomain "github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
	"github.com/nvnhan0810/ecomerce-api/internal/platform/mediaurl"
)

type ListProductsQuery struct {
	Limit               int
	Offset              int
	MerchantID          string
	CategoryID          string
	IncludeOrderFlags   bool
	IncludeAllCategories bool // admin/merchant: show pending categories too
}

type ProductDTO struct {
	ID                   string        `json:"id"`
	MerchantID           string        `json:"merchant_id"`
	Name                 string        `json:"name"`
	Description          string        `json:"description"`
	PriceCents           int64         `json:"price_cents"`
	Currency             string        `json:"currency"`
	Stock                int           `json:"stock"`
	ImageKey             string        `json:"image_key"`
	ImageURL             string        `json:"image_url"`
	Categories           []CategoryDTO `json:"categories"`
	MerchantDisplayName  string        `json:"merchant_display_name,omitempty"`
	MerchantAvatarURL    string        `json:"merchant_avatar_url,omitempty"`
	MerchantProvinceName string        `json:"merchant_province_name,omitempty"`
	HasOrders            *bool         `json:"has_orders,omitempty"`
	CanDelete            *bool         `json:"can_delete,omitempty"`
}

type ListProductsHandler struct {
	repo       domain.ProductRepository
	categories domain.CategoryRepository
	merchants  identitydomain.AccountRepository
	geo        identitydomain.GeoRepository
	publicBase string
}

func NewListProductsHandler(
	repo domain.ProductRepository,
	categories domain.CategoryRepository,
	merchants identitydomain.AccountRepository,
	geo identitydomain.GeoRepository,
	publicBase string,
) *ListProductsHandler {
	return &ListProductsHandler{repo: repo, categories: categories, merchants: merchants, geo: geo, publicBase: publicBase}
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

	categoryID := strings.TrimSpace(q.CategoryID)
	merchantID := strings.TrimSpace(q.MerchantID)

	if categoryID != "" && h.categories != nil {
		cid, err := domain.ParseCategoryID(categoryID)
		if err != nil {
			return nil, err
		}
		// Public category filter only matches approved categories.
		ids, err := h.categories.ListProductIDsByCategory(cid, !q.IncludeAllCategories, limit, offset)
		if err != nil {
			return nil, err
		}
		items, err = h.repo.ListByIDs(ids)
		if err != nil {
			return nil, err
		}
		if merchantID != "" {
			filtered := make([]domain.Product, 0, len(items))
			for _, p := range items {
				if p.MerchantID == merchantID {
					filtered = append(filtered, p)
				}
			}
			items = filtered
		}
	} else if merchantID != "" {
		items, err = h.repo.ListByMerchant(merchantID, limit, offset)
	} else {
		items, err = h.repo.List(limit, offset)
	}
	if err != nil {
		return nil, err
	}

	cache := map[string]merchantPublicInfo{}
	catMap, err := h.loadCategories(items)
	if err != nil {
		return nil, err
	}

	out := make([]ProductDTO, 0, len(items))
	for _, p := range items {
		dto := toDTO(p, h.publicBase)
		dto.Categories = filterCategories(catMap[p.ID], q.IncludeAllCategories)
		h.enrichMerchant(&dto, p.MerchantID, cache)
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

func (h *ListProductsHandler) loadCategories(items []domain.Product) (map[domain.ProductID][]domain.Category, error) {
	if h.categories == nil || len(items) == 0 {
		return map[domain.ProductID][]domain.Category{}, nil
	}
	ids := make([]domain.ProductID, len(items))
	for i, p := range items {
		ids[i] = p.ID
	}
	return h.categories.ListByProductIDs(ids)
}

func filterCategories(cats []domain.Category, includeAll bool) []CategoryDTO {
	out := make([]CategoryDTO, 0, len(cats))
	for _, c := range cats {
		if !includeAll && c.Status != domain.CategoryStatusApproved {
			continue
		}
		out = append(out, toCategoryDTO(c))
	}
	return out
}

type merchantPublicInfo struct {
	displayName  string
	avatarURL    string
	provinceName string
}

func (h *ListProductsHandler) enrichMerchant(dto *ProductDTO, merchantID string, cache map[string]merchantPublicInfo) {
	merchantID = strings.TrimSpace(merchantID)
	if merchantID == "" || h.merchants == nil {
		return
	}
	if info, ok := cache[merchantID]; ok {
		dto.MerchantDisplayName = info.displayName
		dto.MerchantAvatarURL = info.avatarURL
		dto.MerchantProvinceName = info.provinceName
		return
	}
	info := h.lookupMerchant(merchantID)
	cache[merchantID] = info
	dto.MerchantDisplayName = info.displayName
	dto.MerchantAvatarURL = info.avatarURL
	dto.MerchantProvinceName = info.provinceName
}

func (h *ListProductsHandler) lookupMerchant(merchantID string) merchantPublicInfo {
	account, err := h.merchants.FindByID(identitydomain.AccountID(merchantID))
	if err != nil {
		return merchantPublicInfo{}
	}
	info := merchantPublicInfo{
		displayName: account.DisplayName,
		avatarURL:   mediaurl.Absolute(h.publicBase, account.AvatarKey),
	}
	if h.geo != nil && strings.TrimSpace(account.ProvinceCode) != "" {
		if p, err := h.geo.GetProvince(account.ProvinceCode); err == nil {
			info.provinceName = p.Name
		}
	}
	return info
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
		Categories:  []CategoryDTO{},
	}
}
