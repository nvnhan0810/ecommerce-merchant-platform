package commands

import (
	"context"
	"errors"
	"strings"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type CategoryResult struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	StatusLabel         string `json:"status_label"`
	CreatedByMerchantID string `json:"created_by_merchant_id"`
	CreatedAt           string `json:"created_at"`
}

func toCategoryResult(c domain.Category) CategoryResult {
	return CategoryResult{
		ID:                  string(c.ID),
		Name:                c.Name,
		Status:              string(c.Status),
		StatusLabel:         c.Status.LabelVI(),
		CreatedByMerchantID: c.CreatedByMerchantID,
		CreatedAt:           c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

type CreateCategoryCommand struct {
	Name                string
	CreatedByMerchantID string
	// When empty, admin creates as approved; merchant creates as pending.
	AsAdmin bool
}

type CreateCategoryHandler struct {
	repo domain.CategoryRepository
}

func NewCreateCategoryHandler(repo domain.CategoryRepository) *CreateCategoryHandler {
	return &CreateCategoryHandler{repo: repo}
}

func (h *CreateCategoryHandler) Handle(_ context.Context, cmd CreateCategoryCommand) (CategoryResult, error) {
	name := strings.TrimSpace(cmd.Name)
	if existing, err := h.repo.FindByName(name); err == nil {
		return CategoryResult{}, domain.ErrCategoryExists
	} else if !errors.Is(err, domain.ErrCategoryNotFound) {
		return CategoryResult{}, err
	} else {
		_ = existing
	}

	status := domain.CategoryStatusPending
	createdBy := strings.TrimSpace(cmd.CreatedByMerchantID)
	if cmd.AsAdmin {
		status = domain.CategoryStatusApproved
		createdBy = ""
	}
	category, err := domain.NewCategory(name, createdBy, status)
	if err != nil {
		return CategoryResult{}, err
	}
	if err := h.repo.Save(category); err != nil {
		return CategoryResult{}, err
	}
	return toCategoryResult(category), nil
}

type UpdateCategoryCommand struct {
	ID   domain.CategoryID
	Name string
}

type UpdateCategoryHandler struct {
	repo domain.CategoryRepository
}

func NewUpdateCategoryHandler(repo domain.CategoryRepository) *UpdateCategoryHandler {
	return &UpdateCategoryHandler{repo: repo}
}

func (h *UpdateCategoryHandler) Handle(_ context.Context, cmd UpdateCategoryCommand) (CategoryResult, error) {
	category, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return CategoryResult{}, err
	}
	name := strings.TrimSpace(cmd.Name)
	if other, err := h.repo.FindByName(name); err == nil && other.ID != category.ID {
		return CategoryResult{}, domain.ErrCategoryExists
	} else if err != nil && !errors.Is(err, domain.ErrCategoryNotFound) {
		return CategoryResult{}, err
	}
	if err := category.Rename(name); err != nil {
		return CategoryResult{}, err
	}
	if err := h.repo.Save(category); err != nil {
		return CategoryResult{}, err
	}
	return toCategoryResult(category), nil
}

type UpdateCategoryStatusCommand struct {
	ID     domain.CategoryID
	Status domain.CategoryStatus
}

type UpdateCategoryStatusHandler struct {
	repo domain.CategoryRepository
}

func NewUpdateCategoryStatusHandler(repo domain.CategoryRepository) *UpdateCategoryStatusHandler {
	return &UpdateCategoryStatusHandler{repo: repo}
}

func (h *UpdateCategoryStatusHandler) Handle(_ context.Context, cmd UpdateCategoryStatusCommand) (CategoryResult, error) {
	category, err := h.repo.FindByID(cmd.ID)
	if err != nil {
		return CategoryResult{}, err
	}
	if err := category.SetStatus(cmd.Status); err != nil {
		return CategoryResult{}, err
	}
	if err := h.repo.Save(category); err != nil {
		return CategoryResult{}, err
	}
	return toCategoryResult(category), nil
}

type DeleteCategoryCommand struct {
	ID domain.CategoryID
}

type DeleteCategoryHandler struct {
	repo domain.CategoryRepository
}

func NewDeleteCategoryHandler(repo domain.CategoryRepository) *DeleteCategoryHandler {
	return &DeleteCategoryHandler{repo: repo}
}

func (h *DeleteCategoryHandler) Handle(_ context.Context, cmd DeleteCategoryCommand) error {
	return h.repo.Delete(cmd.ID)
}

// SetProductCategoriesCommand replaces all category links for a product.
type SetProductCategoriesCommand struct {
	ProductID       domain.ProductID
	CategoryIDs     []string
	OwnerMerchantID string // empty = admin (any category assignable)
}

type SetProductCategoriesHandler struct {
	products   domain.ProductRepository
	categories domain.CategoryRepository
}

func NewSetProductCategoriesHandler(products domain.ProductRepository, categories domain.CategoryRepository) *SetProductCategoriesHandler {
	return &SetProductCategoriesHandler{products: products, categories: categories}
}

func (h *SetProductCategoriesHandler) Handle(_ context.Context, cmd SetProductCategoriesCommand) error {
	product, err := h.products.FindByID(cmd.ProductID)
	if err != nil {
		return err
	}
	if owner := strings.TrimSpace(cmd.OwnerMerchantID); owner != "" && product.MerchantID != owner {
		return domain.ErrProductNotFound
	}

	ids := make([]domain.CategoryID, 0, len(cmd.CategoryIDs))
	seen := map[domain.CategoryID]struct{}{}
	for _, raw := range cmd.CategoryIDs {
		id, err := domain.ParseCategoryID(raw)
		if err != nil {
			return err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		cat, err := h.categories.FindByID(id)
		if err != nil {
			return err
		}
		if !cat.AssignableBy(cmd.OwnerMerchantID) {
			return domain.ErrCategoryNotAssignable
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return h.categories.SetProductCategories(cmd.ProductID, ids)
}

type RemoveProductCategoryCommand struct {
	ProductID       domain.ProductID
	CategoryID      domain.CategoryID
	OwnerMerchantID string
}

type RemoveProductCategoryHandler struct {
	products   domain.ProductRepository
	categories domain.CategoryRepository
}

func NewRemoveProductCategoryHandler(products domain.ProductRepository, categories domain.CategoryRepository) *RemoveProductCategoryHandler {
	return &RemoveProductCategoryHandler{products: products, categories: categories}
}

func (h *RemoveProductCategoryHandler) Handle(_ context.Context, cmd RemoveProductCategoryCommand) error {
	product, err := h.products.FindByID(cmd.ProductID)
	if err != nil {
		return err
	}
	if owner := strings.TrimSpace(cmd.OwnerMerchantID); owner != "" && product.MerchantID != owner {
		return domain.ErrProductNotFound
	}

	linked, err := h.categories.ListByProductIDs([]domain.ProductID{cmd.ProductID})
	if err != nil {
		return err
	}
	next := make([]domain.CategoryID, 0, len(linked[cmd.ProductID]))
	found := false
	for _, c := range linked[cmd.ProductID] {
		if c.ID == cmd.CategoryID {
			found = true
			continue
		}
		next = append(next, c.ID)
	}
	if !found {
		return domain.ErrCategoryNotFound
	}
	return h.categories.SetProductCategories(cmd.ProductID, next)
}
