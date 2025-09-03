package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

type StudentHandler struct {
	studentService student.StudentService
}

func NewStudentHandler(studentService student.StudentService) *StudentHandler {
	return &StudentHandler{
		studentService: studentService,
	}
}

func (h *StudentHandler) UpdateStudent(c echo.Context) error {
	// Extract URL parameters
	id := c.Param("studentID")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Bind the UpdateStudentRequest struct
	var req models.UpdateStudentRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	// Call service method with partial update request
	err = h.studentService.UpdateStudentPartial(c.Request().Context(), collegeID, studentID, &req)
	if err != nil {
		// Handle specific errors (validation, not found, etc.)
		return helpers.Error(c, "failed to update student", 500)
	}

	// PATCH success - should return 204 No Content
	return helpers.Success(c, "Success", 204)
}