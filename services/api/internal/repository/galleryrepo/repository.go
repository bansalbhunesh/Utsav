package galleryrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateAssetInput struct {
	EventID        uuid.UUID
	UploaderUserID uuid.UUID
	Section        string
	ObjectKey      string
	SubEventID     string
	Status         string
	MimeType       string
	Bytes          int64
}

type Asset struct {
	ID        string `json:"id"`
	Section   string `json:"section"`
	ObjectKey string `json:"object_key"`
	Status    string `json:"status"`
	MimeType  any    `json:"mime_type"`
	Bytes     any    `json:"bytes"`
	CreatedAt any    `json:"created_at"`
}

type Repository interface {
	CreateAsset(ctx context.Context, in CreateAssetInput) error
	ListAssets(ctx context.Context, eventID uuid.UUID, status string) ([]Asset, error)
	UpdateAssetStatus(ctx context.Context, eventID, assetID uuid.UUID, status string) (bool, error)
}

type PGRepository struct{ pool *pgxpool.Pool }

func NewPGRepository(pool *pgxpool.Pool) *PGRepository { return &PGRepository{pool: pool} }

func nullable(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.TrimSpace(s)
}

func (r *PGRepository) CreateAsset(ctx context.Context, in CreateAssetInput) error {
	var sid any
	if strings.TrimSpace(in.SubEventID) != "" {
		if u, err := uuid.Parse(in.SubEventID); err == nil {
			sid = u
		}
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO gallery_assets (event_id, section, object_key, uploader_user_id, sub_event_id, status, mime_type, bytes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		in.EventID, in.Section, in.ObjectKey, in.UploaderUserID, sid, in.Status, nullable(in.MimeType), in.Bytes)
	return err
}

func (r *PGRepository) ListAssets(ctx context.Context, eventID uuid.UUID, status string) ([]Asset, error) {
	q := `
		SELECT id, section, object_key, status, mime_type, bytes, created_at
		FROM gallery_assets
		WHERE event_id=$1`
	args := []any{eventID}
	if strings.TrimSpace(status) != "" {
		q += ` AND status=$2`
		args = append(args, strings.TrimSpace(strings.ToLower(status)))
	}
	q += ` ORDER BY created_at DESC LIMIT 200`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Asset, 0)
	for rows.Next() {
		var id uuid.UUID
		var a Asset
		if err := rows.Scan(&id, &a.Section, &a.ObjectKey, &a.Status, &a.MimeType, &a.Bytes, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan gallery asset row: %w", err)
		}
		a.ID = id.String()
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gallery asset rows: %w", err)
	}
	return out, nil
}

func (r *PGRepository) UpdateAssetStatus(ctx context.Context, eventID, assetID uuid.UUID, status string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE gallery_assets SET status=$1
		WHERE id=$2 AND event_id=$3`, status, assetID, eventID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
