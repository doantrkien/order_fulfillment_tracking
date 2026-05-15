ALTER TABLE reports DROP COLUMN updated_at;
ALTER TABLE order_events DROP COLUMN updated_at;
ALTER TABLE order_events DROP COLUMN event_at;
ALTER TABLE orders ALTER COLUMN total_amount TYPE DECIMAL(12,2) USING total_amount::DECIMAL(12,2);
