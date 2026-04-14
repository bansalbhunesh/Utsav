package eventservice

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/eventrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo eventrepo.Repository
}

func NewService(repo eventrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CheckSlug(ctx context.Context, slug string) (string, bool, *ServiceError) {
	clean := strings.TrimSpace(strings.ToLower(slug))
	if clean == "" {
		return "", false, &ServiceError{Status: http.StatusBadRequest, Code: "MISSING_SLUG", Message: "Slug query is required."}
	}
	available, err := s.repo.IsSlugAvailable(ctx, clean)
	if err != nil {
		return "", false, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to check slug availability."}
	}
	return clean, available, nil
}

func (s *Service) CreateEvent(ctx context.Context, input eventrepo.CreateEventInput) (string, string, *ServiceError) {
	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))
	if input.Slug == "" {
		return "", "", &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_SLUG", Message: "Slug is invalid."}
	}
	if input.EventType == "" {
		input.EventType = "wedding"
	}
	if input.Privacy == "" {
		input.Privacy = "public"
	}
	id, err := s.repo.CreateEventWithOwner(ctx, input)
	if err != nil {
		return "", "", &ServiceError{Status: http.StatusBadRequest, Code: "CREATE_FAILED", Message: "Unable to create event."}
	}
	return id, input.Slug, nil
}

func (s *Service) ListEvents(ctx context.Context, userID uuid.UUID) ([]eventrepo.EventListRow, *ServiceError) {
	rows, err := s.repo.ListEvents(ctx, userID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load events."}
	}
	return rows, nil
}

func (s *Service) GetEvent(ctx context.Context, eventID uuid.UUID) (*eventrepo.EventDetail, *ServiceError) {
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Event not found."}
	}
	return event, nil
}

func (s *Service) PatchEvent(ctx context.Context, eventID uuid.UUID, role string, input eventrepo.PatchEventInput) *ServiceError {
	if err := s.repo.PatchEvent(ctx, eventID, input); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "UPDATE_FAILED", Message: "Unable to update event."}
	}
	_ = role
	return nil
}

func (s *Service) CreateSubEvent(ctx context.Context, eventID uuid.UUID, input eventrepo.CreateSubEventInput) (string, *ServiceError) {
	id, err := s.repo.CreateSubEvent(ctx, eventID, input)
	if err != nil {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "CREATE_FAILED", Message: "Unable to create sub-event."}
	}
	return id, nil
}

func (s *Service) ListSubEvents(ctx context.Context, eventID uuid.UUID) ([]eventrepo.SubEventRow, *ServiceError) {
	rows, err := s.repo.ListSubEvents(ctx, eventID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load sub-events."}
	}
	return rows, nil
}

func (s *Service) InviteMember(ctx context.Context, eventID uuid.UUID, role, invitedPhone string) *ServiceError {
	if err := s.repo.InviteEventMember(ctx, eventID, role, invitedPhone); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVITE_FAILED", Message: "Unable to invite event member."}
	}
	return nil
}
