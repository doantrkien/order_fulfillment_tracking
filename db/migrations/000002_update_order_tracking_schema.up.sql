-- 1. Sửa kiểu dữ liệu bảng orders (Từ số thập phân sang số nguyên)
ALTER TABLE orders ALTER COLUMN total_amount TYPE BIGINT USING total_amount::BIGINT;

-- 2. Thêm các cột thời gian vào bảng order_events
ALTER TABLE order_events ADD COLUMN event_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE order_events ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- 3. Thêm cột thời gian vào bảng reports
ALTER TABLE reports ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
