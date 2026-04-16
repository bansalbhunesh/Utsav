package middleware

import (
	"encoding/json"
	"net/url"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Writer.Header().Set(headerRequestID, rid)
		c.Set("request_id", rid)
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		elapsed := time.Since(start)
		payload := map[string]any{
			"ts":          time.Now().UTC().Format(time.RFC3339),
			"level":       "info",
			"request_id":  c.GetString("request_id"),
			"method":      c.Request.Method,
			"endpoint":    c.FullPath(),
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
			"status_text": http.StatusText(c.Writer.Status()),
			"latency_ms":  elapsed.Milliseconds(),
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"user_id":     c.GetString("user_id"),
			"guest_id":    c.GetString("guest_id"),
			"error_code":  c.GetString("error_code"),
		}
		b, _ := json.Marshal(payload)
		gin.DefaultWriter.Write(append(b, '\n'))
	}
}

func CORS(origins []string) gin.HandlerFunc {
	normalized := make([]string, 0, len(origins))
	for _, o := range origins {
		if no := normalizeOrigin(o); no != "" {
			normalized = append(normalized, no)
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		normalizedOrigin := normalizeOrigin(origin)
		allowed := ""
		for _, o := range normalized {
			if strings.EqualFold(o, normalizedOrigin) {
				allowed = normalizedOrigin
				break
			}
		}
		if allowed != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowed)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, "+headerRequestID)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			if origin != "" && allowed == "" {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return strings.TrimSuffix(origin, "/")
	}
	return u.Scheme + "://" + u.Host
}

func RecoverJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				payload := map[string]any{
					"ts":         time.Now().UTC().Format(time.RFC3339),
					"level":      "error",
					"request_id": c.GetString("request_id"),
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"panic":      rec,
				}
				if b, err := json.Marshal(payload); err == nil {
					gin.DefaultWriter.Write(append(b, '\n'))
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			}
		}()
		c.Next()
	}
}
