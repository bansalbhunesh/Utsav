package memorybookrepo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventSnapshot struct {
	Slug      string
	Title     string
	DateStart any
	DateEnd   any
	Tier      string
}

type Highlights struct {
	GuestsTotal     int64
	RSVPTotal       int64
	RSVPYes         int64
	ShagunCount     int64
	ShagunPaise     int64
	GalleryApproved int64
	BroadcastsCount int64
}

type Repository interface {
	GetEventSnapshot(ctx context.Context, eventID uuid.UUID) (*EventSnapshot, error)
	GetHighlights(ctx context.Context, eventID uuid.UUID) (*Highlights, error)
	UpsertMemoryBook(ctx context.Context, eventID uuid.UUID, slug string, payload map[string]any) error
	GetMemoryBookBySlug(ctx context.Context, slug string) (uuid.UUID, map[string]any, error)
	GetEventTier(ctx context.Context, eventID uuid.UUID) (string, error)
}

type PGRepository struct{ pool *pgxpool.Pool }

func NewPGRepository(pool *pgxpool.Pool) *PGRepository { return &PGRepository{pool: pool} }

func (r *PGRepository) GetEventSnapshot(ctx context.Context, eventID uuid.UUID) (*EventSnapshot, error) {
	var e EventSnapshot
	err := r.pool.QueryRow(ctx, `
		SELECT slug, title, date_start, date_end, tier FROM events WHERE id=$1`, eventID).
		Scan(&e.Slug, &e.Title, &e.DateStart, &e.DateEnd, &e.Tier)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *PGRepository) GetHighlights(ctx context.Context, eventID uuid.UUID) (*Highlights, error) {
	h := &Highlights{}
	_ = r.pool.QueryRow(ctx, `SELECT count(*) FROM guests WHERE event_id=$1`, eventID).Scan(&h.GuestsTotal)
	_ = r.pool.QueryRow(ctx, `SELECT count(*) FROM rsvp_responses WHERE event_id=$1`, eventID).Scan(&h.RSVPTotal)
	_ = r.pool.QueryRow(ctx, `SELECT count(*) FROM rsvp_responses WHERE event_id=$1 AND status='yes'`, eventID).Scan(&h.RSVPYes)
	_ = r.pool.QueryRow(ctx, `SELECT count(*), coalesce(sum(amount_paise),0) FROM shagun_entries WHERE event_id=$1`, eventID).
		Scan(&h.ShagunCount, &h.ShagunPaise)
	_ = r.pool.QueryRow(ctx, `SELECT count(*) FROM gallery_assets WHERE event_id=$1 AND status='approved'`, eventID).Scan(&h.GalleryApproved)
	_ = r.pool.QueryRow(ctx, `SELECT count(*) FROM broadcasts WHERE event_id=$1`, eventID).Scan(&h.BroadcastsCount)
	return h, nil
}

func (r *PGRepository) UpsertMemoryBook(ctx context.Context, eventID uuid.UUID, slug string, payload map[string]any) error {
	blob, _ := json.Marshal(payload)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO memory_books (event_id, slug, payload)
		VALUES ($1,$2,$3::jsonb)
		ON CONFLICT (slug) DO UPDATE SET payload=EXCLUDED.payload, generated_at=now()`,
		eventID, slug, blob)
	return err
}

func (r *PGRepository) GetMemoryBookBySlug(ctx context.Context, slug string) (uuid.UUID, map[string]any, error) {
	var payload []byte
	var eid uuid.UUID
	if err := r.pool.QueryRow(ctx, `
		SELECT mb.event_id, mb.payload
		FROM memory_books mb
		JOIN events e ON e.id = mb.event_id
		WHERE mb.slug=$1 AND lower(trim(coalesce(e.privacy, ''))) = 'public'`, slug).Scan(&eid, &payload); err != nil {
		return uuid.Nil, nil, err
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return uuid.Nil, nil, err
	}
	return eid, decoded, nil
}

func (r *PGRepository) GetEventTier(ctx context.Context, eventID uuid.UUID) (string, error) {
	var tier string
	err := r.pool.QueryRow(ctx, `SELECT tier FROM events WHERE id=$1`, eventID).Scan(&tier)
	return tier, err
}

func BuildPayload(eventID uuid.UUID, snap *EventSnapshot, h *Highlights) map[string]any {
	return map[string]any{
		"version":      1,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"event": map[string]any{
			"id":         eventID.String(),
			"slug":       snap.Slug,
			"title":      snap.Title,
			"date_start": snap.DateStart,
			"date_end":   snap.DateEnd,
			"tier":       snap.Tier,
		},
		"highlights": map[string]any{
			"guest_count":        h.GuestsTotal,
			"rsvp_count":         h.RSVPTotal,
			"rsvp_yes_count":     h.RSVPYes,
			"shagun_count":       h.ShagunCount,
			"shagun_total_paise": h.ShagunPaise,
			"gallery_assets":     h.GalleryApproved,
			"broadcasts":         h.BroadcastsCount,
		},
	}
}
