package publicrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PublicEvent struct {
	ID          uuid.UUID
	Slug        string
	Title       string
	EventType   string
	Privacy     string
	Toggles     []byte
	Branding    []byte
	CoupleNameA any
	CoupleNameB any
	LoveStory   any
	CoverImage  any
	DateStart   any
	DateEnd     any
}

type PublicSubEvent struct {
	ID          uuid.UUID
	Name        string
	SubType     string
	StartsAt    any
	EndsAt      any
	VenueLabel  string
	DressCode   string
	Description string
	SortOrder   int
}

type PublicBroadcast struct {
	ID        uuid.UUID
	Title     string
	Body      string
	ImageURL  string
	Type      string
	CreatedAt any
}

type PublicGalleryAsset struct {
	ID        uuid.UUID
	Section   string
	ObjectKey string
	CreatedAt any
}

type UPIContext struct {
	EventID uuid.UUID
	VPA     string
	Title   string
}

type GuestShagunReportInput struct {
	EventID      uuid.UUID
	AmountPaise  int64
	BlessingNote string
	SubEventID   *string
	GuestPhone   string
}

type Repository interface {
	GetEventBySlug(ctx context.Context, slug string) (*PublicEvent, error)
	GetSlugByEventID(ctx context.Context, eventID uuid.UUID) (string, error)
	ResolveEventIDBySlug(ctx context.Context, slug string) (uuid.UUID, error)
	SubEventBelongsToEvent(ctx context.Context, eventID, subEventID uuid.UUID) (bool, error)
	ListSubEvents(ctx context.Context, eventID uuid.UUID) ([]PublicSubEvent, error)
	ListBroadcasts(ctx context.Context, eventID uuid.UUID) ([]PublicBroadcast, error)
	ListApprovedGallery(ctx context.Context, eventID uuid.UUID) ([]PublicGalleryAsset, error)
	GetUPIContextBySlug(ctx context.Context, slug string) (*UPIContext, error)
	InsertGuestShagunReport(ctx context.Context, in GuestShagunReportInput) error
}

type PGRepository struct {
	write *pgxpool.Pool
	read  *pgxpool.Pool
}

// NewPGRepository uses read for SELECT paths when non-nil (replica); writes use write (primary).
func NewPGRepository(write, read *pgxpool.Pool) *PGRepository {
	if read == nil {
		read = write
	}
	return &PGRepository{write: write, read: read}
}

func (r *PGRepository) GetSlugByEventID(ctx context.Context, eventID uuid.UUID) (string, error) {
	var slug string
	// Primary: used for cache invalidation so slug matches what was just written.
	err := r.write.QueryRow(ctx, `SELECT slug FROM events WHERE id=$1`, eventID).Scan(&slug)
	return slug, err
}

func (r *PGRepository) ResolveEventIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.read.QueryRow(ctx, `SELECT id FROM events WHERE slug=$1`, slug).Scan(&id)
	return id, err
}

func (r *PGRepository) SubEventBelongsToEvent(ctx context.Context, eventID, subEventID uuid.UUID) (bool, error) {
	var ok bool
	err := r.read.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM sub_events WHERE id=$1 AND event_id=$2)`,
		subEventID, eventID).Scan(&ok)
	return ok, err
}

func (r *PGRepository) GetEventBySlug(ctx context.Context, slug string) (*PublicEvent, error) {
	var e PublicEvent
	err := r.read.QueryRow(ctx, `
		SELECT id, title, event_type, privacy, toggles, branding, couple_name_a, couple_name_b, love_story,
			cover_image_url, date_start, date_end
		FROM events WHERE slug=$1 AND lower(trim(coalesce(privacy, ''))) = 'public'`, slug).
		Scan(&e.ID, &e.Title, &e.EventType, &e.Privacy, &e.Toggles, &e.Branding, &e.CoupleNameA, &e.CoupleNameB, &e.LoveStory, &e.CoverImage, &e.DateStart, &e.DateEnd)
	if err != nil {
		return nil, err
	}
	e.Slug = slug
	return &e, nil
}

func (r *PGRepository) ListSubEvents(ctx context.Context, eventID uuid.UUID) ([]PublicSubEvent, error) {
	rows, err := r.read.Query(ctx, `
		SELECT id, name, sub_type, starts_at, ends_at, venue_label, dress_code, description, sort_order
		FROM sub_events WHERE event_id=$1 ORDER BY sort_order, starts_at`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PublicSubEvent, 0)
	for rows.Next() {
		var s PublicSubEvent
		if err := rows.Scan(&s.ID, &s.Name, &s.SubType, &s.StartsAt, &s.EndsAt, &s.VenueLabel, &s.DressCode, &s.Description, &s.SortOrder); err != nil {
			return nil, fmt.Errorf("scan public sub event row: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public sub event rows: %w", err)
	}
	return out, nil
}

func (r *PGRepository) ListBroadcasts(ctx context.Context, eventID uuid.UUID) ([]PublicBroadcast, error) {
	rows, err := r.read.Query(ctx, `
		SELECT id, title, body, image_url, announcement_type, created_at
		FROM broadcasts WHERE event_id=$1 ORDER BY created_at DESC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PublicBroadcast, 0)
	for rows.Next() {
		var b PublicBroadcast
		if err := rows.Scan(&b.ID, &b.Title, &b.Body, &b.ImageURL, &b.Type, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan public broadcast row: %w", err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public broadcast rows: %w", err)
	}
	return out, nil
}

func (r *PGRepository) ListApprovedGallery(ctx context.Context, eventID uuid.UUID) ([]PublicGalleryAsset, error) {
	rows, err := r.read.Query(ctx, `
		SELECT id, section, object_key, created_at
		FROM gallery_assets
		WHERE event_id=$1 AND status='approved'
		ORDER BY created_at DESC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]PublicGalleryAsset, 0)
	for rows.Next() {
		var a PublicGalleryAsset
		if err := rows.Scan(&a.ID, &a.Section, &a.ObjectKey, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan public gallery row: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public gallery rows: %w", err)
	}
	return out, nil
}

func (r *PGRepository) GetUPIContextBySlug(ctx context.Context, slug string) (*UPIContext, error) {
	var u UPIContext
	err := r.read.QueryRow(ctx, `
		SELECT e.id, COALESCE(e.host_upi_vpa,''), COALESCE(NULLIF(e.title,''), e.slug)
		FROM events e WHERE e.slug=$1`, slug).
		Scan(&u.EventID, &u.VPA, &u.Title)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PGRepository) InsertGuestShagunReport(ctx context.Context, in GuestShagunReportInput) error {
	meta := map[string]any{"guest_phone": in.GuestPhone, "reported_at": time.Now().UTC().Format(time.RFC3339)}
	metaJSON, _ := json.Marshal(meta)
	var sid any
	if in.SubEventID != nil {
		if u, err := uuid.Parse(*in.SubEventID); err == nil {
			sid = u
		}
	}
	_, err := r.write.Exec(ctx, `
		INSERT INTO shagun_entries (event_id, channel, amount_paise, blessing_note, status, sub_event_id, meta)
		VALUES ($1,'upi',$2,$3,'guest_reported',$4,$5::jsonb)`,
		in.EventID, in.AmountPaise, in.BlessingNote, sid, metaJSON)
	return err
}
