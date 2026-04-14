package httpserver

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createVendorBody struct {
	Name         string  `json:"name" binding:"required"`
	Category     string  `json:"category"`
	Phone        string  `json:"phone"`
	Email        string  `json:"email"`
	TotalPaise   int64   `json:"total_paise"`
	AdvancePaise int64   `json:"advance_paise"`
	Notes        string  `json:"notes"`
}

func (s *Server) listVendors(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	rows, err := s.Pool.Query(ctx, `
		SELECT id, name, category, phone, email, advance_paise, total_paise, notes, created_at
		FROM event_vendors WHERE event_id=$1 ORDER BY created_at DESC`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	list := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var name, cat, phone, email, notes string
		var adv, tot int64
		var created any
		_ = rows.Scan(&id, &name, &cat, &phone, &email, &adv, &tot, &notes, &created)
		list = append(list, gin.H{
			"id": id.String(), "name": name, "category": cat, "phone": phone,
			"email": email, "advance_paise": adv, "total_paise": tot, "notes": notes,
		})
	}
	c.JSON(http.StatusOK, gin.H{"vendors": list})
}

func (s *Server) postVendor(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	var body createVendorBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	ctx := c.Request.Context()
	var vid uuid.UUID
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO event_vendors (event_id, name, category, phone, email, advance_paise, total_paise, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id`,
		eventID, body.Name, body.Category, body.Phone, body.Email, body.AdvancePaise, body.TotalPaise, body.Notes,
	).Scan(&vid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert_failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": vid.String()})
}

func (s *Server) deleteVendor(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	vid, err := uuid.Parse(c.Param("vendorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}
	ctx := c.Request.Context()
	_, err = s.Pool.Exec(ctx, `DELETE FROM event_vendors WHERE id=$1 AND event_id=$2`, vid, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
