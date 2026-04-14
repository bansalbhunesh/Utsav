package httpserver

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/media"
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

func sanitizeFileName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	repl := strings.NewReplacer(" ", "-", "..", "", "\\", "", "/", "", ":", "", ";", "")
	s = repl.Replace(s)
	if s == "" {
		return "asset"
	}
	return s
}

func (s *Server) postGalleryPresign(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body galleryPresignBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	key := fmt.Sprintf("events/%s/gallery/%d-%s", eventID.String(), time.Now().Unix(), sanitizeFileName(body.FileName))
	resp, err := s.MediaSigner.PresignPut(media.PresignRequest{
		ObjectKey:   key,
		ContentType: body.ContentType,
		ExpiresIn:   10 * time.Minute,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "presign_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"upload": resp})
}

func (s *Server) postGalleryAsset(c *gin.Context) {
	uid, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body galleryAssetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	var sid any
	if body.SubEventID != "" {
		if u, err := uuid.Parse(body.SubEventID); err == nil {
			sid = u
		}
	}
	ctx := c.Request.Context()
	status := strings.TrimSpace(strings.ToLower(body.Status))
	if status == "" {
		status = "pending"
	}
	if status != "pending" && status != "approved" && status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status"})
		return
	}
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO gallery_assets (event_id, section, object_key, uploader_user_id, sub_event_id, status, mime_type, bytes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		eventID, body.Section, body.ObjectKey, uid, sid, status, nullString(body.MimeType), body.Bytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insert_failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (s *Server) listGalleryAssets(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	status := strings.TrimSpace(strings.ToLower(c.Query("status")))
	ctx := c.Request.Context()
	q := `
		SELECT id, section, object_key, status, mime_type, bytes, created_at
		FROM gallery_assets
		WHERE event_id=$1`
	args := []any{eventID}
	if status != "" {
		q += ` AND status=$2`
		args = append(args, status)
	}
	q += ` ORDER BY created_at DESC LIMIT 200`
	rows, err := s.Pool.Query(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var sec, key, st string
		var mt any
		var bytes any
		var created any
		_ = rows.Scan(&id, &sec, &key, &st, &mt, &bytes, &created)
		out = append(out, gin.H{
			"id": id.String(), "section": sec, "object_key": key, "status": st,
			"mime_type": mt, "bytes": bytes, "created_at": created,
			"url": s.MediaSigner.PublicObjectURL(key),
		})
	}
	c.JSON(http.StatusOK, gin.H{"assets": out})
}

func (s *Server) patchGalleryAssetModeration(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	aid, err := uuid.Parse(c.Param("assetId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_asset_id"})
		return
	}
	var body galleryModerateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	status := strings.TrimSpace(strings.ToLower(body.Status))
	if status != "approved" && status != "rejected" && status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status"})
		return
	}
	ctx := c.Request.Context()
	tag, err := s.Pool.Exec(ctx, `
		UPDATE gallery_assets SET status=$1
		WHERE id=$2 AND event_id=$3`, status, aid, eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "update_failed"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
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
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ctx := c.Request.Context()
	rows, err := s.Pool.Query(ctx, `
		SELECT id, title, body, image_url, audience, announcement_type, created_at
		FROM broadcasts
		WHERE event_id=$1
		ORDER BY created_at DESC`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var title, body, atype string
		var img any
		var audience []byte
		var created any
		_ = rows.Scan(&id, &title, &body, &img, &audience, &atype, &created)
		out = append(out, gin.H{
			"id": id.String(), "title": title, "body": body, "image_url": img,
			"audience": string(audience), "announcement_type": atype, "created_at": created,
		})
	}
	c.JSON(http.StatusOK, gin.H{"broadcasts": out})
}

func (s *Server) postBroadcast(c *gin.Context) {
	uid, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageFinancials(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	var body broadcastBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	if body.Type == "" {
		body.Type = "general"
	}
	ctx := c.Request.Context()
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO broadcasts (event_id, title, body, image_url, audience, announcement_type, created_by_user_id)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7)`,
		eventID, body.Title, body.Body, nullString(body.ImageURL), mustJSONB(body.Audience), body.Type, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insert_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
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
	var body organiserProfileBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	ctx := c.Request.Context()
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO organiser_profiles (user_id, company_name, description, logo_url)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id) DO UPDATE SET company_name=EXCLUDED.company_name, description=EXCLUDED.description, logo_url=EXCLUDED.logo_url`,
		uid, body.CompanyName, body.Description, nullString(body.LogoURL))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "upsert_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) getOrganiserMe(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	row := s.Pool.QueryRow(ctx, `
		SELECT id, company_name, description, logo_url, verified FROM organiser_profiles WHERE user_id=$1`, uid)
	var id uuid.UUID
	var name, desc, logo string
	var verified bool
	if err := row.Scan(&id, &name, &desc, &logo, &verified); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no_profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id.String(), "company_name": name, "description": desc, "logo_url": logo, "verified": verified})
}

func (s *Server) getOrganiserEvents(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	var oid uuid.UUID
	if err := s.Pool.QueryRow(ctx, `SELECT id FROM organiser_profiles WHERE user_id=$1`, uid).Scan(&oid); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no_profile"})
		return
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT e.id, e.slug, e.title, e.date_start
		FROM events e
		JOIN organiser_client_events oce ON oce.event_id=e.id
		JOIN organiser_clients oc ON oc.id=oce.organiser_client_id
		WHERE oc.organiser_id=$1`, oid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var slug, title string
		var ds any
		_ = rows.Scan(&id, &slug, &title, &ds)
		out = append(out, gin.H{"id": id.String(), "slug": slug, "title": title, "date_start": ds})
	}
	c.JSON(http.StatusOK, gin.H{"events": out})
}

func (s *Server) organiserIDByUser(ctx context.Context, uid uuid.UUID) (uuid.UUID, error) {
	var oid uuid.UUID
	err := s.Pool.QueryRow(ctx, `SELECT id FROM organiser_profiles WHERE user_id=$1`, uid).Scan(&oid)
	return oid, err
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
	ctx := c.Request.Context()
	oid, err := s.organiserIDByUser(ctx, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no_profile"})
		return
	}
	rows, err := s.Pool.Query(ctx, `
		SELECT id, name, contact_email, contact_phone, notes, created_at
		FROM organiser_clients
		WHERE organiser_id=$1
		ORDER BY created_at DESC`, oid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var name string
		var email, phone, notes, created any
		_ = rows.Scan(&id, &name, &email, &phone, &notes, &created)
		out = append(out, gin.H{
			"id": id.String(), "name": name, "contact_email": email,
			"contact_phone": phone, "notes": notes, "created_at": created,
		})
	}
	c.JSON(http.StatusOK, gin.H{"clients": out})
}

func (s *Server) postOrganiserClient(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body organiserClientBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	ctx := c.Request.Context()
	oid, err := s.organiserIDByUser(ctx, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no_profile"})
		return
	}
	var id uuid.UUID
	err = s.Pool.QueryRow(ctx, `
		INSERT INTO organiser_clients (organiser_id, name, contact_email, contact_phone, notes)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		oid, body.Name, nullString(body.ContactEmail), nullString(body.ContactPhone), nullString(body.Notes)).
		Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "create_failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id.String()})
}

func (s *Server) patchOrganiserClient(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	cid, err := uuid.Parse(c.Param("clientId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client_id"})
		return
	}
	var body organiserClientBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	ctx := c.Request.Context()
	oid, err := s.organiserIDByUser(ctx, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no_profile"})
		return
	}
	tag, err := s.Pool.Exec(ctx, `
		UPDATE organiser_clients
		SET name=$1, contact_email=$2, contact_phone=$3, notes=$4
		WHERE id=$5 AND organiser_id=$6`,
		body.Name, nullString(body.ContactEmail), nullString(body.ContactPhone), nullString(body.Notes), cid, oid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "update_failed"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client_id"})
		return
	}
	var body organiserClientEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	eid, err := uuid.Parse(body.EventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_event_id"})
		return
	}
	ctx := c.Request.Context()
	oid, err := s.organiserIDByUser(ctx, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no_profile"})
		return
	}
	var exists int
	if err := s.Pool.QueryRow(ctx, `
		SELECT 1 FROM organiser_clients WHERE id=$1 AND organiser_id=$2`, cid, oid).Scan(&exists); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client_not_found"})
		return
	}
	if role, ok := s.eventRole(ctx, uid, eid); !ok || (role != "owner" && role != "co_owner" && role != "organiser") {
		c.JSON(http.StatusForbidden, gin.H{"error": "event_not_accessible"})
		return
	}
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO organiser_client_events (organiser_client_id, event_id)
		VALUES ($1,$2)
		ON CONFLICT DO NOTHING`, cid, eid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "link_failed"})
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
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ctx := c.Request.Context()
	var slug, title, tier string
	var ds, de any
	if err := s.Pool.QueryRow(ctx, `
		SELECT slug, title, date_start, date_end, tier FROM events WHERE id=$1`, eventID).
		Scan(&slug, &title, &ds, &de, &tier); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	var guestsTotal, rsvpTotal, rsvpYes, shagunCount, galleryApproved, broadcastsCount int64
	var shagunPaise int64
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM guests WHERE event_id=$1`, eventID).Scan(&guestsTotal)
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM rsvp_responses WHERE event_id=$1`, eventID).Scan(&rsvpTotal)
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM rsvp_responses WHERE event_id=$1 AND status='yes'`, eventID).Scan(&rsvpYes)
	_ = s.Pool.QueryRow(ctx, `SELECT count(*), coalesce(sum(amount_paise),0) FROM shagun_entries WHERE event_id=$1`, eventID).
		Scan(&shagunCount, &shagunPaise)
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM gallery_assets WHERE event_id=$1 AND status='approved'`, eventID).Scan(&galleryApproved)
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM broadcasts WHERE event_id=$1`, eventID).Scan(&broadcastsCount)

	mbSlug := slug + "-memory"
	now := time.Now().UTC().Format(time.RFC3339)
	payload := map[string]any{
		"version":      1,
		"generated_at": now,
		"event": map[string]any{
			"id":         eventID.String(),
			"slug":       slug,
			"title":      title,
			"date_start": ds,
			"date_end":   de,
			"tier":       tier,
		},
		"highlights": map[string]any{
			"guest_count":        guestsTotal,
			"rsvp_count":         rsvpTotal,
			"rsvp_yes_count":     rsvpYes,
			"shagun_count":       shagunCount,
			"shagun_total_paise": shagunPaise,
			"gallery_assets":     galleryApproved,
			"broadcasts":         broadcastsCount,
		},
	}
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO memory_books (event_id, slug, payload)
		VALUES ($1,$2,$3::jsonb)
		ON CONFLICT (slug) DO UPDATE SET payload=EXCLUDED.payload, generated_at=now()`,
		eventID, mbSlug, mustJSONB(payload))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "generate_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"slug": mbSlug, "public_api_path": "/v1/public/memory/" + mbSlug, "payload": payload,
		"export_pdf_available": tier != "free",
	})
}

func (s *Server) getPublicMemoryBook(c *gin.Context) {
	mslug := strings.TrimSpace(c.Param("slug"))
	ctx := c.Request.Context()
	var payload []byte
	var eid uuid.UUID
	err := s.Pool.QueryRow(ctx, `SELECT event_id, payload FROM memory_books WHERE slug=$1`, mslug).Scan(&eid, &payload)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	var decoded any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"event_id": eid.String(), "slug": mslug, "payload": decoded})
}

func (s *Server) postMemoryBookExport(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	ctx := c.Request.Context()
	var tier string
	if err := s.Pool.QueryRow(ctx, `SELECT tier FROM events WHERE id=$1`, eventID).Scan(&tier); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if tier == "free" {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":         "tier_upgrade_required",
			"required_tier": "pro",
			"hint":          "Upgrade tier before PDF export is enabled.",
		})
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
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case "pro":
		return 99000
	case "elite":
		return 249000
	default:
		return 0
	}
}

func (s *Server) postBillingCheckout(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	var body billingCheckoutBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	var eid any
	if body.EventID != "" {
		if u, err := uuid.Parse(body.EventID); err == nil {
			eid = u
		}
	}
	ctx := c.Request.Context()
	var id uuid.UUID
	orderID := "order_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if len(orderID) > 40 {
		orderID = orderID[:40]
	}
	amount := tierPricePaise(body.Tier)
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO billing_checkouts (user_id, event_id, tier, razorpay_order_id, status)
		VALUES ($1,$2,$3,$4,'created') RETURNING id`,
		uid, eid, body.Tier, orderID).Scan(&id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "checkout_failed", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":                 id.String(),
		"order_id":           orderID,
		"key_id":             s.Config.RazorpayKeyID,
		"currency":           "INR",
		"amount_paise":       amount,
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
	rows, err := s.Pool.Query(c.Request.Context(), `
		SELECT id, event_id, tier, razorpay_order_id, status, created_at
		FROM billing_checkouts
		WHERE user_id=$1
		ORDER BY created_at DESC`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}
	defer rows.Close()
	out := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var eventID any
		var tier, orderID, status string
		var created any
		_ = rows.Scan(&id, &eventID, &tier, &orderID, &status, &created)
		out = append(out, gin.H{
			"id": id.String(), "event_id": eventID, "tier": tier, "order_id": orderID, "status": status, "created_at": created,
		})
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
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	sig := c.GetHeader("X-Razorpay-Signature")
	if !verifyRazorpayWebhookSignature(s.Config.RazorpayWebhookSecret, payload, sig) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad_signature"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	if body.Event != "payment.captured" {
		c.JSON(http.StatusOK, gin.H{"ok": true, "ignored": true})
		return
	}
	orderID := strings.TrimSpace(body.Payload.Payment.Entity.OrderID)
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_order_id"})
		return
	}
	ctx := c.Request.Context()
	tag, err := s.Pool.Exec(ctx, `UPDATE billing_checkouts SET status='paid' WHERE razorpay_order_id=$1`, orderID)
	if err != nil || tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "checkout_not_found"})
		return
	}
	var tier string
	var eventID any
	_ = s.Pool.QueryRow(ctx, `
		SELECT tier, event_id FROM billing_checkouts WHERE razorpay_order_id=$1`, orderID).Scan(&tier, &eventID)
	if eventID != nil {
		_, _ = s.Pool.Exec(ctx, `UPDATE events SET tier=$1 WHERE id=$2`, tier, eventID)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "order_id": orderID, "status": "paid"})
}
