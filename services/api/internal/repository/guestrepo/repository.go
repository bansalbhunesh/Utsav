package guestrepo

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Guest struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Phone           string     `json:"phone"`
	Email           any        `json:"email"`
	Relationship    string     `json:"relationship"`
	Side            string     `json:"side"`
	Tags            []string   `json:"tags"`
	GroupID         any        `json:"group_id"`
	RSVPYesCount    int        `json:"-"`
	RSVPTotal       int        `json:"-"`
	SubEventTotal   int        `json:"-"`
	LastRSVPAt      *time.Time `json:"-"`
	ShagunPaise     int64      `json:"-"`
	PriorityScore   int        `json:"priority_score"`
	PriorityTier    string     `json:"priority_tier"`
	PriorityReasons []string   `json:"priority_reasons"`
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
	ListGuests(ctx context.Context, eventID uuid.UUID, limit, offset int, sort, priorityTier string) ([]Guest, error)
	UpsertRelationshipScores(ctx context.Context, eventID uuid.UUID, guests []Guest) error
	RelationshipTierCounts(ctx context.Context, eventID uuid.UUID) (map[string]int, error)
	UpsertGuest(ctx context.Context, eventID uuid.UUID, input GuestInput) (string, error)
	ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, rawCSV string) (*ImportResult, error)
}

type PGRepository struct {
	pool *pgxpool.Pool
}

func NewPGRepository(pool *pgxpool.Pool) *PGRepository {
	return &PGRepository{pool: pool}
}

func (r *PGRepository) ListGuests(ctx context.Context, eventID uuid.UUID, limit, offset int, sort, priorityTier string) ([]Guest, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	orderClause := "g.name ASC"
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "name_desc":
		orderClause = "g.name DESC"
	case "priority_asc":
		orderClause = "priority_score ASC, g.name ASC"
	case "priority_desc":
		orderClause = "priority_score DESC, g.name ASC"
	case "rsvp_desc":
		orderClause = "rsvp_yes_count DESC, g.name ASC"
	case "shagun_desc":
		orderClause = "total_shagun_paise DESC, g.name ASC"
	}
	tier := strings.ToLower(strings.TrimSpace(priorityTier))
	switch tier {
	case "critical", "important", "optional":
	default:
		tier = ""
	}
	rows, err := r.pool.Query(ctx, `
		WITH guest_enriched AS (
			SELECT
				g.id, g.name, g.phone, g.email, g.relationship, g.side, g.tags, g.group_id,
				COALESCE(r.rsvp_yes_count, 0) AS rsvp_yes_count,
				COALESCE(r.rsvp_total_count, 0) AS rsvp_total_count,
				COALESCE(sev.sub_event_total, 0) AS sub_event_total,
				r.latest_rsvp_at AS latest_rsvp_at,
				COALESCE(s.total_shagun_paise, 0) AS total_shagun_paise,
				(
					CASE lower(trim(COALESCE(g.relationship, '')))
						WHEN 'close_family' THEN 45
						WHEN 'immediate_family' THEN 45
						WHEN 'family' THEN 35
						WHEN 'relative' THEN 35
						WHEN 'relatives' THEN 35
						WHEN 'friend' THEN 22
						WHEN 'friends' THEN 22
						WHEN 'colleague' THEN 12
						WHEN 'coworker' THEN 12
						ELSE CASE WHEN trim(COALESCE(g.relationship, '')) <> '' THEN 10 ELSE 5 END
					END
				)
				+ CASE WHEN trim(COALESCE(g.side, '')) <> '' THEN 5 ELSE 0 END
				+ CASE
					WHEN COALESCE(r.rsvp_yes_count, 0) > 0 THEN LEAST(25, 12 + (COALESCE(r.rsvp_yes_count, 0) * 5))
					ELSE 0
				END
				+ CASE
					WHEN COALESCE(s.total_shagun_paise, 0) >= 50000 THEN 20
					WHEN COALESCE(s.total_shagun_paise, 0) >= 15000 THEN 14
					WHEN COALESCE(s.total_shagun_paise, 0) > 0 THEN 8
					ELSE 0
				END AS priority_score
			FROM guests g
			LEFT JOIN LATERAL (
				SELECT
					COUNT(*) FILTER (WHERE rr.status='yes') AS rsvp_yes_count,
					COUNT(*) AS rsvp_total_count,
					MAX(rr.updated_at) AS latest_rsvp_at
				FROM rsvp_responses rr
				WHERE rr.event_id=g.event_id AND rr.guest_phone=g.phone
			) r ON TRUE
			LEFT JOIN LATERAL (
				SELECT COUNT(*) AS sub_event_total
				FROM sub_events sev
				WHERE sev.event_id=g.event_id
			) sev ON TRUE
			LEFT JOIN LATERAL (
				SELECT COALESCE(SUM(se.amount_paise),0) AS total_shagun_paise
				FROM shagun_entries se
				WHERE se.event_id=g.event_id
				  AND (se.guest_id=g.id OR COALESCE(se.meta->>'guest_phone','')=g.phone)
			) s ON TRUE
			WHERE g.event_id=$1
		)
		SELECT
			g.id, g.name, g.phone, g.email, g.relationship, g.side, g.tags, g.group_id,
			g.rsvp_yes_count, g.rsvp_total_count, g.sub_event_total, g.latest_rsvp_at, g.total_shagun_paise
		FROM guest_enriched g
		WHERE
			$2 = '' OR
			($2 = 'critical' AND g.priority_score >= 80) OR
			($2 = 'important' AND g.priority_score >= 50 AND g.priority_score < 80) OR
			($2 = 'optional' AND g.priority_score < 50)
		ORDER BY `+orderClause+`
		LIMIT $3 OFFSET $4`, eventID, tier, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Guest, 0)
	for rows.Next() {
		var g Guest
		if err := rows.Scan(
			&g.ID, &g.Name, &g.Phone, &g.Email, &g.Relationship, &g.Side, &g.Tags, &g.GroupID,
			&g.RSVPYesCount, &g.RSVPTotal, &g.SubEventTotal, &g.LastRSVPAt, &g.ShagunPaise,
		); err != nil {
			return nil, fmt.Errorf("scan guest row: %w", err)
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate guest rows: %w", err)
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

func (r *PGRepository) UpsertRelationshipScores(ctx context.Context, eventID uuid.UUID, guests []Guest) error {
	if len(guests) == 0 {
		return nil
	}
	const q = `
		INSERT INTO guest_relationship_scores
			(event_id, guest_id, priority_score, priority_tier, priority_reasons, source_version, computed_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, 1, $6)
		ON CONFLICT (event_id, guest_id) DO UPDATE SET
			priority_score = EXCLUDED.priority_score,
			priority_tier = EXCLUDED.priority_tier,
			priority_reasons = EXCLUDED.priority_reasons,
			source_version = EXCLUDED.source_version,
			computed_at = EXCLUDED.computed_at
	`
	now := time.Now().UTC()
	for _, g := range guests {
		guestID, err := uuid.Parse(g.ID)
		if err != nil {
			continue
		}
		reasonsRaw, _ := json.Marshal(g.PriorityReasons)
		if _, err := r.pool.Exec(ctx, q, eventID, guestID, g.PriorityScore, strings.ToLower(g.PriorityTier), string(reasonsRaw), now); err != nil {
			return err
		}
	}
	return nil
}

func (r *PGRepository) RelationshipTierCounts(ctx context.Context, eventID uuid.UUID) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			COALESCE(priority_tier, 'optional') AS tier,
			COUNT(*) AS c
		FROM guest_relationship_scores
		WHERE event_id = $1
		GROUP BY COALESCE(priority_tier, 'optional')
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[string]int{
		"critical":  0,
		"important": 0,
		"optional":  0,
	}
	for rows.Next() {
		var tier string
		var count int
		if scanErr := rows.Scan(&tier, &count); scanErr == nil {
			counts[strings.ToLower(strings.TrimSpace(tier))] = count
		}
	}
	return counts, nil
}
