package billingservice

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/repository/billingrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type CreateCheckoutResult struct {
	ID          string
	OrderID     string
	AmountPaise int64
}

type Service struct {
	repo billingrepo.Repository
}

func NewService(repo billingrepo.Repository) *Service {
	return &Service{repo: repo}
}

func TierPricePaise(tier string) int64 {
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case "pro":
		return 99000
	case "elite":
		return 249000
	default:
		return 0
	}
}

func (s *Service) CreateCheckout(ctx context.Context, userID uuid.UUID, tier, eventID, orderID string) (*CreateCheckoutResult, *ServiceError) {
	amount := TierPricePaise(tier)
	if amount <= 0 {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_TIER", Message: "Billing tier is invalid."}
	}
	id, err := s.repo.CreateCheckout(ctx, billingrepo.CreateCheckoutInput{
		UserID:  userID,
		EventID: eventID,
		Tier:    tier,
		OrderID: orderID,
	})
	if err != nil {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "CHECKOUT_FAILED", Message: "Unable to create checkout."}
	}
	return &CreateCheckoutResult{ID: id, OrderID: orderID, AmountPaise: amount}, nil
}

func (s *Service) ListCheckouts(ctx context.Context, userID uuid.UUID) ([]billingrepo.Checkout, *ServiceError) {
	out, err := s.repo.ListCheckouts(ctx, userID)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "QUERY_FAILED", Message: "Failed to load checkouts."}
	}
	return out, nil
}

func (s *Service) MarkOrderPaid(ctx context.Context, orderID string) *ServiceError {
	if strings.TrimSpace(orderID) == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "MISSING_ORDER_ID", Message: "Order id is required."}
	}
	if _, _, err := s.repo.MarkOrderPaidAndFetch(ctx, orderID); err != nil {
		if errors.Is(err, billingrepo.ErrCheckoutNotFound) {
			return &ServiceError{Status: http.StatusNotFound, Code: "CHECKOUT_NOT_FOUND", Message: "Checkout not found for order."}
		}
		return &ServiceError{Status: http.StatusInternalServerError, Code: "BILLING_PERSIST_FAILED", Message: "Unable to persist billing update."}
	}
	return nil
}
