BEGIN;
ALTER TABLE IF EXISTS orders
  ADD COLUMN IF NOT EXISTS created_at timestamptz DEFAULT now(),
  ADD COLUMN IF NOT EXISTS current_status varchar(64),
  ADD COLUMN IF NOT EXISTS total_amount numeric(14,2);

CREATE TABLE IF NOT EXISTS order_events (
  id bigserial PRIMARY KEY,
  order_id bigint NOT NULL,
  new_status varchar(64),
  event_at timestamptz DEFAULT now(),
  payload jsonb,
  created_at timestamptz DEFAULT now()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_order_events_orders'
  ) THEN
    ALTER TABLE IF EXISTS order_events
      ADD CONSTRAINT fk_order_events_orders FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;
  END IF;
END$$;

CREATE TABLE IF NOT EXISTS reports (
  id bigserial PRIMARY KEY,
  date date NOT NULL,
  CONSTRAINT uni_reports_date UNIQUE (date),
  total_orders integer DEFAULT 0,
  total_new integer DEFAULT 0,
  total_delivered integer DEFAULT 0,
  total_cancelled integer DEFAULT 0,
  total_refunded integer DEFAULT 0,
  total_income numeric(18,2) DEFAULT 0,
  avg_deliver_time double precision DEFAULT 0,
  created_at timestamptz DEFAULT now(),
  updated_at timestamptz DEFAULT now()
);

COMMIT;
