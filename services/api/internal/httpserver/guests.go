package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/bhune/utsav/services/api/internal/repository/guestrepo"
	"github.com/bhune/utsav/services/api/internal/repository/shagunrepo"
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
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to manage guests.")
		return
	}
	if s.GuestService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GUEST_SERVICE_UNAVAILABLE", "Guest service unavailable.")
		return
	}
	limit, offset := parseLimitOffset(c)
	sort, priorityTier := parseGuestListQuery(c)
	var cursorStr *string
	if raw := strings.TrimSpace(c.Query("cursor")); raw != "" {
		cursorStr = &raw
	}
	list, nextCursor, svcErr := s.GuestService.ListGuests(c.Request.Context(), eventID, limit, offset, sort, priorityTier, cursorStr)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	out := gin.H{
		"guests": list, "limit": limit, "offset": offset, "sort": sort, "priority_tier": priorityTier,
	}
	if nextCursor != nil {
		out["next_cursor"] = *nextCursor
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) getRelationshipPriorityOverview(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to view guest intelligence.")
		return
	}
	if s.GuestService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GUEST_SERVICE_UNAVAILABLE", "Guest service unavailable.")
		return
	}
	overview, svcErr := s.GuestService.RelationshipScoreOverview(c.Request.Context(), eventID)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"feature":                  "relationship_priority_score",
		"status":                   "active",
		"ranked_guests":            overview.RankedGuests,
		"guests_needing_attention": overview.GuestsNeedingAttention,
		"tier_counts":              overview.TierCounts,
		"coming_next": []string{
			"rsvp_risk_predictor",
			"shagun_signal_intelligence",
		},
	})
}

func (s *Server) postGuest(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to manage guests.")
		return
	}
	var body guestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Guest payload is invalid.")
		return
	}
	if s.GuestService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GUEST_SERVICE_UNAVAILABLE", "Guest service unavailable.")
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		writeAPIError(c, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header is required.")
		return
	}
	rawBody, _ := json.Marshal(body)
	fingerprint := hashFingerprint(eventID.String(), string(rawBody))
	ctx := c.Request.Context()
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "IDEMPOTENCY_FAILED", "Unable to validate idempotency key.")
		return
	}
	defer tx.Rollback(ctx)

	replay, err := reserveIdempotencyInTx(ctx, tx, "guest_upsert", idempotencyKey, fingerprint)
	if err != nil {
		if errors.Is(err, ErrIdempotencyFingerprintMismatch) {
			writeAPIError(c, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency key was already used for a different request.")
			return
		}
		writeAPIError(c, http.StatusInternalServerError, "IDEMPOTENCY_FAILED", "Unable to validate idempotency key.")
		return
	}

	in := guestrepo.GuestInput{
		Name:         body.Name,
		Phone:        body.Phone,
		Email:        body.Email,
		Relationship: body.Relationship,
		Side:         body.Side,
		Tags:         body.Tags,
		GroupID:      body.GroupID,
	}
	var guestID string
	if replay {
		gid, e := s.GuestService.GuestIDByEventPhoneTx(ctx, tx, eventID, in.Phone)
		if e != nil {
			writeAPIError(c, e.Status, e.Code, e.Message)
			return
		}
		guestID = gid
	} else {
		gid, e := s.GuestService.UpsertGuestTx(ctx, tx, eventID, in)
		if e != nil {
			writeAPIError(c, e.Status, e.Code, e.Message)
			return
		}
		guestID = gid
	}
	if err := tx.Commit(ctx); err != nil {
		writeAPIError(c, http.StatusInternalServerError, "UPSERT_FAILED", "Unable to save guest.")
		return
	}
	s.GuestService.InvalidateRelationshipOverview(ctx, eventID)
	c.JSON(http.StatusCreated, gin.H{"id": guestID})
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
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to manage financial entries.")
		return
	}
	var body cashShagunBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Cash shagun payload is invalid.")
		return
	}
	if s.ShagunService == nil {
		writeAPIError(c, http.StatusInternalServerError, "SHAGUN_SERVICE_UNAVAILABLE", "Shagun service unavailable.")
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		writeAPIError(c, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header is required.")
		return
	}
	rawBody, _ := json.Marshal(body)
	fingerprint := hashFingerprint(eventID.String(), string(rawBody))
	ctx := c.Request.Context()
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "IDEMPOTENCY_FAILED", "Unable to validate idempotency key.")
		return
	}
	defer tx.Rollback(ctx)

	replay, err := reserveIdempotencyInTx(ctx, tx, "cash_shagun_create", idempotencyKey, fingerprint)
	if err != nil {
		if errors.Is(err, ErrIdempotencyFingerprintMismatch) {
			writeAPIError(c, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency key was already used for a different request.")
			return
		}
		writeAPIError(c, http.StatusInternalServerError, "IDEMPOTENCY_FAILED", "Unable to validate idempotency key.")
		return
	}
	if !replay {
		svcErr := s.ShagunService.LogCashShagunTx(ctx, tx, eventID, shagunrepo.CashShagunInput{
			GuestID:    body.GuestID,
			GuestPhone: body.GuestPhone,
			AmountINR:  body.AmountINR,
			SubEventID: body.SubEventID,
			Notes:      body.Notes,
		})
		if svcErr != nil {
			writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INSERT_FAILED", "Unable to log cash shagun.")
		return
	}
	if s.GuestService != nil {
		s.GuestService.InvalidateRelationshipOverview(ctx, eventID)
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (s *Server) listRSVPsHost(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to view RSVP responses.")
		return
	}
	if s.RSVPService == nil {
		writeAPIError(c, http.StatusInternalServerError, "RSVP_SERVICE_UNAVAILABLE", "RSVP service unavailable.")
		return
	}
	limit, offset := parseLimitOffset(c)
	rows, svcErr := s.RSVPService.ListHostRSVPs(c.Request.Context(), eventID, limit, offset)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"rsvps": rows, "limit": limit, "offset": offset})
}

func (s *Server) listShagunHost(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageFinancials(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to view shagun entries.")
		return
	}
	if s.ShagunService == nil {
		writeAPIError(c, http.StatusInternalServerError, "SHAGUN_SERVICE_UNAVAILABLE", "Shagun service unavailable.")
		return
	}
	limit, offset := parseLimitOffset(c)
	rows, svcErr := s.ShagunService.ListHostShagun(c.Request.Context(), eventID, limit, offset)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"shagun": rows, "limit": limit, "offset": offset})
}
