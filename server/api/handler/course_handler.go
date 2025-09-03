package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/course"

	"github.com/labstack/echo/v4"
)

type CourseHandler struct {
	courseService course.CourseService
}

func NewCourseHandler(courseService course.CourseService) *CourseHandler {
	return &CourseHandler{
		courseService: courseService,
	}
}

func (h *CourseHandler) UpdateCourse(c echo.Context) error {
	// Extract URL parameters
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Bind the UpdateCourseRequest struct
	var req models.UpdateCourseRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	// Call service method with partial update request
	err = h.courseService.UpdateCoursePartial(c.Request().Context(), collegeID, courseID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Success", 204)
}

func (h *CourseHandler) ListCourses(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Parse pagination parameters
	limit := uint64(10) // default limit
	offset := uint64(0)  // default offset

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseUint(limitParam, 10, 64); err == nil {
			limit = parsedLimit
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.ParseUint(offsetParam, 10, 64); err == nil {
			offset = parsedOffset
		}
	}

	courses, err := h.courseService.FindAllCourses(c.Request().Context(), collegeID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, courses, 200)
}

func (h *CourseHandler) GetCourse(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	course, err := h.courseService.FindCourseByID(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, "course not found", 404)
	}

	return helpers.Success(c, course, 200)
}