package memorybookservice

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/memorybookrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type GenerateResult struct {
	Slug               string
	Payload            map[string]any
	ExportPDFAvailable bool
}

type Service struct{ repo memorybookrepo.Repository }

func NewService(repo memorybookrepo.Repository) *Service { return &Service{repo: repo} }

func (s *Service) Generate(ctx context.Context, eventID uuid.UUID) (*GenerateResult, *ServiceError) {
	snap, err := s.repo.GetEventSnapshot(ctx, eventID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Event not found."}
	}
	highlights, _ := s.repo.GetHighlights(ctx, eventID)
	mbSlug := strings.TrimSpace(snap.Slug) + "-memory"
	payload := memorybookrepo.BuildPayload(eventID, snap, highlights)
	if err := s.repo.UpsertMemoryBook(ctx, eventID, mbSlug, payload); err != nil {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "GENERATE_FAILED", Message: "Failed to generate memory book."}
	}
	return &GenerateResult{
		Slug:               mbSlug,
		Payload:            payload,
		ExportPDFAvailable: snap.Tier != "free",
	}, nil
}

func (s *Service) GetPublic(ctx context.Context, slug string) (string, map[string]any, *ServiceError) {
	eid, payload, err := s.repo.GetMemoryBookBySlug(ctx, strings.TrimSpace(slug))
	if err != nil {
		return "", nil, &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Memory book not found."}
	}
	return eid.String(), payload, nil
}

func (s *Service) Export(ctx context.Context, eventID uuid.UUID) *ServiceError {
	tier, err := s.repo.GetEventTier(ctx, eventID)
	if err != nil {
		return &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Event not found."}
	}
	if tier == "free" {
		return &ServiceError{Status: http.StatusPaymentRequired, Code: "TIER_UPGRADE_REQUIRED", Message: "Upgrade tier before PDF export is enabled."}
	}
	return nil
}
