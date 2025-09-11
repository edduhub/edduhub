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

// QuestionRepository defines the interface for question data operations.
// It provides methods for creating, reading, updating, and deleting question records
// with proper college-based isolation and parameterized queries for security.
type QuestionRepository interface {
	// CreateQuestion creates a new question in the database.
	// It sets the CreatedAt and UpdatedAt timestamps automatically.
	CreateQuestion(ctx context.Context, question *models.Question) error

	// GetQuestionByID retrieves a question by its ID with college isolation.
	// Returns an error if the question is not found or doesn't belong to the college.
	GetQuestionByID(ctx context.Context, collegeID int, questionID int) (*models.Question, error)

	// UpdateQuestion updates all fields of an existing question.
	// It updates the UpdatedAt timestamp automatically.
	UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error

	// DeleteQuestion removes a question from the database.
	// Ensures the question belongs to the specified college for isolation.
	DeleteQuestion(ctx context.Context, collegeID int, questionID int) error

	// FindQuestionsByQuiz retrieves questions for a specific quiz with pagination.
	// Results are ordered by creation date (ascending).
	FindQuestionsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.Question, error)

	// CountQuestionsByQuiz returns the total number of questions for a quiz.
	CountQuestionsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)
}

// questionRepository implements the QuestionRepository interface.
type questionRepository struct {
	DB *DB // Database connection pool
}

// NewQuestionRepository creates a new instance of QuestionRepository.
func NewQuestionRepository(db *DB) QuestionRepository {
	return &questionRepository{DB: db}
}

// Table constants for question operations
const (
	questionTable = "questions"
)

// CreateQuestion creates a new question in the database.
// It automatically sets CreatedAt and UpdatedAt timestamps.
// Uses parameterized queries to prevent SQL injection.
func (r *questionRepository) CreateQuestion(ctx context.Context, question *models.Question) error {
	// Set timestamps
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	// SQL query with parameterized placeholders
	sql := `INSERT INTO questions (quiz_id, text, type, points, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	// Prepare arguments in correct order
	args := []any{question.QuizID, question.Text, question.Type, question.Points,
				 question.CreatedAt, question.UpdatedAt}

	// Execute query and scan the returned ID
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return fmt.Errorf("CreateQuestion: failed to execute query: %w", err)
	}

	// Set the generated ID on the question object
	question.ID = temp.ID
	return nil
}

// GetQuestionByID retrieves a question by its ID with college isolation.
// Uses a JOIN with quizzes table to ensure the question belongs to the college.
func (r *questionRepository) GetQuestionByID(ctx context.Context, collegeID int, questionID int) (*models.Question, error) {
	question := &models.Question{}

	// Query with college isolation through JOIN
	sql := `SELECT q.id, q.quiz_id, q.text, q.type, q.points, q.created_at, q.updated_at
			FROM questions q
			JOIN quizzes qu ON q.quiz_id = qu.id
			WHERE q.id = $1 AND qu.college_id = $2`
	args := []any{questionID, collegeID}

	err := pgxscan.Get(ctx, r.DB.Pool, question, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetQuestionByID: question not found (id: %d for college %d)", questionID, collegeID)
		}
		return nil, fmt.Errorf("GetQuestionByID: failed to execute query: %w", err)
	}

	return question, nil
}

// UpdateQuestion updates all fields of an existing question.
// Updates the UpdatedAt timestamp automatically.
// Uses subquery to ensure college isolation.
func (r *questionRepository) UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	// Update timestamp
	question.UpdatedAt = time.Now()

	// Update query with college isolation through subquery
	sql := `UPDATE questions SET text = $1, type = $2, points = $3, updated_at = $4
			WHERE id = $5 AND quiz_id IN (SELECT id FROM quizzes WHERE college_id = $6)`
	args := []any{question.Text, question.Type, question.Points, question.UpdatedAt,
				 question.ID, collegeID}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateQuestion: failed to execute query: %w", err)
	}

	// Check if any rows were affected
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateQuestion: question not found or no changes (id: %d for college %d)", question.ID, collegeID)
	}

	return nil
}

// DeleteQuestion removes a question from the database.
// Ensures college isolation by checking through quizzes table.
func (r *questionRepository) DeleteQuestion(ctx context.Context, collegeID int, questionID int) error {
	sql := `DELETE FROM questions WHERE id = $1 AND quiz_id IN (SELECT id FROM quizzes WHERE college_id = $2)`
	args := []any{questionID, collegeID}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteQuestion: failed to execute query for college %d: %w", collegeID, err)
	}

	// Check if any rows were affected
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteQuestion: question not found (id: %d for college %d)", questionID, collegeID)
	}

	return nil
}

// FindQuestionsByQuiz retrieves questions for a specific quiz with pagination.
// Results are ordered by creation date (ascending).
// Uses JOIN to ensure college isolation.
func (r *questionRepository) FindQuestionsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.Question, error) {
	questions := []*models.Question{}

	sql := `SELECT q.id, q.quiz_id, q.text, q.type, q.points, q.created_at, q.updated_at
			FROM questions q
			JOIN quizzes qu ON q.quiz_id = qu.id
			WHERE q.quiz_id = $1 AND qu.college_id = $2
			ORDER BY q.created_at ASC
			LIMIT $3 OFFSET $4`
	args := []any{quizID, collegeID, limit, offset}

	err := pgxscan.Select(ctx, r.DB.Pool, &questions, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindQuestionsByQuiz: failed to execute query: %w", err)
	}

	return questions, nil
}

// CountQuestionsByQuiz returns the total count of questions for a specific quiz.
// Used for pagination calculations.
// Uses JOIN to ensure college isolation.
func (r *questionRepository) CountQuestionsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
	sql := `SELECT COUNT(q.id) FROM questions q
			JOIN quizzes qu ON q.quiz_id = qu.id
			WHERE q.quiz_id = $1 AND qu.college_id = $2`
	args := []any{quizID, collegeID}

	temp := struct {
		Count int `db:"count"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("CountQuestionsByQuiz: failed to execute query: %w", err)
	}

	return temp.Count, nil
}