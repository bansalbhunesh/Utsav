package vendorrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Vendor struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	AdvancePaise int64  `json:"advance_paise"`
	TotalPaise   int64  `json:"total_paise"`
	Notes        string `json:"notes"`
}

type CreateInput struct {
	Name         string
	Category     string
	Phone        string
	Email        string
	AdvancePaise int64
	TotalPaise   int64
	Notes        string
}

type Repository interface {
	ListVendors(ctx context.Context, eventID uuid.UUID) ([]Vendor, error)
	CreateVendor(ctx context.Context, eventID uuid.UUID, input CreateInput) (string, error)
	DeleteVendor(ctx context.Context, eventID, vendorID uuid.UUID) error
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) ListVendors(ctx context.Context, eventID uuid.UUID) ([]Vendor, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, category, phone, email, advance_paise, total_paise, notes, created_at
		FROM event_vendors WHERE event_id=$1 ORDER BY created_at DESC`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Vendor, 0)
	for rows.Next() {
		var created any
		var v Vendor
		_ = rows.Scan(&v.ID, &v.Name, &v.Category, &v.Phone, &v.Email, &v.AdvancePaise, &v.TotalPaise, &v.Notes, &created)
		out = append(out, v)
	}
	return out, nil
}

func (r *PGRepository) CreateVendor(ctx context.Context, eventID uuid.UUID, input CreateInput) (string, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		INSERT INTO event_vendors (event_id, name, category, phone, email, advance_paise, total_paise, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id`,
		eventID, input.Name, input.Category, input.Phone, input.Email, input.AdvancePaise, input.TotalPaise, input.Notes,
	).Scan(&id)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (r *PGRepository) DeleteVendor(ctx context.Context, eventID, vendorID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM event_vendors WHERE id=$1 AND event_id=$2`, vendorID, eventID)
	return err
}
