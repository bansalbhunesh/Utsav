package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type shagunReportBody struct {
	AmountINR    float64 `json:"amount_inr" binding:"required"`
	BlessingNote string  `json:"blessing_note"`
	SubEventID   *string `json:"sub_event_id"`
}

// UPI deep link helper (metadata only; funds never touch platform).
func (s *Server) getPublicUPILink(c *gin.Context) {
	if s.PublicService == nil {
		writeAPIError(c, http.StatusInternalServerError, "PUBLIC_SERVICE_UNAVAILABLE", "Public service unavailable.")
		return
	}
	geid, phone, ok := s.guestBearer(c)
	if !ok {
		return
	}
	resp, svcErr := s.PublicService.BuildUPILink(c.Request.Context(), c.Param("slug"), geid, phone)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) postPublicShagunReport(c *gin.Context) {
	geid, phone, ok := s.guestBearer(c)
	if !ok {
		return
	}
	if s.PublicService == nil {
		writeAPIError(c, http.StatusInternalServerError, "PUBLIC_SERVICE_UNAVAILABLE", "Public service unavailable.")
		return
	}
	var body shagunReportBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Shagun report payload is invalid.")
		return
	}
	if svcErr := s.PublicService.ReportShagun(
		c.Request.Context(),
		c.Param("slug"),
		geid,
		phone,
		body.AmountINR,
		body.BlessingNote,
		body.SubEventID,
	); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}
