package broadcastservice

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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

func normalizeBroadcastInput(in broadcastrepo.CreateInput) (broadcastrepo.CreateInput, *ServiceError) {
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Body) == "" {
		return in, &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Broadcast title and body are required."}
	}
	if strings.TrimSpace(in.Type) == "" {
		in.Type = "general"
	}
	return in, nil
}

func (s *Service) Create(ctx context.Context, in broadcastrepo.CreateInput) *ServiceError {
	return s.createWithExecutor(in, func(normalized broadcastrepo.CreateInput) error {
		return s.repo.Create(ctx, normalized)
	})
}

func (s *Service) CreateTx(ctx context.Context, tx pgx.Tx, in broadcastrepo.CreateInput) *ServiceError {
	return s.createWithExecutor(in, func(normalized broadcastrepo.CreateInput) error {
		return s.repo.CreateTx(ctx, tx, normalized)
	})
}

func (s *Service) createWithExecutor(in broadcastrepo.CreateInput, createFn func(in broadcastrepo.CreateInput) error) *ServiceError {
	normalized, svcErr := normalizeBroadcastInput(in)
	if svcErr != nil {
		return svcErr
	}
	if err := createFn(normalized); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INSERT_FAILED", Message: "Failed to create broadcast."}
	}
	return nil
}
