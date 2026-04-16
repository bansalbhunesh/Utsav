package httpserver

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ErrIdempotencyFingerprintMismatch is returned when the same idempotency key was used with a different request body.
var ErrIdempotencyFingerprintMismatch = errors.New("idempotency fingerprint mismatch")

// reserveIdempotencyInTx mirrors rsvprepo.UpsertRSVPResponsesIdempotent: same transaction as the mutation.
// If replay is true, the caller must not repeat side effects (insert); commit the transaction after any read needed for the response.
func reserveIdempotencyInTx(ctx context.Context, tx pgx.Tx, scope, key, fingerprint string) (replay bool, err error) {
	if _, err := tx.Exec(ctx, `
		DELETE FROM idempotency_keys WHERE scope=$1 AND key=$2 AND expires_at < now()
	`, scope, key); err != nil {
		return false, err
	}
	tag, err := tx.Exec(ctx, `
		INSERT INTO idempotency_keys (scope, key, fingerprint, expires_at)
		VALUES ($1, $2, $3, now() + interval '24 hours')
		ON CONFLICT (scope, key) DO NOTHING
	`, scope, key, fingerprint)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		var existing string
		err = tx.QueryRow(ctx, `
			SELECT fingerprint FROM idempotency_keys
			WHERE scope=$1 AND key=$2 AND expires_at > now()
			FOR UPDATE
		`, scope, key).Scan(&existing)
		if err != nil {
			return false, err
		}
		if existing != fingerprint {
			return false, ErrIdempotencyFingerprintMismatch
		}
		return true, nil
	}
	return false, nil
}
