package handler

import (
	"encoding/csv"
	"fmt"
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/batch"

	"github.com/labstack/echo/v4"
)

type BatchHandler struct {
	batchService batch.BatchService
}

func NewBatchHandler(batchService batch.BatchService) *BatchHandler {
	return &BatchHandler{
		batchService: batchService,
	}
}

// ImportStudents imports students from CSV file
func (h *BatchHandler) ImportStudents(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return helpers.Error(c, "file is required", 400)
	}

	src, err := file.Open()
	if err != nil {
		return helpers.Error(c, "failed to open file", 500)
	}
	defer src.Close()

	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		return helpers.Error(c, "failed to parse CSV", 400)
	}

	// Skip header row
	if len(records) < 2 {
		return helpers.Error(c, "CSV file is empty", 400)
	}

	var students []models.Student
	for i, record := range records[1:] {
		if len(record) < 1 {
			return helpers.Error(c, fmt.Sprintf("invalid record at line %d", i+2), 400)
		}

		student := models.Student{
			RollNo:    record[0],
			CollegeID: collegeID,
		}
		students = append(students, student)
	}

	result, err := h.batchService.ImportStudents(c.Request().Context(), collegeID, students)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, result, 200)
}

// ExportStudents exports students to CSV
func (h *BatchHandler) ExportStudents(c echo.Context) error {
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

	csvData, err := h.batchService.ExportStudents(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=students.csv")
	return c.String(200, csvData)
}

// ImportGrades imports grades from CSV file
func (h *BatchHandler) ImportGrades(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseIDStr := c.FormValue("course_id")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "course_id is required", 400)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return helpers.Error(c, "file is required", 400)
	}

	src, err := file.Open()
	if err != nil {
		return helpers.Error(c, "failed to open file", 500)
	}
	defer src.Close()

	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		return helpers.Error(c, "failed to parse CSV", 400)
	}

	result, err := h.batchService.ImportGrades(c.Request().Context(), collegeID, courseID, records)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, result, 200)
}

// ExportGrades exports grades to CSV
func (h *BatchHandler) ExportGrades(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseIDStr := c.QueryParam("course_id")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "course_id is required", 400)
	}

	csvData, err := h.batchService.ExportGrades(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=grades_course_%d.csv", courseID))
	return c.String(200, csvData)
}

// BulkEnroll enrolls multiple students to a course
func (h *BatchHandler) BulkEnroll(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req struct {
		CourseID   int   `json:"course_id"`
		StudentIDs []int `json:"student_ids"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	result, err := h.batchService.BulkEnroll(c.Request().Context(), collegeID, req.CourseID, req.StudentIDs)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, result, 200)
}
