-- +goose Up
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected')),
    created_by_merchant_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_name_lower ON categories (LOWER(name));
CREATE INDEX IF NOT EXISTS idx_categories_status ON categories (status);
CREATE INDEX IF NOT EXISTS idx_categories_created_by ON categories (created_by_merchant_id);

CREATE TABLE IF NOT EXISTS product_categories (
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories (id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_product_categories_category_id ON product_categories (category_id);

-- +goose Down
DROP TABLE IF EXISTS product_categories;
DROP TABLE IF EXISTS categories;
