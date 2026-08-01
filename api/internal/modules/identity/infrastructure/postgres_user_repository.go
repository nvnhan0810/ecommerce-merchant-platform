package infrastructure

import (
	"context"
	"errors"
	"strings"
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
		INSERT INTO users (id, email, display_name, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			role = EXCLUDED.role
	`, user.ID, user.Email, user.DisplayName, string(user.Role), user.CreatedAt)
	return err
}

func (r *PostgresUserRepository) FindByEmail(email string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	email = strings.ToLower(strings.TrimSpace(email))
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, created_at
		FROM users WHERE email = $1
	`, email)
	return scanUser(row)
}

func (r *PostgresUserRepository) ListByRole(role domain.Role) ([]domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, display_name, role, created_at
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
		id, email, displayName, role string
		createdAt                    time.Time
	)
	if err := row.Scan(&id, &email, &displayName, &role, &createdAt); err != nil {
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
		ID:          domain.UserID(id),
		Email:       email,
		DisplayName: displayName,
		Role:        parsedRole,
		CreatedAt:   createdAt.UTC(),
	}, nil
}
