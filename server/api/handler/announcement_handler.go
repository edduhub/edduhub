// Package handler contains the HTTP handlers for the API endpoints.
// This file specifically implements handlers for announcement-related operations,
// bridging the gap between the HTTP server and the announcement service layer.
package handler

import (
	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/announcement"
	"strconv"

	"github.com/labstack/echo/v4"
)

// AnnouncementHandler manages HTTP requests for announcements.
// It relies on an AnnouncementService to perform the actual business logic.
type AnnouncementHandler struct {
	service announcement.AnnouncementService
}

// NewAnnouncementHandler creates a new handler for announcement endpoints.
func NewAnnouncementHandler(service announcement.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{
		service: service,
	}
}

// CreateAnnouncement handles the API endpoint for creating a new announcement.
// It extracts college and user information from the context, binds the request payload,
// and calls the service to create the announcement.
func (h *AnnouncementHandler) CreateAnnouncement(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "missing college_id", 400)
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "missing user_id", 400)
	}

	var req models.CreateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request payload", 400)
	}

	newAnnouncement, err := h.service.CreateAnnouncement(c.Request().Context(), &req, collegeID, userID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, newAnnouncement, 201)
}

// GetAnnouncement handles the retrieval of a single announcement by its ID.
func (h *AnnouncementHandler) GetAnnouncement(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return helpers.Error(c, "invalid announcement ID", 400)
	}

	announcement, err := h.service.GetAnnouncementByID(c.Request().Context(), id)
	if err != nil {
		return helpers.Error(c, "announcement not found", 404)
	}

	return helpers.Success(c, announcement, 200)
}

// ListAnnouncements retrieves a paginated list of announcements for a college.
func (h *AnnouncementHandler) ListAnnouncements(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "missing college_id", 400)
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	announcements, err := h.service.GetAnnouncementsByCollegeID(c.Request().Context(), collegeID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, announcements, 200)
}

// UpdateAnnouncement handles updating an existing announcement's details.
func (h *AnnouncementHandler) UpdateAnnouncement(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return helpers.Error(c, "invalid announcement ID", 400)
	}

	var req models.UpdateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request payload", 400)
	}

	updatedAnnouncement, err := h.service.UpdateAnnouncement(c.Request().Context(), id, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, updatedAnnouncement, 200)
}

// DeleteAnnouncement handles the deletion of an announcement.
func (h *AnnouncementHandler) DeleteAnnouncement(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return helpers.Error(c, "invalid announcement ID", 400)
	}

	if err := h.service.DeleteAnnouncement(c.Request().Context(), id); err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "announcement deleted successfully", 204)
}
