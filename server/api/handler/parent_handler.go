package handler

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
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

	linkedStudents := make([]map[string]any, 0, len(students))
	if role == "admin" || role == "faculty" {
		for _, student := range students {
			linkedStudents = append(linkedStudents, map[string]any{
				"id":             student.StudentID,
				"rollNo":         student.RollNo,
				"enrollmentYear": student.EnrollmentYear,
				"isActive":       student.IsActive,
			})
		}
		return helpers.Success(c, map[string]any{
			"students": linkedStudents,
		}, http.StatusOK)
	}

	kratosID, err := helpers.GetKratosID(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}
	parentUserID, err := h.resolveParentUserID(c.Request().Context(), kratosID)
	if err != nil {
		return helpers.Success(c, map[string]any{
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
		linkedStudents = append(linkedStudents, map[string]any{
			"id":             student.StudentID,
			"rollNo":         student.RollNo,
			"enrollmentYear": student.EnrollmentYear,
			"isActive":       student.IsActive,
		})
	}

	return helpers.Success(c, map[string]any{
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
		return helpers.NotFound(c, map[string]any{"error": "Student not found"}, http.StatusNotFound)
	}

	// Return basic student info
	attendanceRecords, _ := h.attendanceService.GetAttendanceByStudent(c.Request().Context(), collegeID, studentID, 1000, 0)
	attendanceRate := calculateAttendanceRate(attendanceRecords)

	grades, _ := h.gradesService.GetGradesByStudent(c.Request().Context(), collegeID, studentID)
	averageGrade := calculateAverageGrade(grades)

	pendingAssignments, err := h.getPendingAssignmentCount(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, "Failed to compute dashboard metrics", http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]any{
		"student": student,
		"metrics": map[string]any{
			"enrolledCourses":    len(student.Enrollments),
			"attendanceRate":     attendanceRate,
			"pendingAssignments": pendingAssignments,
			"averageGrade":       averageGrade,
			"assessmentsCount":   len(grades),
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
		return helpers.Success(c, map[string]any{
			"attendance": []any{},
			"total":      0,
		}, http.StatusOK)
	}

	return helpers.Success(c, map[string]any{
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
		return helpers.Success(c, map[string]any{
			"grades": []any{},
			"total":  0,
		}, http.StatusOK)
	}

	return helpers.Success(c, map[string]any{
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
		return helpers.Success(c, map[string]any{
			"assignments": []any{},
			"total":       0,
		}, http.StatusOK)
	}

	return helpers.Success(c, map[string]any{
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

// ListParentRelationships returns all parent-student relationships for the admin's college (admin only).
func (h *ParentHandler) ListParentRelationships(c echo.Context) error {
	if h.currentRole(c) != "admin" {
		return helpers.Error(c, "Forbidden", http.StatusForbidden)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	rows, err := h.db.Pool.Query(c.Request().Context(), `
		SELECT
			psr.id,
			psr.parent_user_id,
			u.name AS parent_name,
			u.email AS parent_email,
			psr.student_id,
			s.roll_no AS student_roll_no,
			u2.name AS student_name,
			psr.relation,
			psr.is_primary_contact,
			psr.receive_notifications,
			psr.is_verified,
			psr.created_at
		FROM parent_student_relationships psr
		JOIN users u ON u.id = psr.parent_user_id
		JOIN students s ON s.student_id = psr.student_id
		JOIN users u2 ON u2.id = s.user_id
		WHERE psr.college_id = $1
		ORDER BY psr.created_at DESC`,
		collegeID,
	)
	if err != nil {
		return helpers.Error(c, "Failed to fetch relationships", http.StatusInternalServerError)
	}
	defer rows.Close()

	type rel struct {
		ID                   int    `json:"id"`
		ParentUserID         int    `json:"parentUserId"`
		ParentName           string `json:"parentName"`
		ParentEmail          string `json:"parentEmail"`
		StudentID            int    `json:"studentId"`
		StudentRollNo        string `json:"studentRollNo"`
		StudentName          string `json:"studentName"`
		Relation             string `json:"relation"`
		IsPrimaryContact     bool   `json:"isPrimaryContact"`
		ReceiveNotifications bool   `json:"receiveNotifications"`
		IsVerified           bool   `json:"isVerified"`
		CreatedAt            string `json:"createdAt"`
	}

	var relationships []rel
	for rows.Next() {
		var r rel
		var createdAt time.Time
		if err := rows.Scan(
			&r.ID, &r.ParentUserID, &r.ParentName, &r.ParentEmail,
			&r.StudentID, &r.StudentRollNo, &r.StudentName,
			&r.Relation, &r.IsPrimaryContact, &r.ReceiveNotifications, &r.IsVerified,
			&createdAt,
		); err != nil {
			return helpers.Error(c, "Failed to scan relationship", http.StatusInternalServerError)
		}
		r.CreatedAt = createdAt.Format(time.RFC3339)
		relationships = append(relationships, r)
	}
	if rows.Err() != nil {
		return helpers.Error(c, "Failed to iterate relationships", http.StatusInternalServerError)
	}

	if relationships == nil {
		relationships = []rel{}
	}
	return helpers.Success(c, map[string]any{"relationships": relationships}, http.StatusOK)
}

// CreateParentRelationship creates a parent-student link (admin only).
func (h *ParentHandler) CreateParentRelationship(c echo.Context) error {
	if h.currentRole(c) != "admin" {
		return helpers.Error(c, "Forbidden", http.StatusForbidden)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req struct {
		ParentUserID         int    `json:"parentUserId" validate:"required"`
		StudentID            int    `json:"studentId" validate:"required"`
		Relation             string `json:"relation" validate:"required,oneof=father mother guardian"`
		IsPrimaryContact     bool   `json:"isPrimaryContact"`
		ReceiveNotifications bool   `json:"receiveNotifications"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	// Verify parent user exists and has role=parent
	var parentRole string
	err = h.db.Pool.QueryRow(c.Request().Context(),
		`SELECT role FROM users WHERE id = $1 AND is_active = TRUE`,
		req.ParentUserID,
	).Scan(&parentRole)
	if err != nil {
		if err == pgx.ErrNoRows {
			return helpers.Error(c, "Parent user not found or inactive", http.StatusBadRequest)
		}
		return helpers.Error(c, "Failed to verify parent user", http.StatusInternalServerError)
	}
	if parentRole != "parent" {
		return helpers.Error(c, "User is not a parent", http.StatusBadRequest)
	}

	// Verify student exists and belongs to college
	var studentCollegeID int
	err = h.db.Pool.QueryRow(c.Request().Context(),
		`SELECT college_id FROM students WHERE student_id = $1 AND is_active = TRUE`,
		req.StudentID,
	).Scan(&studentCollegeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return helpers.Error(c, "Student not found or inactive", http.StatusBadRequest)
		}
		return helpers.Error(c, "Failed to verify student", http.StatusInternalServerError)
	}
	if studentCollegeID != collegeID {
		return helpers.Error(c, "Student does not belong to your college", http.StatusBadRequest)
	}

	// Check for duplicate
	var existing int
	dupErr := h.db.Pool.QueryRow(c.Request().Context(),
		`SELECT 1 FROM parent_student_relationships WHERE college_id = $1 AND parent_user_id = $2 AND student_id = $3`,
		collegeID, req.ParentUserID, req.StudentID,
	).Scan(&existing)
	if dupErr == nil {
		return helpers.Error(c, "This parent-student link already exists", http.StatusConflict)
	}
	if dupErr != pgx.ErrNoRows {
		return helpers.Error(c, "Failed to check existing link", http.StatusInternalServerError)
	}

	// Insert
	var id int
	err = h.db.Pool.QueryRow(c.Request().Context(), `
		INSERT INTO parent_student_relationships
			(parent_user_id, student_id, college_id, relation, is_primary_contact, receive_notifications, is_verified, verified_at)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, NOW())
		RETURNING id`,
		req.ParentUserID, req.StudentID, collegeID, req.Relation, req.IsPrimaryContact, req.ReceiveNotifications,
	).Scan(&id)
	if err != nil {
		return helpers.Error(c, "Failed to create link: "+err.Error(), http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]any{"id": id, "message": "Parent-student link created"}, http.StatusCreated)
}

// DeleteParentRelationship removes a parent-student link (admin only).
func (h *ParentHandler) DeleteParentRelationship(c echo.Context) error {
	if h.currentRole(c) != "admin" {
		return helpers.Error(c, "Forbidden", http.StatusForbidden)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	relID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return helpers.Error(c, "Invalid relationship ID", http.StatusBadRequest)
	}

	result, err := h.db.Pool.Exec(c.Request().Context(),
		`DELETE FROM parent_student_relationships WHERE id = $1 AND college_id = $2`,
		relID, collegeID,
	)
	if err != nil {
		return helpers.Error(c, "Failed to delete link", http.StatusInternalServerError)
	}
	if result.RowsAffected() == 0 {
		return helpers.NotFound(c, map[string]any{"error": "Relationship not found"}, http.StatusNotFound)
	}

	return helpers.Success(c, map[string]string{"message": "Link removed"}, http.StatusOK)
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

func (h *ParentHandler) getPendingAssignmentCount(ctx context.Context, collegeID, studentID int) (int, error) {
	var count int
	err := h.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM assignments a
		JOIN enrollments e
		  ON e.course_id = a.course_id
		 AND e.college_id = a.college_id
		 AND e.student_id = $2
		LEFT JOIN assignment_submissions s
		  ON s.assignment_id = a.id
		 AND s.student_id = $2
		WHERE a.college_id = $1
		  AND s.id IS NULL
		  AND a.due_date >= NOW()`,
		collegeID, studentID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func calculateAttendanceRate(records []*models.Attendance) float64 {
	if len(records) == 0 {
		return 0
	}

	present := 0
	for _, record := range records {
		if strings.EqualFold(record.Status, "present") {
			present++
		}
	}

	rate := (float64(present) / float64(len(records))) * 100
	return math.Round(rate*100) / 100
}

func calculateAverageGrade(grades []*models.Grade) float64 {
	if len(grades) == 0 {
		return 0
	}

	total := 0.0
	for _, grade := range grades {
		total += grade.Percentage
	}

	avg := total / float64(len(grades))
	return math.Round(avg*100) / 100
}
