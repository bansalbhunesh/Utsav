package httpserver

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/bhune/utsav/services/api/internal/model"
)

const (
	accessTokenCookieName  = "utsav_access_token"
	refreshTokenCookieName = "utsav_refresh_token"
)

func (s *Server) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	secure := s.Config != nil && strings.EqualFold(strings.TrimSpace(s.Config.Env), "production")
	domain := ""
	if s.Config != nil {
		domain = strings.TrimSpace(s.Config.AuthCookieDomain)
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    accessToken,
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((48 * time.Hour).Seconds()),
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
}

func (s *Server) clearAuthCookies(c *gin.Context) {
	secure := s.Config != nil && strings.EqualFold(strings.TrimSpace(s.Config.Env), "production")
	domain := ""
	if s.Config != nil {
		domain = strings.TrimSpace(s.Config.AuthCookieDomain)
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: accessTokenCookieName, Value: "", Path: "/", Domain: domain, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	http.SetCookie(c.Writer, &http.Cookie{Name: refreshTokenCookieName, Value: "", Path: "/", Domain: domain, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

func (s *Server) postOTPRequest(c *gin.Context) {
	var body model.OTPRequest
	if err := c.ShouldBindJSON(&body); err != nil || s.AuthService == nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Phone is required.")
		return
	}
	if svcErr := s.AuthService.RequestOTP(c.Request.Context(), body.Phone, c.ClientIP()); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	resp := gin.H{"ok": true}
	if s.Config != nil && strings.ToLower(strings.TrimSpace(s.Config.Env)) != "production" {
		resp["dev_hint"] = "use configured DEV_OTP_CODE in non-production docs"
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) postOTPVerify(c *gin.Context) {
	var body model.OTPVerify
	if err := c.ShouldBindJSON(&body); err != nil || s.AuthService == nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Phone and code are required.")
		return
	}
	result, svcErr := s.AuthService.VerifyOTP(c.Request.Context(), body.Phone, body.Code, c.ClientIP())
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	s.setAuthCookies(c, result.AccessToken, result.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"user_id":       result.UserID,
		"authenticated": true,
	})
}

func (s *Server) postRefresh(c *gin.Context) {
	var body model.RefreshBody
	if s.AuthService == nil {
		writeAPIError(c, http.StatusInternalServerError, "AUTH_SERVICE_UNAVAILABLE", "Auth service unavailable.")
		return
	}
	_ = c.ShouldBindJSON(&body)
	refreshToken := strings.TrimSpace(body.RefreshToken)
	if refreshToken == "" {
		if cookieValue, err := c.Cookie(refreshTokenCookieName); err == nil {
			refreshToken = strings.TrimSpace(cookieValue)
		}
	}
	if refreshToken == "" {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Refresh token is required.")
		return
	}
	result, svcErr := s.AuthService.Refresh(c.Request.Context(), refreshToken)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	s.setAuthCookies(c, result.AccessToken, result.RefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
	})
}

func (s *Server) postLogout(c *gin.Context) {
	if tok, err := c.Cookie(refreshTokenCookieName); err == nil && s.AuthService != nil {
		s.AuthService.Logout(c.Request.Context(), strings.TrimSpace(tok))
	}
	s.clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
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
