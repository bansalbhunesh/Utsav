DROP INDEX IF EXISTS idx_webhook_deliveries_retry;

ALTER TABLE webhook_deliveries
    DROP COLUMN IF EXISTS processed_at,
    DROP COLUMN IF EXISTS next_retry_at,
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS attempt_count,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS order_id;
