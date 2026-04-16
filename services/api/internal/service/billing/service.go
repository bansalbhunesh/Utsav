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

// CreateCheckoutIdempotent creates a checkout in the same database transaction as idempotency reservation.
func (s *Service) CreateCheckoutIdempotent(ctx context.Context, userID uuid.UUID, tier, eventID, orderID, idempotencyKey, fingerprint string) (*CreateCheckoutResult, bool, *ServiceError) {
	amount := TierPricePaise(tier)
	if amount <= 0 {
		return nil, false, &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_TIER", Message: "Billing tier is invalid."}
	}
	id, rid, replay, err := s.repo.CreateCheckoutIdempotent(ctx, "billing_checkout", idempotencyKey, fingerprint, billingrepo.CreateCheckoutInput{
		UserID:  userID,
		EventID: eventID,
		Tier:    tier,
		OrderID: orderID,
	})
	if err != nil {
		if errors.Is(err, billingrepo.ErrIdempotencyFingerprintMismatch) {
			return nil, false, &ServiceError{Status: http.StatusConflict, Code: "IDEMPOTENCY_CONFLICT", Message: "Idempotency key was already used for a different request."}
		}
		return nil, false, &ServiceError{Status: http.StatusBadRequest, Code: "CHECKOUT_FAILED", Message: "Unable to create checkout."}
	}
	return &CreateCheckoutResult{ID: id, OrderID: rid, AmountPaise: amount}, replay, nil
}

// MarkOrderPaidFromWebhook applies Razorpay webhook deduplication and order payment in one transaction.
func (s *Service) MarkOrderPaidFromWebhook(ctx context.Context, provider, eventKey, payloadHash, orderID string) *ServiceError {
	if strings.TrimSpace(orderID) == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "MISSING_ORDER_ID", Message: "Order id is required."}
	}
	if err := s.repo.MarkOrderPaidFromWebhook(ctx, provider, eventKey, payloadHash, orderID); err != nil {
		if errors.Is(err, billingrepo.ErrWebhookDedupePayloadMismatch) {
			return &ServiceError{Status: http.StatusConflict, Code: "WEBHOOK_DEDUPE_CONFLICT", Message: "Webhook event id was reused with a different payload."}
		}
		if errors.Is(err, billingrepo.ErrCheckoutNotFound) {
			return &ServiceError{Status: http.StatusNotFound, Code: "CHECKOUT_NOT_FOUND", Message: "Checkout not found for order."}
		}
		return &ServiceError{Status: http.StatusInternalServerError, Code: "BILLING_PERSIST_FAILED", Message: "Unable to persist billing update."}
	}
	return nil
}

// RetryPendingWebhookDeliveries replays failed/queued webhook rows.
func (s *Service) RetryPendingWebhookDeliveries(ctx context.Context, provider string, limit int) (int, *ServiceError) {
	processed, err := s.repo.RetryPendingWebhookDeliveries(ctx, provider, limit)
	if err != nil {
		return 0, &ServiceError{Status: http.StatusInternalServerError, Code: "BILLING_PERSIST_FAILED", Message: "Unable to process pending webhooks."}
	}
	return processed, nil
}
