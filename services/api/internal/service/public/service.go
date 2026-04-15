package publicservice

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bhune/utsav/services/api/internal/cache"
	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/media"
	"github.com/bhune/utsav/services/api/internal/repository/publicrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo   publicrepo.Repository
	signer media.Signer
	cache  cache.Cache
}

func NewService(repo publicrepo.Repository, signer media.Signer, c cache.Cache) *Service {
	return &Service{repo: repo, signer: signer, cache: c}
}

func normalizeSlug(slug string) string {
	return strings.TrimSpace(strings.ToLower(slug))
}

// InvalidatePublicEventCache removes cached public payloads for this event (privacy, title, schedule, etc.).
func (s *Service) InvalidatePublicEventCache(ctx context.Context, eventID uuid.UUID) {
	if s.cache == nil {
		return
	}
	slug, err := s.repo.GetSlugByEventID(ctx, eventID)
	if err != nil || strings.TrimSpace(slug) == "" {
		return
	}
	ns := normalizeSlug(slug)
	if err := s.cache.Delete(ctx,
		"public:event:"+ns,
		"public:schedule:"+ns,
		"public:broadcasts:"+ns,
		"public:gallery:"+ns,
	); err != nil {
		log.Printf("public cache delete: %v", err)
	}
}

func (s *Service) GetEvent(ctx context.Context, slug string) (map[string]any, uuid.UUID, *ServiceError) {
	key := "public:event:" + normalizeSlug(slug)
	if s.cache != nil {
		if raw, err := s.cache.Get(ctx, key); err == nil {
			var payload struct {
				Event map[string]any `json:"event"`
				ID    string         `json:"id"`
			}
			if jsonErr := json.Unmarshal(raw, &payload); jsonErr == nil {
				if eid, parseErr := uuid.Parse(payload.ID); parseErr == nil {
					if pub, _ := payload.Event["privacy"].(string); strings.EqualFold(strings.TrimSpace(pub), "public") {
						return payload.Event, eid, nil
					}
					_ = s.cache.Delete(ctx, key)
				}
			}
		}
	}
	ev, err := s.repo.GetEventBySlug(ctx, normalizeSlug(slug))
	if err != nil {
		return nil, uuid.Nil, &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Public event not found."}
	}
	eventPayload := map[string]any{
		"id":              ev.ID.String(),
		"slug":            ev.Slug,
		"title":           ev.Title,
		"event_type":      ev.EventType,
		"privacy":         ev.Privacy,
		"toggles":         string(ev.Toggles),
		"branding":        string(ev.Branding),
		"couple_name_a":   ev.CoupleNameA,
		"couple_name_b":   ev.CoupleNameB,
		"love_story":      ev.LoveStory,
		"cover_image_url": ev.CoverImage,
		"date_start":      ev.DateStart,
		"date_end":        ev.DateEnd,
	}
	if s.cache != nil {
		raw, _ := json.Marshal(map[string]any{"event": eventPayload, "id": ev.ID.String()})
		_ = s.cache.Set(ctx, key, raw, 60*time.Second)
	}
	return eventPayload, ev.ID, nil
}

func (s *Service) ListSchedule(ctx context.Context, slug string) ([]map[string]any, uuid.UUID, *ServiceError) {
	_, eid, svcErr := s.GetEvent(ctx, slug)
	if svcErr != nil {
		return nil, uuid.Nil, svcErr
	}
	key := "public:schedule:" + normalizeSlug(slug)
	if s.cache != nil {
		if raw, err := s.cache.Get(ctx, key); err == nil {
			var payload struct {
				Rows []map[string]any `json:"rows"`
				ID   string           `json:"id"`
			}
			if jsonErr := json.Unmarshal(raw, &payload); jsonErr == nil {
				if cachedEID, parseErr := uuid.Parse(payload.ID); parseErr == nil && cachedEID == eid {
					return payload.Rows, eid, nil
				}
			}
		}
	}
	rows, err := s.repo.ListSubEvents(ctx, eid)
	if err != nil {
		return nil, uuid.Nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to fetch schedule."}
	}
	now := time.Now()
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		happening := false
		var startTime time.Time
		if t, ok := r.StartsAt.(time.Time); ok {
			startTime = t
		}
		if !startTime.IsZero() {
			endTime := startTime.Add(3 * time.Hour)
			if t, ok := r.EndsAt.(time.Time); ok && !t.IsZero() {
				endTime = t
			}
			happening = (now.Equal(startTime) || now.After(startTime)) && now.Before(endTime)
		}
		out = append(out, map[string]any{
			"id":            r.ID.String(),
			"name":          r.Name,
			"sub_type":      r.SubType,
			"starts_at":     r.StartsAt,
			"ends_at":       r.EndsAt,
			"venue_label":   r.VenueLabel,
			"dress_code":    r.DressCode,
			"description":   r.Description,
			"sort_order":    r.SortOrder,
			"happening_now": happening,
		})
	}
	if s.cache != nil {
		raw, _ := json.Marshal(map[string]any{"rows": out, "id": eid.String()})
		_ = s.cache.Set(ctx, key, raw, 30*time.Second)
	}
	return out, eid, nil
}

func (s *Service) ListBroadcasts(ctx context.Context, slug string) ([]map[string]any, *ServiceError) {
	_, eid, svcErr := s.GetEvent(ctx, slug)
	if svcErr != nil {
		return nil, svcErr
	}
	key := "public:broadcasts:" + normalizeSlug(slug)
	if s.cache != nil {
		if raw, err := s.cache.Get(ctx, key); err == nil {
			var out []map[string]any
			if jsonErr := json.Unmarshal(raw, &out); jsonErr == nil {
				return out, nil
			}
		}
	}
	rows, err := s.repo.ListBroadcasts(ctx, eid)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to fetch broadcasts."}
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"id":         r.ID.String(),
			"title":      r.Title,
			"body":       r.Body,
			"image_url":  r.ImageURL,
			"type":       r.Type,
			"created_at": r.CreatedAt,
		})
	}
	if s.cache != nil {
		raw, _ := json.Marshal(out)
		_ = s.cache.Set(ctx, key, raw, 60*time.Second)
	}
	return out, nil
}

func (s *Service) ListGallery(ctx context.Context, slug string) ([]map[string]any, *ServiceError) {
	_, eid, svcErr := s.GetEvent(ctx, slug)
	if svcErr != nil {
		return nil, svcErr
	}
	key := "public:gallery:" + normalizeSlug(slug)
	if s.cache != nil {
		if raw, err := s.cache.Get(ctx, key); err == nil {
			var out []map[string]any
			if jsonErr := json.Unmarshal(raw, &out); jsonErr == nil {
				return out, nil
			}
		}
	}
	rows, err := s.repo.ListApprovedGallery(ctx, eid)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to fetch gallery."}
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"id":         r.ID.String(),
			"section":    r.Section,
			"object_key": r.ObjectKey,
			"created_at": r.CreatedAt,
			"url":        s.signer.PublicObjectURL(r.ObjectKey),
		})
	}
	if s.cache != nil {
		raw, _ := json.Marshal(out)
		_ = s.cache.Set(ctx, key, raw, 120*time.Second)
	}
	return out, nil
}

func maskPhone(p string) string {
	if len(p) < 4 {
		return "****"
	}
	return "******" + p[len(p)-4:]
}

func (s *Service) BuildUPILink(ctx context.Context, slug string, guestEventID uuid.UUID, guestPhone string) (map[string]any, *ServiceError) {
	up, err := s.repo.GetUPIContextBySlug(ctx, normalizeSlug(slug))
	if err != nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Public event not found."}
	}
	if strings.TrimSpace(up.VPA) == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "HOST_VPA_NOT_CONFIGURED", Message: "Host UPI VPA is not configured."}
	}
	if guestEventID != up.EventID {
		return nil, &ServiceError{Status: http.StatusForbidden, Code: "WRONG_EVENT", Message: "Guest token does not match this event."}
	}
	note := "Shagun from guest for " + up.Title
	upi := "upi://pay?pa=" + url.QueryEscape(up.VPA) +
		"&pn=" + url.QueryEscape(up.Title) +
		"&tn=" + url.QueryEscape(note) +
		"&am=&cu=INR"
	return map[string]any{
		"upi_uri":            upi,
		"payee_vpa":          up.VPA,
		"transaction_note":   note,
		"guest_phone_masked": maskPhone(guestPhone),
	}, nil
}

func (s *Service) ReportShagun(ctx context.Context, slug string, guestEventID uuid.UUID, guestPhone string, amountINR float64, blessingNote string, subEventID *string) *ServiceError {
	resolved, err := s.repo.ResolveEventIDBySlug(ctx, normalizeSlug(slug))
	if err != nil {
		return &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Public event not found."}
	}
	if resolved != guestEventID {
		return &ServiceError{Status: http.StatusForbidden, Code: "WRONG_EVENT", Message: "Guest token does not match this event."}
	}
	eid := guestEventID
	paise := int64(math.Round(amountINR * 100))
	if paise <= 0 {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_AMOUNT", Message: "Shagun amount must be greater than zero."}
	}
	if err := s.repo.InsertGuestShagunReport(ctx, publicrepo.GuestShagunReportInput{
		EventID:      eid,
		AmountPaise:  paise,
		BlessingNote: strings.TrimSpace(blessingNote),
		SubEventID:   subEventID,
		GuestPhone:   guestPhone,
	}); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INSERT_FAILED", Message: "Failed to report shagun payment."}
	}
	return nil
}
