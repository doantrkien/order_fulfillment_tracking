BEGIN;

CREATE TABLE IF NOT EXISTS ai_order_exception_results (
    -- Identity
    id                      BIGSERIAL PRIMARY KEY,
    order_id                BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,

    -- AI Output Schema (AC 1)
    exception_type          VARCHAR(50)     NOT NULL,
    severity                VARCHAR(20)     NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    likely_reason           TEXT            NOT NULL,
    internal_next_action    TEXT            NOT NULL,
    confidence_score        DECIMAL(4,3)    NOT NULL CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),

    -- Fallback Policy (AC 3)
    fallback_used           BOOLEAN         NOT NULL DEFAULT FALSE,
    fallback_reason         TEXT,           -- NULL nếu fallback_used = false

    -- Audit & Observability (AC 4)
    prompt_template_version VARCHAR(20)     NOT NULL DEFAULT 'v1',
    input_size_bytes        INTEGER         NOT NULL DEFAULT 0,  -- kích thước input trước khi build prompt
    duration_ms             INTEGER,                             -- thời gian gọi AI
    request_id              VARCHAR(100),                        -- để truy vết log (AC 4)
    raw_response            JSONB,                               -- lưu full response AI để debug

    -- Timestamps
    evaluated_at            TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_ai_exception_results_order_id
    ON ai_order_exception_results(order_id);

CREATE INDEX IF NOT EXISTS idx_ai_exception_results_exception_type
    ON ai_order_exception_results(exception_type);

CREATE INDEX IF NOT EXISTS idx_ai_exception_results_evaluated_at
    ON ai_order_exception_results(evaluated_at DESC);

CREATE INDEX IF NOT EXISTS idx_ai_exception_results_fallback
    ON ai_order_exception_results(fallback_used)
    WHERE fallback_used = TRUE;  -- partial index, chỉ index các row có fallback

COMMIT;