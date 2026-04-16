package shagunrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CashShagunInput struct {
	GuestID    *string
	GuestPhone string
	AmountINR  float64
	SubEventID *string
	Notes      string
}

type HostShagunRow struct {
	ID          string `json:"id"`
	Channel     string `json:"channel"`
	AmountPaise any    `json:"amount_paise"`
	Blessing    string `json:"blessing_note"`
	Status      string `json:"status"`
	CreatedAt   any    `json:"created_at"`
}

type Repository interface {
	LogCashShagun(ctx context.Context, eventID uuid.UUID, input CashShagunInput) error
	LogCashShagunTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, input CashShagunInput) error
	ListHostShagun(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]HostShagunRow, error)
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func insertCashShagun(ctx context.Context, db execer, eventID uuid.UUID, input CashShagunInput) error {
	paise := int64(math.Round(input.AmountINR * 100))
	var gid any
	if input.GuestID != nil {
		if g, err := uuid.Parse(*input.GuestID); err == nil {
			gid = g
		}
	}
	var sid any
	if input.SubEventID != nil {
		if s2, err := uuid.Parse(*input.SubEventID); err == nil {
			sid = s2
		}
	}
	meta := map[string]any{"notes": input.Notes, "guest_phone": input.GuestPhone}
	metaJSON, _ := json.Marshal(meta)
	_, err := db.Exec(ctx, `
		INSERT INTO shagun_entries (event_id, guest_id, channel, amount_paise, status, sub_event_id, meta)
		VALUES ($1,$2,'cash',$3,'host_verified',$4,$5::jsonb)`,
		eventID, gid, paise, sid, string(metaJSON))
	return err
}

func (r *PGRepository) LogCashShagun(ctx context.Context, eventID uuid.UUID, input CashShagunInput) error {
	return insertCashShagun(ctx, r.pool, eventID, input)
}

func (r *PGRepository) LogCashShagunTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, input CashShagunInput) error {
	return insertCashShagun(ctx, tx, eventID, input)
}

func (r *PGRepository) ListHostShagun(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]HostShagunRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, channel, amount_paise, blessing_note, status, created_at
		FROM shagun_entries WHERE event_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, eventID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]HostShagunRow, 0)
	for rows.Next() {
		var id uuid.UUID
		var row HostShagunRow
		if err := rows.Scan(&id, &row.Channel, &row.AmountPaise, &row.Blessing, &row.Status, &row.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan shagun row: %w", err)
		}
		row.ID = id.String()
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate shagun rows: %w", err)
	}
	return out, nil
}
