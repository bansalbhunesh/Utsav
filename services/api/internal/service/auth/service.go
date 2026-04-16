package authservice

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authtoken "github.com/bhune/utsav/services/api/pkg/auth"
	"github.com/bhune/utsav/services/api/pkg/otp"
	phoneutil "github.com/bhune/utsav/services/api/pkg/phone"
	"github.com/bhune/utsav/services/api/pkg/ratelimit"
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

type repository interface {
	DeletePhoneOTPChallenges(ctx context.Context, phone string) error
	InsertPhoneOTPChallenge(ctx context.Context, phone, codeHash string) error
	GetLatestPhoneOTPChallenge(ctx context.Context, phone string) (*authrepo.OTPChallenge, error)
	IncrementPhoneOTPAttempts(ctx context.Context, id uuid.UUID) error
	ConsumePhoneOTPChallengeByID(ctx context.Context, id uuid.UUID) (bool, error)
	DeletePhoneOTPChallengeByID(ctx context.Context, id uuid.UUID) error
	FindUserIDByPhone(ctx context.Context, phone string) (uuid.UUID, error)
	CreateUserWithPhone(ctx context.Context, phone string) (uuid.UUID, error)
	InsertRefreshTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string) error
	PruneRefreshTokensForUser(ctx context.Context, userID uuid.UUID, maxKeep int) error
	RotateRefreshToken(ctx context.Context, oldTokenHash, newTokenHash string) (uuid.UUID, error)
	RevokeRefreshTokenHash(ctx context.Context, tokenHash string) error
	GetUserProfileByID(ctx context.Context, userID uuid.UUID) (string, string, error)
}

type Service struct {
	repo              repository
	otpRequestLimiter ratelimit.Limiter
	otpVerifyLimiter  ratelimit.Limiter
	devOTP            string
	jwtSecret         []byte
	otpSecret         []byte
	env               string
	otpDispatcher     otp.Dispatcher
	otpMaxAttempts    int
}

func NewService(
	repo repository,
	otpRequestLimiter ratelimit.Limiter,
	otpVerifyLimiter ratelimit.Limiter,
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
		devOTP:            strings.TrimSpace(devOTPCode),
		jwtSecret:         []byte(jwtSecret),
		otpSecret:         []byte(otpSecret),
		env:               strings.TrimSpace(strings.ToLower(env)),
		otpDispatcher:     otpDispatcher,
		otpMaxAttempts:    otpMaxAttempts,
	}
}

func (s *Service) RequestOTP(ctx context.Context, phone, clientIP string) *ServiceError {
	phoneNorm, err := phoneutil.NormalizeE164(phone)
	if err != nil {
		return &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_PHONE", Message: "Phone number is invalid."}
	}
	if s.otpRequestLimiter != nil {
		allowed, err := s.otpRequestLimiter.Allow(ctx, "auth_otp_req:"+clientIP+"|"+phoneNorm)
		if err != nil {
			return &ServiceError{Status: http.StatusInternalServerError, Code: "RATE_LIMIT_FAILED", Message: "Unable to validate OTP rate limits."}
		}
		if !allowed {
			return &ServiceError{
				Status:  http.StatusTooManyRequests,
				Code:    "RATE_LIMITED",
				Message: "Too many OTP requests. Please retry later.",
			}
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
			return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_GENERATE_FAILED", Message: "Unable to generate OTP code."}
		}
	}
	hash, err := otp.HashCode(s.otpSecret, code)
	if err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_HASH_FAILED", Message: "Unable to process OTP request."}
	}
	if err := s.repo.DeletePhoneOTPChallenges(ctx, phoneNorm); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save OTP challenge."}
	}
	if err := s.repo.InsertPhoneOTPChallenge(ctx, phoneNorm, hash); err != nil {
		return &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_PERSIST_FAILED", Message: "Unable to save OTP challenge."}
	}
	if s.otpDispatcher != nil {
		if err := s.otpDispatcher.DispatchOTP(ctx, phoneNorm, code); err != nil {
			_ = s.repo.DeletePhoneOTPChallenges(ctx, phoneNorm)
			return &ServiceError{Status: http.StatusBadGateway, Code: "OTP_SEND_FAILED", Message: "Unable to send OTP code."}
		}
	}
	return nil
}

func (s *Service) VerifyOTP(ctx context.Context, phone, code, clientIP string) (*OTPVerifyResult, *ServiceError) {
	phoneNorm, err := phoneutil.NormalizeE164(phone)
	if err != nil || strings.TrimSpace(code) == "" {
		return nil, &ServiceError{Status: http.StatusBadRequest, Code: "INVALID_BODY", Message: "Phone and code are required."}
	}
	if s.otpVerifyLimiter != nil {
		allowed, err := s.otpVerifyLimiter.Allow(ctx, "auth_otp_verify:"+clientIP+"|"+phoneNorm)
		if err != nil {
			return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "RATE_LIMIT_FAILED", Message: "Unable to validate OTP rate limits."}
		}
		if !allowed {
			return nil, &ServiceError{Status: http.StatusTooManyRequests, Code: "RATE_LIMITED", Message: "Too many OTP verify attempts. Please retry later."}
		}
	}
	ch, err := s.repo.GetLatestPhoneOTPChallenge(ctx, phoneNorm)
	if err != nil {
		if authrepo.IsNoRows(err) {
			return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "NO_CHALLENGE", Message: "No active OTP challenge found."}
		}
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_READ_FAILED", Message: "Unable to read OTP challenge."}
	}
	if time.Now().After(ch.ExpiresAt) {
		return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "OTP_EXPIRED", Message: "OTP has expired. Request a new code."}
	}
	if s.otpMaxAttempts > 0 && ch.Attempts >= s.otpMaxAttempts {
		return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "OTP_LOCKED", Message: "OTP attempts exceeded. Request a new code."}
	}
	if !otp.VerifyCode(s.otpSecret, ch.CodeHash, code) {
		if incErr := s.repo.IncrementPhoneOTPAttempts(ctx, ch.ID); incErr != nil {
			return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_STATE_FAILED", Message: "Unable to update OTP state."}
		}
		return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "INVALID_OTP", Message: "The OTP code is invalid."}
	}
	consumed, consumeErr := s.repo.ConsumePhoneOTPChallengeByID(ctx, ch.ID)
	if consumeErr != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "OTP_STATE_FAILED", Message: "Unable to update OTP state."}
	}
	if !consumed {
		return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "NO_CHALLENGE", Message: "No active OTP challenge found."}
	}

	userID, err := s.repo.FindUserIDByPhone(ctx, phoneNorm)
	if err != nil {
		if authrepo.IsNoRows(err) {
			userID, err = s.repo.CreateUserWithPhone(ctx, phoneNorm)
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
	_ = s.repo.PruneRefreshTokensForUser(ctx, userID, 10)

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
	oldHash := hex.EncodeToString(sum[:])

	rawRefresh := make([]byte, 32)
	if _, err := rand.Read(rawRefresh); err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "ENTROPY_FAILED", Message: "Unable to create refresh token."}
	}
	sum2 := sha256.Sum256(rawRefresh)
	newHash := hex.EncodeToString(sum2[:])

	rotatedUserID, err := s.repo.RotateRefreshToken(ctx, oldHash, newHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &ServiceError{Status: http.StatusUnauthorized, Code: "INVALID_REFRESH", Message: "Refresh token is invalid or expired."}
		}
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "REFRESH_PERSIST_FAILED", Message: "Unable to persist refresh token."}
	}
	access, err := authtoken.SignAccessToken(rotatedUserID, s.jwtSecret, 48*time.Hour)
	if err != nil {
		return nil, &ServiceError{Status: http.StatusInternalServerError, Code: "TOKEN_SIGN_FAILED", Message: "Unable to create access token."}
	}

	return &RefreshResult{
		AccessToken:  access,
		RefreshToken: hex.EncodeToString(rawRefresh),
	}, nil
}

// Logout revokes the refresh token in the database (best-effort). Invalid hex is ignored.
func (s *Service) Logout(ctx context.Context, refreshToken string) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return
	}
	raw, err := hex.DecodeString(refreshToken)
	if err != nil || len(raw) != 32 {
		return
	}
	sum := sha256.Sum256(raw)
	_ = s.repo.RevokeRefreshTokenHash(ctx, hex.EncodeToString(sum[:]))
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
