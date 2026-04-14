package httpserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bhune/utsav/services/api/internal/auth"
	"github.com/bhune/utsav/services/api/internal/config"
	"github.com/bhune/utsav/services/api/internal/media"
	"github.com/bhune/utsav/services/api/internal/ratelimit"
)

type Server struct {
	Pool         *pgxpool.Pool
	Config       *config.Config
	AuthOTPLimit *ratelimit.Window
	RSVPOTPLimit *ratelimit.Window
	MediaSigner  media.Signer
}

func bearerUserID(c *gin.Context, secret []byte) (uuid.UUID, bool) {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return uuid.Nil, false
	}
	raw := strings.TrimSpace(h[7:])
	if raw == "" {
		return uuid.Nil, false
	}
	uid, err := auth.ParseAccessToken(raw, secret)
	if err != nil {
		return uuid.Nil, false
	}
	return uid, true
}

func (s *Server) requireUser(c *gin.Context) (uuid.UUID, bool) {
	uid, ok := bearerUserID(c, []byte(s.Config.JWTSecret))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	return uid, true
}

func (s *Server) eventRole(ctx context.Context, userID, eventID uuid.UUID) (string, bool) {
	var owner uuid.UUID
	err := s.Pool.QueryRow(ctx, `SELECT owner_user_id FROM events WHERE id=$1`, eventID).Scan(&owner)
	if err != nil {
		return "", false
	}
	if owner == userID {
		return "owner", true
	}
	var role string
	err = s.Pool.QueryRow(ctx, `
		SELECT role FROM event_members
		WHERE event_id=$1 AND user_id=$2 AND status='active'`, eventID, userID).Scan(&role)
	if err != nil {
		return "", false
	}
	return role, true
}

func roleCanManageEventData(role string) bool {
	switch role {
	case "owner", "co_owner", "organiser", "contributor":
		return true
	default:
		return false
	}
}

func roleCanManageFinancials(role string) bool {
	switch role {
	case "owner", "co_owner", "organiser":
		return true
	default:
		return false
	}
}

func (s *Server) requireEventAccess(c *gin.Context) (userID uuid.UUID, eventID uuid.UUID, role string, ok bool) {
	userID, ok = s.requireUser(c)
	if !ok {
		return uuid.Nil, uuid.Nil, "", false
	}
	eidStr := c.Param("id")
	if eidStr == "" {
		eidStr = c.Param("eventId")
	}
	eventID, err := uuid.Parse(eidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_event_id"})
		return uuid.Nil, uuid.Nil, "", false
	}
	role, ok = s.eventRole(c.Request.Context(), userID, eventID)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return uuid.Nil, uuid.Nil, "", false
	}
	return userID, eventID, role, true
}

func (s *Server) guestBearer(c *gin.Context) (eventID uuid.UUID, phone string, ok bool) {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, "", false
	}
	raw := strings.TrimSpace(h[7:])
	eid, ph, err := auth.ParseGuestToken(raw, []byte(s.Config.JWTSecret))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_guest_token"})
		return uuid.Nil, "", false
	}
	return eid, ph, true
}
