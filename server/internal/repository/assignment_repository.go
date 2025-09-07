package repository

import (
	"context"
	"eduhub/server/internal/models"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

// AssignmentRepository defines a clean interface for assignment and submission database operations.
type AssignmentRepository interface {
	// Assignment methods
	CreateAssignment(ctx context.Context, assignment *models.Assignment) error
	GetAssignmentByID(ctx context.Context, collegeID int, assignmentID int) (*models.Assignment, error)
	UpdateAssignment(ctx context.Context, assignment *models.Assignment) error
	DeleteAssignment(ctx context.Context, collegeID int, assignmentID int) error
	FindAssignmentsByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Assignment, error)
	CountAssignmentsByCourse(ctx context.Context, collegeID int, courseID int) (int, error)

	// Submission methods
	CreateSubmission(ctx context.Context, submission *models.AssignmentSubmission) error
	GetSubmissionByID(ctx context.Context, submissionID int) (*models.AssignmentSubmission, error)
	GetSubmissionByStudentAndAssignment(ctx context.Context, studentID int, assignmentID int) (*models.AssignmentSubmission, error)
	UpdateSubmission(ctx context.Context, submission *models.AssignmentSubmission) error // For grading/feedback
	FindSubmissionsByAssignment(ctx context.Context, assignmentID int, limit, offset uint64) ([]*models.AssignmentSubmission, error)
	FindSubmissionsByStudent(ctx context.Context, studentID int, limit, offset uint64) ([]*models.AssignmentSubmission, error)
}

type assignmentRepository struct {
	DB *DB
}

func NewAssignmentRepository(db *DB) AssignmentRepository {
	return &assignmentRepository{DB: db}
}

// --- Assignment Methods ---

func (r *assignmentRepository) CreateAssignment(ctx context.Context, assignment *models.Assignment) error {
	now := time.Now()
	assignment.CreatedAt = now
	assignment.UpdatedAt = now

	sql := `INSERT INTO assignments (course_id, college_id, title, description, due_date, max_points, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING id`

	temp := struct {
		ID int `db:"id"`
	}{}

	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql,
		assignment.CourseID, assignment.CollegeID, assignment.Title, assignment.Description,
		assignment.DueDate, assignment.MaxPoints, assignment.CreatedAt, assignment.UpdatedAt)

	if err != nil {
		return fmt.Errorf("CreateAssignment: failed to execute query or scan ID: %w", err)
	}
	assignment.ID = temp.ID
	return nil
}

func (r *assignmentRepository) GetAssignmentByID(ctx context.Context, collegeID int, assignmentID int) (*models.Assignment, error) {
	assignment := &models.Assignment{}
	sql := `SELECT id, course_id, college_id, title, description, due_date, max_points, created_at, updated_at
			 FROM assignments
			 WHERE id = $1 AND college_id = $2`

	err := pgxscan.Get(ctx, r.DB.Pool, assignment, sql, assignmentID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetAssignmentByID: assignment with ID %d not found for college ID %d", assignmentID, collegeID)
		}
		return nil, fmt.Errorf("GetAssignmentByID: failed to execute query or scan: %w", err)
	}
	return assignment, nil
}

func (r *assignmentRepository) UpdateAssignment(ctx context.Context, assignment *models.Assignment) error {
	assignment.UpdatedAt = time.Now()

	sql := `UPDATE assignments
			 SET title = $1, description = $2, due_date = $3, max_points = $4, updated_at = $5
			 WHERE id = $6 AND college_id = $7`

	cmdTag, err := r.DB.Pool.Exec(ctx, sql,
		assignment.Title, assignment.Description, assignment.DueDate,
		assignment.MaxPoints, assignment.UpdatedAt, assignment.ID, assignment.CollegeID)

	if err != nil {
		return fmt.Errorf("UpdateAssignment: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateAssignment: no assignment found with ID %d for college ID %d, or no changes made", assignment.ID, assignment.CollegeID)
	}
	return nil
}

func (r *assignmentRepository) DeleteAssignment(ctx context.Context, collegeID int, assignmentID int) error {
	sql := `DELETE FROM assignments WHERE id = $1 AND college_id = $2`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, assignmentID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteAssignment: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteAssignment: no assignment found with ID %d for college ID %d, or already deleted", assignmentID, collegeID)
	}
	return nil
}

func (r *assignmentRepository) FindAssignmentsByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Assignment, error) {
	assignments := []*models.Assignment{}
	sql := `SELECT id, course_id, college_id, title, description, due_date, max_points, created_at, updated_at
			 FROM assignments
			 WHERE college_id = $1 AND course_id = $2
			 ORDER BY due_date ASC, created_at ASC
			 LIMIT $3 OFFSET $4`

	err := pgxscan.Select(ctx, r.DB.Pool, &assignments, sql, collegeID, courseID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("FindAssignmentsByCourse: failed to execute query or scan: %w", err)
	}
	return assignments, nil
}

func (r *assignmentRepository) CountAssignmentsByCourse(ctx context.Context, collegeID int, courseID int) (int, error) {
	sql := `SELECT COUNT(*) FROM assignments WHERE college_id = $1 AND course_id = $2`
	var count int
	err := r.DB.Pool.QueryRow(ctx, sql, collegeID, courseID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountAssignmentsByCourse: failed to execute query or scan: %w", err)
	}
	return count, nil
}

// --- Submission Methods ---

func (r *assignmentRepository) CreateSubmission(ctx context.Context, submission *models.AssignmentSubmission) error {
	now := time.Now()
	submission.CreatedAt = now
	submission.UpdatedAt = now
	if submission.SubmissionTime.IsZero() {
		submission.SubmissionTime = now
	}

	// ON CONFLICT allows a student to re-submit, updating their submission.
	sql := `INSERT INTO assignment_submissions (assignment_id, student_id, submission_time, content_text, file_path, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (assignment_id, student_id)
			 DO UPDATE SET
				 submission_time = EXCLUDED.submission_time,
				 content_text = EXCLUDED.content_text,
				 file_path = EXCLUDED.file_path,
				 updated_at = EXCLUDED.updated_at
			 RETURNING id`

	temp := struct {
		ID int `db:"id"`
	}{}

	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql,
		submission.AssignmentID, submission.StudentID, submission.SubmissionTime,
		submission.ContentText, submission.FilePath, submission.CreatedAt, submission.UpdatedAt)

	if err != nil {
		return fmt.Errorf("CreateSubmission: failed to execute query or scan ID: %w", err)
	}
	submission.ID = temp.ID
	return nil
}

func (r *assignmentRepository) GetSubmissionByID(ctx context.Context, submissionID int) (*models.AssignmentSubmission, error) {
	submission := &models.AssignmentSubmission{}
	sql := `SELECT id, assignment_id, student_id, submission_time, content_text, file_path, grade, feedback, created_at, updated_at
			 FROM assignment_submissions
			 WHERE id = $1`

	err := pgxscan.Get(ctx, r.DB.Pool, submission, sql, submissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetSubmissionByID: submission with ID %d not found", submissionID)
		}
		return nil, fmt.Errorf("GetSubmissionByID: failed to execute query or scan: %w", err)
	}
	return submission, nil
}

func (r *assignmentRepository) GetSubmissionByStudentAndAssignment(ctx context.Context, studentID int, assignmentID int) (*models.AssignmentSubmission, error) {
	submission := &models.AssignmentSubmission{}
	sql := `SELECT id, assignment_id, student_id, submission_time, content_text, file_path, grade, feedback, created_at, updated_at
			 FROM assignment_submissions
			 WHERE student_id = $1 AND assignment_id = $2`

	err := pgxscan.Get(ctx, r.DB.Pool, submission, sql, studentID, assignmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil if not found, as it's a common check before creating.
		}
		return nil, fmt.Errorf("GetSubmissionByStudentAndAssignment: failed to execute query or scan: %w", err)
	}
	return submission, nil
}

func (r *assignmentRepository) UpdateSubmission(ctx context.Context, submission *models.AssignmentSubmission) error {
	submission.UpdatedAt = time.Now()

	// This update is primarily for grading and feedback.
	sql := `UPDATE assignment_submissions
			 SET grade = $1, feedback = $2, updated_at = $3
			 WHERE id = $4`

	cmdTag, err := r.DB.Pool.Exec(ctx, sql,
		submission.Grade, submission.Feedback, submission.UpdatedAt, submission.ID)

	if err != nil {
		return fmt.Errorf("UpdateSubmission: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateSubmission: no submission found with ID %d, or no changes made", submission.ID)
	}
	return nil
}

func (r *assignmentRepository) FindSubmissionsByAssignment(ctx context.Context, assignmentID int, limit, offset uint64) ([]*models.AssignmentSubmission, error) {
	submissions := []*models.AssignmentSubmission{}
	sql := `SELECT id, assignment_id, student_id, submission_time, content_text, file_path, grade, feedback, created_at, updated_at
			 FROM assignment_submissions
			 WHERE assignment_id = $1
			 ORDER BY submission_time DESC
			 LIMIT $2 OFFSET $3`

	err := pgxscan.Select(ctx, r.DB.Pool, &submissions, sql, assignmentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("FindSubmissionsByAssignment: failed to execute query or scan: %w", err)
	}
	return submissions, nil
}

func (r *assignmentRepository) FindSubmissionsByStudent(ctx context.Context, studentID int, limit, offset uint64) ([]*models.AssignmentSubmission, error) {
	submissions := []*models.AssignmentSubmission{}
	sql := `SELECT id, assignment_id, student_id, submission_time, content_text, file_path, grade, feedback, created_at, updated_at
			 FROM assignment_submissions
			 WHERE student_id = $1
			 ORDER BY submission_time DESC
			 LIMIT $2 OFFSET $3`

	err := pgxscan.Select(ctx, r.DB.Pool, &submissions, sql, studentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("FindSubmissionsByStudent: failed to execute query or scan: %w", err)
	}
	return submissions, nil
}
