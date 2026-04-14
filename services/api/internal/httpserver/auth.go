package httpserver

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/bhune/utsav/services/api/internal/auth"
)

type otpRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type otpVerify struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (s *Server) postOTPRequest(c *gin.Context) {
	var body otpRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Phone is required.")
		return
	}
	if s.AuthOTPLimit != nil && !s.AuthOTPLimit.Allow("auth_otp:"+c.ClientIP()) {
		writeAPIError(c, http.StatusTooManyRequests, "RATE_LIMITED", "Too many OTP requests. Please retry later.")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(s.Config.DevOTPCode), bcrypt.DefaultCost)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "OTP_HASH_FAILED", "Unable to process OTP request.")
		return
	}
	ctx := c.Request.Context()
	_, _ = s.Pool.Exec(ctx, `DELETE FROM phone_otp_challenges WHERE phone=$1`, body.Phone)
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO phone_otp_challenges (phone, code_hash, expires_at)
		VALUES ($1, $2, now() + interval '10 minutes')`, body.Phone, string(hash))
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "OTP_PERSIST_FAILED", "Unable to save OTP challenge.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "dev_hint": "use configured DEV_OTP_CODE in non-production docs"})
}

func (s *Server) postOTPVerify(c *gin.Context) {
	var body otpVerify
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Phone and code are required.")
		return
	}
	ctx := c.Request.Context()
	var id uuid.UUID
	var codeHash string
	var expires time.Time
	err := s.Pool.QueryRow(ctx, `
		SELECT id, code_hash, expires_at FROM phone_otp_challenges
		WHERE phone=$1 ORDER BY created_at DESC LIMIT 1`, body.Phone).Scan(&id, &codeHash, &expires)
	if err != nil {
		writeAPIError(c, http.StatusUnauthorized, "NO_CHALLENGE", "No active OTP challenge found.")
		return
	}
	if time.Now().After(expires) {
		writeAPIError(c, http.StatusUnauthorized, "OTP_EXPIRED", "OTP has expired. Request a new code.")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(codeHash), []byte(body.Code)); err != nil {
		_, _ = s.Pool.Exec(ctx, `UPDATE phone_otp_challenges SET attempts=attempts+1 WHERE id=$1`, id)
		writeAPIError(c, http.StatusUnauthorized, "INVALID_OTP", "The OTP code is invalid.")
		return
	}
	_, _ = s.Pool.Exec(ctx, `DELETE FROM phone_otp_challenges WHERE id=$1`, id)

	var userID uuid.UUID
	err = s.Pool.QueryRow(ctx, `SELECT id FROM users WHERE phone=$1`, body.Phone).Scan(&userID)
	if err == pgx.ErrNoRows {
		err = s.Pool.QueryRow(ctx, `
			INSERT INTO users (phone) VALUES ($1) RETURNING id`, body.Phone).Scan(&userID)
	}
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "USER_UPSERT_FAILED", "Unable to load user profile.")
		return
	}

	access, err := auth.SignAccessToken(userID, []byte(s.Config.JWTSecret), 48*time.Hour)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "TOKEN_SIGN_FAILED", "Unable to create access token.")
		return
	}
	rawRefresh := make([]byte, 32)
	if _, err := rand.Read(rawRefresh); err != nil {
		writeAPIError(c, http.StatusInternalServerError, "ENTROPY_FAILED", "Unable to create refresh token.")
		return
	}
	sum := sha256.Sum256(rawRefresh)
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, now() + interval '30 days')`,
		userID, hex.EncodeToString(sum[:]))
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "REFRESH_PERSIST_FAILED", "Unable to persist refresh token.")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": hex.EncodeToString(rawRefresh),
		"user_id":       userID.String(),
	})
}

func (s *Server) postRefresh(c *gin.Context) {
	var body refreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Refresh token is required.")
		return
	}
	raw, err := hex.DecodeString(body.RefreshToken)
	if err != nil || len(raw) != 32 {
		writeAPIError(c, http.StatusUnauthorized, "INVALID_REFRESH", "Refresh token is invalid.")
		return
	}
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	ctx := c.Request.Context()
	var userID uuid.UUID
	err = s.Pool.QueryRow(ctx, `
		DELETE FROM refresh_tokens WHERE token_hash=$1 AND expires_at > now()
		RETURNING user_id`, hash).Scan(&userID)
	if err != nil {
		writeAPIError(c, http.StatusUnauthorized, "INVALID_REFRESH", "Refresh token is invalid or expired.")
		return
	}
	access, err := auth.SignAccessToken(userID, []byte(s.Config.JWTSecret), 48*time.Hour)
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "TOKEN_SIGN_FAILED", "Unable to create access token.")
		return
	}
	rawRefresh := make([]byte, 32)
	if _, err := rand.Read(rawRefresh); err != nil {
		writeAPIError(c, http.StatusInternalServerError, "ENTROPY_FAILED", "Unable to create refresh token.")
		return
	}
	sum2 := sha256.Sum256(rawRefresh)
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, now() + interval '30 days')`, userID, hex.EncodeToString(sum2[:]))
	if err != nil {
		writeAPIError(c, http.StatusInternalServerError, "REFRESH_PERSIST_FAILED", "Unable to persist refresh token.")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": hex.EncodeToString(rawRefresh),
	})
}

func (s *Server) getMe(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	var phone, name string
	err := s.Pool.QueryRow(ctx, `SELECT phone, COALESCE(display_name,'') FROM users WHERE id=$1`, uid).Scan(&phone, &name)
	if err != nil {
		writeAPIError(c, http.StatusNotFound, "NOT_FOUND", "User profile not found.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": uid.String(), "phone": phone, "display_name": name})
}
