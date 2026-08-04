-- +goose Up
CREATE TABLE IF NOT EXISTS payment_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    onepay_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    onepay_merchant_id TEXT NOT NULL DEFAULT '',
    onepay_access_code TEXT NOT NULL DEFAULT '',
    onepay_hash_secret TEXT NOT NULL DEFAULT '',
    onepay_payment_url TEXT NOT NULL DEFAULT 'https://mtf.onepay.vn/onecomm-pay/vpc.op',
    onepay_return_url TEXT NOT NULL DEFAULT '',
    onepay_ipn_url TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO payment_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    method TEXT NOT NULL CHECK (method IN ('cod', 'onepay')),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    currency TEXT NOT NULL DEFAULT 'VND',
    merch_txn_ref TEXT NOT NULL UNIQUE,
    gateway_txn_no TEXT NOT NULL DEFAULT '',
    response_code TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id);
CREATE INDEX IF NOT EXISTS idx_payments_merch_txn_ref ON payments (merch_txn_ref);

CREATE TABLE IF NOT EXISTS payment_orders (
    payment_id UUID NOT NULL REFERENCES payments (id) ON DELETE CASCADE,
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    PRIMARY KEY (payment_id, order_id)
);

CREATE INDEX IF NOT EXISTS idx_payment_orders_order_id ON payment_orders (order_id);

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS payment_method TEXT NOT NULL DEFAULT 'cod'
        CHECK (payment_method IN ('cod', 'onepay')),
    ADD COLUMN IF NOT EXISTS payment_status TEXT NOT NULL DEFAULT 'unpaid'
        CHECK (payment_status IN ('unpaid', 'pending', 'paid', 'failed', 'cancelled')),
    ADD COLUMN IF NOT EXISTS payment_id UUID,
    ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE orders
    DROP COLUMN IF EXISTS paid_at,
    DROP COLUMN IF EXISTS payment_id,
    DROP COLUMN IF EXISTS payment_status,
    DROP COLUMN IF EXISTS payment_method;

DROP TABLE IF EXISTS payment_orders;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS payment_settings;
