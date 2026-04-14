package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	if err := c.ShouldBindJSON(&body); err != nil || s.AuthService == nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Phone is required.")
		return
	}
	if svcErr := s.AuthService.RequestOTP(c.Request.Context(), body.Phone, c.ClientIP()); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "dev_hint": "use configured DEV_OTP_CODE in non-production docs"})
}

func (s *Server) postOTPVerify(c *gin.Context) {
	var body otpVerify
	if err := c.ShouldBindJSON(&body); err != nil || s.AuthService == nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Phone and code are required.")
		return
	}
	result, svcErr := s.AuthService.VerifyOTP(c.Request.Context(), body.Phone, body.Code)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user_id":       result.UserID,
	})
}

func (s *Server) postRefresh(c *gin.Context) {
	var body refreshBody
	if err := c.ShouldBindJSON(&body); err != nil || s.AuthService == nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Refresh token is required.")
		return
	}
	result, svcErr := s.AuthService.Refresh(c.Request.Context(), body.RefreshToken)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}

func (s *Server) getMe(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.AuthService == nil {
		writeAPIError(c, http.StatusInternalServerError, "AUTH_SERVICE_UNAVAILABLE", "Auth service unavailable.")
		return
	}
	result, svcErr := s.AuthService.GetMe(c.Request.Context(), uid)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": result.ID, "phone": result.Phone, "display_name": result.DisplayName})
}
