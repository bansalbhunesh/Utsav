package rsvpservice

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bhune/utsav/services/api/internal/auth"
	"github.com/bhune/utsav/services/api/internal/otp"
	"github.com/bhune/utsav/services/api/internal/ratelimit"
	"github.com/bhune/utsav/services/api/internal/repository/rsvprepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

type SubmitItemInput struct {
	SubEventID          string
	Status              string
	MealPref            string
	Dietary             string
	AccommodationNeeded bool
	TravelMode          string
	PlusOneNames        string
}

type Service struct {
	repo              rsvprepo.Repository
	otpRequestLimiter ratelimit.Limiter
	otpVerifyLimiter  ratelimit.Limiter
	submitLimiter     ratelimit.Limiter
	devOTP            string
	jwtSecret         []byte
	otpSecret         []byte
	env               string
	otpDispatcher     otp.Dispatcher
	otpMaxAttempts    int
}

var phoneE164Regex = regexp.MustCompile(`^\+?[1-9]\d{7,14}$`)

func NewService(
	repo rsvprepo.Repository,
	otpRequestLimiter ratelimit.Limiter,
	otpVerifyLimiter ratelimit.Limiter,
	submitLimiter ratelimit.Limiter,
	devOTPCode string,
	jwtSecret string,
	otpSecret string,
	env string,
	otpDispatcher otp.Dispatcher,
	otpMaxAttempts int,
) *Service {
	return &Service{
		repo:              repo,
		otpRequestLimiter: otpRequestLimiter,
		otpVerifyLimiter:  otpVerifyLimiter,
		submitLimiter:     submitLimiter,
		devOTP:            strings.TrimSpace(devOTPCode),
		jwtSecret:         []byte(jwtSecret),
		otpSecret:         []byte(otpSecret),
		env:               strings.TrimSpace(strings.ToLower(env)),
		otpDispatcher:     otpDispatcher,
		otpMaxAttempts:    otpMaxAttempts,
	}
}

func (s *Service) EventIDFromSlug(ctx context.Context, slug string) (uuid.UUID, *ServiceError) {
	clean := strings.TrimSpace(strings.ToLower(slug))
	eid, err := s.repo.FindEventIDBySlug(ctx, clean)
	if err != nil {
		if rsvprepo.IsNoRows(err) {
			return uuid.Nil, &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "Event not found."}
		}
		return uuid.Nil, &ServiceError{Status: http.StatusInternalServerError, Code: "EVENT_LOOKUP_FAILED", Message: "Unable to resolve event."}
	}
	return eid, nil
}

func (s *Service) RequestOTP(ctx context.Context, eventID uuid.UUID, slug, phone, clientIP string) *ServiceError {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone is required."}
	}
	if !phoneE164Regex.MatchString(phone) {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_PHONE", Message: "Phone number must be in E.164 format."}
	}
	if s.otpRequestLimiter != nil {
		allowed, err := s.otpRequestLimiter.Allow(ctx, "rsvp_otp_req:"+clientIP+"|"+strings.TrimSpace(strings.ToLower(slug))+"|"+phone)
		if err != nil {
			return &ServiceError{Status: http.StatusInternalServerError, Code: "RATE_LIMIT_FAILED", Message: "Unable to validate RSVP OTP rate limits."}
		}
		if !allowed {
			return &ServiceError{Status: http.StatusTooManyRequests, Code: "RATE_LIMITED", Message: "Too many RSVP OTP requests. Please retry later."}
		}
	}
	if s.env == "production" && s.devOTP != "" {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_CONFIG_INVALID", Message: "DEV_OTP_CODE must be disabled in production."}
	}
	code := s.devOTP
	if code == "" {
		var err error
		code, err = otp.GenerateNumericCode()
		if err != nil {
			return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_GENERATE_FAILED", Message: "Unable to generate RSVP OTP."}
		}
	}
	hash, err := otp.HashCode(s.otpSecret, code)
	if err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_HASH_FAILED", Message: "Unable to process RSVP OTP request."}
	}
	if err := s.repo.DeleteRSVPOTPChallenges(ctx, eventID, phone); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save RSVP OTP challenge."}
	}
	if err := s.repo.InsertRSVPOTPChallenge(ctx, eventID, phone, hash); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save RSVP OTP challenge."}
	}
	if s.otpDispatcher != nil {
		if err := s.otpDispatcher.DispatchOTP(ctx, phone, code); err != nil {
			_ = s.repo.DeleteRSVPOTPChallenges(ctx, eventID, phone)
			return &ServiceError{Status: http.StatusBadGateway, Code: "OTP_SEND_FAILED", Message: "Unable to send RSVP OTP."}
		}
	}
	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, eventID uuid.UUID, phone, code, clientIP string) (string, *ServiceError) {
	phone = strings.TrimSpace(phone)
	if phone == "" || code == "" {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone and code are required."}
	}
	if s.otpVerifyLimiter != nil {
		allowed, err := s.otpVerifyLimiter.Allow(ctx, "rsvp_otp_verify:"+eventID.String()+"|"+clientIP+"|"+phone)
		if err != nil {
			return "", &ServiceError{Status: http.StatusInternalServerError, Code: "RATE_LIMIT_FAILED", Message: "Unable to validate RSVP OTP rate limits."}
		}
		if !allowed {
			return "", &ServiceError{Status: http.StatusTooManyRequests, Code: "RATE_LIMITED", Message: "Too many RSVP OTP verify attempts. Please retry later."}
		}
	}
	ch, err := s.repo.GetLatestRSVPOTPChallenge(ctx, eventID, phone)
	if err != nil {
		if rsvprepo.IsNoRows(err) {
			return "", &ServiceError{Status: http.StatusUnauthorized, Code: "NO_CHALLENGE", Message: "No active RSVP OTP challenge found."}
		}
		return "", &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_READ_FAILED", Message: "Unable to read RSVP OTP challenge."}
	}
	if time.Now().After(ch.ExpiresAt) {
		return "", &ServiceError{Status: http.StatusUnauthorized, Code: "OTP_EXPIRED", Message: "RSVP OTP has expired."}
	}
	if s.otpMaxAttempts > 0 && ch.Attempts >= s.otpMaxAttempts {
		return "", &ServiceError{Status: http.StatusUnauthorized, Code: "OTP_LOCKED", Message: "RSVP OTP attempts exceeded. Request a new code."}
	}
	if !otp.VerifyCode(s.otpSecret, ch.CodeHash, code) {
		if incErr := s.repo.IncrementRSVPOTPAttempts(ctx, ch.ID); incErr != nil {
			return "", &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_STATE_FAILED", Message: "Unable to update RSVP OTP state."}
		}
		return "", &ServiceError{Status: http.StatusUnauthorized, Code: "INVALID_OTP", Message: "Invalid RSVP OTP code."}
	}
	_ = s.repo.DeleteRSVPOTPChallengeByID(ctx, ch.ID)

	tok, err := auth.SignGuestToken(eventID, phone, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return "", &ServiceError{Status: http.StatusInternalServerError, Code: "TOKEN_SIGN_FAILED", Message: "Unable to create guest access token."}
	}
	return tok, nil
}

func (s *Service) SubmitRSVP(ctx context.Context, eventID, guestEventID uuid.UUID, phone, clientIP, idempotencyKey, fingerprint string, items []SubmitItemInput) *ServiceError {
	if strings.TrimSpace(idempotencyKey) == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "MISSING_IDEMPOTENCY_KEY", Message: "Idempotency-Key header is required."}
	}
	if s.submitLimiter != nil {
		allowed, err := s.submitLimiter.Allow(ctx, "public_rsvp_submit:"+eventID.String()+"|"+clientIP+"|"+phone)
		if err != nil {
			return &ServiceError{Status: http.StatusInternalServerError, Code: "RATE_LIMIT_FAILED", Message: "Unable to validate RSVP submit rate limits."}
		}
		if !allowed {
			return &ServiceError{Status: http.StatusTooManyRequests, Code: "RATE_LIMITED", Message: "Too many RSVP submissions. Please retry later."}
		}
	}
	if eventID != guestEventID {
		return &ServiceError{Status: http.StatusForbidden, Code: "WRONG_EVENT", Message: "Guest token does not match this event."}
	}
	if len(items) == 0 {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "RSVP payload is invalid."}
	}

	validStatuses := map[string]bool{"yes": true, "no": true, "maybe": true}
	mapped := make([]rsvprepo.RSVPItem, 0, len(items))
	for _, it := range items {
		sid, err := uuid.Parse(it.SubEventID)
		if err != nil {
			return &ServiceError{Status: http.StatusBadRequest, Code: "BAD_SUB_EVENT", Message: "A sub event id is invalid."}
		}
		status := strings.ToLower(strings.TrimSpace(it.Status))
		if !validStatuses[status] {
			return &ServiceError{
				Status:  http.StatusBadRequest,
				Code:    "INVALID_RSVP_STATUS",
				Message: "Status must be yes, no, or maybe.",
			}
		}
		mapped = append(mapped, rsvprepo.RSVPItem{
			SubEventID:          sid,
			Status:              status,
			MealPref:            it.MealPref,
			Dietary:             it.Dietary,
			AccommodationNeeded: it.AccommodationNeeded,
			TravelMode:          it.TravelMode,
			PlusOneNames:        it.PlusOneNames,
		})
	}

	if err := s.repo.UpsertRSVPResponsesIdempotent(ctx, "public_rsvp_submit", idempotencyKey, fingerprint, eventID, phone, mapped); err != nil {
		switch {
		case errors.Is(err, rsvprepo.ErrIdempotencyConflict):
			return &ServiceError{Status: http.StatusConflict, Code: "IDEMPOTENCY_CONFLICT", Message: "Idempotency key was already used for a different request."}
		case errors.Is(err, rsvprepo.ErrInvalidSubEvents):
			return &ServiceError{Status: http.StatusBadRequest, Code: "BAD_SUB_EVENT", Message: "One or more sub-events are not part of this event."}
		default:
			return &ServiceError{Status: http.StatusBadRequest, Code: "RSVP_UPSERT_FAILED", Message: "Unable to save RSVP response."}
		}
	}
	return nil
}

func (s *Service) ListHostRSVPs(ctx context.Context, eventID uuid.UUID, limit, offset int) ([]rsvprepo.HostRSVPRow, *ServiceError) {
	rows, err := s.repo.ListHostRSVPs(ctx, eventID, limit, offset)
	if err != nil {
		return nil, &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    "QUERY_FAILED",
			Message: "Failed to load RSVP responses.",
		}
	}
	return rows, nil
}
