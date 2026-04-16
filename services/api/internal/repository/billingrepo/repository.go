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

	tag, err := tx.Exec(ctx, `
		INSERT INTO webhook_deliveries (provider, event_key, payload_hash)
		VALUES ($1,$2,$3)
		ON CONFLICT (provider, event_key) DO NOTHING
	`, provider, eventKey, payloadHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var existing string
		err = tx.QueryRow(ctx, `
			SELECT payload_hash FROM webhook_deliveries
			WHERE provider=$1 AND event_key=$2
			FOR UPDATE
		`, provider, eventKey).Scan(&existing)
		if err != nil {
			return err
		}
		if existing != payloadHash {
			return ErrWebhookDedupePayloadMismatch
		}
	}

	if _, _, err := markOrderPaidInTx(ctx, tx, orderID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
