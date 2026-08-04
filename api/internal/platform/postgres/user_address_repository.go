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

const addressSelect = `
	SELECT
		a.id, a.user_id, a.address_line,
		a.country_code, a.province_code, a.ward_code,
		COALESCE(c.name, ''), COALESCE(p.name, ''), COALESCE(w.name, ''),
		a.latitude, a.longitude, a.is_default, a.created_at, a.updated_at
	FROM user_addresses a
	LEFT JOIN countries c ON c.code = a.country_code
	LEFT JOIN provinces p ON p.code = a.province_code
	LEFT JOIN wards w ON w.code = a.ward_code
`

type addressScanner interface {
	Scan(dest ...any) error
}

func scanAddress(row addressScanner) (domain.UserAddress, error) {
	var a domain.UserAddress
	var provinceCode, wardCode *string
	err := row.Scan(
		&a.ID, &a.UserID, &a.AddressLine,
		&a.CountryCode, &provinceCode, &wardCode,
		&a.CountryName, &a.ProvinceName, &a.WardName,
		&a.Latitude, &a.Longitude, &a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return domain.UserAddress{}, err
	}
	if provinceCode != nil {
		a.ProvinceCode = *provinceCode
	}
	if wardCode != nil {
		a.WardCode = *wardCode
	}
	return a, nil
}

func (r *UserAddressRepository) Save(a domain.UserAddress) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO user_addresses (
			id, user_id, address_line,
			country_code, province_code, ward_code, latitude, longitude,
			is_default, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			address_line = EXCLUDED.address_line,
			country_code = EXCLUDED.country_code,
			province_code = EXCLUDED.province_code,
			ward_code = EXCLUDED.ward_code,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			is_default = EXCLUDED.is_default,
			updated_at = EXCLUDED.updated_at
	`
	var provinceCode, wardCode any
	if a.ProvinceCode != "" {
		provinceCode = a.ProvinceCode
	}
	if a.WardCode != "" {
		wardCode = a.WardCode
	}
	_, err := r.pool.Exec(ctx, query,
		a.ID, a.UserID, a.AddressLine,
		a.CountryCode, provinceCode, wardCode, a.Latitude, a.Longitude,
		a.IsDefault, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *UserAddressRepository) FindByID(id domain.AddressID) (domain.UserAddress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a, err := scanAddress(r.pool.QueryRow(ctx, addressSelect+` WHERE a.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.UserAddress{}, domain.ErrAddressNotFound
	}
	return a, err
}

func (r *UserAddressRepository) ListByUserID(userID domain.AccountID) ([]domain.UserAddress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, addressSelect+`
		WHERE a.user_id = $1
		ORDER BY a.is_default DESC, a.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.UserAddress
	for rows.Next() {
		a, err := scanAddress(rows)
		if err != nil {
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
