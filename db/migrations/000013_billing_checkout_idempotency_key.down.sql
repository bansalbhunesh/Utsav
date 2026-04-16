DROP INDEX IF EXISTS idx_billing_checkouts_user_idempotency_key;
ALTER TABLE billing_checkouts DROP COLUMN IF EXISTS idempotency_key;
