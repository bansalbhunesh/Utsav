package middleware

import (
	"net/http"
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
		// structured-friendly single line
		gin.DefaultWriter.Write([]byte(
			time.Now().Format(time.RFC3339) + " " + c.Request.Method + " " + c.Request.URL.Path +
				" " + http.StatusText(c.Writer.Status()) + " " + elapsed.String() + " rid=" + c.GetString("request_id") + "\n",
		))
	}
}

func CORS(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := ""
		for _, o := range origins {
			if o == origin {
				allowed = o
				break
			}
		}
		if allowed != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowed)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, "+headerRequestID)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func RecoverJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			}
		}()
		c.Next()
	}
}
