package handler

import (
	"net/http"
	"strconv"

	"eduhub/server/internal/middleware"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/timetable"

	"github.com/labstack/echo/v4"
)

type TimetableHandler struct {
	timetableService timetable.TimetableService
}

func NewTimetableHandler(timetableService timetable.TimetableService) *TimetableHandler {
	return &TimetableHandler{
		timetableService: timetableService,
	}
}

func (h *TimetableHandler) CreateTimeTableBlock(c echo.Context) error {
	var block models.TimeTableBlock
	if err := middleware.BindAndValidate(c, &block); err != nil {
		return err
	}

	collegeID := c.Get("college_id").(int)
	block.CollegeID = collegeID

	if err := h.timetableService.CreateTimeTableBlock(c.Request().Context(), &block); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create timetable block: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Timetable block created successfully",
		"data":    block,
	})
}

func (h *TimetableHandler) GetTimeTableBlocks(c echo.Context) error {
	collegeID := c.Get("college_id").(int)

	filter := models.TimeTableBlockFilter{
		CollegeID: collegeID,
	}

	// Add optional filters
	if courseIDStr := c.QueryParam("course_id"); courseIDStr != "" {
		courseID, err := strconv.Atoi(courseIDStr)
		if err == nil {
			filter.CourseID = &courseID
		}
	}

	blocks, err := h.timetableService.GetTimeTableBlocks(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get timetable: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": blocks,
	})
}

func (h *TimetableHandler) UpdateTimeTableBlock(c echo.Context) error {
	blockID, err := strconv.Atoi(c.Param("blockID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid block ID")
	}

	var block models.TimeTableBlock
	if err := middleware.BindAndValidate(c, &block); err != nil {
		return err
	}

	block.ID = blockID
	collegeID := c.Get("college_id").(int)
	block.CollegeID = collegeID

	if err := h.timetableService.UpdateTimeTableBlock(c.Request().Context(), &block); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update timetable block: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Timetable block updated successfully",
		"data":    block,
	})
}

func (h *TimetableHandler) DeleteTimeTableBlock(c echo.Context) error {
	blockID, err := strconv.Atoi(c.Param("blockID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid block ID")
	}

	collegeID := c.Get("college_id").(int)

	if err := h.timetableService.DeleteTimeTableBlock(c.Request().Context(), blockID, collegeID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete timetable block: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Timetable block deleted successfully",
	})
}

func (h *TimetableHandler) GetStudentTimetable(c echo.Context) error {
	studentID := c.Get("student_id").(int)

	blocks, err := h.timetableService.GetStudentTimetable(c.Request().Context(), studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get student timetable: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": blocks,
	})
}
