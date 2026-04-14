package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) getPublicEvent(c *gin.Context) {
	if s.PublicService == nil {
		writeAPIError(c, http.StatusInternalServerError, "PUBLIC_SERVICE_UNAVAILABLE", "Public service unavailable.")
		return
	}
	event, _, svcErr := s.PublicService.GetEvent(c.Request.Context(), c.Param("slug"))
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, event)
}

func (s *Server) getPublicSchedule(c *gin.Context) {
	if s.PublicService == nil {
		writeAPIError(c, http.StatusInternalServerError, "PUBLIC_SERVICE_UNAVAILABLE", "Public service unavailable.")
		return
	}
	list, _, svcErr := s.PublicService.ListSchedule(c.Request.Context(), c.Param("slug"))
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sub_events": list})
}

func (s *Server) getPublicBroadcasts(c *gin.Context) {
	if s.PublicService == nil {
		writeAPIError(c, http.StatusInternalServerError, "PUBLIC_SERVICE_UNAVAILABLE", "Public service unavailable.")
		return
	}
	out, svcErr := s.PublicService.ListBroadcasts(c.Request.Context(), c.Param("slug"))
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"broadcasts": out})
}

func (s *Server) getPublicGallery(c *gin.Context) {
	if s.PublicService == nil {
		writeAPIError(c, http.StatusInternalServerError, "PUBLIC_SERVICE_UNAVAILABLE", "Public service unavailable.")
		return
	}
	out, svcErr := s.PublicService.ListGallery(c.Request.Context(), c.Param("slug"))
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"assets": out})
}
