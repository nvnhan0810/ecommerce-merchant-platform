-- +goose Up
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders
    ADD CONSTRAINT orders_status_check CHECK (
        status IN (
            'awaiting_payment',
            'new',
            'paid',
            'confirmed',
            'shipping',
            'succeeded',
            'returning',
            'returned',
            'failed',
            'cancelled'
        )
    );

-- +goose Down
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders
    ADD CONSTRAINT orders_status_check CHECK (
        status IN (
            'new',
            'paid',
            'confirmed',
            'shipping',
            'succeeded',
            'returning',
            'returned',
            'failed',
            'cancelled'
        )
    );
