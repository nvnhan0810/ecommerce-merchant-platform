package queries

import (
	"context"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type CategoryDTO struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	StatusLabel         string `json:"status_label"`
	CreatedByMerchantID string `json:"created_by_merchant_id"`
	CreatedAt           string `json:"created_at"`
}

func toCategoryDTO(c domain.Category) CategoryDTO {
	return CategoryDTO{
		ID:                  string(c.ID),
		Name:                c.Name,
		Status:              string(c.Status),
		StatusLabel:         c.Status.LabelVI(),
		CreatedByMerchantID: c.CreatedByMerchantID,
		CreatedAt:           c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

type ListCategoriesQuery struct {
	Status           string
	MerchantViewerID string // when set with MerchantAssignable, approved + own pending
	MerchantAssignable bool
	ApprovedOnly     bool
}

type ListCategoriesHandler struct {
	repo domain.CategoryRepository
}

func NewListCategoriesHandler(repo domain.CategoryRepository) *ListCategoriesHandler {
	return &ListCategoriesHandler{repo: repo}
}

func (h *ListCategoriesHandler) Handle(_ context.Context, q ListCategoriesQuery) ([]CategoryDTO, error) {
	filter := domain.CategoryListFilter{}
	if q.MerchantAssignable && strings.TrimSpace(q.MerchantViewerID) != "" {
		filter.IncludeApproved = true
		filter.MerchantViewerID = q.MerchantViewerID
	} else if q.ApprovedOnly {
		filter.Status = domain.CategoryStatusApproved
	} else if s := strings.TrimSpace(q.Status); s != "" {
		st, err := domain.ParseCategoryStatus(s)
		if err != nil {
			return nil, err
		}
		filter.Status = st
	}

	items, err := h.repo.List(filter)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryDTO, 0, len(items))
	for _, c := range items {
		out = append(out, toCategoryDTO(c))
	}
	return out, nil
}

type GetCategoryQuery struct {
	ID domain.CategoryID
}

type GetCategoryHandler struct {
	repo domain.CategoryRepository
}

func NewGetCategoryHandler(repo domain.CategoryRepository) *GetCategoryHandler {
	return &GetCategoryHandler{repo: repo}
}

func (h *GetCategoryHandler) Handle(_ context.Context, q GetCategoryQuery) (CategoryDTO, error) {
	c, err := h.repo.FindByID(q.ID)
	if err != nil {
		return CategoryDTO{}, err
	}
	return toCategoryDTO(c), nil
}
