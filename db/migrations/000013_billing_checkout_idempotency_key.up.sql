-- Links a billing checkout row to the client Idempotency-Key so retries return the same order.
ALTER TABLE billing_checkouts ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_billing_checkouts_user_idempotency_key
    ON billing_checkouts (user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
