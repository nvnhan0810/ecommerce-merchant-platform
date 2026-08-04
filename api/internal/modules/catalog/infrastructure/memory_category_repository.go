package infrastructure

import (
	"strings"
	"sync"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type InMemoryCategoryRepository struct {
	mu       sync.RWMutex
	items    map[domain.CategoryID]domain.Category
	order    []domain.CategoryID
	links    map[domain.ProductID][]domain.CategoryID
}

func NewInMemoryCategoryRepository() *InMemoryCategoryRepository {
	return &InMemoryCategoryRepository{
		items: map[domain.CategoryID]domain.Category{},
		order: []domain.CategoryID{},
		links: map[domain.ProductID][]domain.CategoryID{},
	}
}

func (r *InMemoryCategoryRepository) Save(category domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[category.ID]; !exists {
		r.order = append(r.order, category.ID)
	}
	r.items[category.ID] = category
	return nil
}

func (r *InMemoryCategoryRepository) FindByID(id domain.CategoryID) (domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.items[id]
	if !ok {
		return domain.Category{}, domain.ErrCategoryNotFound
	}
	return c, nil
}

func (r *InMemoryCategoryRepository) FindByName(name string) (domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	want := strings.ToLower(strings.TrimSpace(name))
	for _, id := range r.order {
		c := r.items[id]
		if strings.ToLower(c.Name) == want {
			return c, nil
		}
	}
	return domain.Category{}, domain.ErrCategoryNotFound
}

func (r *InMemoryCategoryRepository) List(filter domain.CategoryListFilter) ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Category, 0)
	viewer := strings.TrimSpace(filter.MerchantViewerID)
	for _, id := range r.order {
		c := r.items[id]
		if filter.IncludeApproved && viewer != "" {
			if c.Status == domain.CategoryStatusApproved ||
				(c.Status == domain.CategoryStatusPending && c.CreatedByMerchantID == viewer) {
				out = append(out, c)
			}
			continue
		}
		if filter.Status != "" && c.Status != filter.Status {
			continue
		}
		if mid := strings.TrimSpace(filter.CreatedByMerchantID); mid != "" && c.CreatedByMerchantID != mid {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *InMemoryCategoryRepository) Delete(id domain.CategoryID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return domain.ErrCategoryNotFound
	}
	delete(r.items, id)
	next := make([]domain.CategoryID, 0, len(r.order))
	for _, oid := range r.order {
		if oid != id {
			next = append(next, oid)
		}
	}
	r.order = next
	for pid, cats := range r.links {
		filtered := make([]domain.CategoryID, 0, len(cats))
		for _, cid := range cats {
			if cid != id {
				filtered = append(filtered, cid)
			}
		}
		r.links[pid] = filtered
	}
	return nil
}

func (r *InMemoryCategoryRepository) ListByProductIDs(productIDs []domain.ProductID) (map[domain.ProductID][]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[domain.ProductID][]domain.Category, len(productIDs))
	for _, pid := range productIDs {
		cats := make([]domain.Category, 0)
		for _, cid := range r.links[pid] {
			if c, ok := r.items[cid]; ok {
				cats = append(cats, c)
			}
		}
		out[pid] = cats
	}
	return out, nil
}

func (r *InMemoryCategoryRepository) SetProductCategories(productID domain.ProductID, categoryIDs []domain.CategoryID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]domain.CategoryID, len(categoryIDs))
	copy(cp, categoryIDs)
	r.links[productID] = cp
	return nil
}

func (r *InMemoryCategoryRepository) ListProductIDsByCategory(categoryID domain.CategoryID, approvedOnly bool, limit, offset int) ([]domain.ProductID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if approvedOnly {
		c, ok := r.items[categoryID]
		if !ok || c.Status != domain.CategoryStatusApproved {
			return []domain.ProductID{}, nil
		}
	}
	matched := make([]domain.ProductID, 0)
	for pid, cats := range r.links {
		for _, cid := range cats {
			if cid == categoryID {
				matched = append(matched, pid)
				break
			}
		}
	}
	if offset >= len(matched) {
		return []domain.ProductID{}, nil
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}
	return append([]domain.ProductID(nil), matched[offset:end]...), nil
}
