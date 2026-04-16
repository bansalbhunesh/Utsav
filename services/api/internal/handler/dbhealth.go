package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	dbHealthPollInterval = 10 * time.Second
	dbHealthPingTimeout  = 2 * time.Second
	dbHealthOpenAfter    = 3 // consecutive ping failures before rejecting traffic
)

// startDBHealthPoller runs until process exit. It opens a fail-fast gate after repeated ping failures.
func (s *Server) startDBHealthPoller() {
	if s.Pool == nil {
		return
	}
	s.dbReady.Store(true)
	s.runDBPingOnce()
	ticker := time.NewTicker(dbHealthPollInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.runDBPingOnce()
	}
}

func (s *Server) runDBPingOnce() {
	if s.Pool == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), dbHealthPingTimeout)
	err := s.Pool.Ping(ctx)
	cancel()
	if err == nil {
		s.dbFailStreak.Store(0)
		s.dbReady.Store(true)
		return
	}
	if s.dbFailStreak.Add(1) >= dbHealthOpenAfter {
		s.dbReady.Store(false)
	}
}

func (s *Server) ensureDBHealthPoller() {
	if s.Pool == nil {
		return
	}
	s.dbHealthOnce.Do(func() {
		go s.startDBHealthPoller()
	})
}

func (s *Server) requireDBAvailable() gin.HandlerFunc {
	return func(c *gin.Context) {
		s.ensureDBHealthPoller()
		if !s.dbReady.Load() {
			writeAPIError(c, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Service temporarily unavailable.")
			c.Abort()
			return
		}
		c.Next()
	}
}
