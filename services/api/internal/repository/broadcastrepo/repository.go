package broadcastrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Broadcast struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Body             string `json:"body"`
	ImageURL         any    `json:"image_url"`
	Audience         string `json:"audience"`
	AnnouncementType string `json:"announcement_type"`
	CreatedAt        any    `json:"created_at"`
}

type CreateInput struct {
	EventID         uuid.UUID
	CreatedByUserID uuid.UUID
	Title           string
	Body            string
	ImageURL        string
	Audience        map[string]any
	Type            string
}

type Repository interface {
	List(ctx context.Context, eventID uuid.UUID) ([]Broadcast, error)
	Create(ctx context.Context, in CreateInput) error
}

type PGRepository struct{ pool *pgxpool.Pool }

func NewPGRepository(pool *pgxpool.Pool) *PGRepository { return &PGRepository{pool: pool} }

func nullable(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.TrimSpace(s)
}

func (r *PGRepository) List(ctx context.Context, eventID uuid.UUID) ([]Broadcast, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, body, image_url, audience, announcement_type, created_at
		FROM broadcasts
		WHERE event_id=$1
		ORDER BY created_at DESC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Broadcast, 0)
	for rows.Next() {
		var id uuid.UUID
		var b Broadcast
		var audience []byte
		if err := rows.Scan(&id, &b.Title, &b.Body, &b.ImageURL, &audience, &b.AnnouncementType, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan broadcast row: %w", err)
		}
		b.ID = id.String()
		b.Audience = string(audience)
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate broadcast rows: %w", err)
	}
	return out, nil
}

func (r *PGRepository) Create(ctx context.Context, in CreateInput) error {
	blob, _ := json.Marshal(in.Audience)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO broadcasts (event_id, title, body, image_url, audience, announcement_type, created_by_user_id)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7)`,
		in.EventID, in.Title, in.Body, nullable(in.ImageURL), blob, in.Type, in.CreatedByUserID)
	return err
}
