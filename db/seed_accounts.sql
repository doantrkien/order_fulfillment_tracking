-- Seed accounts with password '123456'
-- Bcrypt hash for 123456: $2a$10$w/lW8qT/51/AwaF877oIL.k9wykL.PMMDe1kFQpfmpQeW0I2URgIi

INSERT INTO accounts (username, email, password, role)
VALUES 
    ('admin', 'admin@demo.com', '$2a$10$w/lW8qT/51/AwaF877oIL.k9wykL.PMMDe1kFQpfmpQeW0I2URgIi', 'admin'),
    ('driver', 'driver@demo.com', '$2a$10$w/lW8qT/51/AwaF877oIL.k9wykL.PMMDe1kFQpfmpQeW0I2URgIi', 'driver')
ON CONFLICT (email) DO UPDATE 
SET password = EXCLUDED.password;
