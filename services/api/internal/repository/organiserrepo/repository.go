package organiserrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileInput struct {
	UserID      uuid.UUID
	CompanyName string
	Description string
	LogoURL     string
}

type Profile struct {
	ID          string `json:"id"`
	CompanyName string `json:"company_name"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
	Verified    bool   `json:"verified"`
}

type Event struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	DateStart any    `json:"date_start"`
}

type Client struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ContactEmail any    `json:"contact_email"`
	ContactPhone any    `json:"contact_phone"`
	Notes        any    `json:"notes"`
	CreatedAt    any    `json:"created_at"`
}

type ClientInput struct {
	Name         string
	ContactEmail string
	ContactPhone string
	Notes        string
}

type Repository interface {
	UpsertProfile(ctx context.Context, in ProfileInput) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error)
	FindOrganiserIDByUser(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	ListOrganiserEvents(ctx context.Context, organiserID uuid.UUID) ([]Event, error)
	ListClients(ctx context.Context, organiserID uuid.UUID) ([]Client, error)
	CreateClient(ctx context.Context, organiserID uuid.UUID, input ClientInput) (string, error)
	UpdateClient(ctx context.Context, organiserID, clientID uuid.UUID, input ClientInput) (bool, error)
	ClientExistsForOrganiser(ctx context.Context, organiserID, clientID uuid.UUID) (bool, error)
	LinkClientEvent(ctx context.Context, clientID, eventID uuid.UUID) error
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository { return &PGRepository{pool: pool} }

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (r *PGRepository) UpsertProfile(ctx context.Context, in ProfileInput) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO organiser_profiles (user_id, company_name, description, logo_url)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id) DO UPDATE SET company_name=EXCLUDED.company_name, description=EXCLUDED.description, logo_url=EXCLUDED.logo_url`,
		in.UserID, in.CompanyName, in.Description, nullable(in.LogoURL))
	return err
}

func (r *PGRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*Profile, error) {
	var id uuid.UUID
	p := &Profile{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, company_name, description, logo_url, verified FROM organiser_profiles WHERE user_id=$1`, userID).
		Scan(&id, &p.CompanyName, &p.Description, &p.LogoURL, &p.Verified)
	if err != nil {
		return nil, err
	}
	p.ID = id.String()
	return p, nil
}

func (r *PGRepository) FindOrganiserIDByUser(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var oid uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM organiser_profiles WHERE user_id=$1`, userID).Scan(&oid)
	return oid, err
}

func (r *PGRepository) ListOrganiserEvents(ctx context.Context, organiserID uuid.UUID) ([]Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.slug, e.title, e.date_start
		FROM events e
		JOIN organiser_client_events oce ON oce.event_id=e.id
		JOIN organiser_clients oc ON oc.id=oce.organiser_client_id
		WHERE oc.organiser_id=$1`, organiserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Event, 0)
	for rows.Next() {
		var id uuid.UUID
		var e Event
		_ = rows.Scan(&id, &e.Slug, &e.Title, &e.DateStart)
		e.ID = id.String()
		out = append(out, e)
	}
	return out, nil
}

func (r *PGRepository) ListClients(ctx context.Context, organiserID uuid.UUID) ([]Client, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, contact_email, contact_phone, notes, created_at
		FROM organiser_clients
		WHERE organiser_id=$1
		ORDER BY created_at DESC`, organiserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Client, 0)
	for rows.Next() {
		var id uuid.UUID
		var c Client
		_ = rows.Scan(&id, &c.Name, &c.ContactEmail, &c.ContactPhone, &c.Notes, &c.CreatedAt)
		c.ID = id.String()
		out = append(out, c)
	}
	return out, nil
}

func (r *PGRepository) CreateClient(ctx context.Context, organiserID uuid.UUID, input ClientInput) (string, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO organiser_clients (organiser_id, name, contact_email, contact_phone, notes)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		organiserID, input.Name, nullable(input.ContactEmail), nullable(input.ContactPhone), nullable(input.Notes)).
		Scan(&id)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (r *PGRepository) UpdateClient(ctx context.Context, organiserID, clientID uuid.UUID, input ClientInput) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE organiser_clients
		SET name=$1, contact_email=$2, contact_phone=$3, notes=$4
		WHERE id=$5 AND organiser_id=$6`,
		input.Name, nullable(input.ContactEmail), nullable(input.ContactPhone), nullable(input.Notes), clientID, organiserID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *PGRepository) ClientExistsForOrganiser(ctx context.Context, organiserID, clientID uuid.UUID) (bool, error) {
	var exists int
	err := r.pool.QueryRow(ctx, `
		SELECT 1 FROM organiser_clients WHERE id=$1 AND organiser_id=$2`, clientID, organiserID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *PGRepository) LinkClientEvent(ctx context.Context, clientID, eventID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO organiser_client_events (organiser_client_id, event_id)
		VALUES ($1,$2)
		ON CONFLICT DO NOTHING`, clientID, eventID)
	return err
}
