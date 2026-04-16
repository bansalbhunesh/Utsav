package rsvprepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrIdempotencyConflict is returned when the same idempotency key was used with a different request body.
var ErrIdempotencyConflict = errors.New("rsvprepo: idempotency fingerprint mismatch")

// ErrInvalidSubEvents is returned when one or more sub_event ids are not part of the event.
var ErrInvalidSubEvents = errors.New("rsvprepo: sub-events not part of event")

type OTPChallenge struct {
	ID        uuid.UUID
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
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
	IncrementRSVPOTPAttempts(ctx context.Context, id uuid.UUID) error
	ConsumeRSVPOTPChallengeByID(ctx context.Context, id uuid.UUID) (bool, error)
	DeleteRSVPOTPChallengeByID(ctx context.Context, id uuid.UUID) error
	UpsertRSVPResponses(ctx context.Context, eventID uuid.UUID, phone string, items []RSVPItem) error
	UpsertRSVPResponsesIdempotent(ctx context.Context, scope, idempotencyKey, fingerprint string, eventID uuid.UUID, phone string, items []RSVPItem) error
	ListHostRSVPs(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]HostRSVPRow, error)
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
		VALUES ($1,$2,$3, now() + interval '5 minutes')`, eventID, phone, codeHash)
	return err
}

func (r *PGRepository) GetLatestRSVPOTPChallenge(ctx context.Context, eventID uuid.UUID, phone string) (*OTPChallenge, error) {
	var ch OTPChallenge
	err := r.pool.QueryRow(ctx, `
		SELECT id, code_hash, expires_at, attempts FROM rsvp_otp_challenges
		WHERE event_id=$1 AND phone=$2 ORDER BY created_at DESC LIMIT 1`, eventID, phone).
		Scan(&ch.ID, &ch.CodeHash, &ch.ExpiresAt, &ch.Attempts)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *PGRepository) DeleteRSVPOTPChallengeByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM rsvp_otp_challenges WHERE id=$1`, id)
	return err
}

func (r *PGRepository) ConsumeRSVPOTPChallengeByID(ctx context.Context, id uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM rsvp_otp_challenges WHERE id=$1`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *PGRepository) IncrementRSVPOTPAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE rsvp_otp_challenges SET attempts=attempts+1 WHERE id=$1`, id)
	return err
}

const upsertRSVPResponseSQL = `
		INSERT INTO rsvp_responses (
			event_id, guest_phone, sub_event_id, status, meal_pref, dietary, accommodation_needed, travel_mode, plus_one_names
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (event_id, guest_phone, sub_event_id) DO UPDATE SET
			status=EXCLUDED.status, meal_pref=EXCLUDED.meal_pref, dietary=EXCLUDED.dietary,
			accommodation_needed=EXCLUDED.accommodation_needed, travel_mode=EXCLUDED.travel_mode,
			plus_one_names=EXCLUDED.plus_one_names, updated_at=now()`

func upsertRSVPBatchTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, phone string, items []RSVPItem) error {
	if len(items) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, it := range items {
		batch.Queue(upsertRSVPResponseSQL, eventID, phone, it.SubEventID, it.Status, it.MealPref, it.Dietary, it.AccommodationNeeded, it.TravelMode, it.PlusOneNames)
	}
	br := tx.SendBatch(ctx, batch)
	defer br.Close()
	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func dedupeSubEventIDs(items []RSVPItem) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(items))
	out := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		if _, ok := seen[it.SubEventID]; ok {
			continue
		}
		seen[it.SubEventID] = struct{}{}
		out = append(out, it.SubEventID)
	}
	return out
}

func countSubEventsForEventTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, subEventIDs []uuid.UUID) (int, error) {
	if len(subEventIDs) == 0 {
		return 0, nil
	}
	var n int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM sub_events
		WHERE event_id=$1 AND id = ANY($2::uuid[])`, eventID, subEventIDs).Scan(&n)
	return n, err
}

func (r *PGRepository) UpsertRSVPResponses(ctx context.Context, eventID uuid.UUID, phone string, items []RSVPItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if len(items) == 0 {
		return tx.Commit(ctx)
	}
	if err := upsertRSVPBatchTx(ctx, tx, eventID, phone, items); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// UpsertRSVPResponsesIdempotent reserves the idempotency key and applies RSVP upserts in a single transaction.
// If the key already exists with the same fingerprint, it commits without re-writing RSVP rows (replay).
func (r *PGRepository) UpsertRSVPResponsesIdempotent(ctx context.Context, scope, idempotencyKey, fingerprint string, eventID uuid.UUID, phone string, items []RSVPItem) error {
	if len(items) == 0 {
		return fmt.Errorf("empty rsvp items")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		DELETE FROM idempotency_keys WHERE scope=$1 AND key=$2 AND expires_at < now()
	`, scope, idempotencyKey); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `
		INSERT INTO idempotency_keys (scope, key, fingerprint, expires_at)
		VALUES ($1, $2, $3, now() + interval '24 hours')
		ON CONFLICT (scope, key) DO NOTHING
	`, scope, idempotencyKey, fingerprint)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		var existing string
		err = tx.QueryRow(ctx, `
			SELECT fingerprint FROM idempotency_keys
			WHERE scope=$1 AND key=$2 AND expires_at > now()
			FOR UPDATE
		`, scope, idempotencyKey).Scan(&existing)
		if err != nil {
			return err
		}
		if existing != fingerprint {
			return ErrIdempotencyConflict
		}
		return tx.Commit(ctx)
	}

	uniq := dedupeSubEventIDs(items)
	n, err := countSubEventsForEventTx(ctx, tx, eventID, uniq)
	if err != nil {
		return err
	}
	if n != len(uniq) {
		return ErrInvalidSubEvents
	}
	if err := upsertRSVPBatchTx(ctx, tx, eventID, phone, items); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PGRepository) ListHostRSVPs(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]HostRSVPRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, guest_phone, sub_event_id, status, meal_pref, dietary, accommodation_needed, travel_mode, plus_one_names, updated_at
		FROM rsvp_responses WHERE event_id=$1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3`, eventID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]HostRSVPRow, 0)
	for rows.Next() {
		var id, sub uuid.UUID
		var row HostRSVPRow
		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("scan host rsvp row: %w", err)
		}
		row.ID = id.String()
		row.SubEventID = sub.String()
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate host rsvp rows: %w", err)
	}
	return out, nil
}

func IsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
