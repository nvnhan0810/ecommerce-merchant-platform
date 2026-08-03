-- +goose Up
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS image_key TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE products
    DROP COLUMN IF EXISTS image_key;
