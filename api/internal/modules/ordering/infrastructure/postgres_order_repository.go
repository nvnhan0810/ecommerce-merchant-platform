package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

	var tracking any
	if strings.TrimSpace(order.DeliveryTrackingCode) != "" {
		tracking = order.DeliveryTrackingCode
	}
	carrier := order.DeliveryCarrier
	if strings.TrimSpace(carrier) == "" {
		carrier = domain.DefaultDeliveryCarrier
	}

		_, err = tx.Exec(ctx, `
			INSERT INTO orders (
				id, code, user_id, merchant_id, status, currency, total_cents, note,
				delivery_tracking_code, delivery_carrier, shipping_name, shipping_phone, shipping_address, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT (id) DO UPDATE SET
				code = EXCLUDED.code,
				user_id = EXCLUDED.user_id,
				merchant_id = EXCLUDED.merchant_id,
				status = EXCLUDED.status,
				currency = EXCLUDED.currency,
				total_cents = EXCLUDED.total_cents,
				note = EXCLUDED.note,
				delivery_tracking_code = EXCLUDED.delivery_tracking_code,
				delivery_carrier = EXCLUDED.delivery_carrier,
				shipping_name = EXCLUDED.shipping_name,
				shipping_phone = EXCLUDED.shipping_phone,
				shipping_address = EXCLUDED.shipping_address,
				updated_at = EXCLUDED.updated_at
		`, order.ID, order.Code, order.UserID, order.MerchantID, order.Status, order.Currency,
			order.TotalCents, order.Note, tracking, carrier, order.ShippingName, order.ShippingPhone, order.ShippingAddress, order.CreatedAt, order.UpdatedAt)
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

	for _, ev := range order.PendingDeliveryEvents() {
		raw := ev.RawPayload
		if len(raw) == 0 {
			raw = json.RawMessage(`{}`)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO delivery_events (
				id, order_id, event_id, delivery_tracking_code, status_code, status_label,
				message, reason, occurred_at, source, raw_payload, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, ev.ID, order.ID, ev.EventID, ev.DeliveryTrackingCode, string(ev.StatusCode), ev.StatusLabel,
			ev.Message, ev.Reason, ev.OccurredAt, ev.Source, raw, ev.CreatedAt)
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
		SELECT id, code, user_id, merchant_id, status, currency, total_cents, note,
		       delivery_tracking_code, delivery_carrier, shipping_name, shipping_phone, shipping_address, created_at, updated_at
		FROM orders WHERE id = $1
	`, id)
	return r.hydrate(ctx, row)
}

func (r *PostgresOrderRepository) FindByCode(code string) (domain.Order, error) {
	parsed, err := domain.ParseOrderCode(code)
	if err != nil {
		return domain.Order{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, code, user_id, merchant_id, status, currency, total_cents, note,
		       delivery_tracking_code, delivery_carrier, shipping_name, shipping_phone, shipping_address, created_at, updated_at
		FROM orders WHERE code = $1
	`, parsed)
	return r.hydrate(ctx, row)
}

func (r *PostgresOrderRepository) FindByDeliveryTrackingCode(code string) (domain.Order, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := r.pool.QueryRow(ctx, `
		SELECT id, code, user_id, merchant_id, status, currency, total_cents, note,
		       delivery_tracking_code, delivery_carrier, shipping_name, shipping_phone, shipping_address, created_at, updated_at
		FROM orders WHERE delivery_tracking_code = $1
	`, code)
	return r.hydrate(ctx, row)
}

func (r *PostgresOrderRepository) HasDeliveryEventID(eventID string) (bool, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM delivery_events WHERE event_id = $1`, eventID).Scan(&n)
	return n > 0, err
}

func (r *PostgresOrderRepository) hydrate(ctx context.Context, row scannable) (domain.Order, error) {
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
	deliveryEvents, err := r.loadDeliveryEvents(ctx, order.ID)
	if err != nil {
		return domain.Order{}, err
	}
	order.DeliveryEvents = deliveryEvents
	return order, nil
}

func (r *PostgresOrderRepository) List(limit, offset int) ([]domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
			SELECT id, code, user_id, merchant_id, status, currency, total_cents, note,
			       delivery_tracking_code, delivery_carrier, shipping_name, shipping_phone, shipping_address, created_at, updated_at
			FROM orders
			ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanList(ctx, rows)
}

func (r *PostgresOrderRepository) ListByMerchant(merchantID string, limit, offset int) ([]domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
			SELECT id, code, user_id, merchant_id, status, currency, total_cents, note,
			       delivery_tracking_code, delivery_carrier, shipping_name, shipping_phone, shipping_address, created_at, updated_at
			FROM orders
			WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, merchantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanList(ctx, rows)
}

func (r *PostgresOrderRepository) ListByUser(userID string, limit, offset int) ([]domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
			SELECT id, code, user_id, merchant_id, status, currency, total_cents, note,
			       delivery_tracking_code, delivery_carrier, shipping_name, shipping_phone, shipping_address, created_at, updated_at
			FROM orders
			WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanList(ctx, rows)
}

func (r *PostgresOrderRepository) scanList(ctx context.Context, rows pgx.Rows) ([]domain.Order, error) {
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

func (r *PostgresOrderRepository) loadDeliveryEvents(ctx context.Context, orderID domain.OrderID) ([]domain.DeliveryEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, event_id, delivery_tracking_code, status_code, status_label,
		       message, reason, occurred_at, source, raw_payload, created_at
		FROM delivery_events
		WHERE order_id = $1
		ORDER BY occurred_at ASC, created_at ASC, id ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.DeliveryEvent, 0)
	for rows.Next() {
		var (
			id, oid, eventID, tracking, statusCode, statusLabel string
			message, reason, source                             string
			raw                                                 []byte
			occurredAt, createdAt                               time.Time
		)
		if err := rows.Scan(
			&id, &oid, &eventID, &tracking, &statusCode, &statusLabel,
			&message, &reason, &occurredAt, &source, &raw, &createdAt,
		); err != nil {
			return nil, err
		}
		if len(raw) == 0 {
			raw = []byte(`{}`)
		}
		events = append(events, domain.DeliveryEvent{
			ID:                   domain.DeliveryEventID(id),
			OrderID:              domain.OrderID(oid),
			EventID:              eventID,
			DeliveryTrackingCode: tracking,
			StatusCode:           domain.DeliveryStatusCode(statusCode),
			StatusLabel:          statusLabel,
			Message:              message,
			Reason:               reason,
			OccurredAt:           occurredAt.UTC(),
			Source:               source,
			RawPayload:           json.RawMessage(raw),
			CreatedAt:            createdAt.UTC(),
		})
	}
	return events, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanOrder(row scannable) (domain.Order, error) {
	var (
		id, code, userID, merchantID, status, currency, note, carrier, shipName, shipPhone, shipAddr string
		tracking                                                                                     *string
		total                                                                                        int64
		createdAt, updatedAt                                                                         time.Time
	)
	if err := row.Scan(
		&id, &code, &userID, &merchantID, &status, &currency, &total, &note,
		&tracking, &carrier, &shipName, &shipPhone, &shipAddr, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, err
	}
	st, err := domain.ParseOrderStatus(status)
	if err != nil {
		return domain.Order{}, fmt.Errorf("invalid stored status: %w", err)
	}
	if carrier == "" {
		carrier = domain.DefaultDeliveryCarrier
	}
	trackingCode := ""
	if tracking != nil {
		trackingCode = *tracking
	}
	return domain.Order{
		ID:                   domain.OrderID(id),
		Code:                 code,
		UserID:               userID,
		MerchantID:           merchantID,
		Status:               st,
		Currency:             currency,
		TotalCents:           total,
		Note:                 note,
		DeliveryTrackingCode: trackingCode,
		DeliveryCarrier:      carrier,
		ShippingName:         shipName,
		ShippingPhone:        shipPhone,
		ShippingAddress:      shipAddr,
		CreatedAt:            createdAt.UTC(),
		UpdatedAt:            updatedAt.UTC(),
	}, nil
}
