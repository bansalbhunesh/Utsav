package shagunrepo

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
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
	ListHostShagun(ctx context.Context, eventID uuid.UUID) ([]HostShagunRow, error)
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) LogCashShagun(ctx context.Context, eventID uuid.UUID, input CashShagunInput) error {
	paise := int64(input.AmountINR * 100)
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
	_, err := r.pool.Exec(ctx, `
		INSERT INTO shagun_entries (event_id, guest_id, channel, amount_paise, status, sub_event_id, meta)
		VALUES ($1,$2,'cash',$3,'host_verified',$4,$5::jsonb)`,
		eventID, gid, paise, sid, string(metaJSON))
	return err
}

func (r *PGRepository) ListHostShagun(ctx context.Context, eventID uuid.UUID) ([]HostShagunRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, channel, amount_paise, blessing_note, status, created_at
		FROM shagun_entries WHERE event_id=$1 ORDER BY created_at DESC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]HostShagunRow, 0)
	for rows.Next() {
		var id uuid.UUID
		var row HostShagunRow
		_ = rows.Scan(&id, &row.Channel, &row.AmountPaise, &row.Blessing, &row.Status, &row.CreatedAt)
		row.ID = id.String()
		out = append(out, row)
	}
	return out, nil
}
