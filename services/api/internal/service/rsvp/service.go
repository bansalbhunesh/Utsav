package rsvpservice

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

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
	env               string
	otpSender         otp.Sender
	otpMaxAttempts    int
}

func NewService(
	repo rsvprepo.Repository,
	otpRequestLimiter ratelimit.Limiter,
	otpVerifyLimiter ratelimit.Limiter,
	submitLimiter ratelimit.Limiter,
	devOTPCode string,
	jwtSecret string,
	env string,
	otpSender otp.Sender,
	otpMaxAttempts int,
) *Service {
	return &Service{
		repo:              repo,
		otpRequestLimiter: otpRequestLimiter,
		otpVerifyLimiter:  otpVerifyLimiter,
		submitLimiter:     submitLimiter,
		devOTP:            strings.TrimSpace(devOTPCode),
		jwtSecret:         []byte(jwtSecret),
		env:               strings.TrimSpace(strings.ToLower(env)),
		otpSender:         otpSender,
		otpMaxAttempts:    otpMaxAttempts,
	}
}

func generateNumericOTP() (string, error) {
	raw := make([]byte, 4)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	n := int(raw[0])<<24 | int(raw[1])<<16 | int(raw[2])<<8 | int(raw[3])
	if n < 0 {
		n = -n
	}
	return fmt.Sprintf("%06d", n%1000000), nil
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
	if phone == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone is required."}
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
		code, err = generateNumericOTP()
		if err != nil {
			return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_GENERATE_FAILED", Message: "Unable to generate RSVP OTP."}
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_HASH_FAILED", Message: "Unable to process RSVP OTP request."}
	}
	if err := s.repo.DeleteRSVPOTPChallenges(ctx, eventID, phone); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save RSVP OTP challenge."}
	}
	if err := s.repo.InsertRSVPOTPChallenge(ctx, eventID, phone, string(hash)); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save RSVP OTP challenge."}
	}
	if s.otpSender != nil {
		if err := s.otpSender.SendOTP(ctx, phone, code); err != nil {
			return &ServiceError{Status: http.StatusBadGateway, Code: "OTP_SEND_FAILED", Message: "Unable to send RSVP OTP."}
		}
	}
	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, eventID uuid.UUID, phone, code, clientIP string) (string, *ServiceError) {
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
	if err := bcrypt.CompareHashAndPassword([]byte(ch.CodeHash), []byte(code)); err != nil {
		_ = s.repo.IncrementRSVPOTPAttempts(ctx, ch.ID)
		return "", &ServiceError{Status: http.StatusUnauthorized, Code: "INVALID_OTP", Message: "Invalid RSVP OTP code."}
	}
	_ = s.repo.DeleteRSVPOTPChallengeByID(ctx, ch.ID)

	tok, err := auth.SignGuestToken(eventID, phone, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return "", &ServiceError{Status: http.StatusInternalServerError, Code: "TOKEN_SIGN_FAILED", Message: "Unable to create guest access token."}
	}
	return tok, nil
}

func (s *Service) SubmitRSVP(ctx context.Context, eventID, guestEventID uuid.UUID, phone, clientIP string, items []SubmitItemInput) *ServiceError {
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

	mapped := make([]rsvprepo.RSVPItem, 0, len(items))
	for _, it := range items {
		sid, err := uuid.Parse(it.SubEventID)
		if err != nil {
			return &ServiceError{Status: http.StatusBadRequest, Code: "BAD_SUB_EVENT", Message: "A sub event id is invalid."}
		}
		mapped = append(mapped, rsvprepo.RSVPItem{
			SubEventID:          sid,
			Status:              it.Status,
			MealPref:            it.MealPref,
			Dietary:             it.Dietary,
			AccommodationNeeded: it.AccommodationNeeded,
			TravelMode:          it.TravelMode,
			PlusOneNames:        it.PlusOneNames,
		})
	}

	if err := s.repo.UpsertRSVPResponses(ctx, eventID, phone, mapped); err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "RSVP_UPSERT_FAILED", Message: "Unable to save RSVP response."}
	}
	return nil
}

func (s *Service) ListHostRSVPs(ctx context.Context, eventID uuid.UUID) ([]rsvprepo.HostRSVPRow, *ServiceError) {
	rows, err := s.repo.ListHostRSVPs(ctx, eventID)
	if err != nil {
		return nil, &ServiceError{
			Status:  http.StatusInternalServerError,
			Code:    "QUERY_FAILED",
			Message: "Failed to load RSVP responses.",
		}
	}
	return rows, nil
}
