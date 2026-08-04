-- +goose Up
CREATE TABLE IF NOT EXISTS payment_callback_events (
    id UUID PRIMARY KEY,
    provider TEXT NOT NULL,
    channel TEXT NOT NULL,
    http_method TEXT NOT NULL DEFAULT '',
    payment_id UUID,
    payment_method TEXT NOT NULL DEFAULT '',
    merch_txn_ref TEXT NOT NULL DEFAULT '',
    response_code TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    paid BOOLEAN NOT NULL DEFAULT FALSE,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    error_message TEXT NOT NULL DEFAULT '',
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payment_callback_events_created
    ON payment_callback_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_callback_events_provider
    ON payment_callback_events (provider, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_callback_events_channel
    ON payment_callback_events (channel, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_callback_events_ref
    ON payment_callback_events (merch_txn_ref)
    WHERE merch_txn_ref <> '';

-- +goose Down
DROP TABLE IF EXISTS payment_callback_events;
