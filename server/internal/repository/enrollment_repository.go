package repository

import (
	"context"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type EnrollmentRepository interface {
	CreateEnrollment(ctx context.Context, enrollment *models.Enrollment) error
	IsStudentEnrolled(ctx context.Context, collegeID int, studentID int, courseID int) (bool, error)
	GetEnrollmentByID(ctx context.Context, collegeID int, enrollmentID int) (*models.Enrollment, error) // Added collegeID for scoping
	UpdateEnrollment(ctx context.Context, enrollment *models.Enrollment) error
	UpdateEnrollmentStatus(ctx context.Context, collegeID int, enrollmentID int, status string) error // Added collegeID for scoping
	UpdateEnrollmentPartial(ctx context.Context, collegeID int, enrollmentID int, req *models.UpdateEnrollmentRequest) error
	DeleteEnrollment(ctx context.Context, collegeID int, enrollmentID int) error

	// Find methods with pagination
	FindEnrollmentsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.Enrollment, error)
	FindEnrollmentsByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Enrollment, error)
	FindEnrollmentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Enrollment, error)

	// Count methods
	CountEnrollmentsByStudent(ctx context.Context, collegeID int, studentID int) (int, error)
	CountEnrollmentsByCourse(ctx context.Context, collegeID int, courseID int) (int, error)
	CountEnrollmentsByCollege(ctx context.Context, collegeID int) (int, error)
}

// enrollmentRepository now holds a direct reference to *DB
type enrollmentRepository struct {
	DB *DB
}

// NewEnrollmentRepository receives the *DB directly
func NewEnrollmentRepository(db *DB) EnrollmentRepository {
	return &enrollmentRepository{
		DB: db,
	}
}

const enrollmentTable = "enrollments" // Define your table name

// CreateEnrollment inserts a new enrollment record into the database.
func (e *enrollmentRepository) CreateEnrollment(ctx context.Context, enrollment *models.Enrollment) error {
	// Set timestamps if they are zero-valued
	now := time.Now()
	if enrollment.CreatedAt.IsZero() {
		enrollment.CreatedAt = now
	}
	if enrollment.UpdatedAt.IsZero() {
		enrollment.UpdatedAt = now
	}

	sql := `INSERT INTO enrollments (student_id, course_id, college_id, enrollment_date, status, grade, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	args := []any{enrollment.StudentID, enrollment.CourseID, enrollment.CollegeID, enrollment.EnrollmentDate, enrollment.Status, enrollment.Grade, enrollment.CreatedAt, enrollment.UpdatedAt}

	// Execute the query and scan the returned ID back into the struct
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, e.DB.Pool, &temp, sql, args...)
	if err != nil {
		return fmt.Errorf("CreateEnrollment: failed to execute query or scan ID: %w", err)
	}
	enrollment.ID = temp.ID

	return nil // Success
}

// IsStudentEnrolled checks if a student is enrolled in a specific course within a college.
func (e *enrollmentRepository) IsStudentEnrolled(ctx context.Context, collegeID int, studentID int, courseID int) (bool, error) {
	sql := `SELECT 1 FROM enrollments WHERE college_id = $1 AND student_id = $2 AND course_id = $3 LIMIT 1`
	args := []any{collegeID, studentID, courseID}

	temp := struct {
		Exists int `db:"1"`
	}{}
	err := pgxscan.Get(ctx, e.DB.Pool, &temp, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("IsStudentEnrolled: failed to execute query: %w", err)
	}

	return true, nil // Record exists
}

// GetEnrollmentByID retrieves a specific enrollment by its ID, scoped by collegeID.
func (e *enrollmentRepository) GetEnrollmentByID(ctx context.Context, collegeID int, enrollmentID int) (*models.Enrollment, error) {
	sql := `SELECT id, student_id, course_id, college_id, enrollment_date, status, grade, created_at, updated_at FROM enrollments WHERE id = $1 AND college_id = $2`
	args := []any{enrollmentID, collegeID}

	enrollment := &models.Enrollment{}

	// Use pgxscan.Get for a single row result
	err := pgxscan.Get(ctx, e.DB.Pool, enrollment, sql, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("GetEnrollmentByID: enrollment with ID %d not found for college ID %d", enrollmentID, collegeID)
		}
		return nil, fmt.Errorf("GetEnrollmentByID: failed to execute query or scan: %w", err)
	}

	return enrollment, nil // Success
}

// FindEnrollmentsByStudent retrieves all enrollment records for a specific student in a college.
func (e *enrollmentRepository) FindEnrollmentsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.Enrollment, error) {
	sql := `SELECT id, student_id, course_id, college_id, enrollment_date, status, grade, created_at, updated_at FROM enrollments WHERE college_id = $1 AND student_id = $2 ORDER BY enrollment_date DESC, course_id ASC LIMIT $3 OFFSET $4`
	args := []any{collegeID, studentID, limit, offset}

	enrollments := []*models.Enrollment{}

	// Use pgxscan.Select for multiple rows
	err := pgxscan.Select(ctx, e.DB.Pool, &enrollments, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindEnrollmentsByStudent: failed to execute query or scan: %w", err)
	}

	return enrollments, nil // Returns slice (empty if no rows) and nil error on success
}

// FindEnrollmentsByCourse retrieves all enrollment records for a specific course in a college with pagination.
func (e *enrollmentRepository) FindEnrollmentsByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Enrollment, error) {
	sql := `SELECT id, student_id, course_id, college_id, enrollment_date, status, grade, created_at, updated_at FROM enrollments WHERE college_id = $1 AND course_id = $2 ORDER BY student_id ASC, enrollment_date DESC LIMIT $3 OFFSET $4`
	args := []any{collegeID, courseID, limit, offset}

	enrollments := []*models.Enrollment{}
	err := pgxscan.Select(ctx, e.DB.Pool, &enrollments, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindEnrollmentsByCourse: failed to execute query or scan: %w", err)
	}

	return enrollments, nil
}

// FindEnrollmentsByCollege retrieves all enrollment records for a specific college with pagination.
func (e *enrollmentRepository) FindEnrollmentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Enrollment, error) {
	sql := `SELECT id, student_id, course_id, college_id, enrollment_date, status, grade, created_at, updated_at FROM enrollments WHERE college_id = $1 ORDER BY course_id ASC, student_id ASC, enrollment_date DESC LIMIT $2 OFFSET $3`
	args := []any{collegeID, limit, offset}

	enrollments := []*models.Enrollment{}
	err := pgxscan.Select(ctx, e.DB.Pool, &enrollments, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindEnrollmentsByCollege: failed to execute query or scan: %w", err)
	}

	return enrollments, nil
}

// CountEnrollmentsByStudent counts the total number of enrollments for a specific student in a college.
func (e *enrollmentRepository) CountEnrollmentsByStudent(ctx context.Context, collegeID int, studentID int) (int, error) {
	return e.countEnrollments(ctx, collegeID, studentID, -1)
}

// CountEnrollmentsByCourse counts the total number of enrollments for a specific course in a college.
func (e *enrollmentRepository) CountEnrollmentsByCourse(ctx context.Context, collegeID int, courseID int) (int, error) {
	return e.countEnrollments(ctx, collegeID, -1, courseID)
}

// CountEnrollmentsByCollege counts the total number of enrollments within a specific college.
func (e *enrollmentRepository) CountEnrollmentsByCollege(ctx context.Context, collegeID int) (int, error) {
	return e.countEnrollments(ctx, collegeID, -1, -1)
}

// countEnrollments is a helper function for counting based on conditions.
// Use -1 for parameters that should not be filtered on.
func (e *enrollmentRepository) countEnrollments(ctx context.Context, collegeID int, studentID int, courseID int) (int, error) {
	args := []any{}
	sql := `SELECT COUNT(*) FROM enrollments WHERE college_id = $1`
	args = append(args, collegeID)

	if studentID != -1 {
		sql += ` AND student_id = $` + fmt.Sprintf("%d", len(args)+1)
		args = append(args, studentID)
	}

	if courseID != -1 {
		sql += ` AND course_id = $` + fmt.Sprintf("%d", len(args)+1)
		args = append(args, courseID)
	}

	temp := struct {
		Count int `db:"count"`
	}{}
	err := pgxscan.Get(ctx, e.DB.Pool, &temp, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("countEnrollments: failed to execute query or scan: %w", err)
	}

	return temp.Count, nil
}

// UpdateEnrollmentStatus updates the status of a specific enrollment record by ID.
func (e *enrollmentRepository) UpdateEnrollmentStatus(ctx context.Context, collegeID int, enrollmentID int, status string) error {
	now := time.Now()
	sql := `UPDATE enrollments SET status = $1, updated_at = $2 WHERE id = $3 AND college_id = $4`
	args := []any{status, now, enrollmentID, collegeID}

	commandTag, err := e.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateEnrollmentStatus: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateEnrollmentStatus: no enrollment found with ID %d for college ID %d, or status unchanged", enrollmentID, collegeID)
	}

	return nil // Success
}

// DeleteEnrollment removes an enrollment record by its ID, scoped by collegeID.
func (e *enrollmentRepository) DeleteEnrollment(ctx context.Context, collegeID, enrollmentID int) error {
	sql := `DELETE FROM enrollments WHERE id = $1 AND college_id = $2`
	args := []any{enrollmentID, collegeID}

	commandTag, err := e.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteEnrollment: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteEnrollment: no enrollment found with ID %d for college ID %d, or already deleted", enrollmentID, collegeID)
	}
	return nil
}

// UpdateEnrollmentPartial updates enrollment fields partially based on the provided request.
func (e *enrollmentRepository) UpdateEnrollmentPartial(ctx context.Context, collegeID int, enrollmentID int, req *models.UpdateEnrollmentRequest) error {
	if enrollmentID <= 0 {
		return fmt.Errorf("UpdateEnrollmentPartial: enrollmentID must be greater than 0")
	}
	if collegeID <= 0 {
		return fmt.Errorf("UpdateEnrollmentPartial: collegeID must be greater than 0")
	}

	// Check if at least one field is provided for update
	hasUpdates := req.StudentID != nil || req.CourseID != nil || req.CollegeID != nil ||
		req.EnrollmentDate != nil || req.Status != nil || req.Grade != nil
	if !hasUpdates {
		return fmt.Errorf("UpdateEnrollmentPartial: at least one field must be provided for update")
	}

	// Build dynamic query based on non-nil fields
	sql := `UPDATE enrollments SET updated_at = NOW()`
	args := []any{}
	argIndex := 1

	if req.StudentID != nil {
		sql += fmt.Sprintf(`, student_id = $%d`, argIndex)
		args = append(args, *req.StudentID)
		argIndex++
	}
	if req.CourseID != nil {
		sql += fmt.Sprintf(`, course_id = $%d`, argIndex)
		args = append(args, *req.CourseID)
		argIndex++
	}
	if req.CollegeID != nil {
		sql += fmt.Sprintf(`, college_id = $%d`, argIndex)
		args = append(args, *req.CollegeID)
		argIndex++
	}
	if req.EnrollmentDate != nil {
		sql += fmt.Sprintf(`, enrollment_date = $%d`, argIndex)
		args = append(args, *req.EnrollmentDate)
		argIndex++
	}
	if req.Status != nil {
		sql += fmt.Sprintf(`, status = $%d`, argIndex)
		args = append(args, *req.Status)
		argIndex++
	}
	if req.Grade != nil {
		sql += fmt.Sprintf(`, grade = $%d`, argIndex)
		args = append(args, *req.Grade)
		argIndex++
	}

	sql += fmt.Sprintf(` WHERE id = $%d AND college_id = $%d`, argIndex, argIndex+1)
	args = append(args, enrollmentID, collegeID)

	commandTag, err := e.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateEnrollmentPartial: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateEnrollmentPartial: no enrollment found with ID %d for college ID %d, or no changes made", enrollmentID, collegeID)
	}

	return nil
}

// UpdateEnrollment updates mutable fields of an existing enrollment record.
func (e *enrollmentRepository) UpdateEnrollment(ctx context.Context, enrollment *models.Enrollment) error {
	enrollment.UpdatedAt = time.Now()
	sql := `UPDATE enrollments SET enrollment_date = $1, status = $2, grade = $3, updated_at = $4 WHERE id = $5 AND college_id = $6`
	args := []any{enrollment.EnrollmentDate, enrollment.Status, enrollment.Grade, enrollment.UpdatedAt, enrollment.ID, enrollment.CollegeID}

	commandTag, err := e.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateEnrollment: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateEnrollment: no enrollment found with ID %d for college ID %d, or no changes made", enrollment.ID, enrollment.CollegeID)
	}

	return nil
}
