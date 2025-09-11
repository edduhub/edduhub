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

// AnswerOptionRepository defines the interface for answer option data operations.
// It provides methods for creating, reading, updating, and deleting answer option records
// with proper college-based isolation and parameterized queries for security.
type AnswerOptionRepository interface {
	// CreateAnswerOption creates a new answer option in the database.
	// It sets the CreatedAt and UpdatedAt timestamps automatically.
	CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error

	// GetAnswerOptionByID retrieves an answer option by its ID with college isolation.
	// Returns an error if the option is not found or doesn't belong to the college.
	GetAnswerOptionByID(ctx context.Context, collegeID int, optionID int) (*models.AnswerOption, error)

	// UpdateAnswerOption updates all fields of an existing answer option.
	// It updates the UpdatedAt timestamp automatically.
	UpdateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error

	// DeleteAnswerOption removes an answer option from the database.
	// Ensures the option belongs to the specified college for isolation.
	DeleteAnswerOption(ctx context.Context, collegeID int, optionID int) error

	// FindAnswerOptionsByQuestion retrieves all answer options for a specific question.
	// Results are ordered by creation date (ascending).
	FindAnswerOptionsByQuestion(ctx context.Context, questionID int) ([]*models.AnswerOption, error)
}

// answerOptionRepository implements the AnswerOptionRepository interface.
type answerOptionRepository struct {
	DB *DB // Database connection pool
}

// NewAnswerOptionRepository creates a new instance of AnswerOptionRepository.
func NewAnswerOptionRepository(db *DB) AnswerOptionRepository {
	return &answerOptionRepository{DB: db}
}

// Table constants for answer option operations
const (
	answerOptionTable = "answer_options"
)

// CreateAnswerOption creates a new answer option in the database.
// It automatically sets CreatedAt and UpdatedAt timestamps.
// Uses parameterized queries to prevent SQL injection.
func (r *answerOptionRepository) CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	// Set timestamps
	now := time.Now()
	option.CreatedAt = now
	option.UpdatedAt = now

	// SQL query with parameterized placeholders
	sql := `INSERT INTO answer_options (question_id, text, is_correct, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`

	// Prepare arguments in correct order
	args := []any{option.QuestionID, option.Text, option.IsCorrect,
				 option.CreatedAt, option.UpdatedAt}

	// Execute query and scan the returned ID
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return fmt.Errorf("CreateAnswerOption: failed to execute query: %w", err)
	}

	// Set the generated ID on the option object
	option.ID = temp.ID
	return nil
}

// GetAnswerOptionByID retrieves an answer option by its ID with college isolation.
// Uses JOINs with questions and quizzes tables to ensure the option belongs to the college.
func (r *answerOptionRepository) GetAnswerOptionByID(ctx context.Context, collegeID int, optionID int) (*models.AnswerOption, error) {
	option := &models.AnswerOption{}

	// Query with college isolation through multiple JOINs
	sql := `SELECT ao.id, ao.question_id, ao.text, ao.is_correct, ao.created_at, ao.updated_at
			FROM answer_options ao
			JOIN questions q ON ao.question_id = q.id
			JOIN quizzes qu ON q.quiz_id = qu.id
			WHERE ao.id = $1 AND qu.college_id = $2`
	args := []any{optionID, collegeID}

	err := pgxscan.Get(ctx, r.DB.Pool, option, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetAnswerOptionByID: answer option not found (id: %d for college %d)", optionID, collegeID)
		}
		return nil, fmt.Errorf("GetAnswerOptionByID: failed to execute query: %w", err)
	}

	return option, nil
}

// UpdateAnswerOption updates all fields of an existing answer option.
// Updates the UpdatedAt timestamp automatically.
// Uses subquery to ensure college isolation through the question-quiz relationship.
func (r *answerOptionRepository) UpdateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error {
	// Update timestamp
	option.UpdatedAt = time.Now()

	// Update query with college isolation through subquery
	sql := `UPDATE answer_options SET text = $1, is_correct = $2, updated_at = $3
			WHERE id = $4 AND question_id IN (
				SELECT q.id FROM questions q
				JOIN quizzes qu ON q.quiz_id = qu.id
				WHERE qu.college_id = $5
			)`
	args := []any{option.Text, option.IsCorrect, option.UpdatedAt, option.ID, collegeID}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateAnswerOption: failed to execute query: %w", err)
	}

	// Check if any rows were affected
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateAnswerOption: answer option not found or no changes (id: %d for college %d)", option.ID, collegeID)
	}

	return nil
}

// DeleteAnswerOption removes an answer option from the database.
// Ensures college isolation by checking through questions and quizzes tables.
func (r *answerOptionRepository) DeleteAnswerOption(ctx context.Context, collegeID int, optionID int) error {
	sql := `DELETE FROM answer_options WHERE id = $1 AND question_id IN (
		SELECT q.id FROM questions q
		JOIN quizzes qu ON q.quiz_id = qu.id
		WHERE qu.college_id = $2
	)`
	args := []any{optionID, collegeID}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteAnswerOption: failed to execute query for college %d: %w", collegeID, err)
	}

	// Check if any rows were affected
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteAnswerOption: answer option not found (id: %d for college %d)", optionID, collegeID)
	}

	return nil
}

// FindAnswerOptionsByQuestion retrieves all answer options for a specific question.
// This method doesn't require college isolation as it's called from contexts
// where the question has already been validated.
// Results are ordered by creation date (ascending).
func (r *answerOptionRepository) FindAnswerOptionsByQuestion(ctx context.Context, questionID int) ([]*models.AnswerOption, error) {
	options := []*models.AnswerOption{}

	sql := `SELECT id, question_id, text, is_correct, created_at, updated_at
			FROM answer_options
			WHERE question_id = $1
			ORDER BY created_at ASC`
	args := []any{questionID}

	err := pgxscan.Select(ctx, r.DB.Pool, &options, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindAnswerOptionsByQuestion: failed to execute query: %w", err)
	}

	return options, nil
}