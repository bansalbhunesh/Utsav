package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	rsvpservice "github.com/bhune/utsav/services/api/internal/service/rsvp"
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
	if s.RSVPService == nil {
		writeAPIError(c, http.StatusInternalServerError, "RSVP_SERVICE_UNAVAILABLE", "RSVP service unavailable.")
		return uuid.Nil, false
	}
	eid, svcErr := s.RSVPService.EventIDFromSlug(c.Request.Context(), c.Param("slug"))
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
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
	if s.RSVPService == nil {
		writeAPIError(c, http.StatusInternalServerError, "RSVP_SERVICE_UNAVAILABLE", "RSVP service unavailable.")
		return
	}
	if svcErr := s.RSVPService.RequestOTP(c.Request.Context(), eid, c.Param("slug"), body.Phone, c.ClientIP()); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
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
	if s.RSVPService == nil {
		writeAPIError(c, http.StatusInternalServerError, "RSVP_SERVICE_UNAVAILABLE", "RSVP service unavailable.")
		return
	}
	tok, svcErr := s.RSVPService.VerifyOTP(c.Request.Context(), eid, body.Phone, body.Code, c.ClientIP())
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
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
	if s.RSVPService == nil {
		writeAPIError(c, http.StatusInternalServerError, "RSVP_SERVICE_UNAVAILABLE", "RSVP service unavailable.")
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		writeAPIError(c, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header is required.")
		return
	}
	rawItems, _ := json.Marshal(body.Items)
	fingerprint := hashFingerprint(eidSlug.String(), phone, string(rawItems))
	ok, err := s.reserveIdempotencyKey(c.Request.Context(), "public_rsvp_submit", idempotencyKey, fingerprint)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "IDEMPOTENCY_FAILED", "Unable to validate idempotency key.")
		return
	}
	if !ok {
		writeAPIError(c, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency key was already used for a different request.")
		return
	}

	inputs := make([]rsvpservice.SubmitItemInput, 0, len(body.Items))
	for _, it := range body.Items {
		inputs = append(inputs, rsvpservice.SubmitItemInput{
			SubEventID:          it.SubEventID,
			Status:              it.Status,
			MealPref:            it.MealPref,
			Dietary:             it.Dietary,
			AccommodationNeeded: it.AccommodationNeeded,
			TravelMode:          it.TravelMode,
			PlusOneNames:        it.PlusOneNames,
		})
	}
	if svcErr := s.RSVPService.SubmitRSVP(c.Request.Context(), eidSlug, geid, phone, c.ClientIP(), inputs); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
