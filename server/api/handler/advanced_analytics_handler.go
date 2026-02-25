package handler

import (
	"strconv"
	"strings"
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/analytics"

	"github.com/labstack/echo/v4"
)

type AdvancedAnalyticsHandler struct {
	advancedAnalyticsService analytics.AdvancedAnalyticsService
}

func NewAdvancedAnalyticsHandler(advancedAnalyticsService analytics.AdvancedAnalyticsService) *AdvancedAnalyticsHandler {
	return &AdvancedAnalyticsHandler{
		advancedAnalyticsService: advancedAnalyticsService,
	}
}

// GetStudentProgression retrieves detailed progression analytics for a student
func (h *AdvancedAnalyticsHandler) GetStudentProgression(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentIDStr := c.Param("studentID")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	progression, err := h.advancedAnalyticsService.GetStudentProgression(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, progression, 200)
}

// GetCourseEngagement retrieves detailed engagement analytics for a course
func (h *AdvancedAnalyticsHandler) GetCourseEngagement(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	engagement, err := h.advancedAnalyticsService.GetCourseEngagement(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, engagement, 200)
}

// GetPredictiveInsights retrieves predictive analytics and insights
func (h *AdvancedAnalyticsHandler) GetPredictiveInsights(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	insights, err := h.advancedAnalyticsService.GetPredictiveInsights(c.Request().Context(), collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, insights, 200)
}

// GetLearningAnalytics retrieves comprehensive learning analytics
func (h *AdvancedAnalyticsHandler) GetLearningAnalytics(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Parse optional date filters
	var startDate, endDate *time.Time

	startDateStr := c.QueryParam("start_date")
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}

	endDateStr := c.QueryParam("end_date")
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	analytics, err := h.advancedAnalyticsService.GetLearningAnalytics(c.Request().Context(), collegeID, startDate, endDate)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, analytics, 200)
}

// GetPerformanceTrends retrieves performance trends over time
func (h *AdvancedAnalyticsHandler) GetPerformanceTrends(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	entityType := c.Param("entityType")
	entityIDStr := c.Param("entityID")
	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		return helpers.Error(c, "invalid entity ID", 400)
	}

	trends, err := h.advancedAnalyticsService.GetPerformanceTrends(c.Request().Context(), collegeID, entityType, entityID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, trends, 200)
}

// GetComparativeAnalysis retrieves comparative analysis between courses
func (h *AdvancedAnalyticsHandler) GetComparativeAnalysis(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseIDsParam := c.QueryParam("course_ids")
	if courseIDsParam == "" {
		return helpers.Error(c, "course_ids parameter is required", 400)
	}

	// Parse comma-separated course IDs
	var courseIDs []int
	for idStr := range strings.SplitSeq(courseIDsParam, ",") {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			courseIDs = append(courseIDs, id)
		}
	}

	if len(courseIDs) < 2 {
		return helpers.Error(c, "at least 2 course IDs are required for comparison", 400)
	}

	analysis, err := h.advancedAnalyticsService.GetComparativeAnalysis(c.Request().Context(), collegeID, courseIDs)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, analysis, 200)
}
