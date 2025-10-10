package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/analytics"

	"github.com/labstack/echo/v4"
)

type AnalyticsHandler struct {
	analyticsService analytics.AnalyticsService
}

func NewAnalyticsHandler(analyticsService analytics.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetStudentPerformance retrieves performance metrics for a student
func (h *AnalyticsHandler) GetStudentPerformance(c echo.Context) error {
	studentIDStr := c.Param("studentID")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseIDStr := c.QueryParam("course_id")
	var courseID *int
	if courseIDStr != "" {
		cid, err := strconv.Atoi(courseIDStr)
		if err == nil {
			courseID = &cid
		}
	}

	metrics, err := h.analyticsService.GetStudentPerformance(c.Request().Context(), collegeID, studentID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, metrics, 200)
}

// GetCourseAnalytics retrieves analytics for a course
func (h *AnalyticsHandler) GetCourseAnalytics(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	analytics, err := h.analyticsService.GetCourseAnalytics(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, analytics, 200)
}

// GetCollegeDashboard retrieves dashboard metrics for college
func (h *AnalyticsHandler) GetCollegeDashboard(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	dashboard, err := h.analyticsService.GetCollegeDashboard(c.Request().Context(), collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, dashboard, 200)
}

// GetAttendanceTrends retrieves attendance trends
func (h *AnalyticsHandler) GetAttendanceTrends(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseIDStr := c.QueryParam("course_id")
	var courseID *int
	if courseIDStr != "" {
		cid, err := strconv.Atoi(courseIDStr)
		if err == nil {
			courseID = &cid
		}
	}

	trends, err := h.analyticsService.GetAttendanceTrends(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, trends, 200)
}

// GetGradeDistribution retrieves grade distribution for a course
func (h *AnalyticsHandler) GetGradeDistribution(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	distribution, err := h.analyticsService.GetGradeDistribution(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, distribution, 200)
}
