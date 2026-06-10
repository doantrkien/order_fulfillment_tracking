BEGIN;
 
DROP INDEX IF EXISTS idx_ai_exception_results_fallback;
DROP INDEX IF EXISTS idx_ai_exception_results_evaluated_at;
DROP INDEX IF EXISTS idx_ai_exception_results_exception_type;
DROP INDEX IF EXISTS idx_ai_exception_results_order_id;
DROP TABLE IF EXISTS ai_order_exception_result;
 
COMMIT;
 
