DO $$
DECLARE
    i INT;
    new_order_id INT;
    statuses TEXT[] := ARRAY['PENDING', 'DELIVERING', 'DELIVERED', 'FAILED'];
    selected_status TEXT;
BEGIN
    -- Comment out these two lines to keep existing data
    TRUNCATE TABLE order_events CASCADE;
    TRUNCATE TABLE orders CASCADE;

    RAISE NOTICE 'Start seed 10,000,000 orders and events...';

    FOR i IN 1..10000000 LOOP
        selected_status := statuses[floor(random() * 4) + 1];
        
        -- 1. Insert order
        INSERT INTO orders (user_info, total_amount, current_status, created_at, updated_at)
        VALUES (
            '{"username": "user_test", "user_phone": "0901234567", "shipping_address": "123 Test St"}'::jsonb,
            (random() * 1000 + 100)::numeric(14,2), 
            selected_status, 
            NOW() - (random() * 30 || ' days')::interval,
            NOW()
        ) RETURNING id INTO new_order_id;
        
        -- Insert events for each order
        INSERT INTO order_events (order_id, new_status, event_at, payload, created_at, updated_by)
        VALUES 
            (new_order_id, 'PENDING', NOW() - '5 hours'::interval, '{"note": "Order created"}'::jsonb, NOW() - '5 hours'::interval, 'system'),
            (new_order_id, 'DELIVERING', NOW() - '2 hours'::interval, '{"note": "Shipper picked up"}'::jsonb, NOW() - '2 hours'::interval, 'system');
            
        -- Insert final event for DELIVERED or FAILED status
        IF selected_status = 'DELIVERED' THEN
            INSERT INTO order_events (order_id, new_status, event_at, payload, created_at, updated_by)
            VALUES (new_order_id, 'DELIVERED', NOW(), '{"note": "Delivered successfully"}'::jsonb, NOW(), 'system');
        ELSIF selected_status = 'FAILED' THEN
            INSERT INTO order_events (order_id, new_status, event_at, payload, created_at, updated_by)
            VALUES (new_order_id, 'FAILED', NOW(), '{"note": "Delivery failed"}'::jsonb, NOW(), 'system');
        END IF;
    END LOOP;

    RAISE NOTICE 'Seed data successfully!';
END $$;
