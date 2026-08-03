-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
