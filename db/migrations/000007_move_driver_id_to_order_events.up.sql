-- Remove driver_id from orders
ALTER TABLE orders DROP COLUMN IF EXISTS driver_id;

-- Add driver_id to order_events (nullable — not every event needs a driver)
ALTER TABLE order_events ADD COLUMN driver_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL;
