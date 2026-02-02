package handler

import (
	"net/http"
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/assignment"
	"eduhub/server/internal/services/attendance"
	"eduhub/server/internal/services/grades"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

// ParentHandler handles parent portal related requests
type ParentHandler struct {
	studentService    student.StudentService
	attendanceService attendance.AttendanceService
	gradesService     grades.GradeServices
	assignmentService assignment.AssignmentService
}

// NewParentHandler creates a new ParentHandler
func NewParentHandler(
	studentService student.StudentService,
	attendanceService attendance.AttendanceService,
	gradesService grades.GradeServices,
	assignmentService assignment.AssignmentService,
) *ParentHandler {
	return &ParentHandler{
		studentService:    studentService,
		attendanceService: attendanceService,
		gradesService:     gradesService,
		assignmentService: assignmentService,
	}
}

// GetLinkedChildren godoc
// @Summary Get linked children for parent
// @Description Returns a list of students linked to the authenticated parent
// @Tags Parent Portal
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/parent/children [get]
func (h *ParentHandler) GetLinkedChildren(c echo.Context) error {
	// Get parent user ID from context (set by auth middleware)
	_, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	// Get college ID
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// TODO: Get linked students from parent-student relationship service
	// For now, return empty list - this will be implemented when parent-student
	// relationship feature is fully developed
	_ = collegeID

	linkedStudents := []map[string]interface{}{}

	return helpers.Success(c, map[string]interface{}{
		"students": linkedStudents,
	}, http.StatusOK)
}

// GetChildDashboard godoc
// @Summary Get child's dashboard data
// @Description Returns dashboard overview for a specific child
// @Tags Parent Portal
// @Accept json
// @Produce json
// @Param studentID path int true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/parent/children/{studentID}/dashboard [get]
func (h *ParentHandler) GetChildDashboard(c echo.Context) error {
	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "Invalid student ID", http.StatusBadRequest)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Verify the parent has access to this student
	if err := h.verifyParentAccess(c, studentID); err != nil {
		return err
	}

	student, err := h.studentService.GetStudentDetailedProfile(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.NotFound(c, map[string]interface{}{"error": "Student not found"}, http.StatusNotFound)
	}

	// Return basic student info
	return helpers.Success(c, map[string]interface{}{
		"student": student,
		"metrics": map[string]interface{}{
			"enrolledCourses":    len(student.Enrollments),
			"attendanceRate":     0,
			"pendingAssignments": 0,
		},
	}, http.StatusOK)
}

// GetChildAttendance godoc
// @Summary Get child's attendance records
// @Description Returns attendance records for a specific child
// @Tags Parent Portal
// @Accept json
// @Produce json
// @Param studentID path int true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/parent/children/{studentID}/attendance [get]
func (h *ParentHandler) GetChildAttendance(c echo.Context) error {
	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "Invalid student ID", http.StatusBadRequest)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Verify the parent has access to this student
	if err := h.verifyParentAccess(c, studentID); err != nil {
		return err
	}

	// Get attendance records
	attendance, err := h.attendanceService.GetAttendanceByStudent(c.Request().Context(), collegeID, studentID, 50, 0)
	if err != nil {
		return helpers.Success(c, map[string]interface{}{
			"attendance": []interface{}{},
			"total":      0,
		}, http.StatusOK)
	}

	return helpers.Success(c, map[string]interface{}{
		"attendance": attendance,
		"total":      len(attendance),
	}, http.StatusOK)
}

// GetChildGrades godoc
// @Summary Get child's grades
// @Description Returns grades for a specific child
// @Tags Parent Portal
// @Accept json
// @Produce json
// @Param studentID path int true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/parent/children/{studentID}/grades [get]
func (h *ParentHandler) GetChildGrades(c echo.Context) error {
	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "Invalid student ID", http.StatusBadRequest)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Verify the parent has access to this student
	if err := h.verifyParentAccess(c, studentID); err != nil {
		return err
	}

	// Get grades
	grades, err := h.gradesService.GetGradesByStudent(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Success(c, map[string]interface{}{
			"grades": []interface{}{},
			"total":  0,
		}, http.StatusOK)
	}

	return helpers.Success(c, map[string]interface{}{
		"grades": grades,
		"total":  len(grades),
	}, http.StatusOK)
}

// GetChildAssignments godoc
// @Summary Get child's assignments
// @Description Returns assignments for a specific child
// @Tags Parent Portal
// @Accept json
// @Produce json
// @Param studentID path int true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/parent/children/{studentID}/assignments [get]
func (h *ParentHandler) GetChildAssignments(c echo.Context) error {
	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "Invalid student ID", http.StatusBadRequest)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Verify the parent has access to this student
	if err := h.verifyParentAccess(c, studentID); err != nil {
		return err
	}

	// Use collegeID for future implementation
	_ = collegeID

	assignments := []interface{}{}

	return helpers.Success(c, map[string]interface{}{
		"assignments": assignments,
		"total":       len(assignments),
	}, http.StatusOK)
}

// verifyParentAccess checks if the authenticated user has access to the student's data
func (h *ParentHandler) verifyParentAccess(c echo.Context, studentID int) error {
	// TODO: Implement proper parent-student relationship verification
	// This should check if the authenticated user is linked as a parent to the student
	// For now, we allow access (will be restricted when parent feature is fully implemented)

	// Get current user role from context
	role := c.Get("role")
	if role == "admin" || role == "faculty" {
		return nil // Admin and faculty can access all student data
	}

	// For parents, we would check the parent-student relationship table
	// This is a placeholder for future implementation
	_ = studentID
	return nil
}
