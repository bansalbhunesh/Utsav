package billingrepo

import (
	"context"
	"strings"

	"github.com/google/uuid"
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
	ListCheckouts(ctx context.Context, userID uuid.UUID) ([]Checkout, error)
	MarkOrderPaidAndFetch(ctx context.Context, orderID string) (string, any, error)
}

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
		_ = rows.Scan(&id, &c.EventID, &c.Tier, &c.OrderID, &c.Status, &c.Created)
		c.ID = id.String()
		out = append(out, c)
	}
	return out, nil
}

func (r *PGRepository) MarkOrderPaidAndFetch(ctx context.Context, orderID string) (string, any, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE billing_checkouts SET status='paid' WHERE razorpay_order_id=$1`, orderID)
	if err != nil || tag.RowsAffected() == 0 {
		return "", nil, err
	}
	var tier string
	var eventID any
	_ = r.pool.QueryRow(ctx, `
		SELECT tier, event_id FROM billing_checkouts WHERE razorpay_order_id=$1`, orderID).Scan(&tier, &eventID)
	if eventID != nil {
		_, _ = r.pool.Exec(ctx, `UPDATE events SET tier=$1 WHERE id=$2`, tier, eventID)
	}
	return tier, eventID, nil
}
