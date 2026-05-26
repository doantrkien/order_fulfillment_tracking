BEGIN;

CREATE INDEX IF NOT EXISTS idx_order_events_order_id ON order_events(order_id);

CREATE INDEX IF NOT EXISTS idx_orders_current_status ON orders(current_status);

CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

COMMIT;
