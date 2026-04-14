package vendorservice

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/vendorrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo vendorrepo.Repository
}

func NewService(repo vendorrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListVendors(ctx context.Context, eventID uuid.UUID) ([]vendorrepo.Vendor, *ServiceError) {
	rows, err := s.repo.ListVendors(ctx, eventID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load vendors."}
	}
	return rows, nil
}

func (s *Service) CreateVendor(ctx context.Context, eventID uuid.UUID, input vendorrepo.CreateInput) (string, *ServiceError) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Vendor name is required."}
	}
	id, err := s.repo.CreateVendor(ctx, eventID, input)
	if err != nil {
		return "", &ServiceError{Status: http.StatusInternalServerError, Code: "INSERT_FAILED", Message: "Unable to create vendor."}
	}
	return id, nil
}

func (s *Service) DeleteVendor(ctx context.Context, eventID uuid.UUID, vendorID string) *ServiceError {
	vid, err := uuid.Parse(vendorID)
	if err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_ID", Message: "Vendor id is invalid."}
	}
	if err := s.repo.DeleteVendor(ctx, eventID, vid); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "DELETE_FAILED", Message: "Unable to delete vendor."}
	}
	return nil
}
