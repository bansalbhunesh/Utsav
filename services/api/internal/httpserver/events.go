package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createEventBody struct {
	Slug        string         `json:"slug" binding:"required"`
	Title       string         `json:"title" binding:"required"`
	EventType   string         `json:"event_type"`
	CoupleA     string         `json:"couple_name_a"`
	CoupleB     string         `json:"couple_name_b"`
	LoveStory   string         `json:"love_story"`
	CoverURL    string         `json:"cover_image_url"`
	DateStart   *string        `json:"date_start"`
	DateEnd     *string        `json:"date_end"`
	Privacy     string         `json:"privacy"`
	Toggles     map[string]any `json:"toggles"`
	Branding    map[string]any `json:"branding"`
	HostUPIVPA  string         `json:"host_upi_vpa"`
}

func (s *Server) getCheckSlug(c *gin.Context) {
	slug := strings.TrimSpace(strings.ToLower(c.Query("slug")))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_slug"})
		return
	}
	ctx := c.Request.Context()
	var exists bool
	_ = s.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM events WHERE slug=$1)`, slug).Scan(&exists)
	c.JSON(http.StatusOK, gin.H{"slug": slug, "available": !exists})
}

func (s *Server) postEvent(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body createEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	slug := strings.TrimSpace(strings.ToLower(body.Slug))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_slug"})
		return
	}
	if body.EventType == "" {
		body.EventType = "wedding"
	}
	if body.Privacy == "" {
		body.Privacy = "public"
	}
	ctx := c.Request.Context()
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tx_begin"})
		return
	}
	defer tx.Rollback(ctx)

	var eid uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO events (
			owner_user_id, slug, title, event_type, couple_name_a, couple_name_b, love_story,
			cover_image_url, date_start, date_end, privacy, toggles, branding, host_upi_vpa
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,coalesce($12::jsonb,'{}'),coalesce($13::jsonb,'{}'),$14)
		RETURNING id`,
		uid, slug, body.Title, body.EventType, nullStr(body.CoupleA), nullStr(body.CoupleB), nullStr(body.LoveStory),
		nullStr(body.CoverURL), body.DateStart, body.DateEnd, body.Privacy, mustJSONB(body.Toggles), mustJSONB(body.Branding), nullStr(body.HostUPIVPA),
	).Scan(&eid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "create_failed", "detail": err.Error()})
		return
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO event_members (event_id, user_id, role, status) VALUES ($1,$2,'owner','active')`, eid, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "member_failed"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit_failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": eid.String(), "slug": slug})
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

func (s *Server) listEvents(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	rows, err := s.Pool.Query(ctx, `
		SELECT e.id, e.slug, e.title, e.event_type, e.date_start, e.updated_at
		FROM events e
		WHERE e.owner_user_id=$1
		UNION
		SELECT e.id, e.slug, e.title, e.event_type, e.date_start, e.updated_at
		FROM events e
		JOIN event_members m ON m.event_id=e.id AND m.user_id=$1 AND m.status='active'
		ORDER BY updated_at DESC`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var slug, title, etype string
		var ds any
		var updated any
		_ = rows.Scan(&id, &slug, &title, &etype, &ds, &updated)
		out = append(out, gin.H{"id": id.String(), "slug": slug, "title": title, "event_type": etype, "date_start": ds})
	}
	c.JSON(http.StatusOK, gin.H{"events": out})
}

func (s *Server) getEvent(c *gin.Context) {
	userID, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	_ = role
	_ = userID
	ctx := c.Request.Context()
	row := s.Pool.QueryRow(ctx, `
		SELECT id, slug, title, event_type, couple_name_a, couple_name_b, love_story, cover_image_url,
			date_start, date_end, privacy, toggles, branding, host_upi_vpa, tier, created_at, updated_at
		FROM events WHERE id=$1`, eventID)
	var e struct {
		ID, Slug, Title, EventType string
		CA, CB, Love, Cover        *string
		DS, DE                      *string
		Privacy                     string
		Toggles, Branding           []byte
		VPA                         *string
		Tier                        string
		Created, Updated            any
	}
	var id uuid.UUID
	if err := row.Scan(&id, &e.Slug, &e.Title, &e.EventType, &e.CA, &e.CB, &e.Love, &e.Cover, &e.DS, &e.DE, &e.Privacy, &e.Toggles, &e.Branding, &e.VPA, &e.Tier, &e.Created, &e.Updated); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id.String(), "slug": e.Slug, "title": e.Title, "event_type": e.EventType,
		"couple_name_a": e.CA, "couple_name_b": e.CB, "love_story": e.Love, "cover_image_url": e.Cover,
		"date_start": e.DS, "date_end": e.DE, "privacy": e.Privacy,
		"toggles": string(e.Toggles), "branding": string(e.Branding), "host_upi_vpa": e.VPA, "tier": e.Tier,
	})
}

type patchEventBody struct {
	Title      *string        `json:"title"`
	Privacy    *string        `json:"privacy"`
	Toggles    map[string]any `json:"toggles"`
	Branding   map[string]any `json:"branding"`
	HostUPIVPA *string        `json:"host_upi_vpa"`
}

func (s *Server) patchEvent(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body patchEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	ctx := c.Request.Context()
	if body.Title != nil {
		_, _ = s.Pool.Exec(ctx, `UPDATE events SET title=$2, updated_at=now() WHERE id=$1`, eventID, *body.Title)
	}
	if body.Privacy != nil {
		_, _ = s.Pool.Exec(ctx, `UPDATE events SET privacy=$2, updated_at=now() WHERE id=$1`, eventID, *body.Privacy)
	}
	if body.HostUPIVPA != nil && roleCanManageFinancials(role) {
		_, _ = s.Pool.Exec(ctx, `UPDATE events SET host_upi_vpa=$2, updated_at=now() WHERE id=$1`, eventID, *body.HostUPIVPA)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type subEventBody struct {
	Name        string  `json:"name" binding:"required"`
	SubType     string  `json:"sub_type"`
	StartsAt    *string `json:"starts_at"`
	EndsAt      *string `json:"ends_at"`
	VenueLabel  string  `json:"venue_label"`
	DressCode   string  `json:"dress_code"`
	Description string  `json:"description"`
}

func (s *Server) postSubEvent(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body subEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	ctx := c.Request.Context()
	var sid uuid.UUID
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO sub_events (event_id, name, sub_type, starts_at, ends_at, venue_label, dress_code, description, sort_order)
		VALUES ($1,$2,$3,CAST($4 AS timestamptz),CAST($5 AS timestamptz),$6,$7,$8,
			(SELECT COALESCE(MAX(sort_order),0)+1 FROM sub_events WHERE event_id=$1))
		RETURNING id`,
		eventID, body.Name, body.SubType, body.StartsAt, body.EndsAt, body.VenueLabel, body.DressCode, body.Description,
	).Scan(&sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "create_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": sid.String()})
}

func (s *Server) listSubEvents(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	rows, err := s.Pool.Query(ctx, `
		SELECT id, name, sub_type, starts_at, ends_at, venue_label, dress_code, description, sort_order
		FROM sub_events WHERE event_id=$1 ORDER BY sort_order, starts_at`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	list := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var name, stype, vl, dc, desc string
		var starts, ends any
		var sort int
		_ = rows.Scan(&id, &name, &stype, &starts, &ends, &vl, &dc, &desc, &sort)
		list = append(list, gin.H{
			"id": id.String(), "name": name, "sub_type": stype, "starts_at": starts, "ends_at": ends,
			"venue_label": vl, "dress_code": dc, "description": desc, "sort_order": sort,
		})
	}
	c.JSON(http.StatusOK, gin.H{"sub_events": list})
}

type inviteMemberBody struct {
	InvitedPhone string `json:"invited_phone" binding:"required"`
	Role         string `json:"role" binding:"required"`
}

func (s *Server) postEventMember(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if role != "owner" && role != "co_owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body inviteMemberBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	ctx := c.Request.Context()
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO event_members (event_id, role, invited_phone, status)
		VALUES ($1,$2,$3,'invited')`, eventID, body.Role, body.InvitedPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invite_failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}
