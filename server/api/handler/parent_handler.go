package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/assignment"
	"eduhub/server/internal/services/attendance"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/email"
	"eduhub/server/internal/services/grades"
	"eduhub/server/internal/services/student"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// ParentHandler handles parent portal related requests
type ParentHandler struct {
	studentService    student.StudentService
	attendanceService attendance.AttendanceService
	gradesService     grades.GradeServices
	assignmentService assignment.AssignmentService
	emailService      email.EmailService
	db                *repository.DB
}

// NewParentHandler creates a new ParentHandler
func NewParentHandler(
	studentService student.StudentService,
	attendanceService attendance.AttendanceService,
	gradesService grades.GradeServices,
	assignmentService assignment.AssignmentService,
	emailService email.EmailService,
	db *repository.DB,
) *ParentHandler {
	return &ParentHandler{
		studentService:    studentService,
		attendanceService: attendanceService,
		gradesService:     gradesService,
		assignmentService: assignmentService,
		emailService:      emailService,
		db:                db,
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
	role := h.currentRole(c)
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	students, err := h.studentService.ListStudents(c.Request().Context(), collegeID, 1000, 0)
	if err != nil {
		return helpers.Error(c, "Failed to fetch students", http.StatusInternalServerError)
	}

	linkedStudents := make([]map[string]interface{}, 0, len(students))
	if role == "admin" || role == "faculty" {
		for _, student := range students {
			linkedStudents = append(linkedStudents, map[string]interface{}{
				"id":             student.StudentID,
				"rollNo":         student.RollNo,
				"enrollmentYear": student.EnrollmentYear,
				"isActive":       student.IsActive,
			})
		}
		return helpers.Success(c, map[string]interface{}{
			"students": linkedStudents,
		}, http.StatusOK)
	}

	kratosID, err := helpers.GetKratosID(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}
	parentUserID, err := h.resolveParentUserID(c.Request().Context(), kratosID)
	if err != nil {
		return helpers.Success(c, map[string]interface{}{
			"students": linkedStudents,
		}, http.StatusOK)
	}

	linkedStudentIDs, err := h.getLinkedStudentIDSet(c.Request().Context(), collegeID, parentUserID)
	if err != nil {
		return helpers.Error(c, "Failed to fetch linked students", http.StatusInternalServerError)
	}

	for _, student := range students {
		if _, ok := linkedStudentIDs[student.StudentID]; !ok {
			continue
		}
		linkedStudents = append(linkedStudents, map[string]interface{}{
			"id":             student.StudentID,
			"rollNo":         student.RollNo,
			"enrollmentYear": student.EnrollmentYear,
			"isActive":       student.IsActive,
		})
	}

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

	// Fetch real assignments for the student
	assignments, err := h.assignmentService.GetAssignmentsByStudent(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Success(c, map[string]interface{}{
			"assignments": []interface{}{},
			"total":       0,
		}, http.StatusOK)
	}

	return helpers.Success(c, map[string]interface{}{
		"assignments": assignments,
		"total":       len(assignments),
	}, http.StatusOK)
}

// verifyParentAccess checks if the authenticated user has access to the student's data
func (h *ParentHandler) verifyParentAccess(c echo.Context, studentID int) error {
	role := h.currentRole(c)
	if role == "admin" || role == "faculty" {
		return nil
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	kratosID, err := helpers.GetKratosID(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	parentUserID, err := h.resolveParentUserID(c.Request().Context(), kratosID)
	if err != nil {
		return helpers.Error(c, "Forbidden: Parent account is not linked", http.StatusForbidden)
	}

	ctx := c.Request().Context()
	var exists bool
	err = h.db.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM parent_student_relationships
			WHERE college_id = $1
			  AND parent_user_id = $2
			  AND student_id = $3
			  AND is_verified = TRUE
		)`,
		collegeID, parentUserID, studentID,
	).Scan(&exists)
	if err != nil {
		return helpers.Error(c, "Failed to verify parent access", http.StatusInternalServerError)
	}

	if exists {
		return nil
	}

	return helpers.Error(c, "Forbidden: You don't have access to this student's data", http.StatusForbidden)
}

// ContactParent sends a direct email to a parent from faculty/admin users.
func (h *ParentHandler) ContactParent(c echo.Context) error {
	role := h.currentRole(c)
	if role != "admin" && role != "faculty" {
		return helpers.Error(c, "Forbidden", http.StatusForbidden)
	}

	var req struct {
		ParentName string `json:"parentName"`
		Email      string `json:"email"`
		Phone      string `json:"phone"`
		Subject    string `json:"subject"`
		Message    string `json:"message"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Subject) == "" || strings.TrimSpace(req.Message) == "" {
		return helpers.Error(c, "Email, subject, and message are required", http.StatusBadRequest)
	}

	body := fmt.Sprintf(
		"<html><body><p><strong>Parent:</strong> %s</p><p><strong>Phone:</strong> %s</p><p>%s</p></body></html>",
		strings.TrimSpace(req.ParentName),
		strings.TrimSpace(req.Phone),
		strings.ReplaceAll(strings.TrimSpace(req.Message), "\n", "<br/>"),
	)
	if err := h.emailService.SendEmail(c.Request().Context(), strings.TrimSpace(req.Email), strings.TrimSpace(req.Subject), body); err != nil {
		return helpers.Error(c, "Failed to send parent contact email", http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]string{"status": "sent"}, http.StatusOK)
}

func (h *ParentHandler) currentRole(c echo.Context) string {
	identity, ok := c.Get("identity").(*auth.Identity)
	if !ok || identity == nil {
		return ""
	}
	return identity.Traits.Role
}

func (h *ParentHandler) resolveParentUserID(ctx context.Context, kratosID string) (int, error) {
	var userID int
	err := h.db.Pool.QueryRow(ctx,
		`SELECT id FROM users WHERE kratos_identity_id = $1 AND is_active = TRUE`,
		kratosID,
	).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("parent user not found")
		}
		return 0, err
	}
	return userID, nil
}

func (h *ParentHandler) getLinkedStudentIDSet(ctx context.Context, collegeID, parentUserID int) (map[int]struct{}, error) {
	rows, err := h.db.Pool.Query(ctx, `
		SELECT student_id
		FROM parent_student_relationships
		WHERE college_id = $1
		  AND parent_user_id = $2
		  AND is_verified = TRUE`,
		collegeID, parentUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make(map[int]struct{})
	for rows.Next() {
		var studentID int
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}
		ids[studentID] = struct{}{}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ids, nil
}
