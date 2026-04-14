package organiserservice

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/organiserrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo organiserrepo.Repository
}

func NewService(repo organiserrepo.Repository) *Service { return &Service{repo: repo} }

func (s *Service) UpsertProfile(ctx context.Context, userID uuid.UUID, companyName, description, logoURL string) *ServiceError {
	if strings.TrimSpace(companyName) == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_COMPANY_NAME", Message: "Company name is required."}
	}
	if err := s.repo.UpsertProfile(ctx, organiserrepo.ProfileInput{
		UserID:      userID,
		CompanyName: strings.TrimSpace(companyName),
		Description: strings.TrimSpace(description),
		LogoURL:     strings.TrimSpace(logoURL),
	}); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "UPSERT_FAILED", Message: "Unable to save organiser profile."}
	}
	return nil
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*organiserrepo.Profile, *ServiceError) {
	p, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Code: "NO_PROFILE", Message: "Organiser profile not found."}
	}
	return p, nil
}

func (s *Service) ListEvents(ctx context.Context, userID uuid.UUID) ([]organiserrepo.Event, *ServiceError) {
	oid, err := s.repo.FindOrganiserIDByUser(ctx, userID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Code: "NO_PROFILE", Message: "Organiser profile not found."}
	}
	events, err := s.repo.ListOrganiserEvents(ctx, oid)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load organiser events."}
	}
	return events, nil
}

func (s *Service) ListClients(ctx context.Context, userID uuid.UUID) ([]organiserrepo.Client, *ServiceError) {
	oid, err := s.repo.FindOrganiserIDByUser(ctx, userID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusNotFound, Code: "NO_PROFILE", Message: "Organiser profile not found."}
	}
	clients, err := s.repo.ListClients(ctx, oid)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load organiser clients."}
	}
	return clients, nil
}

func (s *Service) CreateClient(ctx context.Context, userID uuid.UUID, input organiserrepo.ClientInput) (string, *ServiceError) {
	if strings.TrimSpace(input.Name) == "" {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_NAME", Message: "Client name is required."}
	}
	oid, err := s.repo.FindOrganiserIDByUser(ctx, userID)
	if err != nil {
		return "", &ServiceError{Status: http.StatusNotFound, Code: "NO_PROFILE", Message: "Organiser profile not found."}
	}
	id, err := s.repo.CreateClient(ctx, oid, input)
	if err != nil {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "CREATE_FAILED", Message: "Unable to create organiser client."}
	}
	return id, nil
}

func (s *Service) UpdateClient(ctx context.Context, userID, clientID uuid.UUID, input organiserrepo.ClientInput) *ServiceError {
	if strings.TrimSpace(input.Name) == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_NAME", Message: "Client name is required."}
	}
	oid, err := s.repo.FindOrganiserIDByUser(ctx, userID)
	if err != nil {
		return &ServiceError{Status: http.StatusNotFound, Code: "NO_PROFILE", Message: "Organiser profile not found."}
	}
	ok, err := s.repo.UpdateClient(ctx, oid, clientID, input)
	if err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "UPDATE_FAILED", Message: "Unable to update organiser client."}
	}
	if !ok {
		return &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Organiser client not found."}
	}
	return nil
}

func (s *Service) LinkClientEvent(ctx context.Context, userID, clientID, eventID uuid.UUID, canAccessEvent bool) *ServiceError {
	oid, err := s.repo.FindOrganiserIDByUser(ctx, userID)
	if err != nil {
		return &ServiceError{Status: http.StatusNotFound, Code: "NO_PROFILE", Message: "Organiser profile not found."}
	}
	exists, err := s.repo.ClientExistsForOrganiser(ctx, oid, clientID)
	if err != nil || !exists {
		return &ServiceError{Status: http.StatusNotFound, Code: "CLIENT_NOT_FOUND", Message: "Organiser client not found."}
	}
	if !canAccessEvent {
		return &ServiceError{Status: http.StatusForbidden, Code: "EVENT_NOT_ACCESSIBLE", Message: "Event is not accessible."}
	}
	if err := s.repo.LinkClientEvent(ctx, clientID, eventID); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "LINK_FAILED", Message: "Unable to link event to client."}
	}
	return nil
}
