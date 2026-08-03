-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id),
    merchant_id UUID NOT NULL REFERENCES merchants (id),
    status TEXT NOT NULL CHECK (
        status IN (
            'new',
            'paid',
            'confirmed',
            'shipping',
            'succeeded',
            'failed',
            'cancelled'
        )
    ),
    currency TEXT NOT NULL DEFAULT 'VND',
    total_cents BIGINT NOT NULL CHECK (total_cents >= 0),
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT orders_code_format CHECK (code ~ '^[A-Z0-9]{10}$')
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_code ON orders (code);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
CREATE INDEX IF NOT EXISTS idx_orders_merchant_id ON orders (merchant_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders (created_at DESC);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products (id),
    product_name TEXT NOT NULL,
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents > 0),
    quantity INT NOT NULL CHECK (quantity > 0),
    line_total_cents BIGINT NOT NULL CHECK (line_total_cents > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items (product_id);

-- +goose Down
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP TABLE IF EXISTS order_items;

DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_merchant_id;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_code;
DROP TABLE IF EXISTS orders;
