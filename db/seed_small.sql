DO $$
DECLARE
    i INT;
    new_order_id INT;
    selected_status TEXT;
    driver_user_id BIGINT;
    base_time TIMESTAMPTZ;
BEGIN
    -- Comment out these lines if you want to keep existing data
    TRUNCATE TABLE order_events CASCADE;
    TRUNCATE TABLE orders CASCADE;
    TRUNCATE TABLE accounts CASCADE;

    RAISE NOTICE 'Seeding accounts...';
    
    -- Insert admin_demo
    INSERT INTO accounts (username, email, password, role)
    VALUES ('admin_demo', 'admin@demo.com', '$2a$10$w/lW8qT/51/AwaF877oIL.k9wykL.PMMDe1kFQpfmpQeW0I2URgIi', 'admin');

    -- Insert driver_demo and get its ID
    INSERT INTO accounts (username, email, password, role)
    VALUES ('driver_demo', 'driver@demo.com', '$2a$10$w/lW8qT/51/AwaF877oIL.k9wykL.PMMDe1kFQpfmpQeW0I2URgIi', 'driver')
    RETURNING id INTO driver_user_id;

    RAISE NOTICE 'Start seed 10 orders (shipped & delivered)...';

    FOR i IN 1..10 LOOP
        -- Randomly choose between shipped and delivered
        IF random() > 0.5 THEN
            selected_status := 'shipped';
        ELSE
            selected_status := 'delivered';
        END IF;
        
        -- Create a base timestamp for this order (some time in the last 12 hours)
        base_time := NOW() - (random() * 12) * '1 hour'::interval;
        
        -- Insert order
        INSERT INTO orders (user_info, total_amount, current_status, created_at, updated_at)
        VALUES (
            '{"username": "user_test", "user_phone": "0901234567", "shipping_address": "123 Test St"}'::jsonb,
            (random() * 1000 + 100)::numeric(14,2), 
            selected_status,
            base_time,
            base_time + '5 hours'::interval
        ) RETURNING id INTO new_order_id;
        
        -- Event 1: created
        INSERT INTO order_events (order_id, previous_status, new_status, event_at, created_at, updated_by)
        VALUES (new_order_id, NULL, 'created', base_time, base_time, 'system');
        
        -- Event 2: paid
        INSERT INTO order_events (order_id, previous_status, new_status, event_at, created_at, updated_by)
        VALUES (new_order_id, 'created', 'paid', base_time + '30 minutes'::interval, base_time + '30 minutes'::interval, 'system');
        
        -- Event 3: packed
        INSERT INTO order_events (order_id, previous_status, new_status, event_at, created_at, updated_by)
        VALUES (new_order_id, 'paid', 'packed', base_time + '2 hours'::interval, base_time + '2 hours'::interval, 'system');
        
        -- Event 4: shipped
        INSERT INTO order_events (order_id, previous_status, new_status, event_at, created_at, updated_by, driver_id)
        VALUES (new_order_id, 'packed', 'shipped', base_time + '4 hours'::interval, base_time + '4 hours'::interval, 'driver_demo', driver_user_id);
        
        IF selected_status = 'delivered' THEN
            -- Event 5: delivered
            INSERT INTO order_events (order_id, previous_status, new_status, event_at, created_at, updated_by, driver_id)
            VALUES (new_order_id, 'shipped', 'delivered', base_time + '5 hours'::interval, base_time + '5 hours'::interval, 'driver_demo', driver_user_id);
        END IF;
        
    END LOOP;

    RAISE NOTICE 'Seed data successfully!';
END $$;
