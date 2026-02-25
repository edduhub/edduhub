package handler

import (
	"math"
	"sort"
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/course"
	"eduhub/server/internal/services/grades"

	"github.com/labstack/echo/v4"
)

type GradeHandler struct {
	gradeService  grades.GradeServices
	courseService course.CourseService
}

func NewGradeHandler(gradeService grades.GradeServices, courseService course.CourseService) *GradeHandler {
	return &GradeHandler{
		gradeService:  gradeService,
		courseService: courseService,
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

// GetMyGrades returns all grades for the currently authenticated student
func (h *GradeHandler) GetMyGrades(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "student ID required", 400)
	}

	grades, err := h.gradeService.GetGradesByStudent(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, grades, 200)
}

// GetMyCourseGrades returns aggregated grades by course for current student
func (h *GradeHandler) GetMyCourseGrades(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "student ID required", 400)
	}

	grades, err := h.gradeService.GetGradesByStudent(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	// Group grades by course
	type aggregatedCourse struct {
		totalScore float64
		maxScore   float64
	}

	aggByCourse := make(map[int]*aggregatedCourse)
	for _, grade := range grades {
		agg, ok := aggByCourse[grade.CourseID]
		if !ok {
			agg = &aggregatedCourse{}
			aggByCourse[grade.CourseID] = agg
		}
		agg.totalScore += float64(grade.ObtainedMarks)
		agg.maxScore += float64(grade.TotalMarks)
	}

	courseIDs := make([]int, 0, len(aggByCourse))
	for id := range aggByCourse {
		courseIDs = append(courseIDs, id)
	}
	sort.Ints(courseIDs)

	result := make([]map[string]any, 0, len(courseIDs))
	for _, courseID := range courseIDs {
		agg := aggByCourse[courseID]
		percentage := 0.0
		letter := ""
		if agg.maxScore > 0 {
			percentage = math.Round((agg.totalScore/agg.maxScore)*10000) / 100
			switch {
			case percentage >= 90:
				letter = "A+"
			case percentage >= 85:
				letter = "A"
			case percentage >= 80:
				letter = "A-"
			case percentage >= 75:
				letter = "B+"
			case percentage >= 70:
				letter = "B"
			case percentage >= 65:
				letter = "B-"
			case percentage >= 60:
				letter = "C+"
			case percentage >= 55:
				letter = "C"
			case percentage >= 50:
				letter = "D"
			default:
				letter = "F"
			}
		}

		courseName := "Course " + strconv.Itoa(courseID)
		courseCode := "COURSE-" + strconv.Itoa(courseID)
		credits := 3
		courseDetails, err := h.courseService.FindCourseByID(c.Request().Context(), collegeID, courseID)
		if err == nil && courseDetails != nil {
			courseName = courseDetails.Name
			if courseDetails.Credits > 0 {
				credits = courseDetails.Credits
			}
		}

		result = append(result, map[string]any{
			"courseId":    courseID,
			"courseName":  courseName,
			"courseCode":  courseCode,
			"credits":     credits,
			"totalScore":  agg.totalScore,
			"maxScore":    agg.maxScore,
			"percentage":  percentage,
			"letterGrade": letter,
		})
	}

	return helpers.Success(c, result, 200)
}
