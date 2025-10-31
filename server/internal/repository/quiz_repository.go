package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

// QuizRepository defines the interface for quiz data operations.
// It provides methods for creating, reading, updating, and deleting quiz records
// with proper college-based isolation and parameterized queries for security.
type QuizRepository interface {
	// CreateQuiz creates a new quiz in the database.
	// It sets the CreatedAt and UpdatedAt timestamps automatically.
	CreateQuiz(ctx context.Context, quiz *models.Quiz) error

	// GetQuizByID retrieves a quiz by its ID and college ID for isolation.
	// Returns an error if the quiz is not found or doesn't belong to the college.
	GetQuizByID(ctx context.Context, collegeID int, quizID int) (*models.Quiz, error)

	// UpdateQuiz updates all fields of an existing quiz.
	// It updates the UpdatedAt timestamp automatically.
	UpdateQuiz(ctx context.Context, quiz *models.Quiz) error

	// UpdateQuizPartial performs a partial update of quiz fields.
	// Only non-nil fields in the request will be updated.
	UpdateQuizPartial(ctx context.Context, collegeID int, quizID int, req *models.UpdateQuizRequest) error

	// DeleteQuiz removes a quiz from the database.
	// Ensures the quiz belongs to the specified college for isolation.
	DeleteQuiz(ctx context.Context, collegeID int, quizID int) error

	// FindQuizzesByCourse retrieves quizzes for a specific course with pagination.
	// Results are ordered by due date and creation date.
	FindQuizzesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Quiz, error)

	// CountQuizzesByCourse returns the total number of quizzes for a course.
	CountQuizzesByCourse(ctx context.Context, collegeID int, courseID int) (int, error)
}

// quizRepository implements the QuizRepository interface.
type quizRepository struct {
	DB *DB // Database connection pool
}

// NewQuizRepository creates a new instance of QuizRepository.
func NewQuizRepository(db *DB) QuizRepository {
	return &quizRepository{DB: db}
}

// Table constants for quiz operations
const (
	quizTable = "quizzes"
)

// CreateQuiz creates a new quiz in the database.
// It automatically sets CreatedAt and UpdatedAt timestamps.
// Uses parameterized queries to prevent SQL injection.
func (r *quizRepository) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	// Check if database connection is available
	if r.DB == nil || r.DB.Pool == nil {
		return fmt.Errorf("database connection is required")
	}
	
	// Set timestamps
	now := time.Now()
	quiz.CreatedAt = now
	quiz.UpdatedAt = now

	// SQL query with parameterized placeholders
	sql := `INSERT INTO quizzes (college_id, course_id, title, description, time_limit_minutes, due_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	// Prepare arguments in correct order
	args := []any{quiz.CollegeID, quiz.CourseID, quiz.Title, quiz.Description,
				 quiz.TimeLimitMinutes, quiz.DueDate, quiz.CreatedAt, quiz.UpdatedAt}

	// Execute query and scan the returned ID
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return fmt.Errorf("CreateQuiz: failed to execute query: %w", err)
	}

	// Set the generated ID on the quiz object
	quiz.ID = temp.ID
	return nil
}

// GetQuizByID retrieves a quiz by its ID with college isolation.
// Returns an error if the quiz doesn't exist or doesn't belong to the college.
func (r *quizRepository) GetQuizByID(ctx context.Context, collegeID int, quizID int) (*models.Quiz, error) {
	// Check if database connection is available
	if r.DB == nil || r.DB.Pool == nil {
		return nil, fmt.Errorf("database connection is required")
	}
	
	quiz := &models.Quiz{}

	// Query with college isolation
	sql := `SELECT id, college_id, course_id, title, description, time_limit_minutes, due_date, created_at, updated_at
			FROM quizzes WHERE id = $1 AND college_id = $2`
	args := []any{quizID, collegeID}

	err := pgxscan.Get(ctx, r.DB.Pool, quiz, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetQuizByID: quiz not found (id: %d, college: %d)", quizID, collegeID)
		}
		return nil, fmt.Errorf("GetQuizByID: failed to execute query: %w", err)
	}

	return quiz, nil
}

// UpdateQuiz updates all fields of an existing quiz.
// Updates the UpdatedAt timestamp automatically.
func (r *quizRepository) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	// Check if database connection is available
	if r.DB == nil || r.DB.Pool == nil {
		return fmt.Errorf("database connection is required")
	}
	
	// Update timestamp
	quiz.UpdatedAt = time.Now()

	// Update query with college isolation
	sql := `UPDATE quizzes SET title = $1, description = $2, time_limit_minutes = $3, due_date = $4, updated_at = $5
			WHERE id = $6 AND college_id = $7`
	args := []any{quiz.Title, quiz.Description, quiz.TimeLimitMinutes, quiz.DueDate,
				 quiz.UpdatedAt, quiz.ID, quiz.CollegeID}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateQuiz: failed to execute query: %w", err)
	}

	// Check if any rows were affected
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateQuiz: quiz not found or no changes (id: %d, college: %d)", quiz.ID, quiz.CollegeID)
	}

	return nil
}

// UpdateQuizPartial performs a partial update of quiz fields.
// Only fields that are non-nil in the request will be updated.
// Uses dynamic SQL building for efficiency.
func (r *quizRepository) UpdateQuizPartial(ctx context.Context, collegeID int, quizID int, req *models.UpdateQuizRequest) error {
	// Check if database connection is available
	if r.DB == nil || r.DB.Pool == nil {
		return fmt.Errorf("database connection is required")
	}
	
	// Input validation
	if collegeID <= 0 {
		return fmt.Errorf("UpdateQuizPartial: collegeID must be greater than 0")
	}
	if quizID <= 0 {
		return fmt.Errorf("UpdateQuizPartial: quizID must be greater than 0")
	}
	if req == nil {
		return fmt.Errorf("UpdateQuizPartial: UpdateQuizRequest cannot be nil")
	}

	// Check if at least one field is being updated
	hasUpdates := req.Title != nil || req.Description != nil || req.TimeLimitMinutes != nil ||
				 req.DueDate != nil || req.CollegeID != nil || req.CourseID != nil
	if !hasUpdates {
		return fmt.Errorf("UpdateQuizPartial: at least one field must be provided for update")
	}

	// Build dynamic SET clause
	setClauses := []string{"updated_at = NOW()"}
	args := []any{}
	paramCount := 0

	// Add fields conditionally with parameterized queries
	if req.CollegeID != nil {
		paramCount++
		setClauses = append(setClauses, fmt.Sprintf("college_id = $%d", paramCount))
		args = append(args, *req.CollegeID)
	}
	if req.CourseID != nil {
		paramCount++
		setClauses = append(setClauses, fmt.Sprintf("course_id = $%d", paramCount))
		args = append(args, *req.CourseID)
	}
	if req.Title != nil {
		paramCount++
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", paramCount))
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		paramCount++
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", paramCount))
		args = append(args, *req.Description)
	}
	if req.TimeLimitMinutes != nil {
		paramCount++
		setClauses = append(setClauses, fmt.Sprintf("time_limit_minutes = $%d", paramCount))
		args = append(args, *req.TimeLimitMinutes)
	}
	if req.DueDate != nil {
		paramCount++
		setClauses = append(setClauses, fmt.Sprintf("due_date = $%d", paramCount))
		args = append(args, *req.DueDate)
	}

	// Add WHERE clause parameters
	args = append(args, quizID, collegeID)

	// Build final SQL query
	sql := fmt.Sprintf(`UPDATE quizzes SET %s WHERE id = $%d AND college_id = $%d`,
		strings.Join(setClauses, ", "),
		len(setClauses)+1,
		len(setClauses)+2)

	// Execute the update
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateQuizPartial: failed to execute query: %w", err)
	}

	// Check if the quiz was found and updated
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateQuizPartial: quiz not found or no changes (id: %d, college: %d)", quizID, collegeID)
	}

	return nil
}

// DeleteQuiz removes a quiz from the database.
// Ensures college isolation by checking college_id in the WHERE clause.
func (r *quizRepository) DeleteQuiz(ctx context.Context, collegeID int, quizID int) error {
	// Check if database connection is available
	if r.DB == nil || r.DB.Pool == nil {
		return fmt.Errorf("database connection is required")
	}
	
	sql := `DELETE FROM quizzes WHERE id = $1 AND college_id = $2`
	args := []any{quizID, collegeID}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteQuiz: failed to execute query: %w", err)
	}

	// Check if any rows were affected
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteQuiz: quiz not found (id: %d, college: %d)", quizID, collegeID)
	}

	return nil
}

// FindQuizzesByCourse retrieves quizzes for a specific course with pagination.
// Results are ordered by due date (descending) then creation date (descending).
func (r *quizRepository) FindQuizzesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Quiz, error) {
	// Check if database connection is available
	if r.DB == nil || r.DB.Pool == nil {
		return nil, fmt.Errorf("database connection is required")
	}
	
	quizzes := []*models.Quiz{}

	sql := `SELECT id, college_id, course_id, title, description, time_limit_minutes, due_date, created_at, updated_at
			FROM quizzes
			WHERE college_id = $1 AND course_id = $2
			ORDER BY due_date DESC, created_at DESC
			LIMIT $3 OFFSET $4`
	args := []any{collegeID, courseID, limit, offset}

	err := pgxscan.Select(ctx, r.DB.Pool, &quizzes, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindQuizzesByCourse: failed to execute query: %w", err)
	}

	return quizzes, nil
}

// CountQuizzesByCourse returns the total count of quizzes for a specific course.
// Used for pagination calculations.
func (r *quizRepository) CountQuizzesByCourse(ctx context.Context, collegeID int, courseID int) (int, error) {
	// Check if database connection is available
	if r.DB == nil || r.DB.Pool == nil {
		return 0, fmt.Errorf("database connection is required")
	}
	
	sql := `SELECT COUNT(*) FROM quizzes WHERE college_id = $1 AND course_id = $2`
	args := []any{collegeID, courseID}

	temp := struct {
		Count int `db:"count"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("CountQuizzesByCourse: failed to execute query: %w", err)
	}

	return temp.Count, nil
}