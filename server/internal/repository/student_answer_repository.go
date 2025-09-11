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

// StudentAnswerRepository defines the interface for student answer data operations.
// It provides methods for creating, reading, updating, and querying student answer records
// with proper college-based isolation and parameterized queries for security.
type StudentAnswerRepository interface {
	// CreateStudentAnswer creates a new student answer in the database.
	// It sets the CreatedAt and UpdatedAt timestamps automatically.
	// Uses UPSERT to handle conflicts on (quiz_attempt_id, question_id).
	CreateStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error

	// GetStudentAnswerByID retrieves a student answer by its ID with college isolation.
	// Returns an error if the answer is not found or doesn't belong to the college.
	GetStudentAnswerByID(ctx context.Context, collegeID int, answerID int) (*models.StudentAnswer, error)

	// UpdateStudentAnswer updates the is_correct and points_awarded fields of an existing student answer.
	// It updates the UpdatedAt timestamp automatically.
	UpdateStudentAnswer(ctx context.Context, collegeID int, answer *models.StudentAnswer) error

	// FindStudentAnswersByAttempt retrieves student answers for a specific quiz attempt with pagination.
	// Results are ordered by question ID (ascending).
	FindStudentAnswersByAttempt(ctx context.Context, collegeID int, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error)

	// GetStudentAnswerForQuestion retrieves a student answer for a specific question in a quiz attempt.
	// Returns an error if the answer is not found or doesn't belong to the college.
	GetStudentAnswerForQuestion(ctx context.Context, collegeID int, attemptID int, questionID int) (*models.StudentAnswer, error)
}

// studentAnswerRepository implements the StudentAnswerRepository interface.
type studentAnswerRepository struct {
	DB *DB // Database connection pool
}

// NewStudentAnswerRepository creates a new instance of StudentAnswerRepository.
func NewStudentAnswerRepository(db *DB) StudentAnswerRepository {
	return &studentAnswerRepository{DB: db}
}

// Table constants for student answer operations
const (
	studentAnswerTable = "student_answers"
)

// CreateStudentAnswer creates a new student answer in the database.
// It automatically sets CreatedAt and UpdatedAt timestamps.
// Uses UPSERT (ON CONFLICT) to handle duplicate answers for the same question in the same attempt.
// Uses parameterized queries to prevent SQL injection.
func (r *studentAnswerRepository) CreateStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error {
	// Set timestamps
	now := time.Now()
	answer.CreatedAt = now
	answer.UpdatedAt = now

	// SQL query with UPSERT to handle conflicts
	sql := `INSERT INTO student_answers (quiz_attempt_id, question_id, selected_option_id, answer_text, is_correct, points_awarded, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (quiz_attempt_id, question_id)
			DO UPDATE SET selected_option_id = EXCLUDED.selected_option_id,
						 answer_text = EXCLUDED.answer_text,
						 is_correct = EXCLUDED.is_correct,
						 points_awarded = EXCLUDED.points_awarded,
						 updated_at = EXCLUDED.updated_at
			RETURNING id`

	// Prepare arguments in correct order
	args := []any{answer.QuizAttemptID, answer.QuestionID, answer.SelectedOptionID,
				 answer.AnswerText, answer.IsCorrect, answer.PointsAwarded,
				 answer.CreatedAt, answer.UpdatedAt}

	// Execute query and scan the returned ID
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return fmt.Errorf("CreateStudentAnswer: failed to execute query: %w", err)
	}

	// Set the generated ID on the answer object
	answer.ID = temp.ID
	return nil
}

// GetStudentAnswerByID retrieves a student answer by its ID with college isolation.
// Uses JOIN with quiz_attempts table to ensure the answer belongs to the college.
func (r *studentAnswerRepository) GetStudentAnswerByID(ctx context.Context, collegeID int, answerID int) (*models.StudentAnswer, error) {
	answer := &models.StudentAnswer{}

	// Query with college isolation through JOIN
	sql := `SELECT sa.id, sa.quiz_attempt_id, sa.question_id, sa.selected_option_id, sa.answer_text, sa.is_correct, sa.points_awarded, sa.created_at, sa.updated_at
			FROM student_answers sa
			JOIN quiz_attempts qa ON sa.quiz_attempt_id = qa.id
			WHERE sa.id = $1 AND qa.college_id = $2`
	args := []any{answerID, collegeID}

	err := pgxscan.Get(ctx, r.DB.Pool, answer, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetStudentAnswerByID: student answer not found (id: %d for college %d)", answerID, collegeID)
		}
		return nil, fmt.Errorf("GetStudentAnswerByID: failed to execute query: %w", err)
	}

	return answer, nil
}

// UpdateStudentAnswer updates the is_correct and points_awarded fields of an existing student answer.
// Updates the UpdatedAt timestamp automatically.
// Uses subquery to ensure college isolation.
func (r *studentAnswerRepository) UpdateStudentAnswer(ctx context.Context, collegeID int, answer *models.StudentAnswer) error {
	// Update timestamp
	answer.UpdatedAt = time.Now()

	// Update query with college isolation through subquery
	sql := `UPDATE student_answers SET is_correct = $1, points_awarded = $2, updated_at = $3
			WHERE id = $4 AND quiz_attempt_id IN (SELECT id FROM quiz_attempts WHERE college_id = $5)`
	args := []any{answer.IsCorrect, answer.PointsAwarded, answer.UpdatedAt, answer.ID, collegeID}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateStudentAnswer: failed to execute query for college %d: %w", collegeID, err)
	}

	// Check if any rows were affected
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateStudentAnswer: student answer not found or no changes (id: %d for college %d)", answer.ID, collegeID)
	}

	return nil
}

// FindStudentAnswersByAttempt retrieves student answers for a specific quiz attempt with pagination.
// Results are ordered by question ID (ascending).
// Uses JOIN with quiz_attempts table to ensure college isolation.
func (r *studentAnswerRepository) FindStudentAnswersByAttempt(ctx context.Context, collegeID int, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error) {
	answers := []*models.StudentAnswer{}

	sql := `SELECT sa.id, sa.quiz_attempt_id, sa.question_id, sa.selected_option_id, sa.answer_text, sa.is_correct, sa.points_awarded, sa.created_at, sa.updated_at
			FROM student_answers sa
			JOIN quiz_attempts qa ON sa.quiz_attempt_id = qa.id
			WHERE sa.quiz_attempt_id = $1 AND qa.college_id = $2
			ORDER BY sa.question_id ASC
			LIMIT $3 OFFSET $4`
	args := []any{attemptID, collegeID, limit, offset}

	err := pgxscan.Select(ctx, r.DB.Pool, &answers, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindStudentAnswersByAttempt: failed to execute query: %w", err)
	}

	return answers, nil
}

// GetStudentAnswerForQuestion retrieves a student answer for a specific question in a quiz attempt.
// Returns an error if the answer is not found or doesn't belong to the college.
// Uses JOIN with quiz_attempts table to ensure college isolation.
func (r *studentAnswerRepository) GetStudentAnswerForQuestion(ctx context.Context, collegeID int, attemptID int, questionID int) (*models.StudentAnswer, error) {
	answer := &models.StudentAnswer{}

	sql := `SELECT sa.id, sa.quiz_attempt_id, sa.question_id, sa.selected_option_id, sa.answer_text, sa.is_correct, sa.points_awarded, sa.created_at, sa.updated_at
			FROM student_answers sa
			JOIN quiz_attempts qa ON sa.quiz_attempt_id = qa.id
			WHERE sa.quiz_attempt_id = $1 AND sa.question_id = $2 AND qa.college_id = $3`
	args := []any{attemptID, questionID, collegeID}

	err := pgxscan.Get(ctx, r.DB.Pool, answer, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetStudentAnswerForQuestion: student answer not found (attempt: %d, question: %d for college %d)", attemptID, questionID, collegeID)
		}
		return nil, fmt.Errorf("GetStudentAnswerForQuestion: failed to execute query: %w", err)
	}

	return answer, nil
}