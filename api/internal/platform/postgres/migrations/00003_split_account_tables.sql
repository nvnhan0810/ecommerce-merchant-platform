-- +goose Up
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL DEFAULT '',
    password_hash TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Move legacy role-based rows out of users (no-op if already split).
INSERT INTO admins (id, email, display_name, password_hash, created_at)
SELECT id, email, display_name, COALESCE(password_hash, ''), created_at
FROM users
WHERE role = 'admin'
ON CONFLICT (id) DO NOTHING;

INSERT INTO merchants (id, email, display_name, password_hash, created_at)
SELECT id, email, display_name, COALESCE(password_hash, ''), created_at
FROM users
WHERE role = 'merchant'
ON CONFLICT (id) DO NOTHING;

DELETE FROM users WHERE role IN ('admin', 'merchant');

DROP INDEX IF EXISTS idx_users_role;

ALTER TABLE users DROP COLUMN IF EXISTS role;

-- +goose Down
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user';

UPDATE users SET role = 'user' WHERE role IS NULL OR role = '';

INSERT INTO users (id, email, display_name, role, password_hash, created_at)
SELECT id, email, display_name, 'admin', password_hash, created_at
FROM admins
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, email, display_name, role, password_hash, created_at)
SELECT id, email, display_name, 'merchant', password_hash, created_at
FROM merchants
ON CONFLICT (id) DO NOTHING;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('user', 'merchant', 'admin'));

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

DROP TABLE IF EXISTS merchants;
DROP TABLE IF EXISTS admins;
