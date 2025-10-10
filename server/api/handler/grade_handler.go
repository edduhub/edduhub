package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/grades"

	"github.com/labstack/echo/v4"
)

type GradeHandler struct {
	gradeService grades.GradeServices
}

func NewGradeHandler(gradeService grades.GradeServices) *GradeHandler {
	return &GradeHandler{
		gradeService: gradeService,
	}
}

func (h *GradeHandler) GetGradesByCourse(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	grades, err := h.gradeService.GetGradesByCourse(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, grades, 200)
}

func (h *GradeHandler) CreateAssessment(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var grade models.Grade
	if err := c.Bind(&grade); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	grade.CourseID = courseID
	grade.CollegeID = collegeID

	if grade.StudentID == 0 {
		return helpers.Error(c, "student_id is required", 400)
	}

	if grade.AssessmentName == "" {
		return helpers.Error(c, "assessment_name is required", 400)
	}

	if grade.AssessmentType == "" {
		return helpers.Error(c, "assessment_type is required", 400)
	}

	err = h.gradeService.CreateGrade(c.Request().Context(), &grade)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, grade, 201)
}

func (h *GradeHandler) UpdateAssessment(c echo.Context) error {
	assessmentIDStr := c.Param("assessmentID")
	assessmentID, err := strconv.Atoi(assessmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid assessment ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateGradeRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.gradeService.UpdateGradePartial(c.Request().Context(), collegeID, assessmentID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]string{"message": "assessment updated successfully"}, 200)
}

func (h *GradeHandler) DeleteAssessment(c echo.Context) error {
	assessmentIDStr := c.Param("assessmentID")
	assessmentID, err := strconv.Atoi(assessmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid assessment ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.gradeService.DeleteGrade(c.Request().Context(), collegeID, assessmentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]string{"message": "assessment deleted successfully"}, 200)
}

func (h *GradeHandler) SubmitScores(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	_, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	assessmentIDStr := c.Param("assessmentID")
	assessmentID, err := strconv.Atoi(assessmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid assessment ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateGradeRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.gradeService.UpdateGradePartial(c.Request().Context(), collegeID, assessmentID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]string{"message": "scores submitted successfully"}, 200)
}

func (h *GradeHandler) GetStudentGrades(c echo.Context) error {
	studentIDStr := c.Param("studentID")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	grades, err := h.gradeService.GetGradesByStudent(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, grades, 200)
}
