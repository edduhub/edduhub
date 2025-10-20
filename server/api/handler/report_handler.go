package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/report"

	"github.com/labstack/echo/v4"
)

type ReportHandler struct {
	reportService report.ReportService
}

func NewReportHandler(reportService report.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// GenerateGradeCard generates a PDF grade card for a student
func (h *ReportHandler) GenerateGradeCard(c echo.Context) error {
	studentIDStr := c.Param("studentID")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	semesterStr := c.QueryParam("semester")
	var semester *int
	if semesterStr != "" {
		sem, err := strconv.Atoi(semesterStr)
		if err == nil {
			semester = &sem
		}
	}

	pdfBytes, err := h.reportService.GenerateGradeCard(c.Request().Context(), collegeID, studentID, semester)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=grade_card.pdf")
	return c.Blob(200, "application/pdf", pdfBytes)
}

// GenerateMyGradeCard generates a PDF grade card for the current student user
func (h *ReportHandler) GenerateMyGradeCard(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Extract student ID from context (set by LoadStudentProfile middleware)
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "student profile not found", 400)
	}

	semesterStr := c.QueryParam("semester")
	var semester *int
	if semesterStr != "" {
		sem, err := strconv.Atoi(semesterStr)
		if err == nil {
			semester = &sem
		}
	}

	pdfBytes, err := h.reportService.GenerateGradeCard(c.Request().Context(), collegeID, studentID, semester)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=grade_card.pdf")
	return c.Blob(200, "application/pdf", pdfBytes)
}

// GenerateMyTranscript generates an official transcript for the current student user
func (h *ReportHandler) GenerateMyTranscript(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Extract student ID from context (set by LoadStudentProfile middleware)
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "student profile not found", 400)
	}

	pdfBytes, err := h.reportService.GenerateTranscript(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=transcript.pdf")
	return c.Blob(200, "application/pdf", pdfBytes)
}

// GenerateTranscript generates an official transcript for a student
func (h *ReportHandler) GenerateTranscript(c echo.Context) error {
	studentIDStr := c.Param("studentID")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	pdfBytes, err := h.reportService.GenerateTranscript(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=transcript.pdf")
	return c.Blob(200, "application/pdf", pdfBytes)
}

// GenerateAttendanceReport generates attendance report for a course
func (h *ReportHandler) GenerateAttendanceReport(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	pdfBytes, err := h.reportService.GenerateAttendanceReport(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=attendance_report.pdf")
	return c.Blob(200, "application/pdf", pdfBytes)
}

// GenerateCourseReport generates comprehensive course report
func (h *ReportHandler) GenerateCourseReport(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	pdfBytes, err := h.reportService.GenerateCourseReport(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=course_report.pdf")
	return c.Blob(200, "application/pdf", pdfBytes)
}
