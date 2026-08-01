package infrastructure

import (
	"sync"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

// InMemoryProductRepository is used in unit/feature tests.
type InMemoryProductRepository struct {
	mu    sync.RWMutex
	items map[domain.ProductID]domain.Product
	order []domain.ProductID
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		items: map[domain.ProductID]domain.Product{},
		order: []domain.ProductID{},
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

func (r *InMemoryProductRepository) Count() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items), nil
}
