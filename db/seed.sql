DO $$
DECLARE
    i INT;
    new_order_id INT;
    statuses TEXT[] := ARRAY['created', 'paid', 'packed', 'shipped', 'delivered', 'cancelled', 'refunded'];
    selected_status TEXT;
    driver_user_id BIGINT;
    base_time TIMESTAMPTZ;
BEGIN
    -- Comment out these lines to keep existing data
    TRUNCATE TABLE order_events CASCADE;
    TRUNCATE TABLE orders CASCADE;
    TRUNCATE TABLE accounts CASCADE;

    RAISE NOTICE 'Seeding accounts...';
    
    -- Insert admin_demo
    INSERT INTO accounts (username, email, password, role)
    VALUES ('admin_demo', 'admin@demo.com', '$2a$10$q.GA5YzP/E..uN5vJQekk.p69YadgIshe5GvvnTr48jVB.oyMSHt2', 'admin');

    -- Insert driver_demo and get its ID
    INSERT INTO accounts (username, email, password, role)
    VALUES ('driver_demo', 'driver@demo.com', '$2a$10$SnvlwGCRvI8QuUaps9Z2I.Um3QDNQlU/VvSnG84OqhC5UIwaTdU.C', 'driver')
    RETURNING id INTO driver_user_id;

    RAISE NOTICE 'Start seed 1,000,000 orders and events...';

    FOR i IN 1..1000000 LOOP
        selected_status := statuses[floor(random() * 7) + 1];
        
        -- Create a base timestamp for this order (some time in the last 30 days)
        base_time := NOW() - (random() * 30) * '1 day'::interval;
        
        -- Insert order
        INSERT INTO orders (user_info, total_amount, current_status, driver_id, created_at, updated_at)
        VALUES (
            '{"username": "user_test", "user_phone": "0901234567", "shipping_address": "123 Test St"}'::jsonb,
            (random() * 1000 + 100)::numeric(14,2), 
            selected_status,
            driver_user_id,
            base_time,
            base_time + '5 hours'::interval
        ) RETURNING id INTO new_order_id;
        
        -- Insert events sequentially to represent real transitions
        -- Initial event: created
        INSERT INTO order_events (order_id, new_status, event_at, created_at, updated_by)
        VALUES (new_order_id, 'created', base_time, base_time, 'system');
        
        -- Sequential events depending on selected_status
        IF selected_status = 'cancelled' THEN
            INSERT INTO order_events (order_id, new_status, event_at, created_at, updated_by)
            VALUES (new_order_id, 'cancelled', base_time + '10 minutes'::interval, base_time + '10 minutes'::interval, 'system');
            
        ELSIF selected_status IN ('paid', 'packed', 'shipped', 'delivered', 'refunded') THEN
            INSERT INTO order_events (order_id, new_status, event_at, created_at, updated_by)
            VALUES (new_order_id, 'paid', base_time + '30 minutes'::interval, base_time + '30 minutes'::interval, 'system');
            
            IF selected_status = 'refunded' THEN
                INSERT INTO order_events (order_id, new_status, event_at, created_at, updated_by)
                VALUES (new_order_id, 'refunded', base_time + '1 hour'::interval, base_time + '1 hour'::interval, 'system');
                
            ELSIF selected_status IN ('packed', 'shipped', 'delivered') THEN
                INSERT INTO order_events (order_id, new_status, event_at, created_at, updated_by)
                VALUES (new_order_id, 'packed', base_time + '2 hours'::interval, base_time + '2 hours'::interval, 'system');
                
                IF selected_status IN ('shipped', 'delivered') THEN
                    INSERT INTO order_events (order_id, new_status, event_at, created_at, updated_by)
                    VALUES (new_order_id, 'shipped', base_time + '4 hours'::interval, base_time + '4 hours'::interval, 'system');
                    
                    IF selected_status = 'delivered' THEN
                        INSERT INTO order_events (order_id, new_status, event_at, created_at, updated_by)
                        VALUES (new_order_id, 'delivered', base_time + '5 hours'::interval, base_time + '5 hours'::interval, 'system');
                    END IF;
                END IF;
            END IF;
        END IF;
        
    END LOOP;

    RAISE NOTICE 'Seed data successfully!';
END $$;
