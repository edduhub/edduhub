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

func (h *StudentHandler) ListStudents(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	limit := uint64(10)
	offset := uint64(0)
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

	students, err := h.studentService.ListStudents(c.Request().Context(), collegeID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, students, 200)
}

func (h *StudentHandler) CreateStudent(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var student models.Student
	if err := c.Bind(&student); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	student.CollegeID = collegeID

	err = h.studentService.CreateStudent(c.Request().Context(), &student)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, student, 201)
}

func (h *StudentHandler) GetStudent(c echo.Context) error {
	id := c.Param("studentID")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	profile, err := h.studentService.GetStudentDetailedProfile(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 404)
	}

	return helpers.Success(c, profile, 200)
}

func (h *StudentHandler) UpdateStudent(c echo.Context) error {
	id := c.Param("studentID")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateStudentRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.studentService.UpdateStudentPartial(c.Request().Context(), collegeID, studentID, &req)
	if err != nil {
		return helpers.Error(c, "failed to update student", 500)
	}

	return helpers.Success(c, "Success", 204)
}

func (h *StudentHandler) DeleteStudent(c echo.Context) error {
	id := c.Param("studentID")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.studentService.DeleteStudent(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Student deleted successfully", 204)
}

func (h *StudentHandler) FreezeStudent(c echo.Context) error {
	id := c.Param("studentID")
	studentID, err := strconv.Atoi(id)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.studentService.FreezeStudent(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Student frozen successfully", 200)
}