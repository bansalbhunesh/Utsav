package rsvpservice

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/bhune/utsav/services/api/internal/auth"
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
	repo      rsvprepo.Repository
	otpWindow *ratelimit.Window
	devOTP    string
	jwtSecret []byte
}

func NewService(repo rsvprepo.Repository, otpWindow *ratelimit.Window, devOTPCode, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		otpWindow: otpWindow,
		devOTP:    devOTPCode,
		jwtSecret: []byte(jwtSecret),
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
	if phone == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone is required."}
	}
	if s.otpWindow != nil && !s.otpWindow.Allow("rsvp_otp:"+clientIP+"|"+strings.TrimSpace(strings.ToLower(slug))+"|"+phone) {
		return &ServiceError{Status: http.StatusTooManyRequests, Code: "RATE_LIMITED", Message: "Too many RSVP OTP requests. Please retry later."}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.devOTP), bcrypt.DefaultCost)
	if err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_HASH_FAILED", Message: "Unable to process RSVP OTP request."}
	}
	if err := s.repo.DeleteRSVPOTPChallenges(ctx, eventID, phone); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save RSVP OTP challenge."}
	}
	if err := s.repo.InsertRSVPOTPChallenge(ctx, eventID, phone, string(hash)); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save RSVP OTP challenge."}
	}
	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, eventID uuid.UUID, phone, code string) (string, *ServiceError) {
	if phone == "" || code == "" {
		return "", &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone and code are required."}
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
	if err := bcrypt.CompareHashAndPassword([]byte(ch.CodeHash), []byte(code)); err != nil {
		return "", &ServiceError{Status: http.StatusUnauthorized, Code: "INVALID_OTP", Message: "Invalid RSVP OTP code."}
	}
	_ = s.repo.DeleteRSVPOTPChallengeByID(ctx, ch.ID)

	tok, err := auth.SignGuestToken(eventID, phone, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return "", &ServiceError{Status: http.StatusInternalServerError, Code: "TOKEN_SIGN_FAILED", Message: "Unable to create guest access token."}
	}
	return tok, nil
}

func (s *Service) SubmitRSVP(ctx context.Context, eventID, guestEventID uuid.UUID, phone string, items []SubmitItemInput) *ServiceError {
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
