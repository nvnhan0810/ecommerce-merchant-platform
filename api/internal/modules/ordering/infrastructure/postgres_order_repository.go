package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) Save(order domain.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, code, user_id, merchant_id, status, currency, total_cents, note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			code = EXCLUDED.code,
			user_id = EXCLUDED.user_id,
			merchant_id = EXCLUDED.merchant_id,
			status = EXCLUDED.status,
			currency = EXCLUDED.currency,
			total_cents = EXCLUDED.total_cents,
			note = EXCLUDED.note,
			updated_at = EXCLUDED.updated_at
	`, order.ID, order.Code, order.UserID, order.MerchantID, order.Status, order.Currency,
		order.TotalCents, order.Note, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID); err != nil {
		return err
	}
	for _, item := range order.Items {
		_, err := tx.Exec(ctx, `
			INSERT INTO order_items (
				id, order_id, product_id, product_name, unit_price_cents, quantity, line_total_cents, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, item.ID, order.ID, item.ProductID, item.ProductName, item.UnitPriceCents,
			item.Quantity, item.LineTotalCents, item.CreatedAt)
		if err != nil {
			return err
		}
	}

	for _, ev := range order.PendingEvents() {
		var from any
		if ev.FromStatus != "" {
			from = string(ev.FromStatus)
		}
		var to any
		if ev.ToStatus != "" {
			to = string(ev.ToStatus)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO order_events (
				id, order_id, event_type, from_status, to_status, message,
				actor_id, actor_email, actor_role, actor_name, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, ev.ID, order.ID, ev.Type, from, to, ev.Message,
			ev.ActorID, ev.ActorEmail, ev.ActorRole, ev.ActorName, ev.CreatedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) FindByID(id domain.OrderID) (domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, code, user_id, merchant_id, status, currency, total_cents, note, created_at, updated_at
		FROM orders WHERE id = $1
	`, id)
	order, err := scanOrder(row)
	if err != nil {
		return domain.Order{}, err
	}
	items, err := r.loadItems(ctx, order.ID)
	if err != nil {
		return domain.Order{}, err
	}
	order.Items = items
	history, err := r.loadEvents(ctx, order.ID)
	if err != nil {
		return domain.Order{}, err
	}
	order.History = history
	return order, nil
}

func (r *PostgresOrderRepository) FindByCode(code string) (domain.Order, error) {
	parsed, err := domain.ParseOrderCode(code)
	if err != nil {
		return domain.Order{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, code, user_id, merchant_id, status, currency, total_cents, note, created_at, updated_at
		FROM orders WHERE code = $1
	`, parsed)
	order, err := scanOrder(row)
	if err != nil {
		return domain.Order{}, err
	}
	items, err := r.loadItems(ctx, order.ID)
	if err != nil {
		return domain.Order{}, err
	}
	order.Items = items
	history, err := r.loadEvents(ctx, order.ID)
	if err != nil {
		return domain.Order{}, err
	}
	order.History = history
	return order, nil
}

func (r *PostgresOrderRepository) List(limit, offset int) ([]domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, user_id, merchant_id, status, currency, total_cents, note, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items, err := r.loadItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *PostgresOrderRepository) ListByMerchant(merchantID string, limit, offset int) ([]domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, user_id, merchant_id, status, currency, total_cents, note, created_at, updated_at
		FROM orders
		WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, merchantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items, err := r.loadItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *PostgresOrderRepository) ListByUser(userID string, limit, offset int) ([]domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, user_id, merchant_id, status, currency, total_cents, note, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Order, 0)
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items, err := r.loadItems(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		o.Items = items
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *PostgresOrderRepository) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&n)
	return n, err
}

func (r *PostgresOrderRepository) loadItems(ctx context.Context, orderID domain.OrderID) ([]domain.OrderItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, product_id, product_name, unit_price_cents, quantity, line_total_cents, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.OrderItem, 0)
	for rows.Next() {
		var (
			id, productID, name string
			unit, line          int64
			qty                 int
			createdAt           time.Time
		)
		if err := rows.Scan(&id, &productID, &name, &unit, &qty, &line, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, domain.OrderItem{
			ID:             domain.OrderItemID(id),
			ProductID:      productID,
			ProductName:    name,
			UnitPriceCents: unit,
			Quantity:       qty,
			LineTotalCents: line,
			CreatedAt:      createdAt.UTC(),
		})
	}
	return items, rows.Err()
}

func (r *PostgresOrderRepository) loadEvents(ctx context.Context, orderID domain.OrderID) ([]domain.OrderEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, event_type, from_status, to_status, message,
		       actor_id, actor_email, actor_role, actor_name, created_at
		FROM order_events
		WHERE order_id = $1
		ORDER BY created_at ASC, id ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.OrderEvent, 0)
	for rows.Next() {
		var (
			id, oid, eventType, message    string
			actorID, actorEmail, actorRole string
			actorName                      string
			fromStatus, toStatus           *string
			createdAt                      time.Time
		)
		if err := rows.Scan(
			&id, &oid, &eventType, &fromStatus, &toStatus, &message,
			&actorID, &actorEmail, &actorRole, &actorName, &createdAt,
		); err != nil {
			return nil, err
		}
		ev := domain.OrderEvent{
			ID:         domain.OrderEventID(id),
			OrderID:    domain.OrderID(oid),
			Type:       domain.OrderEventType(eventType),
			Message:    message,
			ActorID:    actorID,
			ActorEmail: actorEmail,
			ActorRole:  actorRole,
			ActorName:  actorName,
			CreatedAt:  createdAt.UTC(),
		}
		if fromStatus != nil && *fromStatus != "" {
			if st, err := domain.ParseOrderStatus(*fromStatus); err == nil {
				ev.FromStatus = st
			}
		}
		if toStatus != nil && *toStatus != "" {
			if st, err := domain.ParseOrderStatus(*toStatus); err == nil {
				ev.ToStatus = st
			}
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanOrder(row scannable) (domain.Order, error) {
	var (
		id, code, userID, merchantID, status, currency, note string
		total                                                int64
		createdAt, updatedAt                                 time.Time
	)
	if err := row.Scan(&id, &code, &userID, &merchantID, &status, &currency, &total, &note, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, err
	}
	st, err := domain.ParseOrderStatus(status)
	if err != nil {
		return domain.Order{}, fmt.Errorf("invalid stored status: %w", err)
	}
	return domain.Order{
		ID:         domain.OrderID(id),
		Code:       code,
		UserID:     userID,
		MerchantID: merchantID,
		Status:     st,
		Currency:   currency,
		TotalCents: total,
		Note:       note,
		CreatedAt:  createdAt.UTC(),
		UpdatedAt:  updatedAt.UTC(),
	}, nil
}
