package httpserver

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bhune/utsav/services/api/internal/auth"
	"github.com/bhune/utsav/services/api/internal/config"
	"github.com/bhune/utsav/services/api/internal/media"
	authservice "github.com/bhune/utsav/services/api/internal/service/auth"
	billingservice "github.com/bhune/utsav/services/api/internal/service/billing"
	broadcastservice "github.com/bhune/utsav/services/api/internal/service/broadcast"
	eventservice "github.com/bhune/utsav/services/api/internal/service/event"
	galleryservice "github.com/bhune/utsav/services/api/internal/service/gallery"
	guestservice "github.com/bhune/utsav/services/api/internal/service/guest"
	memorybookservice "github.com/bhune/utsav/services/api/internal/service/memorybook"
	organiserservice "github.com/bhune/utsav/services/api/internal/service/organiser"
	publicservice "github.com/bhune/utsav/services/api/internal/service/public"
	rsvpservice "github.com/bhune/utsav/services/api/internal/service/rsvp"
	shagunservice "github.com/bhune/utsav/services/api/internal/service/shagun"
	vendorservice "github.com/bhune/utsav/services/api/internal/service/vendor"
)

type apiErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiErrorEnvelope struct {
	Success bool           `json:"success"`
	Error   apiErrorDetail `json:"error"`
}

func writeAPIError(c *gin.Context, status int, code string, message string) {
	c.Set("error_code", code)
	logPayload := map[string]any{
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"level":      "warn",
		"request_id": c.GetString("request_id"),
		"user_id":    c.GetString("user_id"),
		"guest_id":   c.GetString("guest_id"),
		"status":     status,
		"error_code": code,
		"error":      message,
	}
	if c.Request != nil {
		logPayload["method"] = c.Request.Method
		logPayload["path"] = c.Request.URL.Path
		logPayload["endpoint"] = c.FullPath()
	}
	if b, err := json.Marshal(logPayload); err == nil {
		gin.DefaultWriter.Write(append(b, '\n'))
	}
	c.JSON(status, apiErrorEnvelope{
		Success: false,
		Error: apiErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

type Server struct {
	Pool         *pgxpool.Pool
	Config       *config.Config
	MediaSigner  media.Signer
	AuthService  *authservice.Service
	BillingService *billingservice.Service
	BroadcastService *broadcastservice.Service
	EventService *eventservice.Service
	GalleryService *galleryservice.Service
	MemoryBookService *memorybookservice.Service
	OrganiserService *organiserservice.Service
	PublicService *publicservice.Service
	RSVPService  *rsvpservice.Service
	GuestService *guestservice.Service
	ShagunService *shagunservice.Service
	VendorService *vendorservice.Service
}

func (s *Server) requireUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := s.requireUser(c); !ok {
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) invalidatePublicEventCache(ctx context.Context, eventID uuid.UUID) {
	if s.PublicService == nil {
		return
	}
	s.PublicService.InvalidatePublicEventCache(ctx, eventID)
}

func (s *Server) requireEventAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, eventID, role, ok := s.requireEventAccess(c)
		if !ok {
			c.Abort()
			return
		}
		c.Set("auth_user_id", userID.String())
		c.Set("auth_event_id", eventID.String())
		c.Set("auth_event_role", role)
		c.Next()
	}
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

func cookieUserID(c *gin.Context, secret []byte) (uuid.UUID, bool) {
	raw, err := c.Cookie("utsav_access_token")
	if err != nil || strings.TrimSpace(raw) == "" {
		return uuid.Nil, false
	}
	uid, parseErr := auth.ParseAccessToken(raw, secret)
	if parseErr != nil {
		return uuid.Nil, false
	}
	return uid, true
}

func (s *Server) requireUser(c *gin.Context) (uuid.UUID, bool) {
	uid, ok := bearerUserID(c, []byte(s.Config.JWTSecret))
	if !ok {
		uid, ok = cookieUserID(c, []byte(s.Config.JWTSecret))
	}
	if !ok {
		writeAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid access token.")
		return uuid.Nil, false
	}
	c.Set("user_id", uid.String())
	return uid, true
}

func (s *Server) eventRole(ctx context.Context, userID, eventID uuid.UUID) (string, bool) {
	var owner uuid.UUID
	var memberRole sql.NullString
	err := s.Pool.QueryRow(ctx, `
		SELECT e.owner_user_id, m.role
		FROM events e
		LEFT JOIN event_members m ON m.event_id = e.id AND m.user_id = $2 AND m.status = 'active'
		WHERE e.id = $1`, eventID, userID).Scan(&owner, &memberRole)
	if err != nil {
		return "", false
	}
	if owner == userID {
		return "owner", true
	}
	if !memberRole.Valid {
		return "", false
	}
	return memberRole.String, true
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
		writeAPIError(c, http.StatusBadRequest, "INVALID_EVENT_ID", "Event id is not valid.")
		return uuid.Nil, uuid.Nil, "", false
	}
	role, ok = s.eventRole(c.Request.Context(), userID, eventID)
	if !ok {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have access to this event.")
		return uuid.Nil, uuid.Nil, "", false
	}
	return userID, eventID, role, true
}

func (s *Server) guestBearer(c *gin.Context) (eventID uuid.UUID, phone string, ok bool) {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		writeAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing guest access token.")
		return uuid.Nil, "", false
	}
	raw := strings.TrimSpace(h[7:])
	eid, ph, err := auth.ParseGuestToken(raw, []byte(s.Config.JWTSecret))
	if err != nil {
		writeAPIError(c, http.StatusUnauthorized, "INVALID_GUEST_TOKEN", "Guest session is invalid or expired.")
		return uuid.Nil, "", false
	}
	c.Set("guest_id", ph)
	return eid, ph, true
}

func parseLimitOffset(c *gin.Context) (int, int) {
	limit := 50
	offset := 0
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			offset = n
		}
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func parseGuestListQuery(c *gin.Context) (sort string, priorityTier string) {
	sort = strings.ToLower(strings.TrimSpace(c.Query("sort")))
	switch sort {
	case "", "name_asc", "name_desc", "priority_desc", "priority_asc", "rsvp_desc", "shagun_desc":
	default:
		sort = "name_asc"
	}
	if sort == "" {
		sort = "name_asc"
	}
	priorityTier = strings.ToLower(strings.TrimSpace(c.Query("priority_tier")))
	switch priorityTier {
	case "", "critical", "important", "optional":
	default:
		priorityTier = ""
	}
	return sort, priorityTier
}

func hashFingerprint(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Server) reserveIdempotencyKey(ctx context.Context, scope, key, fingerprint string) (bool, error) {
	if _, err := s.Pool.Exec(ctx, `
		DELETE FROM idempotency_keys WHERE scope=$1 AND key=$2 AND expires_at < now()
	`, scope, key); err != nil {
		return false, err
	}
	tag, err := s.Pool.Exec(ctx, `
		INSERT INTO idempotency_keys (scope, key, fingerprint, expires_at)
		VALUES ($1,$2,$3, now() + interval '24 hours')
		ON CONFLICT (scope, key) DO NOTHING
	`, scope, key, fingerprint)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		var existing string
		readErr := s.Pool.QueryRow(ctx, `
			SELECT fingerprint FROM idempotency_keys
			WHERE scope=$1 AND key=$2 AND expires_at > now()
		`, scope, key).Scan(&existing)
		if readErr != nil {
			return false, readErr
		}
		return existing == fingerprint, nil
	}
	return true, nil
}

func (s *Server) reserveWebhookDelivery(ctx context.Context, provider, eventKey, payloadHash string) (bool, error) {
	tag, err := s.Pool.Exec(ctx, `
		INSERT INTO webhook_deliveries (provider, event_key, payload_hash)
		VALUES ($1,$2,$3)
		ON CONFLICT (provider, event_key) DO NOTHING
	`, provider, eventKey, payloadHash)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		var existing string
		readErr := s.Pool.QueryRow(ctx, `
			SELECT payload_hash FROM webhook_deliveries WHERE provider=$1 AND event_key=$2
		`, provider, eventKey).Scan(&existing)
		if readErr != nil {
			return false, readErr
		}
		return existing == payloadHash, nil
	}
	return true, nil
}
