package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type UserAddressRepository struct {
	pool *pgxpool.Pool
}

func NewUserAddressRepository(pool *pgxpool.Pool) *UserAddressRepository {
	return &UserAddressRepository{pool: pool}
}

func (r *UserAddressRepository) Save(a domain.UserAddress) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO user_addresses (id, user_id, recipient_name, phone_number, address_line, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			recipient_name = EXCLUDED.recipient_name,
			phone_number = EXCLUDED.phone_number,
			address_line = EXCLUDED.address_line,
			is_default = EXCLUDED.is_default,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, query,
		a.ID, a.UserID, a.RecipientName, a.PhoneNumber, a.AddressLine, a.IsDefault, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *UserAddressRepository) FindByID(id domain.AddressID) (domain.UserAddress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, recipient_name, phone_number, address_line, is_default, created_at, updated_at
		FROM user_addresses
		WHERE id = $1
	`
	var a domain.UserAddress
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.RecipientName, &a.PhoneNumber, &a.AddressLine, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.UserAddress{}, domain.ErrAddressNotFound
	}
	return a, err
}

func (r *UserAddressRepository) ListByUserID(userID domain.AccountID) ([]domain.UserAddress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, recipient_name, phone_number, address_line, is_default, created_at, updated_at
		FROM user_addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.UserAddress
	for rows.Next() {
		var a domain.UserAddress
		if err := rows.Scan(&a.ID, &a.UserID, &a.RecipientName, &a.PhoneNumber, &a.AddressLine, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *UserAddressRepository) Delete(id domain.AddressID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.pool.Exec(ctx, "DELETE FROM user_addresses WHERE id = $1", id)
	return err
}

func (r *UserAddressRepository) ClearDefault(userID domain.AccountID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.pool.Exec(ctx, "UPDATE user_addresses SET is_default = false WHERE user_id = $1", userID)
	return err
}
