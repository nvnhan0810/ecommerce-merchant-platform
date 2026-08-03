package infrastructure

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Save(user domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, display_name, role, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			role = EXCLUDED.role,
			password_hash = EXCLUDED.password_hash
	`, user.ID, user.Email, user.DisplayName, string(user.Role), user.PasswordHash, user.CreatedAt)
	return err
}

func (r *PostgresUserRepository) FindByEmail(email string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	email = strings.ToLower(strings.TrimSpace(email))
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, password_hash, created_at
		FROM users WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *PostgresUserRepository) FindByID(id domain.UserID) (domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, password_hash, created_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *PostgresUserRepository) ListByRole(role domain.Role) ([]domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, display_name, role, password_hash, created_at
		FROM users WHERE role = $1
		ORDER BY created_at ASC
	`, string(role))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *PostgresUserRepository) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

type userScannable interface {
	Scan(dest ...any) error
}

func scanUser(row userScannable) (domain.User, error) {
	var (
		id, email, displayName, role, passwordHash string
		createdAt                                  time.Time
	)
	if err := row.Scan(&id, &email, &displayName, &role, &passwordHash, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}
	parsedRole, err := domain.ParseRole(role)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		ID:           domain.UserID(id),
		Email:        email,
		DisplayName:  displayName,
		Role:         parsedRole,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt.UTC(),
	}, nil
}

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

func (r *InMemoryUserRepository) FindByID(id domain.UserID) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.items[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return u, nil
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
