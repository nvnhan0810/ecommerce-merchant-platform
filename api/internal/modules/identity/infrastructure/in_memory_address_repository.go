package infrastructure

import (
	"sort"
	"sync"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type InMemoryAddressRepository struct {
	mu   sync.RWMutex
	data map[domain.AddressID]domain.UserAddress
}

func NewInMemoryAddressRepository() *InMemoryAddressRepository {
	return &InMemoryAddressRepository{
		data: make(map[domain.AddressID]domain.UserAddress),
	}
}

func (r *InMemoryAddressRepository) Save(a domain.UserAddress) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[a.ID] = a
	return nil
}

func (r *InMemoryAddressRepository) FindByID(id domain.AddressID) (domain.UserAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.data[id]
	if !ok {
		return domain.UserAddress{}, domain.ErrAddressNotFound
	}
	return a, nil
}

func (r *InMemoryAddressRepository) ListByUserID(userID domain.AccountID) ([]domain.UserAddress, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var items []domain.UserAddress
	for _, a := range r.data {
		if a.UserID == userID {
			items = append(items, a)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDefault != items[j].IsDefault {
			return items[i].IsDefault
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

func (r *InMemoryAddressRepository) Delete(id domain.AddressID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

func (r *InMemoryAddressRepository) ClearDefault(userID domain.AccountID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, a := range r.data {
		if a.UserID == userID && a.IsDefault {
			a.IsDefault = false
			r.data[id] = a
		}
	}
	return nil
}
