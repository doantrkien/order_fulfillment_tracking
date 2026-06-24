BEGIN;

-- ============================================================
-- Table 1: ai_evaluation_runs
-- Lưu trữ trạng thái chạy batch đánh giá AI
-- ============================================================
DROP TABLE IF EXISTS ai_evaluation_results_detail CASCADE;
DROP TABLE IF EXISTS ai_evaluation_runs CASCADE;

CREATE TABLE IF NOT EXISTS ai_evaluation_runs (
    id              BIGSERIAL       PRIMARY KEY,
    dataset_name    VARCHAR(255)    NOT NULL,
    status          VARCHAR(20)     NOT NULL DEFAULT 'PENDING'
                        CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED')),
    total_cases     INTEGER         NOT NULL DEFAULT 0,
    passed_cases    INTEGER         NOT NULL DEFAULT 0,
    failed_cases    INTEGER         NOT NULL DEFAULT 0,
    fallback_count  INTEGER         NOT NULL DEFAULT 0,
    accuracy_rate   DECIMAL(5,2)    DEFAULT 0.00,
    avg_latency_ms  INTEGER         DEFAULT 0,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ai_eval_runs_status
    ON ai_evaluation_runs(status);

CREATE INDEX IF NOT EXISTS idx_ai_eval_runs_created_at
    ON ai_evaluation_runs(created_at DESC);

-- ============================================================
-- Table 2: ai_evaluation_results_detail
-- Lưu chi tiết từng case trong batch
-- ============================================================
CREATE TABLE IF NOT EXISTS ai_evaluation_results_detail (
    id               BIGSERIAL       PRIMARY KEY,
    run_id           BIGINT          NOT NULL REFERENCES ai_evaluation_runs(id) ON DELETE CASCADE,
    input            JSONB           NOT NULL,
    expected_output  JSONB           NOT NULL,
    actual_output    JSONB,
    status           VARCHAR(20)     NOT NULL DEFAULT 'PENDING'
                        CHECK (status IN ('PENDING', 'PASSED', 'FAILED')),
    latency_ms       INTEGER         DEFAULT 0,
    error_message    TEXT,
    created_at       TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ai_eval_details_run_id
    ON ai_evaluation_results_detail(run_id);

COMMIT;
