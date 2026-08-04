-- +goose Up
ALTER TABLE merchants
    ADD COLUMN IF NOT EXISTS avatar_key TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS address_line TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS country_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS province_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ward_code TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE merchants
    DROP COLUMN IF EXISTS avatar_key,
    DROP COLUMN IF EXISTS address_line,
    DROP COLUMN IF EXISTS country_code,
    DROP COLUMN IF EXISTS province_code,
    DROP COLUMN IF EXISTS ward_code;
