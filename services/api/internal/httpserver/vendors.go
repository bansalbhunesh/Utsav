package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/bhune/utsav/services/api/internal/repository/vendorrepo"
)

type createVendorBody struct {
	Name         string  `json:"name" binding:"required"`
	Category     string  `json:"category"`
	Phone        string  `json:"phone"`
	Email        string  `json:"email"`
	TotalPaise   int64   `json:"total_paise"`
	AdvancePaise int64   `json:"advance_paise"`
	Notes        string  `json:"notes"`
}

func (s *Server) listVendors(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if s.VendorService == nil {
		writeAPIError(c, http.StatusInternalServerError, "VENDOR_SERVICE_UNAVAILABLE", "Vendor service unavailable.")
		return
	}
	rows, svcErr := s.VendorService.ListVendors(c.Request.Context(), eventID)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"vendors": rows})
}

func (s *Server) postVendor(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	var body createVendorBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Vendor payload is invalid.")
		return
	}
	if s.VendorService == nil {
		writeAPIError(c, http.StatusInternalServerError, "VENDOR_SERVICE_UNAVAILABLE", "Vendor service unavailable.")
		return
	}
	id, svcErr := s.VendorService.CreateVendor(c.Request.Context(), eventID, vendorrepo.CreateInput{
		Name:         body.Name,
		Category:     body.Category,
		Phone:        body.Phone,
		Email:        body.Email,
		AdvancePaise: body.AdvancePaise,
		TotalPaise:   body.TotalPaise,
		Notes:        body.Notes,
	})
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (s *Server) deleteVendor(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if s.VendorService == nil {
		writeAPIError(c, http.StatusInternalServerError, "VENDOR_SERVICE_UNAVAILABLE", "Vendor service unavailable.")
		return
	}
	if svcErr := s.VendorService.DeleteVendor(c.Request.Context(), eventID, c.Param("vendorId")); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
