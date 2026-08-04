package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool}
}

func (r *PostgresProductRepository) Save(product domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO products (id, merchant_id, name, description, price_cents, currency, stock, image_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			merchant_id = EXCLUDED.merchant_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			price_cents = EXCLUDED.price_cents,
			currency = EXCLUDED.currency,
			stock = EXCLUDED.stock,
			image_key = EXCLUDED.image_key
	`, product.ID, product.MerchantID, product.Name, product.Description,
		product.Price.AmountCents, product.Price.Currency, product.Stock, product.ImageKey, product.CreatedAt)
	return err
}

func (r *PostgresProductRepository) FindByID(id domain.ProductID) (domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := r.pool.QueryRow(ctx, `
		SELECT id, merchant_id, name, description, price_cents, currency, stock, image_key, created_at
		FROM products WHERE id = $1
	`, id)
	return scanProduct(row)
}

func (r *PostgresProductRepository) List(limit, offset int) ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, name, description, price_cents, currency, stock, image_key, created_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Product, 0)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PostgresProductRepository) ListByMerchant(merchantID string, limit, offset int) ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := r.pool.Query(ctx, `
		SELECT id, merchant_id, name, description, price_cents, currency, stock, image_key, created_at
		FROM products
		WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, merchantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Product, 0)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PostgresProductRepository) HasOrderItems(id domain.ProductID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM order_items WHERE product_id = $1)
	`, id).Scan(&exists)
	return exists, err
}

func (r *PostgresProductRepository) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&n)
	return n, err
}

func (r *PostgresProductRepository) Delete(id domain.ProductID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tag, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanProduct(row scannable) (domain.Product, error) {
	var (
		id, merchantID, name, description, currency, imageKey string
		priceCents                                            int64
		stock                                                 int
		createdAt                                             time.Time
	)
	if err := row.Scan(&id, &merchantID, &name, &description, &priceCents, &currency, &stock, &imageKey, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}
		return domain.Product{}, err
	}
	price, err := domain.NewMoney(priceCents, currency)
	if err != nil {
		return domain.Product{}, fmt.Errorf("invalid stored price: %w", err)
	}
	return domain.Product{
		ID:          domain.ProductID(id),
		MerchantID:  merchantID,
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		ImageKey:    imageKey,
		CreatedAt:   createdAt.UTC(),
	}, nil
}
