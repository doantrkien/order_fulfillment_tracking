<<<<<<< HEAD
ALTER TABLE order_events ADD COLUMN updated_by VARCHAR(100) NOT NULL DEFAULT '';
=======
ALTER TABLE order_events ADD COLUMN IF NOT EXISTS updated_by VARCHAR(100) NOT NULL DEFAULT '';
>>>>>>> dev
