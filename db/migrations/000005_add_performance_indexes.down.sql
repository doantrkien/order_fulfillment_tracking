BEGIN;

DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_current_status;
DROP INDEX IF EXISTS idx_order_events_order_id;

COMMIT;
