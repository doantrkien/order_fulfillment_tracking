BEGIN;

-- Drop indexes trước, sau đó drop table
-- Thứ tự: bảng có FK trỏ đến bảng khác phải drop trước

-- ai_customer_update_drafts indexes
DROP INDEX IF EXISTS idx_ai_drafts_pending;
DROP INDEX IF EXISTS idx_ai_drafts_exception_result_id;
DROP INDEX IF EXISTS idx_ai_drafts_review_status;
DROP INDEX IF EXISTS idx_ai_drafts_order_id;
DROP TABLE IF EXISTS ai_customer_update_drafts;

COMMIT;
