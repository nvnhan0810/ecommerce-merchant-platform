-- +goose Up
CREATE TABLE IF NOT EXISTS order_events (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (
        event_type IN ('created', 'status_changed', 'cancelled')
    ),
    from_status TEXT,
    to_status TEXT,
    message TEXT NOT NULL DEFAULT '',
    actor_id TEXT NOT NULL DEFAULT '',
    actor_email TEXT NOT NULL DEFAULT '',
    actor_role TEXT NOT NULL DEFAULT '',
    actor_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_events_order_id_created_at
    ON order_events (order_id, created_at ASC);

-- Backfill history for orders seeded before this migration.
INSERT INTO order_events (
    id, order_id, event_type, from_status, to_status, message,
    actor_id, actor_email, actor_role, actor_name, created_at
)
SELECT
    gen_random_uuid(),
    o.id,
    'created',
    NULL,
    'new',
    'Tạo đơn hàng',
    o.user_id::text,
    '',
    'user',
    '',
    o.created_at
FROM orders o
WHERE NOT EXISTS (
    SELECT 1 FROM order_events e WHERE e.order_id = o.id AND e.event_type = 'created'
);

INSERT INTO order_events (
    id, order_id, event_type, from_status, to_status, message,
    actor_id, actor_email, actor_role, actor_name, created_at
)
SELECT
    gen_random_uuid(),
    o.id,
    CASE WHEN o.status = 'cancelled' THEN 'cancelled' ELSE 'status_changed' END,
    'new',
    o.status,
    CASE
        WHEN o.status = 'cancelled' THEN 'Huỷ đơn hàng'
        ELSE 'Đổi trạng thái từ Mới sang ' || CASE o.status
            WHEN 'paid' THEN 'Đã thanh toán'
            WHEN 'confirmed' THEN 'Đã xác nhận'
            WHEN 'shipping' THEN 'Đang vận chuyển'
            WHEN 'succeeded' THEN 'Thành công'
            WHEN 'failed' THEN 'Thất bại'
            ELSE o.status
        END
    END,
    '',
    '',
    'system',
    'Hệ thống',
    o.updated_at
FROM orders o
WHERE o.status <> 'new'
  AND NOT EXISTS (
      SELECT 1
      FROM order_events e
      WHERE e.order_id = o.id
        AND e.event_type IN ('status_changed', 'cancelled')
  );

-- +goose Down
DROP INDEX IF EXISTS idx_order_events_order_id_created_at;
DROP TABLE IF EXISTS order_events;
