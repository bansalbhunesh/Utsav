package guestrepo

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	phoneutil "github.com/bhune/utsav/services/api/pkg/phone"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

type ListGuestsParams struct {
	EventID      uuid.UUID
	Limit        int
	Offset       int
	Sort         string
	Cursor       *ListGuestsCursor
	PriorityTier string // lowercase critical|important|optional; empty = no filter
}

type Repository interface {
	ListGuests(ctx context.Context, p ListGuestsParams) ([]Guest, error)
	UpsertGuest(ctx context.Context, eventID uuid.UUID, input GuestInput) (string, error)
	UpsertGuestTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, input GuestInput) (string, error)
	GuestIDByEventPhoneTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, phone string) (string, error)
	ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, src io.Reader) (*ImportResult, error)
}

type PGRepository struct {
	write *pgxpool.Pool
	read  *pgxpool.Pool
}

// NewPGRepository creates a repository. read may be a replica for ListGuests; if nil, write is used for reads.
func NewPGRepository(write, read *pgxpool.Pool) *PGRepository {
	if read == nil {
		read = write
	}
	return &PGRepository{write: write, read: read}
}

func normalizeGuestListSort(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "name_desc", "priority_asc", "priority_desc", "rsvp_desc", "shagun_desc":
		return strings.ToLower(strings.TrimSpace(sort))
	default:
		return "name_asc"
	}
}

func (r *PGRepository) ListGuests(ctx context.Context, p ListGuestsParams) ([]Guest, error) {
	limit := p.Limit
	if limit <= 0 {
		limit = 50
	}
	// Align with API max page size (httpserver parseLimitOffset caps at 200).
	if limit > 200 {
		limit = 200
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}
	sort := normalizeGuestListSort(p.Sort)

	orderClause := "g.name ASC, g.id ASC"
	switch sort {
	case "name_desc":
		orderClause = "g.name DESC, g.id DESC"
	case "priority_asc":
		orderClause = "priority_score ASC, g.name ASC, g.id ASC"
	case "priority_desc":
		orderClause = "priority_score DESC, g.name ASC, g.id ASC"
	case "rsvp_desc":
		orderClause = "rsvp_yes_count DESC, g.name ASC, g.id ASC"
	case "shagun_desc":
		orderClause = "total_shagun_paise DESC, g.name ASC, g.id ASC"
	}

	args := []any{p.EventID}
	argPos := 2
	seekSQL := ""
	if p.Cursor != nil {
		c := p.Cursor
		switch sort {
		case "name_asc":
			seekSQL = fmt.Sprintf(` AND (g.name, g.id) > ($%d::text, $%d::uuid)`, argPos, argPos+1)
			args = append(args, c.Name, c.ID)
			argPos += 2
		case "name_desc":
			seekSQL = fmt.Sprintf(` AND (g.name, g.id) < ($%d::text, $%d::uuid)`, argPos, argPos+1)
			args = append(args, c.Name, c.ID)
			argPos += 2
		case "priority_desc":
			seekSQL = fmt.Sprintf(` AND (
				g.priority_score < $%d
				OR (g.priority_score = $%d AND g.name > $%d::text)
				OR (g.priority_score = $%d AND g.name = $%d::text AND g.id::uuid > $%d::uuid)
			)`, argPos, argPos+1, argPos+2, argPos+3, argPos+4, argPos+5)
			args = append(args, c.PriorityScore, c.PriorityScore, c.Name, c.PriorityScore, c.Name, c.ID)
			argPos += 6
		case "priority_asc":
			seekSQL = fmt.Sprintf(` AND (
				g.priority_score > $%d
				OR (g.priority_score = $%d AND g.name > $%d::text)
				OR (g.priority_score = $%d AND g.name = $%d::text AND g.id::uuid > $%d::uuid)
			)`, argPos, argPos+1, argPos+2, argPos+3, argPos+4, argPos+5)
			args = append(args, c.PriorityScore, c.PriorityScore, c.Name, c.PriorityScore, c.Name, c.ID)
			argPos += 6
		case "rsvp_desc":
			seekSQL = fmt.Sprintf(` AND (
				g.rsvp_yes_count < $%d
				OR (g.rsvp_yes_count = $%d AND g.name > $%d::text)
				OR (g.rsvp_yes_count = $%d AND g.name = $%d::text AND g.id::uuid > $%d::uuid)
			)`, argPos, argPos+1, argPos+2, argPos+3, argPos+4, argPos+5)
			args = append(args, c.RSVPYes, c.RSVPYes, c.Name, c.RSVPYes, c.Name, c.ID)
			argPos += 6
		case "shagun_desc":
			seekSQL = fmt.Sprintf(` AND (
				g.total_shagun_paise < $%d
				OR (g.total_shagun_paise = $%d AND g.name > $%d::text)
				OR (g.total_shagun_paise = $%d AND g.name = $%d::text AND g.id::uuid > $%d::uuid)
			)`, argPos, argPos+1, argPos+2, argPos+3, argPos+4, argPos+5)
			args = append(args, c.ShagunPaise, c.ShagunPaise, c.Name, c.ShagunPaise, c.Name, c.ID)
			argPos += 6
		}
	}
	if p.Cursor != nil && seekSQL == "" {
		return nil, fmt.Errorf("guest list cursor is not valid for sort %q", sort)
	}

	tierSQL := ""
	switch strings.ToLower(strings.TrimSpace(p.PriorityTier)) {
	case "critical", "important", "optional":
		t := strings.ToLower(strings.TrimSpace(p.PriorityTier))
		tierSQL = fmt.Sprintf(` AND (
			CASE WHEN g.priority_score >= 80 THEN 'critical'
			     WHEN g.priority_score >= 50 THEN 'important'
			     ELSE 'optional' END
		) = $%d::text`, argPos)
		args = append(args, t)
		argPos++
	}

	limitPos := argPos
	args = append(args, limit)
	argPos++

	limitOffsetSQL := fmt.Sprintf(`LIMIT $%d`, limitPos)
	if p.Cursor == nil {
		offPos := argPos
		args = append(args, offset)
		limitOffsetSQL = fmt.Sprintf(`LIMIT $%d OFFSET $%d`, limitPos, offPos)
	}

	rows, err := r.read.Query(ctx, `
		WITH guest_enriched AS (
			SELECT
				g.id, g.name, g.phone, g.email, g.relationship, g.side, g.tags, g.group_id,
				COALESCE(r.rsvp_yes_count, 0) AS rsvp_yes_count,
				COALESCE(r.rsvp_total_count, 0) AS rsvp_total_count,
				COALESCE(evt.sub_event_total, 0) AS sub_event_total,
				r.latest_rsvp_at AS latest_rsvp_at,
				COALESCE(s.total_shagun_paise, 0) AS total_shagun_paise,
				LEAST(100, GREATEST(0,
					ROUND(100.0 * (
						(0.30 * prio.rel_w + 0.20 * prio.rc + 0.15 * prio.rs + 0.15 * prio.ec + 0.10 * prio.hr + 0.10 * prio.ho)
						* prio.decay * prio.unc
					))::double precision
				))::int AS priority_score
			FROM guests g
			CROSS JOIN (
				SELECT COUNT(*)::int AS sub_event_total FROM sub_events WHERE event_id=$1
			) evt
			LEFT JOIN LATERAL (
				SELECT
					COUNT(*) FILTER (WHERE rr.status='yes') AS rsvp_yes_count,
					COUNT(*) AS rsvp_total_count,
					MAX(rr.updated_at) AS latest_rsvp_at
				FROM rsvp_responses rr
				WHERE rr.event_id=g.event_id AND rr.guest_phone=g.phone
			) r ON TRUE
			LEFT JOIN LATERAL (
				SELECT COALESCE(SUM(se.amount_paise),0) AS total_shagun_paise
				FROM shagun_entries se
				WHERE se.event_id=g.event_id
				  AND (se.guest_id=g.id OR COALESCE(se.meta->>'guest_phone','')=g.phone)
			) s ON TRUE
			CROSS JOIN LATERAL (
				SELECT
					(CASE lower(trim(COALESCE(g.relationship, '')))
						WHEN 'close_family' THEN 1.0::float8
						WHEN 'immediate_family' THEN 1.0::float8
						WHEN 'family' THEN 0.85::float8
						WHEN 'relative' THEN 0.85::float8
						WHEN 'relatives' THEN 0.85::float8
						WHEN 'friend' THEN 0.65::float8
						WHEN 'friends' THEN 0.65::float8
						WHEN 'colleague' THEN 0.45::float8
						WHEN 'coworker' THEN 0.45::float8
						ELSE CASE WHEN trim(COALESCE(g.relationship, '')) = '' THEN 0.2::float8 ELSE 0.35::float8 END
					END) AS rel_w,
					LEAST(1.0::float8, GREATEST(0.0::float8, COALESCE(r.rsvp_yes_count, 0)::float8 / 3.0)) AS rc,
					CASE WHEN r.latest_rsvp_at IS NULL THEN 0.0::float8
						ELSE LEAST(1.0::float8, GREATEST(0.0::float8, 1.0::float8 - ((EXTRACT(EPOCH FROM (now() - r.latest_rsvp_at)) / 86400.0) / 14.0)))
					END AS rs,
					CASE WHEN COALESCE(evt.sub_event_total, 0) = 0 THEN 0.0::float8
						ELSE LEAST(1.0::float8, GREATEST(0.0::float8, COALESCE(r.rsvp_total_count, 0)::float8 / NULLIF(evt.sub_event_total, 0)::float8))
					END AS ec,
					CASE WHEN COALESCE(r.rsvp_total_count, 0) = 0 THEN 0.0::float8
						ELSE LEAST(1.0::float8, GREATEST(0.0::float8, COALESCE(r.rsvp_yes_count, 0)::float8 / NULLIF(r.rsvp_total_count, 0)::float8))
					END AS hr,
					(CASE WHEN EXISTS (
						SELECT 1 FROM unnest(COALESCE(g.tags, ARRAY[]::text[])) AS t(tag)
						WHERE lower(trim(tag)) IN ('vip', 'priority', 'must_call')
					) THEN 1.0::float8 ELSE 0.0::float8 END) AS ho,
					CASE WHEN r.latest_rsvp_at IS NULL THEN 0.90::float8
						ELSE 0.75::float8 + 0.25::float8 * exp(-(EXTRACT(EPOCH FROM (now() - r.latest_rsvp_at)) / 86400.0) / 30.0)
					END AS decay,
					GREATEST(0.70::float8, LEAST(1.0::float8, 1.0::float8 - 0.08::float8 * (
						(CASE WHEN COALESCE(r.rsvp_total_count, 0) = 0 THEN 3 ELSE 0 END) +
						(CASE WHEN COALESCE(evt.sub_event_total, 0) = 0 THEN 1 ELSE 0 END)
					)::float8)) AS unc
			) prio
			WHERE g.event_id=$1
		)
		SELECT
			g.id, g.name, g.phone, g.email, g.relationship, g.side, g.tags, g.group_id,
			g.rsvp_yes_count, g.rsvp_total_count, g.sub_event_total, g.latest_rsvp_at, g.total_shagun_paise,
			g.priority_score
		FROM guest_enriched g WHERE 1=1`+seekSQL+tierSQL+`
		ORDER BY `+orderClause+`
		`+limitOffsetSQL, args...)
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
			&g.PriorityScore,
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
	err := r.write.QueryRow(ctx, `
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

func (r *PGRepository) UpsertGuestTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, input GuestInput) (string, error) {
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
	err := tx.QueryRow(ctx, `
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

func (r *PGRepository) GuestIDByEventPhoneTx(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, phone string) (string, error) {
	var guestID string
	err := tx.QueryRow(ctx, `
		SELECT id::text FROM guests WHERE event_id=$1 AND phone=$2 LIMIT 1`,
		eventID, phone,
	).Scan(&guestID)
	if err != nil {
		return "", err
	}
	return guestID, nil
}

func (r *PGRepository) ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, src io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(src)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	result := &ImportResult{Imported: 0, Errors: []ImportError{}}

	type csvRow struct {
		line  int
		name  string
		phone string
		email string
		rel   string
		side  string
	}

	valid := make([]csvRow, 0, 256)
	lineNo := 0
	nameIdx, phoneIdx, emailIdx, relIdx, sideIdx := 0, 1, -1, -1, -1
	headerChecked := false
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		lineNo++
		if lineNo == 1 && len(row) > 0 && strings.HasPrefix(row[0], "\ufeff") {
			row[0] = strings.TrimPrefix(row[0], "\ufeff")
		}
		if !headerChecked {
			headerChecked = true
			joined := strings.ToLower(strings.Join(row, "|"))
			if strings.Contains(joined, "phone") && strings.Contains(joined, "name") {
				nameIdx, phoneIdx, emailIdx, relIdx, sideIdx = -1, -1, -1, -1, -1
				for i, h := range row {
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
				continue
			}
		}
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
		phoneRaw := strings.TrimSpace(row[phoneIdx])
		phone, phoneErr := phoneutil.NormalizeE164(phoneRaw)
		if phoneRaw == "" {
			result.Errors = append(result.Errors, ImportError{Line: lineNo, Error: "missing_phone"})
			continue
		}
		if phoneErr != nil {
			result.Errors = append(result.Errors, ImportError{Line: lineNo, Error: "invalid_phone"})
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

		valid = append(valid, csvRow{line: lineNo, name: name, phone: phone, email: email, rel: rel, side: side})
	}
	if lineNo == 0 {
		return &ImportResult{Imported: 0, Errors: []ImportError{{Line: 1, Error: "empty_csv"}}}, nil
	}
	result.Imported = len(valid)
	if len(valid) == 0 {
		return result, nil
	}

	tx, err := r.write.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE csv_import_staging (
		name text NOT NULL,
		phone text NOT NULL,
		email text NOT NULL DEFAULT '',
		relationship text NOT NULL DEFAULT '',
		side text NOT NULL DEFAULT ''
	) ON COMMIT DROP`); err != nil {
		return nil, err
	}

	byPhone := make(map[string]csvRow, len(valid))
	for _, r := range valid {
		byPhone[r.phone] = r
	}
	copyRows := make([][]any, 0, len(byPhone))
	for _, r := range byPhone {
		copyRows = append(copyRows, []any{r.name, r.phone, r.email, r.rel, r.side})
	}

	_, err = tx.CopyFrom(ctx, pgx.Identifier{"csv_import_staging"},
		[]string{"name", "phone", "email", "relationship", "side"},
		pgx.CopyFromRows(copyRows))
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO guests (event_id, group_id, name, phone, email, relationship, side, tags)
		SELECT $1::uuid, NULL, s.name, s.phone, NULLIF(s.email,''), NULLIF(s.relationship,''), NULLIF(s.side,''), '{}'::text[]
		FROM csv_import_staging s
		ON CONFLICT (event_id, phone) DO UPDATE SET name=EXCLUDED.name, email=EXCLUDED.email,
			relationship=EXCLUDED.relationship, side=EXCLUDED.side, updated_at=now()`,
		eventID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
