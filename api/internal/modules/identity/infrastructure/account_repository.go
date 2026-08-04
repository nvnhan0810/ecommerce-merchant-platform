package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/identity/domain"
)

type PostgresAccountRepository struct {
	pool  *pgxpool.Pool
	table string
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresAccountRepository {
	return &PostgresAccountRepository{pool: pool, table: "users"}
}

func NewPostgresMerchantRepository(pool *pgxpool.Pool) *PostgresAccountRepository {
	return &PostgresAccountRepository{pool: pool, table: "merchants"}
}

func NewPostgresAdminRepository(pool *pgxpool.Pool) *PostgresAccountRepository {
	return &PostgresAccountRepository{pool: pool, table: "admins"}
}

func (r *PostgresAccountRepository) isMerchant() bool {
	return r.table == "merchants"
}

func (r *PostgresAccountRepository) Save(account domain.Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if r.isMerchant() {
		q := `
			INSERT INTO merchants (
				id, email, display_name, password_hash, created_at,
				avatar_key, address_line, country_code, province_code, ward_code
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (id) DO UPDATE SET
				email = EXCLUDED.email,
				display_name = EXCLUDED.display_name,
				password_hash = EXCLUDED.password_hash,
				avatar_key = EXCLUDED.avatar_key,
				address_line = EXCLUDED.address_line,
				country_code = EXCLUDED.country_code,
				province_code = EXCLUDED.province_code,
				ward_code = EXCLUDED.ward_code
		`
		_, err := r.pool.Exec(ctx, q,
			account.ID, account.Email, account.DisplayName, account.PasswordHash, account.CreatedAt,
			account.AvatarKey, account.AddressLine, account.CountryCode, account.ProvinceCode, account.WardCode,
		)
		return err
	}
	q := fmt.Sprintf(`
		INSERT INTO %s (id, email, display_name, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name,
			password_hash = EXCLUDED.password_hash
	`, r.table)
	_, err := r.pool.Exec(ctx, q, account.ID, account.Email, account.DisplayName, account.PasswordHash, account.CreatedAt)
	return err
}

func (r *PostgresAccountRepository) FindByEmail(email string) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	email = strings.ToLower(strings.TrimSpace(email))
	if r.isMerchant() {
		q := `
			SELECT id, email, display_name, password_hash, created_at,
				avatar_key, address_line, country_code, province_code, ward_code
			FROM merchants WHERE email = $1
		`
		return scanMerchantAccount(r.pool.QueryRow(ctx, q, email))
	}
	q := fmt.Sprintf(`
		SELECT id, email, display_name, password_hash, created_at
		FROM %s WHERE email = $1
	`, r.table)
	return scanAccount(r.pool.QueryRow(ctx, q, email))
}

func (r *PostgresAccountRepository) FindByID(id domain.AccountID) (domain.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if r.isMerchant() {
		q := `
			SELECT id, email, display_name, password_hash, created_at,
				avatar_key, address_line, country_code, province_code, ward_code
			FROM merchants WHERE id = $1
		`
		return scanMerchantAccount(r.pool.QueryRow(ctx, q, id))
	}
	q := fmt.Sprintf(`
		SELECT id, email, display_name, password_hash, created_at
		FROM %s WHERE id = $1
	`, r.table)
	return scanAccount(r.pool.QueryRow(ctx, q, id))
}

func (r *PostgresAccountRepository) List() ([]domain.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if r.isMerchant() {
		q := `
			SELECT id, email, display_name, password_hash, created_at,
				avatar_key, address_line, country_code, province_code, ward_code
			FROM merchants
			ORDER BY created_at ASC
		`
		rows, err := r.pool.Query(ctx, q)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := make([]domain.Account, 0)
		for rows.Next() {
			a, err := scanMerchantAccount(rows)
			if err != nil {
				return nil, err
			}
			out = append(out, a)
		}
		return out, rows.Err()
	}
	q := fmt.Sprintf(`
		SELECT id, email, display_name, password_hash, created_at
		FROM %s
		ORDER BY created_at ASC
	`, r.table)
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Account, 0)
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *PostgresAccountRepository) Delete(id domain.AccountID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.table)
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *PostgresAccountRepository) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, r.table)
	var n int
	err := r.pool.QueryRow(ctx, q).Scan(&n)
	return n, err
}

type accountScannable interface {
	Scan(dest ...any) error
}

func scanAccount(row accountScannable) (domain.Account, error) {
	var (
		id, email, displayName, passwordHash string
		createdAt                            time.Time
	)
	if err := row.Scan(&id, &email, &displayName, &passwordHash, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Account{}, domain.ErrAccountNotFound
		}
		return domain.Account{}, err
	}
	return domain.Account{
		ID:           domain.AccountID(id),
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt.UTC(),
	}, nil
}

func scanMerchantAccount(row accountScannable) (domain.Account, error) {
	var (
		id, email, displayName, passwordHash string
		createdAt                            time.Time
		avatarKey, addressLine, countryCode  string
		provinceCode, wardCode               string
	)
	if err := row.Scan(
		&id, &email, &displayName, &passwordHash, &createdAt,
		&avatarKey, &addressLine, &countryCode, &provinceCode, &wardCode,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Account{}, domain.ErrAccountNotFound
		}
		return domain.Account{}, err
	}
	return domain.Account{
		ID:           domain.AccountID(id),
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt.UTC(),
		AvatarKey:    avatarKey,
		AddressLine:  addressLine,
		CountryCode:  countryCode,
		ProvinceCode: provinceCode,
		WardCode:     wardCode,
	}, nil
}

type InMemoryAccountRepository struct {
	mu    sync.RWMutex
	items map[domain.AccountID]domain.Account
}

func NewInMemoryAccountRepository() *InMemoryAccountRepository {
	return &InMemoryAccountRepository{items: map[domain.AccountID]domain.Account{}}
}

func (r *InMemoryAccountRepository) Save(account domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[account.ID] = account
	return nil
}

func (r *InMemoryAccountRepository) FindByEmail(email string) (domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	email = strings.ToLower(strings.TrimSpace(email))
	for _, a := range r.items {
		if a.Email == email {
			return a, nil
		}
	}
	return domain.Account{}, domain.ErrAccountNotFound
}

func (r *InMemoryAccountRepository) FindByID(id domain.AccountID) (domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.items[id]
	if !ok {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return a, nil
}

func (r *InMemoryAccountRepository) List() ([]domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Account, 0, len(r.items))
	for _, a := range r.items {
		out = append(out, a)
	}
	return out, nil
}

func (r *InMemoryAccountRepository) Delete(id domain.AccountID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return domain.ErrAccountNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *InMemoryAccountRepository) Count() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items), nil
}
