package infrastructure

import (
	"sync"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

// InMemoryProductRepository is used in unit/feature tests.
type InMemoryProductRepository struct {
	mu      sync.RWMutex
	items   map[domain.ProductID]domain.Product
	order   []domain.ProductID
	ordered map[domain.ProductID]bool
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		items:   map[domain.ProductID]domain.Product{},
		order:   []domain.ProductID{},
		ordered: map[domain.ProductID]bool{},
	}
}

func (r *InMemoryProductRepository) Save(product domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[product.ID]; !exists {
		r.order = append(r.order, product.ID)
	}
	r.items[product.ID] = product
	return nil
}

func (r *InMemoryProductRepository) FindByID(id domain.ProductID) (domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.items[id]
	if !ok {
		return domain.Product{}, domain.ErrProductNotFound
	}
	return p, nil
}

func (r *InMemoryProductRepository) List(limit, offset int) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if offset >= len(r.order) {
		return []domain.Product{}, nil
	}
	end := offset + limit
	if end > len(r.order) {
		end = len(r.order)
	}
	out := make([]domain.Product, 0, end-offset)
	for _, id := range r.order[offset:end] {
		out = append(out, r.items[id])
	}
	return out, nil
}

func (r *InMemoryProductRepository) ListByMerchant(merchantID string, limit, offset int) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.Product, 0)
	for _, id := range r.order {
		p := r.items[id]
		if p.MerchantID == merchantID {
			filtered = append(filtered, p)
		}
	}
	if offset >= len(filtered) {
		return []domain.Product{}, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return append([]domain.Product(nil), filtered[offset:end]...), nil
}

func (r *InMemoryProductRepository) HasOrderItems(id domain.ProductID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ordered[id], nil
}

// MarkOrdered marks a product as referenced by an order (tests only).
func (r *InMemoryProductRepository) MarkOrdered(id domain.ProductID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ordered == nil {
		r.ordered = map[domain.ProductID]bool{}
	}
	r.ordered[id] = true
}

func (r *InMemoryProductRepository) Count() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items), nil
}

func (r *InMemoryProductRepository) Delete(id domain.ProductID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return domain.ErrProductNotFound
	}
	delete(r.items, id)
	next := make([]domain.ProductID, 0, len(r.order))
	for _, existing := range r.order {
		if existing != id {
			next = append(next, existing)
		}
	}
	r.order = next
	return nil
}
