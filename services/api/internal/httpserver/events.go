package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/bhune/utsav/services/api/internal/repository/eventrepo"
)

type createEventBody struct {
	Slug        string         `json:"slug" binding:"required"`
	Title       string         `json:"title" binding:"required"`
	EventType   string         `json:"event_type"`
	CoupleA     string         `json:"couple_name_a"`
	CoupleB     string         `json:"couple_name_b"`
	LoveStory   string         `json:"love_story"`
	CoverURL    string         `json:"cover_image_url"`
	DateStart   *string        `json:"date_start"`
	DateEnd     *string        `json:"date_end"`
	Privacy     string         `json:"privacy"`
	Toggles     map[string]any `json:"toggles"`
	Branding    map[string]any `json:"branding"`
	HostUPIVPA  string         `json:"host_upi_vpa"`
}

func (s *Server) getCheckSlug(c *gin.Context) {
	if s.EventService == nil {
		writeAPIError(c, http.StatusInternalServerError, "EVENT_SERVICE_UNAVAILABLE", "Event service unavailable.")
		return
	}
	slug, available, svcErr := s.EventService.CheckSlug(c.Request.Context(), c.Query("slug"))
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"slug": slug, "available": available})
}

func (s *Server) postEvent(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.EventService == nil {
		writeAPIError(c, http.StatusInternalServerError, "EVENT_SERVICE_UNAVAILABLE", "Event service unavailable.")
		return
	}
	var body createEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Event payload is invalid.")
		return
	}
	id, slug, svcErr := s.EventService.CreateEvent(c.Request.Context(), eventrepo.CreateEventInput{
		OwnerUserID: uid,
		Slug:        body.Slug,
		Title:       body.Title,
		EventType:   body.EventType,
		CoupleA:     body.CoupleA,
		CoupleB:     body.CoupleB,
		LoveStory:   body.LoveStory,
		CoverURL:    body.CoverURL,
		DateStart:   body.DateStart,
		DateEnd:     body.DateEnd,
		Privacy:     body.Privacy,
		Toggles:     body.Toggles,
		Branding:    body.Branding,
		HostUPIVPA:  body.HostUPIVPA,
	})
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "slug": slug})
}

func (s *Server) listEvents(c *gin.Context) {
	uid, ok := s.requireUser(c)
	if !ok {
		return
	}
	if s.EventService == nil {
		writeAPIError(c, http.StatusInternalServerError, "EVENT_SERVICE_UNAVAILABLE", "Event service unavailable.")
		return
	}
	rows, svcErr := s.EventService.ListEvents(c.Request.Context(), uid)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": rows})
}

func (s *Server) getEvent(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if s.EventService == nil {
		writeAPIError(c, http.StatusInternalServerError, "EVENT_SERVICE_UNAVAILABLE", "Event service unavailable.")
		return
	}
	event, svcErr := s.EventService.GetEvent(c.Request.Context(), eventID)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": event.ID, "slug": event.Slug, "title": event.Title, "event_type": event.EventType,
		"couple_name_a": event.CoupleA, "couple_name_b": event.CoupleB, "love_story": event.LoveStory, "cover_image_url": event.CoverURL,
		"date_start": event.DateStart, "date_end": event.DateEnd, "privacy": event.Privacy,
		"toggles": event.Toggles, "branding": event.Branding, "host_upi_vpa": event.HostUPIVPA, "tier": event.Tier,
	})
}

type patchEventBody struct {
	Title      *string        `json:"title"`
	Privacy    *string        `json:"privacy"`
	Toggles    map[string]any `json:"toggles"`
	Branding   map[string]any `json:"branding"`
	HostUPIVPA *string        `json:"host_upi_vpa"`
}

func (s *Server) patchEvent(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to update this event.")
		return
	}
	if s.EventService == nil {
		writeAPIError(c, http.StatusInternalServerError, "EVENT_SERVICE_UNAVAILABLE", "Event service unavailable.")
		return
	}
	var body patchEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Event patch payload is invalid.")
		return
	}
	hostVPA := body.HostUPIVPA
	if hostVPA != nil && !roleCanManageFinancials(role) {
		hostVPA = nil
	}
	if svcErr := s.EventService.PatchEvent(c.Request.Context(), eventID, role, eventrepo.PatchEventInput{
		Title:      body.Title,
		Privacy:    body.Privacy,
		HostUPIVPA: hostVPA,
	}); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type subEventBody struct {
	Name        string  `json:"name" binding:"required"`
	SubType     string  `json:"sub_type"`
	StartsAt    *string `json:"starts_at"`
	EndsAt      *string `json:"ends_at"`
	VenueLabel  string  `json:"venue_label"`
	DressCode   string  `json:"dress_code"`
	Description string  `json:"description"`
}

func (s *Server) postSubEvent(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if !roleCanManageEventData(role) {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to create sub-events.")
		return
	}
	if s.EventService == nil {
		writeAPIError(c, http.StatusInternalServerError, "EVENT_SERVICE_UNAVAILABLE", "Event service unavailable.")
		return
	}
	var body subEventBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Sub-event payload is invalid.")
		return
	}
	id, svcErr := s.EventService.CreateSubEvent(c.Request.Context(), eventID, eventrepo.CreateSubEventInput{
		Name:        body.Name,
		SubType:     body.SubType,
		StartsAt:    body.StartsAt,
		EndsAt:      body.EndsAt,
		VenueLabel:  body.VenueLabel,
		DressCode:   body.DressCode,
		Description: body.Description,
	})
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (s *Server) listSubEvents(c *gin.Context) {
	_, eventID, _, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if s.EventService == nil {
		writeAPIError(c, http.StatusInternalServerError, "EVENT_SERVICE_UNAVAILABLE", "Event service unavailable.")
		return
	}
	list, svcErr := s.EventService.ListSubEvents(c.Request.Context(), eventID)
	if svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sub_events": list})
}

type inviteMemberBody struct {
	InvitedPhone string `json:"invited_phone" binding:"required"`
	Role         string `json:"role" binding:"required"`
}

func (s *Server) postEventMember(c *gin.Context) {
	_, eventID, role, ok := s.requireEventAccess(c)
	if !ok {
		return
	}
	if role != "owner" && role != "co_owner" {
		writeAPIError(c, http.StatusForbidden, "FORBIDDEN", "Only owners can invite members.")
		return
	}
	if s.EventService == nil {
		writeAPIError(c, http.StatusInternalServerError, "EVENT_SERVICE_UNAVAILABLE", "Event service unavailable.")
		return
	}
	var body inviteMemberBody
	if err := c.ShouldBindJSON(&body); err != nil {
		writeAPIError(c, http.StatusBadRequest, "INVALID_BODY", "Invite payload is invalid.")
		return
	}
	if svcErr := s.EventService.InviteMember(c.Request.Context(), eventID, body.Role, body.InvitedPhone); svcErr != nil {
		writeAPIError(c, svcErr.Status, svcErr.Code, svcErr.Message)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}
