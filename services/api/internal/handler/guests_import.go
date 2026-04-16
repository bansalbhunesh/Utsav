package httpserver

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/bhune/utsav/services/api/internal/model"
)

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
	if s.GuestService == nil {
		writeAPIError(c, http.StatusInternalServerError, "GUEST_SERVICE_UNAVAILABLE", "Guest service unavailable.")
		return
	}
	contentType := strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type")))
	var src io.Reader
	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "CSV file is required.")
			return
		}
		defer file.Close()
		src = file
	} else {
		var body model.GuestsImportBody
		if err := c.ShouldBindJSON(&body); err != nil {
			writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "CSV payload is required.")
			return
		}
		src = bytes.NewBufferString(body.CSV)
	}
	result, svcErr := s.GuestService.ImportGuestsCSV(c.Request.Context(), eventID, src)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported": result.Imported, "errors": result.Errors})
}
