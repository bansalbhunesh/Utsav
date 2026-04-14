package httpserver

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) getPublicEvent(c *gin.Context) {
	slug := strings.TrimSpace(strings.ToLower(c.Param("slug")))
	ctx := c.Request.Context()
	var eid uuid.UUID
	var title, etype, privacy string
	var toggles, branding []byte
	var ca, cb, love, cover any
	var ds, de any
	err := s.Pool.QueryRow(ctx, `
		SELECT id, title, event_type, privacy, toggles, branding, couple_name_a, couple_name_b, love_story,
			cover_image_url, date_start, date_end
		FROM events WHERE slug=$1`, slug).Scan(
		&eid, &title, &etype, &privacy, &toggles, &branding, &ca, &cb, &love, &cover, &ds, &de)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": eid.String(), "slug": slug, "title": title, "event_type": etype, "privacy": privacy,
		"toggles": string(toggles), "branding": string(branding),
		"couple_name_a": ca, "couple_name_b": cb, "love_story": love, "cover_image_url": cover,
		"date_start": ds, "date_end": de,
	})
}

func (s *Server) getPublicSchedule(c *gin.Context) {
	slug := strings.TrimSpace(strings.ToLower(c.Param("slug")))
	ctx := c.Request.Context()
	var eid uuid.UUID
	if err := s.Pool.QueryRow(ctx, `SELECT id FROM events WHERE slug=$1`, slug).Scan(&eid); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT id, name, sub_type, starts_at, ends_at, venue_label, dress_code, description, sort_order
		FROM sub_events WHERE event_id=$1 ORDER BY sort_order, starts_at`, eid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	list := []gin.H{}
	now := time.Now()
	for rows.Next() {
		var id uuid.UUID
		var name, stype, vl, dc, desc string
		var starts, ends any
		var sort int
		_ = rows.Scan(&id, &name, &stype, &starts, &ends, &vl, &dc, &desc, &sort)
		happening := false
		var tStart time.Time
		switch v := starts.(type) {
		case time.Time:
			tStart = v
		}
		if !tStart.IsZero() {
			var endT time.Time
			if v, ok2 := ends.(time.Time); ok2 && !v.IsZero() {
				endT = v
			} else {
				endT = tStart.Add(3 * time.Hour)
			}
			if (now.Equal(tStart) || now.After(tStart)) && now.Before(endT) {
				happening = true
			}
		}
		list = append(list, gin.H{
			"id": id.String(), "name": name, "sub_type": stype, "starts_at": starts, "ends_at": ends,
			"venue_label": vl, "dress_code": dc, "description": desc, "sort_order": sort,
			"happening_now": happening,
		})
	}
	c.JSON(http.StatusOK, gin.H{"sub_events": list})
}

func (s *Server) getPublicBroadcasts(c *gin.Context) {
	slug := strings.TrimSpace(strings.ToLower(c.Param("slug")))
	ctx := c.Request.Context()
	var eid uuid.UUID
	if err := s.Pool.QueryRow(ctx, `SELECT id FROM events WHERE slug=$1`, slug).Scan(&eid); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT id, title, body, image_url, announcement_type, created_at
		FROM broadcasts WHERE event_id=$1 ORDER BY created_at DESC`, eid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var title, body, img, atype string
		var created any
		_ = rows.Scan(&id, &title, &body, &img, &atype, &created)
		out = append(out, gin.H{"id": id.String(), "title": title, "body": body, "image_url": img, "type": atype, "created_at": created})
	}
	c.JSON(http.StatusOK, gin.H{"broadcasts": out})
}

func (s *Server) getPublicGallery(c *gin.Context) {
	slug := strings.TrimSpace(strings.ToLower(c.Param("slug")))
	ctx := c.Request.Context()
	var eid uuid.UUID
	if err := s.Pool.QueryRow(ctx, `SELECT id FROM events WHERE slug=$1`, slug).Scan(&eid); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT id, section, object_key, created_at
		FROM gallery_assets
		WHERE event_id=$1 AND status='approved'
		ORDER BY created_at DESC`, eid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var sec, key string
		var created any
		_ = rows.Scan(&id, &sec, &key, &created)
		out = append(out, gin.H{
			"id": id.String(), "section": sec, "object_key": key, "created_at": created,
			"url": s.MediaSigner.PublicObjectURL(key),
		})
	}
	c.JSON(http.StatusOK, gin.H{"assets": out})
}
