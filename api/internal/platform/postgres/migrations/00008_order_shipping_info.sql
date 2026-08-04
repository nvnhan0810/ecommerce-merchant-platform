-- +goose Up
CREATE TABLE IF NOT EXISTS user_addresses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    recipient_name TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    address_line TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_addresses_user_id ON user_addresses (user_id);

ALTER TABLE orders ADD COLUMN shipping_name TEXT NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN shipping_phone TEXT NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN shipping_address TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE orders DROP COLUMN shipping_address;
ALTER TABLE orders DROP COLUMN shipping_phone;
ALTER TABLE orders DROP COLUMN shipping_name;

DROP INDEX IF EXISTS idx_user_addresses_user_id;
DROP TABLE IF EXISTS user_addresses;
