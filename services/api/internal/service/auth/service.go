package authservice

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	authtoken "github.com/bhune/utsav/services/api/internal/auth"
	"github.com/bhune/utsav/services/api/internal/ratelimit"
	"github.com/bhune/utsav/services/api/internal/repository/authrepo"
)

type ServiceError struct {
	Status  int
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

type OTPVerifyResult struct {
	AccessToken  string
	RefreshToken string
	UserID       string
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

type MeResult struct {
	ID          string
	Phone       string
	DisplayName string
}

type Service struct {
	repo      authrepo.Repository
	otpWindow *ratelimit.Window
	devOTP    string
	jwtSecret []byte
}

func NewService(repo authrepo.Repository, otpWindow *ratelimit.Window, devOTPCode, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		otpWindow: otpWindow,
		devOTP:    devOTPCode,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) RequestOTP(ctx context.Context, phone, clientIP string) *ServiceError {
	if phone == "" {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone is required."}
	}
	if s.otpWindow != nil && !s.otpWindow.Allow("auth_otp:"+clientIP) {
		return &ServiceError{
			Status:  http.StatusTooManyRequests,
			Code:    "RATE_LIMITED",
			Message: "Too many OTP requests. Please retry later.",
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(s.devOTP), bcrypt.DefaultCost)
	if err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_HASH_FAILED", Message: "Unable to process OTP request."}
	}
	if err := s.repo.DeletePhoneOTPChallenges(ctx, phone); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save OTP challenge."}
	}
	if err := s.repo.InsertPhoneOTPChallenge(ctx, phone, string(hash)); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save OTP challenge."}
	}
	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, phone, code string) (*OTPVerifyResult, *ServiceError) {
	if phone == "" || code == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone and code are required."}
	}
	ch, err := s.repo.GetLatestPhoneOTPChallenge(ctx, phone)
	if err != nil {
		if authrepo.IsNoRows(err) {
			return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "NO_CHALLENGE", Message: "No active OTP challenge found."}
		}
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_READ_FAILED", Message: "Unable to read OTP challenge."}
	}
	if time.Now().After(ch.ExpiresAt) {
		return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "OTP_EXPIRED", Message: "OTP has expired. Request a new code."}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(ch.CodeHash), []byte(code)); err != nil {
		_ = s.repo.IncrementPhoneOTPAttempts(ctx, ch.ID)
		return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "INVALID_OTP", Message: "The OTP code is invalid."}
	}
	_ = s.repo.DeletePhoneOTPChallengeByID(ctx, ch.ID)

	userID, err := s.repo.FindUserIDByPhone(ctx, phone)
	if err != nil {
		if authrepo.IsNoRows(err) {
			userID, err = s.repo.CreateUserWithPhone(ctx, phone)
		}
		if err != nil {
			return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "USER_UPSERT_FAILED", Message: "Unable to load user profile."}
		}
	}

	access, err := authtoken.SignAccessToken(userID, s.jwtSecret, 48*time.Hour)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "TOKEN_SIGN_FAILED", Message: "Unable to create access token."}
	}
	rawRefresh := make([]byte, 32)
	if _, err := rand.Read(rawRefresh); err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "ENTROPY_FAILED", Message: "Unable to create refresh token."}
	}
	sum := sha256.Sum256(rawRefresh)
	if err := s.repo.InsertRefreshTokenHash(ctx, userID, hex.EncodeToString(sum[:])); err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "REFRESH_PERSIST_FAILED", Message: "Unable to persist refresh token."}
	}

	return &OTPVerifyResult{
		AccessToken:  access,
		RefreshToken: hex.EncodeToString(rawRefresh),
		UserID:       userID.String(),
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, *ServiceError) {
	if refreshToken == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Refresh token is required."}
	}
	raw, err := hex.DecodeString(refreshToken)
	if err != nil || len(raw) != 32 {
		return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "INVALID_REFRESH", Message: "Refresh token is invalid."}
	}

	sum := sha256.Sum256(raw)
	userID, err := s.repo.ConsumeRefreshTokenHash(ctx, hex.EncodeToString(sum[:]))
	if err != nil {
		return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "INVALID_REFRESH", Message: "Refresh token is invalid or expired."}
	}

	access, err := authtoken.SignAccessToken(userID, s.jwtSecret, 48*time.Hour)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "TOKEN_SIGN_FAILED", Message: "Unable to create access token."}
	}
	rawRefresh := make([]byte, 32)
	if _, err := rand.Read(rawRefresh); err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "ENTROPY_FAILED", Message: "Unable to create refresh token."}
	}
	sum2 := sha256.Sum256(rawRefresh)
	if err := s.repo.InsertRefreshTokenHash(ctx, userID, hex.EncodeToString(sum2[:])); err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "REFRESH_PERSIST_FAILED", Message: "Unable to persist refresh token."}
	}

	return &RefreshResult{
		AccessToken:  access,
		RefreshToken: hex.EncodeToString(rawRefresh),
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID) (*MeResult, *ServiceError) {
	phone, displayName, err := s.repo.GetUserProfileByID(ctx, userID)
	if err != nil {
		if authrepo.IsNoRows(err) {
			return nil, &ServiceError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "User profile not found."}
		}
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "USER_READ_FAILED", Message: "Unable to load user profile."}
	}
	return &MeResult{
		ID:          userID.String(),
		Phone:       phone,
		DisplayName: displayName,
	}, nil
}
