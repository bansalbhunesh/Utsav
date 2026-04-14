package eventrepo

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateEventInput struct {
	OwnerUserID uuid.UUID
	Slug        string
	Title       string
	EventType   string
	CoupleA     string
	CoupleB     string
	LoveStory   string
	CoverURL    string
	DateStart   *string
	DateEnd     *string
	Privacy     string
	Toggles     map[string]any
	Branding    map[string]any
	HostUPIVPA  string
}

type EventListRow struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	EventType string `json:"event_type"`
	DateStart any    `json:"date_start"`
}

type EventDetail struct {
	ID          string
	Slug        string
	Title       string
	EventType   string
	CoupleA     *string
	CoupleB     *string
	LoveStory   *string
	CoverURL    *string
	DateStart   *string
	DateEnd     *string
	Privacy     string
	Toggles     string
	Branding    string
	HostUPIVPA  *string
	Tier        string
}

type PatchEventInput struct {
	Title      *string
	Privacy    *string
	HostUPIVPA *string
}

type CreateSubEventInput struct {
	Name        string
	SubType     string
	StartsAt    *string
	EndsAt      *string
	VenueLabel  string
	DressCode   string
	Description string
}

type SubEventRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SubType     string `json:"sub_type"`
	StartsAt    any    `json:"starts_at"`
	EndsAt      any    `json:"ends_at"`
	VenueLabel  string `json:"venue_label"`
	DressCode   string `json:"dress_code"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type Repository interface {
	IsSlugAvailable(ctx context.Context, slug string) (bool, error)
	CreateEventWithOwner(ctx context.Context, input CreateEventInput) (string, error)
	ListEvents(ctx context.Context, userID uuid.UUID) ([]EventListRow, error)
	GetEventByID(ctx context.Context, eventID uuid.UUID) (*EventDetail, error)
	PatchEvent(ctx context.Context, eventID uuid.UUID, input PatchEventInput) error
	CreateSubEvent(ctx context.Context, eventID uuid.UUID, input CreateSubEventInput) (string, error)
	ListSubEvents(ctx context.Context, eventID uuid.UUID) ([]SubEventRow, error)
	InviteEventMember(ctx context.Context, eventID uuid.UUID, role, invitedPhone string) error
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) IsSlugAvailable(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM events WHERE slug=$1)`, slug).Scan(&exists)
	return !exists, err
}

func (r *PGRepository) CreateEventWithOwner(ctx context.Context, input CreateEventInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var eid uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO events (
			owner_user_id, slug, title, event_type, couple_name_a, couple_name_b, love_story,
			cover_image_url, date_start, date_end, privacy, toggles, branding, host_upi_vpa
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,coalesce($12::jsonb,'{}'),coalesce($13::jsonb,'{}'),$14)
		RETURNING id`,
		input.OwnerUserID, input.Slug, input.Title, input.EventType, nullStr(input.CoupleA), nullStr(input.CoupleB), nullStr(input.LoveStory),
		nullStr(input.CoverURL), input.DateStart, input.DateEnd, input.Privacy, mustJSONB(input.Toggles), mustJSONB(input.Branding), nullStr(input.HostUPIVPA),
	).Scan(&eid)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, `INSERT INTO event_members (event_id, user_id, role, status) VALUES ($1,$2,'owner','active')`, eid, input.OwnerUserID)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return eid.String(), nil
}

func (r *PGRepository) ListEvents(ctx context.Context, userID uuid.UUID) ([]EventListRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.slug, e.title, e.event_type, e.date_start, e.updated_at
		FROM events e
		WHERE e.owner_user_id=$1
		UNION
		SELECT e.id, e.slug, e.title, e.event_type, e.date_start, e.updated_at
		FROM events e
		JOIN event_members m ON m.event_id=e.id AND m.user_id=$1 AND m.status='active'
		ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EventListRow, 0)
	for rows.Next() {
		var id uuid.UUID
		var row EventListRow
		var updated any
		_ = rows.Scan(&id, &row.Slug, &row.Title, &row.EventType, &row.DateStart, &updated)
		row.ID = id.String()
		out = append(out, row)
	}
	return out, nil
}

func (r *PGRepository) GetEventByID(ctx context.Context, eventID uuid.UUID) (*EventDetail, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, title, event_type, couple_name_a, couple_name_b, love_story, cover_image_url,
			date_start, date_end, privacy, toggles, branding, host_upi_vpa, tier, created_at, updated_at
		FROM events WHERE id=$1`, eventID)
	var e EventDetail
	var id uuid.UUID
	var toggles, branding []byte
	var created, updated any
	if err := row.Scan(&id, &e.Slug, &e.Title, &e.EventType, &e.CoupleA, &e.CoupleB, &e.LoveStory, &e.CoverURL, &e.DateStart, &e.DateEnd, &e.Privacy, &toggles, &branding, &e.HostUPIVPA, &e.Tier, &created, &updated); err != nil {
		return nil, err
	}
	e.ID = id.String()
	e.Toggles = string(toggles)
	e.Branding = string(branding)
	return &e, nil
}

func (r *PGRepository) PatchEvent(ctx context.Context, eventID uuid.UUID, input PatchEventInput) error {
	if input.Title != nil {
		_, _ = r.pool.Exec(ctx, `UPDATE events SET title=$2, updated_at=now() WHERE id=$1`, eventID, *input.Title)
	}
	if input.Privacy != nil {
		_, _ = r.pool.Exec(ctx, `UPDATE events SET privacy=$2, updated_at=now() WHERE id=$1`, eventID, *input.Privacy)
	}
	if input.HostUPIVPA != nil {
		_, _ = r.pool.Exec(ctx, `UPDATE events SET host_upi_vpa=$2, updated_at=now() WHERE id=$1`, eventID, *input.HostUPIVPA)
	}
	return nil
}

func (r *PGRepository) CreateSubEvent(ctx context.Context, eventID uuid.UUID, input CreateSubEventInput) (string, error) {
	var sid uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO sub_events (event_id, name, sub_type, starts_at, ends_at, venue_label, dress_code, description, sort_order)
		VALUES ($1,$2,$3,CAST($4 AS timestamptz),CAST($5 AS timestamptz),$6,$7,$8,
			(SELECT COALESCE(MAX(sort_order),0)+1 FROM sub_events WHERE event_id=$1))
		RETURNING id`,
		eventID, input.Name, input.SubType, input.StartsAt, input.EndsAt, input.VenueLabel, input.DressCode, input.Description,
	).Scan(&sid)
	if err != nil {
		return "", err
	}
	return sid.String(), nil
}

func (r *PGRepository) ListSubEvents(ctx context.Context, eventID uuid.UUID) ([]SubEventRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, sub_type, starts_at, ends_at, venue_label, dress_code, description, sort_order
		FROM sub_events WHERE event_id=$1 ORDER BY sort_order, starts_at`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SubEventRow, 0)
	for rows.Next() {
		var id uuid.UUID
		var row SubEventRow
		_ = rows.Scan(&id, &row.Name, &row.SubType, &row.StartsAt, &row.EndsAt, &row.VenueLabel, &row.DressCode, &row.Description, &row.SortOrder)
		row.ID = id.String()
		out = append(out, row)
	}
	return out, nil
}

func (r *PGRepository) InviteEventMember(ctx context.Context, eventID uuid.UUID, role, invitedPhone string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO event_members (event_id, role, invited_phone, status)
		VALUES ($1,$2,$3,'invited')`, eventID, role, invitedPhone)
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func mustJSONB(m map[string]any) []byte {
	if m == nil {
		return []byte("{}")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return []byte("{}")
	}
	return b
}
