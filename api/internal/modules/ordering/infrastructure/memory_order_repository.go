package infrastructure

import (
	"sync"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type InMemoryOrderRepository struct {
	mu    sync.RWMutex
	items map[domain.OrderID]domain.Order
	order []domain.OrderID
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		items: map[domain.OrderID]domain.Order{},
		order: []domain.OrderID{},
	}
}

func (r *InMemoryOrderRepository) Save(order domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[order.ID]; !ok {
		r.order = append(r.order, order.ID)
	}
	r.items[order.ID] = order
	return nil
}

func (r *InMemoryOrderRepository) FindByID(id domain.OrderID) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.items[id]
	if !ok {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return o, nil
}

func (r *InMemoryOrderRepository) List(limit, offset int) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if offset >= len(r.order) {
		return []domain.Order{}, nil
	}
	end := offset + limit
	if end > len(r.order) {
		end = len(r.order)
	}
	out := make([]domain.Order, 0, end-offset)
	for _, id := range r.order[offset:end] {
		out = append(out, r.items[id])
	}
	return out, nil
}

func (r *InMemoryOrderRepository) Count() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items), nil
}
