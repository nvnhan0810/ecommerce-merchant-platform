-- +goose Up
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS delivery_tracking_code TEXT,
    ADD COLUMN IF NOT EXISTS delivery_carrier TEXT NOT NULL DEFAULT 'internal';

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders
    ADD CONSTRAINT orders_status_check CHECK (
        status IN (
            'new',
            'paid',
            'confirmed',
            'shipping',
            'succeeded',
            'returning',
            'returned',
            'failed',
            'cancelled'
        )
    );

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_delivery_tracking_code
    ON orders (delivery_tracking_code)
    WHERE delivery_tracking_code IS NOT NULL;

CREATE TABLE IF NOT EXISTS delivery_events (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    event_id TEXT NOT NULL DEFAULT '',
    delivery_tracking_code TEXT NOT NULL DEFAULT '',
    status_code TEXT NOT NULL,
    status_label TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_delivery_events_order_id_occurred_at
    ON delivery_events (order_id, occurred_at ASC, created_at ASC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_delivery_events_event_id
    ON delivery_events (event_id)
    WHERE event_id <> '';

-- +goose Down
DROP INDEX IF EXISTS idx_delivery_events_event_id;
DROP INDEX IF EXISTS idx_delivery_events_order_id_occurred_at;
DROP TABLE IF EXISTS delivery_events;

DROP INDEX IF EXISTS idx_orders_delivery_tracking_code;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders
    ADD CONSTRAINT orders_status_check CHECK (
        status IN (
            'new',
            'paid',
            'confirmed',
            'shipping',
            'succeeded',
            'failed',
            'cancelled'
        )
    );

ALTER TABLE orders
    DROP COLUMN IF EXISTS delivery_carrier,
    DROP COLUMN IF EXISTS delivery_tracking_code;
