DROP INDEX IF EXISTS idx_idempotency_keys_expires_at;
ALTER TABLE idempotency_keys DROP COLUMN IF EXISTS expires_at;
