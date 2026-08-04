package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/catalog/domain"
)

type PostgresCategoryRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresCategoryRepository(pool *pgxpool.Pool) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{pool: pool}
}

func (r *PostgresCategoryRepository) Save(category domain.Category) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO categories (id, name, status, created_by_merchant_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			created_by_merchant_id = EXCLUDED.created_by_merchant_id
	`, category.ID, category.Name, category.Status, category.CreatedByMerchantID, category.CreatedAt)
	return err
}

func (r *PostgresCategoryRepository) FindByID(id domain.CategoryID) (domain.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, status, created_by_merchant_id, created_at
		FROM categories WHERE id = $1
	`, id)
	return scanCategory(row)
}

func (r *PostgresCategoryRepository) FindByName(name string) (domain.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, status, created_by_merchant_id, created_at
		FROM categories WHERE LOWER(name) = LOWER($1)
	`, strings.TrimSpace(name))
	return scanCategory(row)
}

func (r *PostgresCategoryRepository) List(filter domain.CategoryListFilter) ([]domain.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, status, created_by_merchant_id, created_at
		FROM categories
		WHERE 1=1
	`
	args := make([]any, 0, 4)
	argN := 1

	if filter.IncludeApproved && strings.TrimSpace(filter.MerchantViewerID) != "" {
		query += ` AND (status = 'approved' OR (created_by_merchant_id = $` + itoa(argN) + ` AND status = 'pending'))`
		args = append(args, strings.TrimSpace(filter.MerchantViewerID))
		argN++
	} else {
		if filter.Status != "" {
			query += ` AND status = $` + itoa(argN)
			args = append(args, filter.Status)
			argN++
		}
		if strings.TrimSpace(filter.CreatedByMerchantID) != "" {
			query += ` AND created_by_merchant_id = $` + itoa(argN)
			args = append(args, strings.TrimSpace(filter.CreatedByMerchantID))
			argN++
		}
	}

	query += ` ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Category, 0)
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PostgresCategoryRepository) Delete(id domain.CategoryID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}

func (r *PostgresCategoryRepository) ListByProductIDs(productIDs []domain.ProductID) (map[domain.ProductID][]domain.Category, error) {
	out := make(map[domain.ProductID][]domain.Category, len(productIDs))
	if len(productIDs) == 0 {
		return out, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ids := make([]string, len(productIDs))
	for i, id := range productIDs {
		ids[i] = string(id)
		out[id] = []domain.Category{}
	}

	rows, err := r.pool.Query(ctx, `
		SELECT pc.product_id, c.id, c.name, c.status, c.created_by_merchant_id, c.created_at
		FROM product_categories pc
		JOIN categories c ON c.id = pc.category_id
		WHERE pc.product_id = ANY($1::uuid[])
		ORDER BY c.name ASC
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productID string
		var id, name, status, createdBy string
		var createdAt time.Time
		if err := rows.Scan(&productID, &id, &name, &status, &createdBy, &createdAt); err != nil {
			return nil, err
		}
		pid := domain.ProductID(productID)
		out[pid] = append(out[pid], domain.Category{
			ID:                  domain.CategoryID(id),
			Name:                name,
			Status:              domain.CategoryStatus(status),
			CreatedByMerchantID: createdBy,
			CreatedAt:           createdAt.UTC(),
		})
	}
	return out, rows.Err()
}

func (r *PostgresCategoryRepository) SetProductCategories(productID domain.ProductID, categoryIDs []domain.CategoryID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM product_categories WHERE product_id = $1`, productID); err != nil {
		return err
	}
	for _, cid := range categoryIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO product_categories (product_id, category_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, productID, cid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresCategoryRepository) ListProductIDsByCategory(categoryID domain.CategoryID, approvedOnly bool, limit, offset int) ([]domain.ProductID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT pc.product_id
		FROM product_categories pc
		JOIN categories c ON c.id = pc.category_id
		WHERE pc.category_id = $1
	`
	args := []any{categoryID}
	if approvedOnly {
		query += ` AND c.status = 'approved'`
	}
	query += ` ORDER BY pc.product_id LIMIT $2 OFFSET $3`
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ProductID, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, domain.ProductID(id))
	}
	return out, rows.Err()
}

func scanCategory(row scannable) (domain.Category, error) {
	var id, name, status, createdBy string
	var createdAt time.Time
	if err := row.Scan(&id, &name, &status, &createdBy, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Category{}, domain.ErrCategoryNotFound
		}
		return domain.Category{}, err
	}
	return domain.Category{
		ID:                  domain.CategoryID(id),
		Name:                name,
		Status:              domain.CategoryStatus(status),
		CreatedByMerchantID: createdBy,
		CreatedAt:           createdAt.UTC(),
	}, nil
}

func itoa(n int) string {
	const digits = "0123456789"
	if n < 10 {
		return string(digits[n])
	}
	return itoa(n/10) + string(digits[n%10])
}
