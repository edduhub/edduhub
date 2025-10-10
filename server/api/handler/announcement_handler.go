package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/announcement"

	"github.com/labstack/echo/v4"
)

type AnnouncementHandler struct {
	announcementService announcement.AnnouncementService
}

func NewAnnouncementHandler(announcementService announcement.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{
		announcementService: announcementService,
	}
}

// ListAnnouncements retrieves announcements with optional filters
func (h *AnnouncementHandler) ListAnnouncements(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseIDStr := c.QueryParam("course_id")
	priorityStr := c.QueryParam("priority")
	publishedStr := c.QueryParam("published")
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	filter := models.AnnouncementFilter{
		CollegeID: &collegeID,
		Limit:     20,
		Offset:    0,
	}

	if courseIDStr != "" {
		courseID, err := strconv.Atoi(courseIDStr)
		if err == nil {
			filter.CourseID = &courseID
		}
	}

	if priorityStr != "" {
		filter.Priority = &priorityStr
	}

	if publishedStr != "" {
		published := publishedStr == "true"
		filter.IsPublished = &published
	}

	if limitStr != "" {
		limit, err := strconv.ParseUint(limitStr, 10, 64)
		if err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr != "" {
		offset, err := strconv.ParseUint(offsetStr, 10, 64)
		if err == nil {
			filter.Offset = offset
		}
	}

	announcements, err := h.announcementService.GetAnnouncements(c.Request().Context(), filter)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, announcements, 200)
}

// CreateAnnouncement creates a new announcement
func (h *AnnouncementHandler) CreateAnnouncement(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var announcement models.Announcement
	if err := c.Bind(&announcement); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	announcement.CollegeID = collegeID

	err = h.announcementService.CreateAnnouncement(c.Request().Context(), &announcement)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, announcement, 201)
}

// GetAnnouncement retrieves a specific announcement
func (h *AnnouncementHandler) GetAnnouncement(c echo.Context) error {
	announcementIDStr := c.Param("announcementID")
	announcementID, err := strconv.Atoi(announcementIDStr)
	if err != nil {
		return helpers.Error(c, "invalid announcement ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	announcement, err := h.announcementService.GetAnnouncement(c.Request().Context(), collegeID, announcementID)
	if err != nil {
		return helpers.Error(c, "announcement not found", 404)
	}

	return helpers.Success(c, announcement, 200)
}

// UpdateAnnouncement updates an announcement
func (h *AnnouncementHandler) UpdateAnnouncement(c echo.Context) error {
	announcementIDStr := c.Param("announcementID")
	announcementID, err := strconv.Atoi(announcementIDStr)
	if err != nil {
		return helpers.Error(c, "invalid announcement ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.announcementService.UpdateAnnouncement(c.Request().Context(), collegeID, announcementID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Announcement updated successfully", 200)
}

// DeleteAnnouncement deletes an announcement
func (h *AnnouncementHandler) DeleteAnnouncement(c echo.Context) error {
	announcementIDStr := c.Param("announcementID")
	announcementID, err := strconv.Atoi(announcementIDStr)
	if err != nil {
		return helpers.Error(c, "invalid announcement ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.announcementService.DeleteAnnouncement(c.Request().Context(), collegeID, announcementID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Announcement deleted successfully", 200)
}
