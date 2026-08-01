package infrastructure

import (
	"strings"
	"sync"

	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	items map[domain.UserID]domain.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{items: map[domain.UserID]domain.User{}}
}

func (r *InMemoryUserRepository) Save(user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindByEmail(email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	email = strings.ToLower(strings.TrimSpace(email))
	for _, u := range r.items {
		if u.Email == email {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}

func (r *InMemoryUserRepository) ListByRole(role domain.Role) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.User, 0)
	for _, u := range r.items {
		if u.Role == role {
			out = append(out, u)
		}
	}
	return out, nil
}

func (r *InMemoryUserRepository) Count() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items), nil
}
