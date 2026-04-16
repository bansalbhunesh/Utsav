package billingrepo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Checkout struct {
	ID      string `json:"id"`
	EventID any    `json:"event_id"`
	Tier    string `json:"tier"`
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Created any    `json:"created_at"`
}

type CreateCheckoutInput struct {
	UserID  uuid.UUID
	EventID string
	Tier    string
	OrderID string
}

type Repository interface {
	CreateCheckout(ctx context.Context, input CreateCheckoutInput) (string, error)
	CreateCheckoutIdempotent(ctx context.Context, scope, idempotencyKey, fingerprint string, input CreateCheckoutInput) (checkoutID string, orderID string, replay bool, err error)
	ListCheckouts(ctx context.Context, userID uuid.UUID) ([]Checkout, error)
	MarkOrderPaidAndFetch(ctx context.Context, orderID string) (string, any, error)
	MarkOrderPaidFromWebhook(ctx context.Context, provider, eventKey, payloadHash, orderID string) error
	RetryPendingWebhookDeliveries(ctx context.Context, provider string, limit int) (int, error)
}

var (
	ErrCheckoutNotFound               = errors.New("checkout not found")
	ErrWebhookDedupePayloadMismatch   = errors.New("webhook event_key replay with different payload")
	ErrIdempotencyFingerprintMismatch = errors.New("billingrepo: idempotency fingerprint mismatch")
)

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) CreateCheckout(ctx context.Context, input CreateCheckoutInput) (string, error) {
	var eid any
	if input.EventID != "" {
		if u, err := uuid.Parse(input.EventID); err == nil {
			eid = u
		}
	}
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO billing_checkouts (user_id, event_id, tier, razorpay_order_id, status)
		VALUES ($1,$2,$3,$4,'created') RETURNING id`,
		input.UserID, eid, strings.TrimSpace(strings.ToLower(input.Tier)), input.OrderID).Scan(&id)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

// CreateCheckoutIdempotent reserves the idempotency key and creates the checkout in one transaction.
// On replay (same key and fingerprint), returns the existing checkout id and order id without inserting again.
func (r *PGRepository) CreateCheckoutIdempotent(ctx context.Context, scope, idempotencyKey, fingerprint string, input CreateCheckoutInput) (checkoutID string, orderID string, replay bool, err error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return "", "", false, fmt.Errorf("missing idempotency key")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", "", false, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		DELETE FROM idempotency_keys WHERE scope=$1 AND key=$2 AND expires_at < now()
	`, scope, idempotencyKey); err != nil {
		return "", "", false, err
	}
	tag, err := tx.Exec(ctx, `
		INSERT INTO idempotency_keys (scope, key, fingerprint, expires_at)
		VALUES ($1, $2, $3, now() + interval '24 hours')
		ON CONFLICT (scope, key) DO NOTHING
	`, scope, idempotencyKey, fingerprint)
	if err != nil {
		return "", "", false, err
	}
	if tag.RowsAffected() == 0 {
		var existing string
		err = tx.QueryRow(ctx, `
			SELECT fingerprint FROM idempotency_keys
			WHERE scope=$1 AND key=$2 AND expires_at > now()
			FOR UPDATE
		`, scope, idempotencyKey).Scan(&existing)
		if err != nil {
			return "", "", false, err
		}
		if existing != fingerprint {
			return "", "", false, ErrIdempotencyFingerprintMismatch
		}
		var cid string
		var rid string
		err = tx.QueryRow(ctx, `
			SELECT id::text, razorpay_order_id FROM billing_checkouts
			WHERE user_id=$1 AND idempotency_key=$2
			FOR UPDATE`,
			input.UserID, idempotencyKey,
		).Scan(&cid, &rid)
		if err != nil {
			return "", "", false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return "", "", false, err
		}
		return cid, rid, true, nil
	}

	var eid any
	if input.EventID != "" {
		if u, err := uuid.Parse(input.EventID); err == nil {
			eid = u
		}
	}
	var id uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO billing_checkouts (user_id, event_id, tier, razorpay_order_id, status, idempotency_key)
		VALUES ($1,$2,$3,$4,'created',$5) RETURNING id`,
		input.UserID, eid, strings.TrimSpace(strings.ToLower(input.Tier)), input.OrderID, idempotencyKey,
	).Scan(&id)
	if err != nil {
		return "", "", false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", "", false, err
	}
	return id.String(), input.OrderID, false, nil
}

func (r *PGRepository) ListCheckouts(ctx context.Context, userID uuid.UUID) ([]Checkout, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, event_id, tier, razorpay_order_id, status, created_at
		FROM billing_checkouts
		WHERE user_id=$1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Checkout, 0)
	for rows.Next() {
		var id uuid.UUID
		var c Checkout
		if err := rows.Scan(&id, &c.EventID, &c.Tier, &c.OrderID, &c.Status, &c.Created); err != nil {
			return nil, fmt.Errorf("scan checkout row: %w", err)
		}
		c.ID = id.String()
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate checkout rows: %w", err)
	}
	return out, nil
}

func markOrderPaidInTx(ctx context.Context, tx pgx.Tx, orderID string) (tier string, eventID *uuid.UUID, err error) {
	var status string
	err = tx.QueryRow(ctx, `
		SELECT tier, event_id, status FROM billing_checkouts WHERE razorpay_order_id=$1
		FOR UPDATE`, orderID).Scan(&tier, &eventID, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, ErrCheckoutNotFound
		}
		return "", nil, err
	}
	if status != "paid" {
		if _, err := tx.Exec(ctx, `UPDATE billing_checkouts SET status='paid' WHERE razorpay_order_id=$1`, orderID); err != nil {
			return "", nil, err
		}
	}
	if eventID != nil {
		if _, err := tx.Exec(ctx, `UPDATE events SET tier=$1 WHERE id=$2`, tier, *eventID); err != nil {
			return "", nil, err
		}
	}
	return tier, eventID, nil
}

func (r *PGRepository) MarkOrderPaidAndFetch(ctx context.Context, orderID string) (string, any, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", nil, err
	}
	defer tx.Rollback(ctx)

	tier, eventID, err := markOrderPaidInTx(ctx, tx, orderID)
	if err != nil {
		return "", nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", nil, err
	}
	return tier, eventID, nil
}

// MarkOrderPaidFromWebhook records webhook deduplication and marks the order paid in a single transaction.
func (r *PGRepository) MarkOrderPaidFromWebhook(ctx context.Context, provider, eventKey, payloadHash, orderID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var existingHash string
	var status string
	var attempts int
	err = tx.QueryRow(ctx, `
		SELECT payload_hash, status, attempt_count
		FROM webhook_deliveries
		WHERE provider=$1 AND event_key=$2
		FOR UPDATE
	`, provider, eventKey).Scan(&existingHash, &status, &attempts)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		if _, err := tx.Exec(ctx, `
			INSERT INTO webhook_deliveries (
				provider, event_key, payload_hash, order_id, status, attempt_count, next_retry_at
			)
			VALUES ($1,$2,$3,$4,'received',0,now())
		`, provider, eventKey, payloadHash, orderID); err != nil {
			return err
		}
	} else {
		if existingHash != payloadHash {
			return ErrWebhookDedupePayloadMismatch
		}
		if _, err := tx.Exec(ctx, `
			UPDATE webhook_deliveries
			SET order_id = CASE WHEN order_id = '' THEN $3 ELSE order_id END
			WHERE provider=$1 AND event_key=$2
		`, provider, eventKey, orderID); err != nil {
			return err
		}
		if status == "processed" {
			return tx.Commit(ctx)
		}
	}

	if _, _, err := markOrderPaidInTx(ctx, tx, orderID); err != nil {
		_, _ = tx.Exec(ctx, `
			UPDATE webhook_deliveries
			SET status='failed',
				attempt_count = attempt_count + 1,
				last_error = LEFT($3, 1024),
				next_retry_at = now() + interval '2 minutes'
			WHERE provider=$1 AND event_key=$2
		`, provider, eventKey, err.Error())
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE webhook_deliveries
		SET status='processed',
			attempt_count = attempt_count + 1,
			last_error = '',
			next_retry_at = now(),
			processed_at = now()
		WHERE provider=$1 AND event_key=$2
	`, provider, eventKey); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PGRepository) RetryPendingWebhookDeliveries(ctx context.Context, provider string, limit int) (int, error) {
	if limit <= 0 {
		limit = 25
	}
	pv := strings.TrimSpace(strings.ToLower(provider))

	// Single transaction: claim rows with SKIP LOCKED so each API replica gets a disjoint batch
	// (no duplicate work / double payment attempts across Render instances).
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	type row struct {
		eventKey string
		orderID  string
	}
	rows, err := tx.Query(ctx, `
		SELECT event_key, order_id
		FROM webhook_deliveries
		WHERE provider = $1
		  AND status IN ('received', 'failed')
		  AND order_id <> ''
		  AND next_retry_at <= now()
		ORDER BY next_retry_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`, pv, limit)
	if err != nil {
		return 0, err
	}
	pending := make([]row, 0, limit)
	for rows.Next() {
		var rw row
		if err := rows.Scan(&rw.eventKey, &rw.orderID); err != nil {
			rows.Close()
			return 0, err
		}
		pending = append(pending, rw)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()

	processed := 0
	for _, p := range pending {
		if _, _, err := markOrderPaidInTx(ctx, tx, p.orderID); err != nil {
			_, _ = tx.Exec(ctx, `
				UPDATE webhook_deliveries
				SET status='failed',
					attempt_count = attempt_count + 1,
					last_error = LEFT($3, 1024),
					next_retry_at = now() + interval '5 minutes'
				WHERE provider=$1 AND event_key=$2
			`, pv, p.eventKey, err.Error())
			continue
		}
		if _, err := tx.Exec(ctx, `
			UPDATE webhook_deliveries
			SET status='processed',
				attempt_count = attempt_count + 1,
				last_error = '',
				next_retry_at = now(),
				processed_at = now()
			WHERE provider=$1 AND event_key=$2
		`, pv, p.eventKey); err != nil {
			continue
		}
		processed++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return processed, nil
}
