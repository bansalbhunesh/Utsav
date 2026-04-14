package httpserver

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/bhune/utsav/services/api/internal/auth"
)

type rsvpOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type rsvpOTPVerify struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type rsvpItem struct {
	SubEventID           string `json:"sub_event_id" binding:"required"`
	Status               string `json:"status" binding:"required"`
	MealPref             string `json:"meal_pref"`
	Dietary              string `json:"dietary"`
	AccommodationNeeded  bool   `json:"accommodation_needed"`
	TravelMode           string `json:"travel_mode"`
	PlusOneNames         string `json:"plus_one_names"`
}

type rsvpSubmit struct {
	Items []rsvpItem `json:"items" binding:"required"`
}

func (s *Server) eventIDFromSlug(c *gin.Context) (uuid.UUID, bool) {
	slug := strings.TrimSpace(strings.ToLower(c.Param("slug")))
	ctx := c.Request.Context()
	var eid uuid.UUID
	if err := s.Pool.QueryRow(ctx, `SELECT id FROM events WHERE slug=$1`, slug).Scan(&eid); err != nil {
		writeAPIError(c, http.StatusNotFound, "NOT_FOUND", "Event not found.")
		return uuid.Nil, false
	}
	return eid, true
}

func (s *Server) postPublicRSVPOTPRequest(c *gin.Context) {
	eid, ok := s.eventIDFromSlug(c)
	if !ok {
		return
	}
	var body rsvpOTPRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Phone is required.")
		return
	}
	slug := strings.TrimSpace(strings.ToLower(c.Param("slug")))
	if s.RSVPOTPLimit != nil && !s.RSVPOTPLimit.Allow("rsvp_otp:"+c.ClientIP()+"|"+slug+"|"+body.Phone) {
		writeAPIError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many RSVP OTP requests. Please retry later.")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.Config.DevOTPCode), bcrypt.DefaultCost)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "OTP_HASH_FAILED", "Unable to process RSVP OTP request.")
		return
	}
	ctx := c.Request.Context()
	_, _ = s.Pool.Exec(ctx, `DELETE FROM rsvp_otp_challenges WHERE event_id=$1 AND phone=$2`, eid, body.Phone)
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO rsvp_otp_challenges (event_id, phone, code_hash, expires_at)
		VALUES ($1,$2,$3, now() + interval '10 minutes')`, eid, body.Phone, string(hash))
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "OTP_PERSIST_FAILED", "Unable to save RSVP OTP challenge.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) postPublicRSVPOTPVerify(c *gin.Context) {
	eid, ok := s.eventIDFromSlug(c)
	if !ok {
		return
	}
	var body rsvpOTPVerify
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Phone and code are required.")
		return
	}
	ctx := c.Request.Context()
	var id uuid.UUID
	var codeHash string
	var expires time.Time
	err := s.Pool.QueryRow(ctx, `
		SELECT id, code_hash, expires_at FROM rsvp_otp_challenges
		WHERE event_id=$1 AND phone=$2 ORDER BY created_at DESC LIMIT 1`, eid, body.Phone).Scan(&id, &codeHash, &expires)
	if err != nil {
		writeAPIError(c, http.StatusUnauthorized, "NO_CHALLENGE", "No active RSVP OTP challenge found.")
		return
	}
	if time.Now().After(expires) {
		writeAPIError(c, http.StatusUnauthorized, "OTP_EXPIRED", "RSVP OTP has expired.")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(codeHash), []byte(body.Code)); err != nil {
		writeAPIError(c, http.StatusUnauthorized, "INVALID_OTP", "Invalid RSVP OTP code.")
		return
	}
	_, _ = s.Pool.Exec(ctx, `DELETE FROM rsvp_otp_challenges WHERE id=$1`, id)
	tok, err := auth.SignGuestToken(eid, body.Phone, []byte(s.Config.JWTSecret), 24*time.Hour)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "TOKEN_SIGN_FAILED", "Unable to create guest access token.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"guest_access_token": tok})
}

func (s *Server) postPublicRSVP(c *gin.Context) {
	eidSlug, ok := s.eventIDFromSlug(c)
	if !ok {
		return
	}
	geid, phone, ok := s.guestBearer(c)
	if !ok {
		return
	}
	if geid != eidSlug {
		writeAPIError(c, http.StatusForbidden, "WRONG_EVENT", "Guest token does not match this event.")
		return
	}
	var body rsvpSubmit
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "RSVP payload is invalid.")
		return
	}
	ctx := c.Request.Context()
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "TX_BEGIN_FAILED", "Unable to start RSVP transaction.")
		return
	}
	defer tx.Rollback(ctx)
	for _, it := range body.Items {
		sid, err := uuid.Parse(it.SubEventID)
		if err != nil {
			writeAPIError(c, http.StatusBadRequest, "BAD_SUB_EVENT", "A sub event id is invalid.")
			return
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO rsvp_responses (
				event_id, guest_phone, sub_event_id, status, meal_pref, dietary, accommodation_needed, travel_mode, plus_one_names
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (event_id, guest_phone, sub_event_id) DO UPDATE SET
				status=EXCLUDED.status, meal_pref=EXCLUDED.meal_pref, dietary=EXCLUDED.dietary,
				accommodation_needed=EXCLUDED.accommodation_needed, travel_mode=EXCLUDED.travel_mode,
				plus_one_names=EXCLUDED.plus_one_names, updated_at=now()`,
			eidSlug, phone, sid, it.Status, it.MealPref, it.Dietary, it.AccommodationNeeded, it.TravelMode, it.PlusOneNames)
		if err != nil {
			writeAPIError(c, http.StatusBadRequest, "RSVP_UPSERT_FAILED", "Unable to save RSVP response.")
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		writeAPIError(c, http.StatusInternalServerError, "TX_COMMIT_FAILED", "Unable to commit RSVP transaction.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
