BEGIN;

-- ============================================================
-- Table 1: ai_customer_update_drafts
-- Lưu bản nháp tin nhắn khách hàng do AI tạo ra.
-- Yêu cầu Phase 3: AI chỉ được DRAFT, không được tự gửi.
-- Lifecycle: PENDING -> APPROVED -> SENT | REJECTED
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_customer_update_drafts (
    -- Identity
    id                      BIGSERIAL       PRIMARY KEY,
    order_id                BIGINT          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    ai_exception_result_id  BIGINT          REFERENCES ai_order_exception_results(id) ON DELETE SET NULL,

    -- Draft Content (AC 1)
    draft_message           TEXT            NOT NULL,
    tone                    VARCHAR(20)     NOT NULL DEFAULT 'neutral'
                                CHECK (tone IN ('neutral', 'apologetic', 'informative', 'proactive')),
    channel                 VARCHAR(50)     NOT NULL DEFAULT 'email'
                                CHECK (channel IN ('email', 'sms', 'push_notification')),

    -- Human Review State — AI không được tự chuyển sang SENT (No auto-critical actions)
    review_status           VARCHAR(20)     NOT NULL DEFAULT 'PENDING'
                                CHECK (review_status IN ('PENDING', 'APPROVED', 'REJECTED', 'SENT')),
    reviewed_by             VARCHAR(100),                        -- NULL nếu chưa review
    reviewed_at             TIMESTAMPTZ,                         -- NULL nếu chưa review
    sent_at                 TIMESTAMPTZ,                         -- NULL nếu chưa gửi

    -- AI Metadata
    confidence_score        DECIMAL(4,3)    CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    fallback_used           BOOLEAN         NOT NULL DEFAULT FALSE,
    fallback_reason         TEXT,                                -- NULL nếu fallback_used = false
    prompt_template_version VARCHAR(20)     NOT NULL DEFAULT 'v1',
    input_size_bytes        INTEGER         NOT NULL DEFAULT 0,
    duration_ms             INTEGER,
    request_id              VARCHAR(100),                        -- liên kết với exception analysis request
    raw_response            JSONB,                               -- lưu full response AI để debug

    -- Timestamps
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_ai_drafts_order_id
    ON ai_customer_update_drafts(order_id);

CREATE INDEX IF NOT EXISTS idx_ai_drafts_review_status
    ON ai_customer_update_drafts(review_status);

CREATE INDEX IF NOT EXISTS idx_ai_drafts_exception_result_id
    ON ai_customer_update_drafts(ai_exception_result_id)
    WHERE ai_exception_result_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ai_drafts_pending
    ON ai_customer_update_drafts(created_at DESC)
    WHERE review_status = 'PENDING';              -- partial index: operator queue

COMMIT;
