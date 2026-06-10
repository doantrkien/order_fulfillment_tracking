BEGIN;
ALTER TABLE order_events ADD COLUMN driver_note TEXT;
COMMIT;