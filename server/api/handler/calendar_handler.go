package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/calendar"

	"github.com/labstack/echo/v4"
)

type CalendarHandler struct {
	calendarService calendar.CalendarService
}

func NewCalendarHandler(calendarService calendar.CalendarService) *CalendarHandler {
	return &CalendarHandler{
		calendarService: calendarService,
	}
}

func (h *CalendarHandler) GetEvents(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	filter := models.CalendarBlockFilter{
		CollegeID: &collegeID,
		Limit:     10,
		Offset:    0,
	}

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseUint(limitParam, 10, 64); err == nil {
			filter.Limit = parsedLimit
		}
	}
	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.ParseUint(offsetParam, 10, 64); err == nil {
			filter.Offset = parsedOffset
		}
	}

	events, err := h.calendarService.GetEvents(c.Request().Context(), filter)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, events, 200)
}

func (h *CalendarHandler) CreateEvent(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var event models.CalendarBlock
	if err := c.Bind(&event); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	event.CollegeID = collegeID

	err = h.calendarService.CreateEvent(c.Request().Context(), &event)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, event, 201)
}

func (h *CalendarHandler) UpdateEvent(c echo.Context) error {
	eventIDStr := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return helpers.Error(c, "invalid event ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateCalendarRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.calendarService.UpdateEvent(c.Request().Context(), collegeID, eventID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Event updated successfully", 204)
}

func (h *CalendarHandler) DeleteEvent(c echo.Context) error {
	eventIDStr := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return helpers.Error(c, "invalid event ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.calendarService.DeleteEvent(c.Request().Context(), collegeID, eventID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Event deleted successfully", 204)
}
