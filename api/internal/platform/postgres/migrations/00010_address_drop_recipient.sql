-- +goose Up
ALTER TABLE user_addresses DROP COLUMN IF EXISTS recipient_name;
ALTER TABLE user_addresses DROP COLUMN IF EXISTS phone_number;

-- +goose Down
ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS recipient_name TEXT NOT NULL DEFAULT '';
ALTER TABLE user_addresses ADD COLUMN IF NOT EXISTS phone_number TEXT NOT NULL DEFAULT '';
