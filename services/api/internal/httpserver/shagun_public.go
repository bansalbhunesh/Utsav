package httpserver

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type shagunReportBody struct {
	AmountINR    float64 `json:"amount_inr" binding:"required"`
	BlessingNote string  `json:"blessing_note"`
	SubEventID   *string `json:"sub_event_id"`
}

// UPI deep link helper (metadata only; funds never touch platform).
func (s *Server) getPublicUPILink(c *gin.Context) {
	slug := strings.TrimSpace(strings.ToLower(c.Param("slug")))
	ctx := c.Request.Context()
	var eid uuid.UUID
	var vpa, title string
	if err := s.Pool.QueryRow(ctx, `
		SELECT e.id, COALESCE(e.host_upi_vpa,''), COALESCE(NULLIF(e.title,''), e.slug)
		FROM events e WHERE e.slug=$1`, slug).Scan(&eid, &vpa, &title); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if vpa == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host_vpa_not_configured"})
		return
	}
	geid, phone, ok := s.guestBearer(c)
	if !ok {
		return
	}
	if geid != eid {
		c.JSON(http.StatusForbidden, gin.H{"error": "wrong_event"})
		return
	}
	note := "Shagun from guest for " + title
	upi := "upi://pay?pa=" + url.QueryEscape(vpa) +
		"&pn=" + url.QueryEscape(title) +
		"&tn=" + url.QueryEscape(note) +
		"&am=&cu=INR"
	c.JSON(http.StatusOK, gin.H{
		"upi_uri":            upi,
		"payee_vpa":          vpa,
		"transaction_note":   note,
		"guest_phone_masked": maskPhone(phone),
	})
}

func maskPhone(p string) string {
	if len(p) < 4 {
		return "****"
	}
	return "******" + p[len(p)-4:]
}

func (s *Server) postPublicShagunReport(c *gin.Context) {
	eidSlug, ok := s.eventIDFromSlug(c)
	if !ok {
		return
	}
	geid, phone, ok := s.guestBearer(c)
	if !ok {
		return
	}
	if geid != eidSlug {
		c.JSON(http.StatusForbidden, gin.H{"error": "wrong_event"})
		return
	}
	var body shagunReportBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	paise := int64(body.AmountINR * 100)
	var sid any
	if body.SubEventID != nil {
		if u, err := uuid.Parse(*body.SubEventID); err == nil {
			sid = u
		}
	}
	ctx := c.Request.Context()
	meta := map[string]any{"guest_phone": phone, "reported_at": time.Now().UTC().Format(time.RFC3339)}
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO shagun_entries (event_id, channel, amount_paise, blessing_note, status, sub_event_id, meta)
		VALUES ($1,'upi',$2,$3,'guest_reported',$4,$5::jsonb)`,
		eidSlug, paise, body.BlessingNote, sid, mustJSONB(meta))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insert_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}
