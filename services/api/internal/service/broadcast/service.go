package broadcastservice

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/broadcastrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct{ repo broadcastrepo.Repository }

func NewService(repo broadcastrepo.Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, eventID uuid.UUID) ([]broadcastrepo.Broadcast, *ServiceError) {
	out, err := s.repo.List(ctx, eventID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to list broadcasts."}
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, in broadcastrepo.CreateInput) *ServiceError {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Body) == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Broadcast title and body are required."}
	}
	if strings.TrimSpace(in.Type) == "" {
		in.Type = "general"
	}
	if err := s.repo.Create(ctx, in); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INSERT_FAILED", Message: "Failed to create broadcast."}
	}
	return nil
}
