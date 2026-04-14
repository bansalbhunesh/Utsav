package publicservice

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

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
}

func NewService(repo publicrepo.Repository, signer media.Signer) *Service {
	return &Service{repo: repo, signer: signer}
}

func normalizeSlug(slug string) string {
	return strings.TrimSpace(strings.ToLower(slug))
}

func (s *Service) GetEvent(ctx context.Context, slug string) (map[string]any, uuid.UUID, *ServiceError) {
	ev, err := s.repo.GetEventBySlug(ctx, normalizeSlug(slug))
	if err != nil {
		return nil, uuid.Nil, &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Public event not found."}
	}
	return map[string]any{
		"id":             ev.ID.String(),
		"slug":           ev.Slug,
		"title":          ev.Title,
		"event_type":     ev.EventType,
		"privacy":        ev.Privacy,
		"toggles":        string(ev.Toggles),
		"branding":       string(ev.Branding),
		"couple_name_a":  ev.CoupleNameA,
		"couple_name_b":  ev.CoupleNameB,
		"love_story":     ev.LoveStory,
		"cover_image_url": ev.CoverImage,
		"date_start":     ev.DateStart,
		"date_end":       ev.DateEnd,
	}, ev.ID, nil
}

func (s *Service) ListSchedule(ctx context.Context, slug string) ([]map[string]any, uuid.UUID, *ServiceError) {
	_, eid, svcErr := s.GetEvent(ctx, slug)
	if svcErr != nil {
		return nil, uuid.Nil, svcErr
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
	return out, eid, nil
}

func (s *Service) ListBroadcasts(ctx context.Context, slug string) ([]map[string]any, *ServiceError) {
	_, eid, svcErr := s.GetEvent(ctx, slug)
	if svcErr != nil {
		return nil, svcErr
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
	return out, nil
}

func (s *Service) ListGallery(ctx context.Context, slug string) ([]map[string]any, *ServiceError) {
	_, eid, svcErr := s.GetEvent(ctx, slug)
	if svcErr != nil {
		return nil, svcErr
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
	_, eid, svcErr := s.GetEvent(ctx, slug)
	if svcErr != nil {
		return svcErr
	}
	if guestEventID != eid {
		return &ServiceError{Status: http.StatusForbidden, Code: "WRONG_EVENT", Message: "Guest token does not match this event."}
	}
	paise := int64(amountINR * 100)
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
