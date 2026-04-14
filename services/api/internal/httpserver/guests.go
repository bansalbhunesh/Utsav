package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type guestBody struct {
	Name         string   `json:"name" binding:"required"`
	Phone        string   `json:"phone" binding:"required"`
	Email        string   `json:"email"`
	Relationship string   `json:"relationship"`
	Side         string   `json:"side"`
	Tags         []string `json:"tags"`
	GroupID      *string  `json:"group_id"`
}

func (s *Server) listGuests(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ctx := c.Request.Context()
	rows, err := s.Pool.Query(ctx, `
		SELECT id, name, phone, email, relationship, side, tags, group_id
		FROM guests WHERE event_id=$1 ORDER BY name`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var name, phone, rel, side string
		var email any
		var tags []string
		var gid any
		_ = rows.Scan(&id, &name, &phone, &email, &rel, &side, &tags, &gid)
		out = append(out, gin.H{
			"id": id.String(), "name": name, "phone": phone, "email": email,
			"relationship": rel, "side": side, "tags": tags, "group_id": gid,
		})
	}
	c.JSON(http.StatusOK, gin.H{"guests": out})
}

func (s *Server) postGuest(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body guestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	ctx := c.Request.Context()
	var gid any
	if body.GroupID != nil && *body.GroupID != "" {
		g, err := uuid.Parse(*body.GroupID)
		if err == nil {
			gid = g
		}
	}
	tags := any([]string{})
	if body.Tags != nil {
		tags = body.Tags
	}
	var guestID uuid.UUID
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO guests (event_id, group_id, name, phone, email, relationship, side, tags)
		VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),$8::text[])
		ON CONFLICT (event_id, phone) DO UPDATE SET name=EXCLUDED.name, email=EXCLUDED.email,
			relationship=EXCLUDED.relationship, side=EXCLUDED.side, tags=EXCLUDED.tags, updated_at=now()
		RETURNING id`,
		eventID, gid, body.Name, body.Phone, body.Email, body.Relationship, body.Side, tags,
	).Scan(&guestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "upsert_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": guestID.String()})
}

type cashShagunBody struct {
	GuestID    *string `json:"guest_id"`
	GuestPhone string  `json:"guest_phone"`
	AmountINR  float64 `json:"amount_inr" binding:"required"`
	SubEventID *string `json:"sub_event_id"`
	Notes      string  `json:"notes"`
}

func (s *Server) postCashShagun(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageFinancials(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body cashShagunBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	paise := int64(body.AmountINR * 100)
	ctx := c.Request.Context()
	var gid any
	if body.GuestID != nil {
		if g, err := uuid.Parse(*body.GuestID); err == nil {
			gid = g
		}
	}
	var sid any
	if body.SubEventID != nil {
		if s2, err := uuid.Parse(*body.SubEventID); err == nil {
			sid = s2
		}
	}
	meta := map[string]any{"notes": body.Notes, "guest_phone": body.GuestPhone}
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO shagun_entries (event_id, guest_id, channel, amount_paise, status, sub_event_id, meta)
		VALUES ($1,$2,'cash',$3,'host_verified',$4,$5::jsonb)`,
		eventID, gid, paise, sid, mustJSONB(meta))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insert_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (s *Server) listRSVPsHost(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ctx := c.Request.Context()
	rows, err := s.Pool.Query(ctx, `
		SELECT id, guest_phone, sub_event_id, status, meal_pref, dietary, accommodation_needed, travel_mode, plus_one_names, updated_at
		FROM rsvp_responses WHERE event_id=$1`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id, sub uuid.UUID
		var phone, st, meal, diet, travel, plus string
		var acc bool
		var updated any
		_ = rows.Scan(&id, &phone, &sub, &st, &meal, &diet, &acc, &travel, &plus, &updated)
		out = append(out, gin.H{
			"id": id.String(), "guest_phone": phone, "sub_event_id": sub.String(), "status": st,
			"meal_pref": meal, "dietary": diet, "accommodation_needed": acc, "travel_mode": travel,
			"plus_one_names": plus, "updated_at": updated,
		})
	}
	c.JSON(http.StatusOK, gin.H{"rsvps": out})
}

func (s *Server) listShagunHost(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageFinancials(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ctx := c.Request.Context()
	rows, err := s.Pool.Query(ctx, `
		SELECT id, channel, amount_paise, blessing_note, status, created_at
		FROM shagun_entries WHERE event_id=$1 ORDER BY created_at DESC`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var ch, bless, st string
		var amt any
		var created any
		_ = rows.Scan(&id, &ch, &amt, &bless, &st, &created)
		out = append(out, gin.H{"id": id.String(), "channel": ch, "amount_paise": amt, "blessing_note": bless, "status": st, "created_at": created})
	}
	c.JSON(http.StatusOK, gin.H{"shagun": out})
}
