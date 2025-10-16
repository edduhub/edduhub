package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

// QuizAttemptRepository defines the interface for quiz attempt data operations.
// It provides methods for creating, reading, updating, and querying quiz attempt records
// with proper college-based isolation and parameterized queries for security.
type QuizAttemptRepository interface {
	// CreateQuizAttempt creates a new quiz attempt in the database.
	// It sets the CreatedAt and UpdatedAt timestamps automatically.
	// Sets default values for StartTime and Status if not provided.
	CreateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error

	// GetQuizAttemptByID retrieves a quiz attempt by its ID with college isolation.
	// Returns an error if the attempt is not found or doesn't belong to the college.
	GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)

	// UpdateQuizAttempt updates the end_time, score, and status of an existing quiz attempt.
	// It updates the UpdatedAt timestamp automatically.
	UpdateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error

	// FindQuizAttemptsByStudent retrieves quiz attempts for a specific student with pagination.
	// Results are ordered by start time (descending).
	FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error)

	// FindQuizAttemptsByQuiz retrieves quiz attempts for a specific quiz with pagination.
	// Results are ordered by student ID then start time (descending).
	FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error)

	// CountQuizAttemptsByQuiz returns the total number of attempts for a quiz.
	// Used for pagination calculations.
	CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)
}

// quizAttemptRepository implements the QuizAttemptRepository interface.
type quizAttemptRepository struct {
	DB *DB // Database connection pool
}

// NewQuizAttemptRepository creates a new instance of QuizAttemptRepository.
func NewQuizAttemptRepository(db *DB) QuizAttemptRepository {
	return &quizAttemptRepository{DB: db}
}

// Table constants for quiz attempt operations
const (
	quizAttemptTable = "quiz_attempts"
)

// CreateQuizAttempt creates a new quiz attempt in the database.
// It automatically sets CreatedAt and UpdatedAt timestamps.
// Sets default values for StartTime (current time) and Status ("InProgress") if not provided.
// Uses parameterized queries to prevent SQL injection.
func (r *quizAttemptRepository) CreateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error {
	// Set timestamps
	now := time.Now()
	attempt.CreatedAt = now
	attempt.UpdatedAt = now

	// Set default values if not provided
	if attempt.StartTime.IsZero() {
		attempt.StartTime = now
	}
	if attempt.Status == "" {
		attempt.Status = "InProgress"
	}

	// SQL query with parameterized placeholders
	sql := `INSERT INTO quiz_attempts (student_id, quiz_id, college_id, start_time, end_time, score, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	// Prepare arguments in correct order
	args := []any{attempt.StudentID, attempt.QuizID, attempt.CollegeID, attempt.StartTime,
				 attempt.EndTime, attempt.Score, attempt.Status, attempt.CreatedAt, attempt.UpdatedAt}

	// Execute query and scan the returned ID
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return fmt.Errorf("CreateQuizAttempt: failed to execute query: %w", err)
	}

	// Set the generated ID on the attempt object
	attempt.ID = temp.ID
	return nil
}

// GetQuizAttemptByID retrieves a quiz attempt by its ID with college isolation.
// Ensures the attempt belongs to the specified college.
func (r *quizAttemptRepository) GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error) {
	attempt := &models.QuizAttempt{}

	// Query with college isolation
	sql := `SELECT id, student_id, quiz_id, college_id, start_time, end_time, score, status, created_at, updated_at
			FROM quiz_attempts WHERE id = $1 AND college_id = $2`
	args := []any{attemptID, collegeID}

	err := pgxscan.Get(ctx, r.DB.Pool, attempt, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetQuizAttemptByID: quiz attempt not found (id: %d, college: %d)", attemptID, collegeID)
		}
		return nil, fmt.Errorf("GetQuizAttemptByID: failed to execute query: %w", err)
	}

	return attempt, nil
}

// UpdateQuizAttempt updates the end_time, score, and status of an existing quiz attempt.
// Updates the UpdatedAt timestamp automatically.
// Only updates specific fields to maintain data integrity.
func (r *quizAttemptRepository) UpdateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error {
	// Update timestamp
	attempt.UpdatedAt = time.Now()

	// Update query with college isolation
	sql := `UPDATE quiz_attempts SET end_time = $1, score = $2, status = $3, updated_at = $4
			WHERE id = $5 AND college_id = $6`
	args := []any{attempt.EndTime, attempt.Score, attempt.Status, attempt.UpdatedAt,
				 attempt.ID, attempt.CollegeID}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateQuizAttempt: failed to execute query: %w", err)
	}

	// Check if any rows were affected
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateQuizAttempt: quiz attempt not found or no changes (id: %d)", attempt.ID)
	}

	return nil
}

// FindQuizAttemptsByStudent retrieves quiz attempts for a specific student with pagination.
// Results are ordered by start time (descending).
// Ensures college isolation.
func (r *quizAttemptRepository) FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
	attempts := []*models.QuizAttempt{}

	sql := `SELECT id, student_id, quiz_id, college_id, start_time, end_time, score, status, created_at, updated_at
			FROM quiz_attempts
			WHERE college_id = $1 AND student_id = $2
			ORDER BY start_time DESC
			LIMIT $3 OFFSET $4`
	args := []any{collegeID, studentID, limit, offset}

	err := pgxscan.Select(ctx, r.DB.Pool, &attempts, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindQuizAttemptsByStudent: failed to execute query: %w", err)
	}

	return attempts, nil
}

// FindQuizAttemptsByQuiz retrieves quiz attempts for a specific quiz with pagination.
// Results are ordered by student ID (ascending) then start time (descending).
// Ensures college isolation.
func (r *quizAttemptRepository) FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
	attempts := []*models.QuizAttempt{}

	sql := `SELECT id, student_id, quiz_id, college_id, start_time, end_time, score, status, created_at, updated_at
			FROM quiz_attempts
			WHERE college_id = $1 AND quiz_id = $2
			ORDER BY student_id ASC, start_time DESC
			LIMIT $3 OFFSET $4`
	args := []any{collegeID, quizID, limit, offset}

	err := pgxscan.Select(ctx, r.DB.Pool, &attempts, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindQuizAttemptsByQuiz: failed to execute query: %w", err)
	}

	return attempts, nil
}

// CountQuizAttemptsByQuiz returns the total count of quiz attempts for a quiz.
// Ensures college isolation.
func (r *quizAttemptRepository) CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
	var count int

	// Query to count quiz attempts for a specific quiz within a college
	sql := `SELECT COUNT(*) FROM quiz_attempts WHERE college_id = $1 AND quiz_id = $2`
	args := []any{collegeID, quizID}

	err := pgxscan.Get(ctx, r.DB.Pool, &count, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("CountQuizAttemptsByQuiz: failed to count quiz attempts: %w", err)
	}

	return count, nil
}
