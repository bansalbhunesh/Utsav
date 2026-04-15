package shagunservice

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/shagunrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type Service struct {
	repo shagunrepo.Repository
}

func NewService(repo shagunrepo.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LogCashShagun(ctx context.Context, eventID uuid.UUID, input shagunrepo.CashShagunInput) *ServiceError {
	if input.AmountINR <= 0 {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Amount must be greater than zero."}
	}
	if err := s.repo.LogCashShagun(ctx, eventID, input); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INSERT_FAILED", Message: "Unable to log cash shagun."}
	}
	return nil
}

func (s *Service) ListHostShagun(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]shagunrepo.HostShagunRow, *ServiceError) {
	rows, err := s.repo.ListHostShagun(ctx, eventID, limit, offset)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load shagun entries."}
	}
	return rows, nil
}
