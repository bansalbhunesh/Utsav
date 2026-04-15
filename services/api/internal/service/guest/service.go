package guestservice

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/guestrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo guestrepo.Repository
}

func NewService(repo guestrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListGuests(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]guestrepo.Guest, *ServiceError) {
	list, err := s.repo.ListGuests(ctx, eventID, limit, offset)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load guests."}
	}
	return list, nil
}

func (s *Service) UpsertGuest(ctx context.Context, eventID uuid.UUID, input guestrepo.GuestInput) (string, *ServiceError) {
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	if input.Name == "" || input.Phone == "" {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Name and phone are required."}
	}
	guestID, err := s.repo.UpsertGuest(ctx, eventID, input)
	if err != nil {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "UPSERT_FAILED", Message: "Unable to save guest."}
	}
	return guestID, nil
}

func (s *Service) ImportGuestsCSV(ctx context.Context, eventID uuid.UUID, csv string) (*guestrepo.ImportResult, *ServiceError) {
	if strings.TrimSpace(csv) == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "EMPTY_CSV", Message: "CSV payload cannot be empty."}
	}
	result, err := s.repo.ImportGuestsCSV(ctx, eventID, csv)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "CSV_IMPORT_FAILED", Message: "Unable to import guest CSV."}
	}
	return result, nil
}
