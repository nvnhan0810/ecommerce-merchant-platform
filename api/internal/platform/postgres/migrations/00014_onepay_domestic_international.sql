-- +goose Up
ALTER TABLE payment_settings
    ADD COLUMN IF NOT EXISTS onepay_domestic_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS onepay_domestic_merchant_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_domestic_access_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_domestic_hash_secret TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_domestic_payment_url TEXT NOT NULL DEFAULT 'https://mtf.onepay.vn/onecomm-pay/vpc.op',
    ADD COLUMN IF NOT EXISTS onepay_international_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS onepay_international_merchant_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_international_access_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_international_hash_secret TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_international_payment_url TEXT NOT NULL DEFAULT 'https://mtf.onepay.vn/vpcpay/vpcpay.op';

-- Migrate legacy single OnePay config into domestic gateway.
UPDATE payment_settings SET
    onepay_domestic_enabled = COALESCE(onepay_enabled, FALSE),
    onepay_domestic_merchant_id = COALESCE(onepay_merchant_id, ''),
    onepay_domestic_access_code = COALESCE(onepay_access_code, ''),
    onepay_domestic_hash_secret = COALESCE(onepay_hash_secret, ''),
    onepay_domestic_payment_url = CASE
        WHEN COALESCE(onepay_payment_url, '') = '' THEN 'https://mtf.onepay.vn/onecomm-pay/vpc.op'
        ELSE onepay_payment_url
    END
WHERE id = 1;

ALTER TABLE payment_settings
    DROP COLUMN IF EXISTS onepay_enabled,
    DROP COLUMN IF EXISTS onepay_merchant_id,
    DROP COLUMN IF EXISTS onepay_access_code,
    DROP COLUMN IF EXISTS onepay_hash_secret,
    DROP COLUMN IF EXISTS onepay_payment_url;

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_method_check;
ALTER TABLE payments
    ADD CONSTRAINT payments_method_check
    CHECK (method IN ('cod', 'onepay', 'onepay_domestic', 'onepay_international'));

UPDATE payments SET method = 'onepay_domestic' WHERE method = 'onepay';

ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_method_check;
ALTER TABLE payments
    ADD CONSTRAINT payments_method_check
    CHECK (method IN ('cod', 'onepay_domestic', 'onepay_international'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_method_check;
ALTER TABLE orders
    ADD CONSTRAINT orders_payment_method_check
    CHECK (payment_method IN ('cod', 'onepay', 'onepay_domestic', 'onepay_international'));

UPDATE orders SET payment_method = 'onepay_domestic' WHERE payment_method = 'onepay';

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_method_check;
ALTER TABLE orders
    ADD CONSTRAINT orders_payment_method_check
    CHECK (payment_method IN ('cod', 'onepay_domestic', 'onepay_international'));

-- +goose Down
ALTER TABLE payment_settings
    ADD COLUMN IF NOT EXISTS onepay_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS onepay_merchant_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_access_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_hash_secret TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS onepay_payment_url TEXT NOT NULL DEFAULT 'https://mtf.onepay.vn/onecomm-pay/vpc.op';

UPDATE payment_settings SET
    onepay_enabled = onepay_domestic_enabled,
    onepay_merchant_id = onepay_domestic_merchant_id,
    onepay_access_code = onepay_domestic_access_code,
    onepay_hash_secret = onepay_domestic_hash_secret,
    onepay_payment_url = onepay_domestic_payment_url
WHERE id = 1;

ALTER TABLE payment_settings
    DROP COLUMN IF EXISTS onepay_domestic_enabled,
    DROP COLUMN IF EXISTS onepay_domestic_merchant_id,
    DROP COLUMN IF EXISTS onepay_domestic_access_code,
    DROP COLUMN IF EXISTS onepay_domestic_hash_secret,
    DROP COLUMN IF EXISTS onepay_domestic_payment_url,
    DROP COLUMN IF EXISTS onepay_international_enabled,
    DROP COLUMN IF EXISTS onepay_international_merchant_id,
    DROP COLUMN IF EXISTS onepay_international_access_code,
    DROP COLUMN IF EXISTS onepay_international_hash_secret,
    DROP COLUMN IF EXISTS onepay_international_payment_url;

UPDATE payments SET method = 'onepay' WHERE method IN ('onepay_domestic', 'onepay_international');
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_method_check;
ALTER TABLE payments
    ADD CONSTRAINT payments_method_check
    CHECK (method IN ('cod', 'onepay'));

UPDATE orders SET payment_method = 'onepay' WHERE payment_method IN ('onepay_domestic', 'onepay_international');
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_method_check;
ALTER TABLE orders
    ADD CONSTRAINT orders_payment_method_check
    CHECK (payment_method IN ('cod', 'onepay'));
