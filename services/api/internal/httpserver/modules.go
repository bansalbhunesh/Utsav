package httpserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/broadcastrepo"
	"github.com/bhune/utsav/services/api/internal/repository/galleryrepo"
	"github.com/bhune/utsav/services/api/internal/repository/organiserrepo"
	billingservice "github.com/bhune/utsav/services/api/internal/service/billing"
)

type galleryAssetBody struct {
	Section    string `json:"section" binding:"required"`
	ObjectKey  string `json:"object_key" binding:"required"`
	SubEventID string `json:"sub_event_id"`
	Status     string `json:"status"`
	MimeType   string `json:"mime_type"`
	Bytes      int64  `json:"bytes"`
}

type galleryPresignBody struct {
	Section     string `json:"section" binding:"required"`
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type"`
	SubEventID  string `json:"sub_event_id"`
}

type galleryModerateBody struct {
	Status string `json:"status" binding:"required"`
}

func (s *Server) postGalleryPresign(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to manage gallery.")
		return
	}
	if s.GalleryService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GALLERY_SERVICE_UNAVAILABLE", "Gallery service unavailable.")
		return
	}
	var body galleryPresignBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Gallery presign payload is invalid.")
		return
	}
	upload, svcErr := s.GalleryService.PresignPut(eventID, body.FileName, body.ContentType)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"upload": upload})
}

func (s *Server) postGalleryAsset(c *gin.Context) {
	uid, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to manage gallery.")
		return
	}
	if s.GalleryService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GALLERY_SERVICE_UNAVAILABLE", "Gallery service unavailable.")
		return
	}
	var body galleryAssetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Gallery asset payload is invalid.")
		return
	}
	if svcErr := s.GalleryService.CreateAsset(c.Request.Context(), galleryrepo.CreateAssetInput{
		EventID:        eventID,
		UploaderUserID: uid,
		Section:        body.Section,
		ObjectKey:      body.ObjectKey,
		SubEventID:     body.SubEventID,
		Status:         body.Status,
		MimeType:       body.MimeType,
		Bytes:          body.Bytes,
	}); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	s.invalidatePublicEventCache(c.Request.Context(), eventID)
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (s *Server) listGalleryAssets(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to manage gallery.")
		return
	}
	if s.GalleryService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GALLERY_SERVICE_UNAVAILABLE", "Gallery service unavailable.")
		return
	}
	assets, svcErr := s.GalleryService.ListAssets(c.Request.Context(), eventID, c.Query("status"))
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"assets": assets})
}

func (s *Server) patchGalleryAssetModeration(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to moderate gallery.")
		return
	}
	aid, err := uuid.Parse(c.Param("assetId"))
	if err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_ASSET_ID", "Asset id is invalid.")
		return
	}
	if s.GalleryService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GALLERY_SERVICE_UNAVAILABLE", "Gallery service unavailable.")
		return
	}
	var body galleryModerateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Gallery moderation payload is invalid.")
		return
	}
	if svcErr := s.GalleryService.ModerateAsset(c.Request.Context(), eventID, aid, body.Status); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	s.invalidatePublicEventCache(c.Request.Context(), eventID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type broadcastBody struct {
	Title    string         `json:"title" binding:"required"`
	Body     string         `json:"body" binding:"required"`
	ImageURL string         `json:"image_url"`
	Audience map[string]any `json:"audience"`
	Type     string         `json:"announcement_type"`
}

func (s *Server) listBroadcastsHost(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to manage broadcasts.")
		return
	}
	if s.BroadcastService == nil {
		writeAPIError(c, http.StatusInternalServerError, "BROADCAST_SERVICE_UNAVAILABLE", "Broadcast service unavailable.")
		return
	}
	out, svcErr := s.BroadcastService.List(c.Request.Context(), eventID)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"broadcasts": out})
}

func (s *Server) postBroadcast(c *gin.Context) {
	uid, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to create broadcasts.")
		return
	}
	if s.BroadcastService == nil {
		writeAPIError(c, http.StatusInternalServerError, "BROADCAST_SERVICE_UNAVAILABLE", "Broadcast service unavailable.")
		return
	}
	var body broadcastBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Broadcast payload is invalid.")
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		writeAPIError(c, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header is required.")
		return
	}
	rawBody, _ := json.Marshal(body)
	fingerprint := hashFingerprint(eventID.String(), uid.String(), string(rawBody))
	ctx := c.Request.Context()
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "IDEMPOTENCY_FAILED", "Unable to validate idempotency key.")
		return
	}
	defer tx.Rollback(ctx)

	replay, err := reserveIdempotencyInTx(ctx, tx, "broadcast_create", idempotencyKey, fingerprint)
	if err != nil {
		if errors.Is(err, ErrIdempotencyFingerprintMismatch) {
			writeAPIError(c, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency key was already used for a different request.")
			return
		}
		writeAPIError(c, http.StatusInternalServerError, "IDEMPOTENCY_FAILED", "Unable to validate idempotency key.")
		return
	}
	if !replay {
		if svcErr := s.BroadcastService.CreateTx(ctx, tx, broadcastrepo.CreateInput{
			EventID:         eventID,
			CreatedByUserID: uid,
			Title:           body.Title,
			Body:            body.Body,
			ImageURL:        body.ImageURL,
			Audience:        body.Audience,
			Type:            body.Type,
		}); svcErr != nil {
			writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		writeAPIError(c, http.StatusInternalServerError, "INSERT_FAILED", "Failed to create broadcast.")
		return
	}
	s.invalidatePublicEventCache(ctx, eventID)
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

type organiserProfileBody struct {
	CompanyName string `json:"company_name" binding:"required"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
}

func (s *Server) postOrganiserProfile(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.OrganiserService == nil {
		writeAPIError(c, http.StatusInternalServerError, "ORGANISER_SERVICE_UNAVAILABLE", "Organiser service unavailable.")
		return
	}
	var body organiserProfileBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Organiser profile payload is invalid.")
		return
	}
	if svcErr := s.OrganiserService.UpsertProfile(c.Request.Context(), uid, body.CompanyName, body.Description, body.LogoURL); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) getOrganiserMe(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.OrganiserService == nil {
		writeAPIError(c, http.StatusInternalServerError, "ORGANISER_SERVICE_UNAVAILABLE", "Organiser service unavailable.")
		return
	}
	profile, svcErr := s.OrganiserService.GetMe(c.Request.Context(), uid)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (s *Server) getOrganiserEvents(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.OrganiserService == nil {
		writeAPIError(c, http.StatusInternalServerError, "ORGANISER_SERVICE_UNAVAILABLE", "Organiser service unavailable.")
		return
	}
	events, svcErr := s.OrganiserService.ListEvents(c.Request.Context(), uid)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

type organiserClientBody struct {
	Name         string `json:"name" binding:"required"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	Notes        string `json:"notes"`
}

func (s *Server) listOrganiserClients(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.OrganiserService == nil {
		writeAPIError(c, http.StatusInternalServerError, "ORGANISER_SERVICE_UNAVAILABLE", "Organiser service unavailable.")
		return
	}
	clients, svcErr := s.OrganiserService.ListClients(c.Request.Context(), uid)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"clients": clients})
}

func (s *Server) postOrganiserClient(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.OrganiserService == nil {
		writeAPIError(c, http.StatusInternalServerError, "ORGANISER_SERVICE_UNAVAILABLE", "Organiser service unavailable.")
		return
	}
	var body organiserClientBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Organiser client payload is invalid.")
		return
	}
	id, svcErr := s.OrganiserService.CreateClient(c.Request.Context(), uid, organiserrepo.ClientInput{
		Name:         body.Name,
		ContactEmail: body.ContactEmail,
		ContactPhone: body.ContactPhone,
		Notes:        body.Notes,
	})
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (s *Server) patchOrganiserClient(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	cid, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_CLIENT_ID", "Client id is invalid.")
		return
	}
	if s.OrganiserService == nil {
		writeAPIError(c, http.StatusInternalServerError, "ORGANISER_SERVICE_UNAVAILABLE", "Organiser service unavailable.")
		return
	}
	var body organiserClientBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Organiser client payload is invalid.")
		return
	}
	if svcErr := s.OrganiserService.UpdateClient(c.Request.Context(), uid, cid, organiserrepo.ClientInput{
		Name:         body.Name,
		ContactEmail: body.ContactEmail,
		ContactPhone: body.ContactPhone,
		Notes:        body.Notes,
	}); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type organiserClientEventBody struct {
	EventID string `json:"event_id" binding:"required"`
}

func (s *Server) postOrganiserClientEvent(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	cid, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_CLIENT_ID", "Client id is invalid.")
		return
	}
	if s.OrganiserService == nil {
		writeAPIError(c, http.StatusInternalServerError, "ORGANISER_SERVICE_UNAVAILABLE", "Organiser service unavailable.")
		return
	}
	var body organiserClientEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Organiser event-link payload is invalid.")
		return
	}
	eid, err := uuid.Parse(body.EventID)
	if err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_EVENT_ID", "Event id is invalid.")
		return
	}
	role, hasRole := s.eventRole(c.Request.Context(), uid, eid)
	canAccessEvent := hasRole && (role == "owner" || role == "co_owner" || role == "organiser")
	if svcErr := s.OrganiserService.LinkClientEvent(c.Request.Context(), uid, cid, eid, canAccessEvent); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (s *Server) postMemoryBookGenerate(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to generate memory book.")
		return
	}
	if s.MemoryBookService == nil {
		writeAPIError(c, http.StatusInternalServerError, "MEMORY_BOOK_SERVICE_UNAVAILABLE", "Memory book service unavailable.")
		return
	}
	result, svcErr := s.MemoryBookService.Generate(c.Request.Context(), eventID)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"slug":                 result.Slug,
		"public_api_path":      "/v1/public/memory/" + result.Slug,
		"payload":              result.Payload,
		"export_pdf_available": result.ExportPDFAvailable,
	})
}

func (s *Server) getPublicMemoryBook(c *gin.Context) {
	mslug := strings.TrimSpace(c.Param("slug"))
	if s.MemoryBookService == nil {
		writeAPIError(c, http.StatusInternalServerError, "MEMORY_BOOK_SERVICE_UNAVAILABLE", "Memory book service unavailable.")
		return
	}
	eid, payload, svcErr := s.MemoryBookService.GetPublic(c.Request.Context(), mslug)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"event_id": eid, "slug": mslug, "payload": payload})
}

func (s *Server) postMemoryBookExport(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You are not allowed to export memory book.")
		return
	}
	if s.MemoryBookService == nil {
		writeAPIError(c, http.StatusInternalServerError, "MEMORY_BOOK_SERVICE_UNAVAILABLE", "Memory book service unavailable.")
		return
	}
	if svcErr := s.MemoryBookService.Export(c.Request.Context(), eventID); svcErr != nil {
		if svcErr.Status == http.StatusPaymentRequired {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":         "tier_upgrade_required",
				"required_tier": "pro",
				"hint":          "Upgrade tier before PDF export is enabled.",
			})
			return
		}
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"stub":      true,
		"status":    "queued",
		"next_step": "Render payload to PDF and upload to object storage.",
	})
}

type billingCheckoutBody struct {
	Tier    string `json:"tier" binding:"required"`
	EventID string `json:"event_id"`
}

func tierPricePaise(tier string) int64 {
	return billingservice.TierPricePaise(tier)
}

func (s *Server) postBillingCheckout(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.BillingService == nil {
		writeAPIError(c, http.StatusInternalServerError, "BILLING_SERVICE_UNAVAILABLE", "Billing service unavailable.")
		return
	}
	var body billingCheckoutBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Checkout payload is invalid.")
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		writeAPIError(c, http.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY", "Idempotency-Key header is required.")
		return
	}
	fingerprint := hashFingerprint(uid.String(), strings.TrimSpace(strings.ToLower(body.Tier)), strings.TrimSpace(body.EventID))
	orderID := "order_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if len(orderID) > 40 {
		orderID = orderID[:40]
	}
	result, _, svcErr := s.BillingService.CreateCheckoutIdempotent(c.Request.Context(), uid, body.Tier, body.EventID, orderID, idempotencyKey, fingerprint)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":                 result.ID,
		"order_id":           result.OrderID,
		"key_id":             s.Config.RazorpayKeyID,
		"currency":           "INR",
		"amount_paise":       result.AmountPaise,
		"status":             "created",
		"razorpay_stub":      true,
		"webhook_ready":      true,
		"webhook_event_name": "payment.captured",
	})
}

func (s *Server) listBillingCheckouts(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.BillingService == nil {
		writeAPIError(c, http.StatusInternalServerError, "BILLING_SERVICE_UNAVAILABLE", "Billing service unavailable.")
		return
	}
	out, svcErr := s.BillingService.ListCheckouts(c.Request.Context(), uid)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"checkouts": out})
}

func verifyRazorpayWebhookSignature(secret string, payload []byte, sig string) bool {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(payload)
	got := m.Sum(nil)
	want, err := hexString(sig)
	if err != nil {
		return false
	}
	return hmac.Equal(got, want)
}

func hexString(s string) ([]byte, error) {
	const hex = "0123456789abcdef"
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("invalid hex length")
	}
	out := make([]byte, 0, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		hi := strings.IndexByte(hex, s[i])
		lo := strings.IndexByte(hex, s[i+1])
		if hi < 0 || lo < 0 {
			return nil, fmt.Errorf("invalid hex")
		}
		out = append(out, byte((hi<<4)|lo))
	}
	return out, nil
}

func (s *Server) postRazorpayWebhook(c *gin.Context) {
	const maxWebhookBody = 512 << 10
	payload, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBody+1))
	if err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Webhook payload is invalid.")
		return
	}
	if len(payload) > maxWebhookBody {
		writeAPIError(c, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE", "Webhook payload is too large.")
		return
	}
	sig := c.GetHeader("X-Razorpay-Signature")
	if !verifyRazorpayWebhookSignature(s.Config.RazorpayWebhookSecret, payload, sig) {
		writeAPIError(c, http.StatusUnauthorized, "BAD_SIGNATURE", "Webhook signature verification failed.")
		return
	}
	var body struct {
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					OrderID string `json:"order_id"`
				} `json:"entity"`
			} `json:"payment"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_JSON", "Webhook JSON payload is invalid.")
		return
	}
	if body.Event != "payment.captured" {
		c.JSON(http.StatusOK, gin.H{"ok": true, "ignored": true})
		return
	}
	orderID := strings.TrimSpace(body.Payload.Payment.Entity.OrderID)
	if orderID == "" {
		writeAPIError(c, http.StatusBadRequest, "MISSING_ORDER_ID", "Order id is required.")
		return
	}
	eventKey := strings.TrimSpace(c.GetHeader("X-Razorpay-Event-Id"))
	if eventKey == "" {
		eventKey = hashFingerprint(body.Event, orderID, sig)
	}
	payloadHash := hashFingerprint(string(payload))
	if s.BillingService == nil {
		writeAPIError(c, http.StatusInternalServerError, "BILLING_SERVICE_UNAVAILABLE", "Billing service unavailable.")
		return
	}
	var svcErr *billingservice.ServiceError
	backoff := 200 * time.Millisecond
	for attempt := 1; attempt <= 3; attempt++ {
		svcErr = s.BillingService.MarkOrderPaidFromWebhook(c.Request.Context(), "razorpay", eventKey, payloadHash, orderID)
		if svcErr == nil || svcErr.Status < 500 {
			break
		}
		if attempt < 3 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "order_id": orderID, "status": "paid"})
}
