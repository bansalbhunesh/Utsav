package guestrepo

import (
	"context"
	"encoding/csv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Guest struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Phone        string   `json:"phone"`
	Email        any      `json:"email"`
	Relationship string   `json:"relationship"`
	Side         string   `json:"side"`
	Tags         []string `json:"tags"`
	GroupID      any      `json:"group_id"`
}

type ImportError struct {
	Line  int    `json:"line"`
	Error string `json:"error"`
}

type ImportResult struct {
	Imported int           `json:"imported"`
	Errors   []ImportError `json:"errors"`
}

type GuestInput struct {
	Name         string
	Phone        string
	Email        string
	Relationship string
	Side         string
	Tags         []string
	GroupID      *string
}

type Repository interface {
	ListGuests(ctx context.Context, eventID uuid.UUID) ([]Guest, error)
	UpsertGuest(ctx context.Context, eventID uuid.UUID, input GuestInput) (string, error)
	ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, rawCSV string) (*ImportResult, error)
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) ListGuests(ctx context.Context, eventID uuid.UUID) ([]Guest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, phone, email, relationship, side, tags, group_id
		FROM guests WHERE event_id=$1 ORDER BY name`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Guest, 0)
	for rows.Next() {
		var g Guest
		_ = rows.Scan(&g.ID, &g.Name, &g.Phone, &g.Email, &g.Relationship, &g.Side, &g.Tags, &g.GroupID)
		out = append(out, g)
	}
	return out, nil
}

func (r *PGRepository) UpsertGuest(ctx context.Context, eventID uuid.UUID, input GuestInput) (string, error) {
	var gid any
	if input.GroupID != nil && *input.GroupID != "" {
		if g, err := uuid.Parse(*input.GroupID); err == nil {
			gid = g
		}
	}
	tags := any([]string{})
	if input.Tags != nil {
		tags = input.Tags
	}
	var guestID string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO guests (event_id, group_id, name, phone, email, relationship, side, tags)
		VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),$8::text[])
		ON CONFLICT (event_id, phone) DO UPDATE SET name=EXCLUDED.name, email=EXCLUDED.email,
			relationship=EXCLUDED.relationship, side=EXCLUDED.side, tags=EXCLUDED.tags, updated_at=now()
		RETURNING id`,
		eventID, gid, input.Name, input.Phone, input.Email, input.Relationship, input.Side, tags,
	).Scan(&guestID)
	if err != nil {
		return "", err
	}
	return guestID, nil
}

func (r *PGRepository) ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, rawCSV string) (*ImportResult, error) {
	raw := strings.TrimSpace(rawCSV)
	if strings.HasPrefix(raw, "\ufeff") {
		raw = strings.TrimPrefix(raw, "\ufeff")
	}
	reader := csv.NewReader(strings.NewReader(raw))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &ImportResult{Imported: 0, Errors: []ImportError{{Line: 1, Error: "empty_csv"}}}, nil
	}

	start := 0
	nameIdx, phoneIdx, emailIdx, relIdx, sideIdx := 0, 1, -1, -1, -1
	first := records[0]
	joined := strings.ToLower(strings.Join(first, "|"))
	if strings.Contains(joined, "phone") && strings.Contains(joined, "name") {
		start = 1
		nameIdx, phoneIdx, emailIdx, relIdx, sideIdx = -1, -1, -1, -1, -1
		for i, h := range first {
			hl := strings.ToLower(strings.TrimSpace(h))
			switch hl {
			case "name", "guest", "guest_name":
				nameIdx = i
			case "phone", "mobile", "contact":
				phoneIdx = i
			case "email", "e-mail":
				emailIdx = i
			case "relationship", "relation":
				relIdx = i
			case "side", "bride_groom", "family_side":
				sideIdx = i
			}
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result := &ImportResult{Imported: 0, Errors: []ImportError{}}

	for i := start; i < len(records); i++ {
		row := records[i]
		lineNo := i + 1

		maxIdx := phoneIdx
		if nameIdx > maxIdx {
			maxIdx = nameIdx
		}
		if emailIdx > maxIdx {
			maxIdx = emailIdx
		}
		if relIdx > maxIdx {
			maxIdx = relIdx
		}
		if sideIdx > maxIdx {
			maxIdx = sideIdx
		}
		if len(row) <= maxIdx || phoneIdx < 0 || nameIdx < 0 {
			result.Errors = append(result.Errors, ImportError{Line: lineNo, Error: "too_few_columns"})
			continue
		}

		name := strings.TrimSpace(row[nameIdx])
		phone := strings.TrimSpace(row[phoneIdx])
		if phone == "" {
			result.Errors = append(result.Errors, ImportError{Line: lineNo, Error: "missing_phone"})
			continue
		}
		if name == "" {
			name = phone
		}

		email := ""
		if emailIdx >= 0 && emailIdx < len(row) {
			email = strings.TrimSpace(row[emailIdx])
		}
		rel := ""
		if relIdx >= 0 && relIdx < len(row) {
			rel = strings.TrimSpace(row[relIdx])
		}
		side := ""
		if sideIdx >= 0 && sideIdx < len(row) {
			side = strings.TrimSpace(row[sideIdx])
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO guests (event_id, group_id, name, phone, email, relationship, side, tags)
			VALUES ($1,NULL,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),$7::text[])
			ON CONFLICT (event_id, phone) DO UPDATE SET name=EXCLUDED.name, email=EXCLUDED.email,
				relationship=EXCLUDED.relationship, side=EXCLUDED.side, updated_at=now()`,
			eventID, name, phone, email, rel, side, []string{})
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Line: lineNo, Error: err.Error()})
			continue
		}
		result.Imported++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
