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

    -- Timestamps
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP
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


-- ============================================================
-- Table 2: ai_evaluation_runs
-- Lưu kết quả mỗi lần chạy evaluation batch (>= 20 test cases).
-- Yêu cầu Phase 3: phải có evaluation evidence cho demo và peer review.
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_evaluation_runs (
    -- Identity
    id                      BIGSERIAL       PRIMARY KEY,

    -- Run Metadata
    workflow_name           VARCHAR(100)    NOT NULL,            -- e.g. 'order-exception-analysis'
    prompt_template_version VARCHAR(20)     NOT NULL DEFAULT 'v1',
    run_by                  VARCHAR(100),                        -- người/service trigger eval
    environment             VARCHAR(20)     NOT NULL DEFAULT 'dev'
                                CHECK (environment IN ('dev', 'staging', 'prod')),

    -- Aggregate Results
    total_cases             INTEGER         NOT NULL DEFAULT 0,
    passed_cases            INTEGER         NOT NULL DEFAULT 0,
    failed_cases            INTEGER         NOT NULL DEFAULT 0,
    fallback_cases          INTEGER         NOT NULL DEFAULT 0,
    avg_confidence          DECIMAL(4,3)    CHECK (avg_confidence >= 0.0 AND avg_confidence <= 1.0),
    pass_rate               DECIMAL(5,4)    CHECK (pass_rate >= 0.0 AND pass_rate <= 1.0),  -- passed/total

    -- Detail & Debug
    notes                   TEXT,
    raw_results             JSONB,                               -- mảng chi tiết từng test case
    triggered_by            VARCHAR(50)     NOT NULL DEFAULT 'manual'
                                CHECK (triggered_by IN ('manual', 'ci', 'scheduled')),

    -- Timestamps
    run_at                  TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_ai_eval_runs_workflow
    ON ai_evaluation_runs(workflow_name);

CREATE INDEX IF NOT EXISTS idx_ai_eval_runs_prompt_version
    ON ai_evaluation_runs(prompt_template_version);

CREATE INDEX IF NOT EXISTS idx_ai_eval_runs_run_at
    ON ai_evaluation_runs(run_at DESC);

COMMIT;