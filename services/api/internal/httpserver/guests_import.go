package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type importGuestsBody struct {
	CSV string `json:"csv" binding:"required"`
}

// postGuestsImport accepts CSV with header row optional. Columns: name, phone (required);
// optional: email, relationship, side (matched by header name).
func (s *Server) postGuestsImport(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to import guests.")
		return
	}
	var body importGuestsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "CSV payload is required.")
		return
	}
	if s.GuestService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GUEST_SERVICE_UNAVAILABLE", "Guest service unavailable.")
		return
	}
	result, svcErr := s.GuestService.ImportGuestsCSV(c.Request.Context(), eventID, body.CSV)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported": result.Imported, "errors": result.Errors})
}
