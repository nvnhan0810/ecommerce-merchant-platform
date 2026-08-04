package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nvnhan0810/ecomerce-api/internal/modules/ordering/domain"
)

type PostgresPaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPaymentRepository(pool *pgxpool.Pool) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{pool: pool}
}

func (r *PostgresPaymentRepository) Save(payment domain.Payment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (
			id, user_id, method, status, amount_cents, currency, merch_txn_ref,
			gateway_txn_no, response_code, message, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			method = EXCLUDED.method,
			status = EXCLUDED.status,
			amount_cents = EXCLUDED.amount_cents,
			currency = EXCLUDED.currency,
			merch_txn_ref = EXCLUDED.merch_txn_ref,
			gateway_txn_no = EXCLUDED.gateway_txn_no,
			response_code = EXCLUDED.response_code,
			message = EXCLUDED.message,
			updated_at = EXCLUDED.updated_at
	`, payment.ID, payment.UserID, payment.Method, payment.Status, payment.AmountCents, payment.Currency,
		payment.MerchTxnRef, payment.GatewayTxnNo, payment.ResponseCode, payment.Message,
		payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM payment_orders WHERE payment_id = $1`, payment.ID); err != nil {
		return err
	}
	for _, oid := range payment.OrderIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO payment_orders (payment_id, order_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, payment.ID, oid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresPaymentRepository) FindByID(id domain.PaymentID) (domain.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.scanPayment(ctx, `
		SELECT id, user_id, method, status, amount_cents, currency, merch_txn_ref,
			gateway_txn_no, response_code, message, created_at, updated_at
		FROM payments WHERE id = $1
	`, id)
}

func (r *PostgresPaymentRepository) FindByMerchTxnRef(ref string) (domain.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.scanPayment(ctx, `
		SELECT id, user_id, method, status, amount_cents, currency, merch_txn_ref,
			gateway_txn_no, response_code, message, created_at, updated_at
		FROM payments WHERE merch_txn_ref = $1
	`, strings.TrimSpace(ref))
}

func (r *PostgresPaymentRepository) scanPayment(ctx context.Context, query string, arg any) (domain.Payment, error) {
	var (
		id, userID, method, status, currency, merchRef, gatewayTxn, respCode, message string
		amount                                                                        int64
		createdAt, updatedAt                                                          time.Time
	)
	err := r.pool.QueryRow(ctx, query, arg).Scan(
		&id, &userID, &method, &status, &amount, &currency, &merchRef,
		&gatewayTxn, &respCode, &message, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Payment{}, domain.ErrPaymentNotFound
		}
		return domain.Payment{}, err
	}
	m, err := domain.ParsePaymentMethod(method)
	if err != nil {
		return domain.Payment{}, err
	}
	st, err := domain.ParsePaymentStatus(status)
	if err != nil {
		return domain.Payment{}, err
	}
	orderIDs, err := r.loadOrderIDs(ctx, domain.PaymentID(id))
	if err != nil {
		return domain.Payment{}, err
	}
	return domain.Payment{
		ID:           domain.PaymentID(id),
		UserID:       userID,
		Method:       m,
		Status:       st,
		AmountCents:  amount,
		Currency:     currency,
		MerchTxnRef:  merchRef,
		GatewayTxnNo: gatewayTxn,
		ResponseCode: respCode,
		Message:      message,
		OrderIDs:     orderIDs,
		CreatedAt:    createdAt.UTC(),
		UpdatedAt:    updatedAt.UTC(),
	}, nil
}

func (r *PostgresPaymentRepository) loadOrderIDs(ctx context.Context, id domain.PaymentID) ([]domain.OrderID, error) {
	rows, err := r.pool.Query(ctx, `SELECT order_id FROM payment_orders WHERE payment_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.OrderID, 0)
	for rows.Next() {
		var oid string
		if err := rows.Scan(&oid); err != nil {
			return nil, err
		}
		out = append(out, domain.OrderID(oid))
	}
	return out, rows.Err()
}

func (r *PostgresPaymentRepository) GetSettings() (domain.PaymentSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var s domain.PaymentSettings
	err := r.pool.QueryRow(ctx, `
		SELECT onepay_return_url, onepay_ipn_url,
			onepay_domestic_enabled, onepay_domestic_merchant_id, onepay_domestic_access_code,
			onepay_domestic_hash_secret, onepay_domestic_payment_url,
			onepay_international_enabled, onepay_international_merchant_id, onepay_international_access_code,
			onepay_international_hash_secret, onepay_international_payment_url,
			updated_at
		FROM payment_settings WHERE id = 1
	`).Scan(
		&s.OnePayReturnURL, &s.OnePayIPNURL,
		&s.OnePayDomestic.Enabled, &s.OnePayDomestic.MerchantID, &s.OnePayDomestic.AccessCode,
		&s.OnePayDomestic.HashSecret, &s.OnePayDomestic.PaymentURL,
		&s.OnePayInternational.Enabled, &s.OnePayInternational.MerchantID, &s.OnePayInternational.AccessCode,
		&s.OnePayInternational.HashSecret, &s.OnePayInternational.PaymentURL,
		&s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PaymentSettings{
				OnePayDomestic: domain.OnePayGatewaySettings{
					PaymentURL: DefaultOnePayDomesticPaymentURL,
				},
				OnePayInternational: domain.OnePayGatewaySettings{
					PaymentURL: DefaultOnePayInternationalPaymentURL,
				},
				UpdatedAt: time.Now().UTC(),
			}, nil
		}
		return domain.PaymentSettings{}, err
	}
	s.UpdatedAt = s.UpdatedAt.UTC()
	if strings.TrimSpace(s.OnePayDomestic.PaymentURL) == "" {
		s.OnePayDomestic.PaymentURL = DefaultOnePayDomesticPaymentURL
	}
	if strings.TrimSpace(s.OnePayInternational.PaymentURL) == "" {
		s.OnePayInternational.PaymentURL = DefaultOnePayInternationalPaymentURL
	}
	return s, nil
}

func (r *PostgresPaymentRepository) SaveSettings(settings domain.PaymentSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now().UTC()
	domURL := strings.TrimSpace(settings.OnePayDomestic.PaymentURL)
	if domURL == "" {
		domURL = DefaultOnePayDomesticPaymentURL
	}
	intlURL := strings.TrimSpace(settings.OnePayInternational.PaymentURL)
	if intlURL == "" {
		intlURL = DefaultOnePayInternationalPaymentURL
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO payment_settings (
			id, onepay_return_url, onepay_ipn_url,
			onepay_domestic_enabled, onepay_domestic_merchant_id, onepay_domestic_access_code,
			onepay_domestic_hash_secret, onepay_domestic_payment_url,
			onepay_international_enabled, onepay_international_merchant_id, onepay_international_access_code,
			onepay_international_hash_secret, onepay_international_payment_url,
			updated_at
		) VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			onepay_return_url = EXCLUDED.onepay_return_url,
			onepay_ipn_url = EXCLUDED.onepay_ipn_url,
			onepay_domestic_enabled = EXCLUDED.onepay_domestic_enabled,
			onepay_domestic_merchant_id = EXCLUDED.onepay_domestic_merchant_id,
			onepay_domestic_access_code = EXCLUDED.onepay_domestic_access_code,
			onepay_domestic_hash_secret = EXCLUDED.onepay_domestic_hash_secret,
			onepay_domestic_payment_url = EXCLUDED.onepay_domestic_payment_url,
			onepay_international_enabled = EXCLUDED.onepay_international_enabled,
			onepay_international_merchant_id = EXCLUDED.onepay_international_merchant_id,
			onepay_international_access_code = EXCLUDED.onepay_international_access_code,
			onepay_international_hash_secret = EXCLUDED.onepay_international_hash_secret,
			onepay_international_payment_url = EXCLUDED.onepay_international_payment_url,
			updated_at = EXCLUDED.updated_at
	`, strings.TrimSpace(settings.OnePayReturnURL),
		strings.TrimSpace(settings.OnePayIPNURL),
		settings.OnePayDomestic.Enabled,
		strings.TrimSpace(settings.OnePayDomestic.MerchantID),
		strings.TrimSpace(settings.OnePayDomestic.AccessCode),
		strings.TrimSpace(settings.OnePayDomestic.HashSecret),
		domURL,
		settings.OnePayInternational.Enabled,
		strings.TrimSpace(settings.OnePayInternational.MerchantID),
		strings.TrimSpace(settings.OnePayInternational.AccessCode),
		strings.TrimSpace(settings.OnePayInternational.HashSecret),
		intlURL,
		now)
	return err
}

func (r *PostgresPaymentRepository) SaveCallbackEvent(event domain.PaymentCallbackEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw := event.RawPayload
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	var paymentID *string
	if strings.TrimSpace(string(event.PaymentID)) != "" {
		id := string(event.PaymentID)
		paymentID = &id
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO payment_callback_events (
			id, provider, channel, http_method, payment_id, payment_method,
			merch_txn_ref, response_code, message, paid, success, error_message,
			raw_payload, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, event.ID, event.Provider, string(event.Channel), event.HTTPMethod, paymentID, string(event.PaymentMethod),
		event.MerchTxnRef, event.ResponseCode, event.Message, event.Paid, event.Success, event.ErrorMessage,
		raw, event.CreatedAt.UTC())
	return err
}

func (r *PostgresPaymentRepository) FindCallbackEventByID(id domain.PaymentCallbackEventID) (domain.PaymentCallbackEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.scanCallbackEvent(ctx, `
		SELECT id, provider, channel, http_method, payment_id, payment_method,
			merch_txn_ref, response_code, message, paid, success, error_message,
			raw_payload, created_at
		FROM payment_callback_events WHERE id = $1
	`, string(id))
}

func (r *PostgresPaymentRepository) ListCallbackEvents(filter domain.PaymentCallbackListFilter) ([]domain.PaymentCallbackEvent, error) {
	filter.Normalize()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	query := `
		SELECT id, provider, channel, http_method, payment_id, payment_method,
			merch_txn_ref, response_code, message, paid, success, error_message,
			raw_payload, created_at
		FROM payment_callback_events
		WHERE ($1 = '' OR provider = $1)
		  AND ($2 = '' OR channel = $2)
		  AND ($3 = '' OR merch_txn_ref ILIKE '%' || $3 || '%')
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.pool.Query(ctx, query, filter.Provider, filter.Channel, filter.MerchTxnRef, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.PaymentCallbackEvent, 0)
	for rows.Next() {
		ev, err := scanCallbackEventRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

type callbackScanner interface {
	Scan(dest ...any) error
}

func (r *PostgresPaymentRepository) scanCallbackEvent(ctx context.Context, query string, arg any) (domain.PaymentCallbackEvent, error) {
	row := r.pool.QueryRow(ctx, query, arg)
	ev, err := scanCallbackEventRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PaymentCallbackEvent{}, domain.ErrPaymentCallbackNotFound
		}
		return domain.PaymentCallbackEvent{}, err
	}
	return ev, nil
}

func scanCallbackEventRow(row callbackScanner) (domain.PaymentCallbackEvent, error) {
	var (
		id, provider, channel, httpMethod, paymentMethod, merchRef, respCode, message, errMsg string
		paymentID                                                                             *string
		paid, success                                                                         bool
		raw                                                                                   []byte
		createdAt                                                                             time.Time
	)
	if err := row.Scan(
		&id, &provider, &channel, &httpMethod, &paymentID, &paymentMethod,
		&merchRef, &respCode, &message, &paid, &success, &errMsg,
		&raw, &createdAt,
	); err != nil {
		return domain.PaymentCallbackEvent{}, err
	}
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	var pid domain.PaymentID
	if paymentID != nil {
		pid = domain.PaymentID(*paymentID)
	}
	var method domain.PaymentMethod
	if strings.TrimSpace(paymentMethod) != "" {
		if m, err := domain.ParsePaymentMethod(paymentMethod); err == nil {
			method = m
		}
	}
	return domain.PaymentCallbackEvent{
		ID:            domain.PaymentCallbackEventID(id),
		Provider:      provider,
		Channel:       domain.PaymentCallbackChannel(channel),
		HTTPMethod:    httpMethod,
		PaymentID:     pid,
		PaymentMethod: method,
		MerchTxnRef:   merchRef,
		ResponseCode:  respCode,
		Message:       message,
		Paid:          paid,
		Success:       success,
		ErrorMessage:  errMsg,
		RawPayload:    raw,
		CreatedAt:     createdAt.UTC(),
	}, nil
}
