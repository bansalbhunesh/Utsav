package rsvprepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OTPChallenge struct {
	ID        uuid.UUID
	CodeHash  string
	ExpiresAt time.Time
}

type RSVPItem struct {
	SubEventID          uuid.UUID
	Status              string
	MealPref            string
	Dietary             string
	AccommodationNeeded bool
	TravelMode          string
	PlusOneNames        string
}

type HostRSVPRow struct {
	ID                  string `json:"id"`
	GuestPhone          string `json:"guest_phone"`
	SubEventID          string `json:"sub_event_id"`
	Status              string `json:"status"`
	MealPref            string `json:"meal_pref"`
	Dietary             string `json:"dietary"`
	AccommodationNeeded bool   `json:"accommodation_needed"`
	TravelMode          string `json:"travel_mode"`
	PlusOneNames        string `json:"plus_one_names"`
	UpdatedAt           any    `json:"updated_at"`
}

type Repository interface {
	FindEventIDBySlug(ctx context.Context, slug string) (uuid.UUID, error)
	DeleteRSVPOTPChallenges(ctx context.Context, eventID uuid.UUID, phone string) error
	InsertRSVPOTPChallenge(ctx context.Context, eventID uuid.UUID, phone, codeHash string) error
	GetLatestRSVPOTPChallenge(ctx context.Context, eventID uuid.UUID, phone string) (*OTPChallenge, error)
	DeleteRSVPOTPChallengeByID(ctx context.Context, id uuid.UUID) error
	UpsertRSVPResponses(ctx context.Context, eventID uuid.UUID, phone string, items []RSVPItem) error
	ListHostRSVPs(ctx context.Context, eventID uuid.UUID) ([]HostRSVPRow, error)
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) FindEventIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	var eid uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM events WHERE slug=$1`, slug).Scan(&eid)
	return eid, err
}

func (r *PGRepository) DeleteRSVPOTPChallenges(ctx context.Context, eventID uuid.UUID, phone string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM rsvp_otp_challenges WHERE event_id=$1 AND phone=$2`, eventID, phone)
	return err
}

func (r *PGRepository) InsertRSVPOTPChallenge(ctx context.Context, eventID uuid.UUID, phone, codeHash string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rsvp_otp_challenges (event_id, phone, code_hash, expires_at)
		VALUES ($1,$2,$3, now() + interval '10 minutes')`, eventID, phone, codeHash)
	return err
}

func (r *PGRepository) GetLatestRSVPOTPChallenge(ctx context.Context, eventID uuid.UUID, phone string) (*OTPChallenge, error) {
	var ch OTPChallenge
	err := r.pool.QueryRow(ctx, `
		SELECT id, code_hash, expires_at FROM rsvp_otp_challenges
		WHERE event_id=$1 AND phone=$2 ORDER BY created_at DESC LIMIT 1`, eventID, phone).
		Scan(&ch.ID, &ch.CodeHash, &ch.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *PGRepository) DeleteRSVPOTPChallengeByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM rsvp_otp_challenges WHERE id=$1`, id)
	return err
}

func (r *PGRepository) UpsertRSVPResponses(ctx context.Context, eventID uuid.UUID, phone string, items []RSVPItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, it := range items {
		_, err = tx.Exec(ctx, `
			INSERT INTO rsvp_responses (
				event_id, guest_phone, sub_event_id, status, meal_pref, dietary, accommodation_needed, travel_mode, plus_one_names
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (event_id, guest_phone, sub_event_id) DO UPDATE SET
				status=EXCLUDED.status, meal_pref=EXCLUDED.meal_pref, dietary=EXCLUDED.dietary,
				accommodation_needed=EXCLUDED.accommodation_needed, travel_mode=EXCLUDED.travel_mode,
				plus_one_names=EXCLUDED.plus_one_names, updated_at=now()`,
			eventID, phone, it.SubEventID, it.Status, it.MealPref, it.Dietary, it.AccommodationNeeded, it.TravelMode, it.PlusOneNames)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PGRepository) ListHostRSVPs(ctx context.Context, eventID uuid.UUID) ([]HostRSVPRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, guest_phone, sub_event_id, status, meal_pref, dietary, accommodation_needed, travel_mode, plus_one_names, updated_at
		FROM rsvp_responses WHERE event_id=$1`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]HostRSVPRow, 0)
	for rows.Next() {
		var id, sub uuid.UUID
		var row HostRSVPRow
		_ = rows.Scan(
			&id,
			&row.GuestPhone,
			&sub,
			&row.Status,
			&row.MealPref,
			&row.Dietary,
			&row.AccommodationNeeded,
			&row.TravelMode,
			&row.PlusOneNames,
			&row.UpdatedAt,
		)
		row.ID = id.String()
		row.SubEventID = sub.String()
		out = append(out, row)
	}
	return out, nil
}

func IsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
