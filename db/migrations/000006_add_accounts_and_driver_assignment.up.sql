-- System accounts (admin and driver roles only)
CREATE TABLE accounts (
    id         BIGSERIAL PRIMARY KEY,
    username   VARCHAR(100) NOT NULL UNIQUE,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,          -- bcrypt hash
    role       VARCHAR(20)  NOT NULL,          -- 'admin' or 'driver'
    created_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Assign each order to a driver
ALTER TABLE orders ADD COLUMN driver_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL;
