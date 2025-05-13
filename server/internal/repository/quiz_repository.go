package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

// Interface Definitions
type QuizRepository interface {
	// Quiz Methods
	CreateQuiz(ctx context.Context, quiz *models.Quiz) error
	GetQuizByID(ctx context.Context, collegeID int, quizID int) (*models.Quiz, error)
	UpdateQuiz(ctx context.Context, quiz *models.Quiz) error
	DeleteQuiz(ctx context.Context, collegeID int, quizID int) error
	FindQuizzesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Quiz, error)
	CountQuizzesByCourse(ctx context.Context, collegeID int, courseID int) (int, error)

	// Question Methods
	CreateQuestion(ctx context.Context, question *models.Question) error
	GetQuestionByID(ctx context.Context, questionID int) (*models.Question, error) // Assuming question ID is globally unique or scoped by quiz later
	UpdateQuestion(ctx context.Context, question *models.Question) error
	DeleteQuestion(ctx context.Context, questionID int) error
	FindQuestionsByQuiz(ctx context.Context, quizID int, limit, offset uint64) ([]*models.Question, error)
	CountQuestionsByQuiz(ctx context.Context, quizID int) (int, error)

	// AnswerOption Methods
	CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error
	GetAnswerOptionByID(ctx context.Context, optionID int) (*models.AnswerOption, error)
	UpdateAnswerOption(ctx context.Context, option *models.AnswerOption) error
	DeleteAnswerOption(ctx context.Context, optionID int) error
	FindAnswerOptionsByQuestion(ctx context.Context, questionID int) ([]*models.AnswerOption, error) // No pagination usually needed here

	// QuizAttempt Methods
	CreateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error
	GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)
	UpdateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error // Updated to return the attempt
	FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error)
	FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error)
	CountQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int) (int, error)
	CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)

	// StudentAnswer Methods
	CreateStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error
	UpdateStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error // For grading
	FindStudentAnswersByAttempt(ctx context.Context, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error)
	GetStudentAnswerForQuestion(ctx context.Context, attemptID int, questionID int) (*models.StudentAnswer, error)
	GetStudentAnswerByID(ctx context.Context, answerID int) (*models.StudentAnswer, error) // Fixed return type

	// Advanced Quiz Methods
	GradeQuizAttempt(ctx context.Context, attemptID int) (*models.QuizAttempt, error)
	FindCompletedQuizAttempts(ctx context.Context, collegeID int, quizID int) ([]*models.QuizAttempt, error)
	GetQuizStatistics(ctx context.Context, collegeID int, quizID int) (*models.QuizStatistics, error)
	FindIncompleteQuizAttemptByStudent(ctx context.Context, collegeID int, studentID int, quizID int) (*models.QuizAttempt, error)
	GetQuestionWithAnswerOptions(ctx context.Context, questionID int) (*models.QuestionWithOptions, error)
}

type quizRepository struct {
	DB *DB
}

func NewQuizRepository(db *DB) QuizRepository {
	return &quizRepository{DB: db}
}

// Table Constants
const (
	quizTable          = "quizzes"
	questionTable      = "questions"
	answerOptionTable  = "answer_options"
	quizAttemptTable   = "quiz_attempts"
	studentAnswerTable = "student_answers"
)

// --- Quiz Methods ---

func (r *quizRepository) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	now := time.Now()
	quiz.CreatedAt = now
	quiz.UpdatedAt = now
	query := r.DB.SQ.Insert(quizTable).
		Columns("college_id", "course_id", "title", "description", "time_limit_minutes", "due_date", "created_at", "updated_at").
		Values(quiz.CollegeID, quiz.CourseID, quiz.Title, quiz.Description, quiz.TimeLimitMinutes, quiz.DueDate, quiz.CreatedAt, quiz.UpdatedAt).
		Suffix("RETURNING id")
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("CreateQuiz: build query: %w", err)
	}
	err = r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&quiz.ID)
	if err != nil {
		return fmt.Errorf("CreateQuiz: exec/scan: %w", err)
	}
	return nil
}

func (r *quizRepository) GetQuizByID(ctx context.Context, collegeID int, quizID int) (*models.Quiz, error) {
	quiz := &models.Quiz{}
	query := r.DB.SQ.Select("id", "college_id", "course_id", "title", "description", "time_limit_minutes", "due_date", "created_at", "updated_at").
		From(quizTable).Where(squirrel.Eq{"id": quizID, "college_id": collegeID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetQuizByID: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, quiz, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetQuizByID: not found (id: %d, college: %d)", quizID, collegeID)
		}
		return nil, fmt.Errorf("GetQuizByID: exec/scan: %w", err)
	}
	return quiz, nil
}

func (r *quizRepository) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	quiz.UpdatedAt = time.Now()
	query := r.DB.SQ.Update(quizTable).
		Set("title", quiz.Title).Set("description", quiz.Description).Set("time_limit_minutes", quiz.TimeLimitMinutes).
		Set("due_date", quiz.DueDate).Set("updated_at", quiz.UpdatedAt).
		Where(squirrel.Eq{"id": quiz.ID, "college_id": quiz.CollegeID})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UpdateQuiz: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateQuiz: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateQuiz: not found or no changes (id: %d, college: %d)", quiz.ID, quiz.CollegeID)
	}
	return nil
}

func (r *quizRepository) DeleteQuiz(ctx context.Context, collegeID int, quizID int) error {
	// Note: Consider cascading deletes or handling related questions/attempts in the service layer
	query := r.DB.SQ.Delete(quizTable).Where(squirrel.Eq{"id": quizID, "college_id": collegeID})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("DeleteQuiz: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteQuiz: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteQuiz: not found (id: %d, college: %d)", quizID, collegeID)
	}
	return nil
}

func (r *quizRepository) FindQuizzesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Quiz, error) {
	quizzes := []*models.Quiz{}
	query := r.DB.SQ.Select("id", "college_id", "course_id", "title", "description", "time_limit_minutes", "due_date", "created_at", "updated_at").
		From(quizTable).Where(squirrel.Eq{"college_id": collegeID, "course_id": courseID}).
		OrderBy("due_date DESC", "created_at DESC").Limit(limit).Offset(offset)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("FindQuizzesByCourse: build query: %w", err)
	}
	err = pgxscan.Select(ctx, r.DB.Pool, &quizzes, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindQuizzesByCourse: exec/scan: %w", err)
	}
	return quizzes, nil
}

func (r *quizRepository) CountQuizzesByCourse(ctx context.Context, collegeID int, courseID int) (int, error) {
	var count int
	query := r.DB.SQ.Select("COUNT(*)").From(quizTable).Where(squirrel.Eq{"college_id": collegeID, "course_id": courseID})
	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("CountQuizzesByCourse: build query: %w", err)
	}
	err = r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountQuizzesByCourse: exec/scan: %w", err)
	}
	return count, nil
}

func (r *quizRepository) CreateQuestion(ctx context.Context, question *models.Question) error {
	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now
	query := r.DB.SQ.Insert(questionTable).
		Columns("quiz_id", "text", "type", "points", "created_at", "updated_at").
		Values(question.QuizID, question.Text, question.Type, question.Points, question.CreatedAt, question.UpdatedAt).
		Suffix("RETURNING id")
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("CreateQuestion: build query: %w", err)
	}
	err = r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&question.ID)
	if err != nil {
		return fmt.Errorf("CreateQuestion: exec/scan: %w", err)
	}
	return nil
}

func (r *quizRepository) GetQuestionByID(ctx context.Context, questionID int) (*models.Question, error) {
	question := &models.Question{}
	query := r.DB.SQ.Select("id", "quiz_id", "text", "type", "points", "created_at", "updated_at").
		From(questionTable).Where(squirrel.Eq{"id": questionID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetQuestionByID: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, question, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetQuestionByID: not found (id: %d)", questionID)
		}
		return nil, fmt.Errorf("GetQuestionByID: exec/scan: %w", err)
	}
	return question, nil
}

func (r *quizRepository) UpdateQuestion(ctx context.Context, question *models.Question) error {
	question.UpdatedAt = time.Now()
	query := r.DB.SQ.Update(questionTable).
		Set("text", question.Text).Set("type", question.Type).Set("points", question.Points).
		Set("updated_at", question.UpdatedAt).Where(squirrel.Eq{"id": question.ID})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UpdateQuestion: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateQuestion: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateQuestion: not found or no changes (id: %d)", question.ID)
	}
	return nil
}

func (r *quizRepository) DeleteQuestion(ctx context.Context, questionID int) error {
	// Note: Consider cascading deletes for options/student answers
	query := r.DB.SQ.Delete(questionTable).Where(squirrel.Eq{"id": questionID})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("DeleteQuestion: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteQuestion: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteQuestion: not found (id: %d)", questionID)
	}
	return nil
}

func (r *quizRepository) FindQuestionsByQuiz(ctx context.Context, quizID int, limit, offset uint64) ([]*models.Question, error) {
	questions := []*models.Question{}
	query := r.DB.SQ.Select("id", "quiz_id", "text", "type", "points", "created_at", "updated_at").
		From(questionTable).Where(squirrel.Eq{"quiz_id": quizID}).
		OrderBy("created_at ASC").Limit(limit).Offset(offset) // Order might be based on question number if added
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("FindQuestionsByQuiz: build query: %w", err)
	}
	err = pgxscan.Select(ctx, r.DB.Pool, &questions, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindQuestionsByQuiz: exec/scan: %w", err)
	}
	return questions, nil
}

func (r *quizRepository) CountQuestionsByQuiz(ctx context.Context, quizID int) (int, error) {
	var count int
	query := r.DB.SQ.Select("COUNT(*)").From(questionTable).Where(squirrel.Eq{"quiz_id": quizID})
	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("CountQuestionsByQuiz: build query: %w", err)
	}
	err = r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountQuestionsByQuiz: exec/scan: %w", err)
	}
	return count, nil
}

// --- AnswerOption Methods --- (Simplified - add Get/Update/Delete similarly if needed) ---

func (r *quizRepository) CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	now := time.Now()
	option.CreatedAt = now
	option.UpdatedAt = now
	query := r.DB.SQ.Insert(answerOptionTable).
		Columns("question_id", "text", "is_correct", "created_at", "updated_at").
		Values(option.QuestionID, option.Text, option.IsCorrect, option.CreatedAt, option.UpdatedAt).
		Suffix("RETURNING id")
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("CreateAnswerOption: build query: %w", err)
	}
	err = r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&option.ID)
	if err != nil {
		return fmt.Errorf("CreateAnswerOption: exec/scan: %w", err)
	}
	return nil
}

func (r *quizRepository) FindAnswerOptionsByQuestion(ctx context.Context, questionID int) ([]*models.AnswerOption, error) {
	options := []*models.AnswerOption{}
	query := r.DB.SQ.Select("id", "question_id", "text", "is_correct", "created_at", "updated_at").
		From(answerOptionTable).Where(squirrel.Eq{"question_id": questionID}).OrderBy("created_at ASC")
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("FindAnswerOptionsByQuestion: build query: %w", err)
	}
	err = pgxscan.Select(ctx, r.DB.Pool, &options, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindAnswerOptionsByQuestion: exec/scan: %w", err)
	}
	return options, nil
}

func (r *quizRepository) GetAnswerOptionByID(ctx context.Context, optionID int) (*models.AnswerOption, error) {
	option := &models.AnswerOption{}
	query := r.DB.SQ.Select("id", "question_id", "text", "is_correct", "created_at", "updated_at").
		From(answerOptionTable).Where(squirrel.Eq{"id": optionID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetAnswerOptionByID: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, option, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetAnswerOptionByID: not found (id: %d)", optionID)
		}
		return nil, fmt.Errorf("GetAnswerOptionByID: exec/scan: %w", err)
	}
	return option, nil
}

func (r *quizRepository) UpdateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	option.UpdatedAt = time.Now()
	query := r.DB.SQ.Update(answerOptionTable).
		Set("text", option.Text).Set("is_correct", option.IsCorrect).
		Set("updated_at", option.UpdatedAt).Where(squirrel.Eq{"id": option.ID})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UpdateAnswerOption: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateAnswerOption: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateAnswerOption: not found or no changes (id: %d)", option.ID)
	}
	return nil
}

func (r *quizRepository) DeleteAnswerOption(ctx context.Context, optionID int) error {
	query := r.DB.SQ.Delete(answerOptionTable).Where(squirrel.Eq{"id": optionID})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("DeleteAnswerOption: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteAnswerOption: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteAnswerOption: not found (id: %d)", optionID)
	}
	return nil
}

func (r *quizRepository) CreateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error {
	now := time.Now()
	attempt.CreatedAt = now
	attempt.UpdatedAt = now
	if attempt.StartTime.IsZero() {
		attempt.StartTime = now
	} // Default start time
	if attempt.Status == "" {
		attempt.Status = "InProgress"
	} // Default status

	query := r.DB.SQ.Insert(quizAttemptTable).
		Columns("student_id", "quiz_id", "college_id", "start_time", "end_time", "score", "status", "created_at", "updated_at").
		Values(attempt.StudentID, attempt.QuizID, attempt.CollegeID, attempt.StartTime, attempt.EndTime, attempt.Score, attempt.Status, attempt.CreatedAt, attempt.UpdatedAt).
		Suffix("RETURNING id")
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("CreateQuizAttempt: build query: %w", err)
	}
	err = r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&attempt.ID)
	if err != nil {
		return fmt.Errorf("CreateQuizAttempt: exec/scan: %w", err)
	}
	return nil
}

func (r *quizRepository) GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error) {
	attempt := &models.QuizAttempt{}
	query := r.DB.SQ.Select("id", "student_id", "quiz_id", "college_id", "start_time", "end_time", "score", "status", "created_at", "updated_at").
		From(quizAttemptTable).Where(squirrel.Eq{"id": attemptID, "college_id": collegeID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetQuizAttemptByID: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, attempt, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetQuizAttemptByID: not found (id: %d, college: %d)", attemptID, collegeID)
		}
		return nil, fmt.Errorf("GetQuizAttemptByID: exec/scan: %w", err)
	}
	return attempt, nil
}

func (r *quizRepository) UpdateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error {
	attempt.UpdatedAt = time.Now()
	query := r.DB.SQ.Update(quizAttemptTable).
		Set("end_time", attempt.EndTime).Set("score", attempt.Score).Set("status", attempt.Status).
		Set("updated_at", attempt.UpdatedAt).
		Where(squirrel.Eq{"id": attempt.ID, "college_id": attempt.CollegeID}) // Ensure scoping
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UpdateQuizAttempt: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateQuizAttempt: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateQuizAttempt: not found or no changes (id: %d)", attempt.ID)
	}
	return nil
}

// --- StudentAnswer Methods --- (Simplified Get/Update/Find - add similarly) ---

func (r *quizRepository) CreateStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error {
	// This often acts like an Upsert: Insert or Update if exists
	now := time.Now()
	answer.CreatedAt = now
	answer.UpdatedAt = now

	query := r.DB.SQ.Insert(studentAnswerTable).
		Columns("quiz_attempt_id", "question_id", "selected_option_id", "answer_text", "is_correct", "points_awarded", "created_at", "updated_at").
		Values(answer.QuizAttemptID, answer.QuestionID, answer.SelectedOptionID, answer.AnswerText, answer.IsCorrect, answer.PointsAwarded, answer.CreatedAt, answer.UpdatedAt).
		Suffix(`ON CONFLICT (quiz_attempt_id, question_id) DO UPDATE SET
                selected_option_id = EXCLUDED.selected_option_id,
                answer_text = EXCLUDED.answer_text,
                is_correct = EXCLUDED.is_correct,
                points_awarded = EXCLUDED.points_awarded,
                updated_at = EXCLUDED.updated_at
              RETURNING id`) // Return ID whether inserted or updated

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("CreateStudentAnswer: build query: %w", err)
	}
	err = r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&answer.ID)
	if err != nil {
		return fmt.Errorf("CreateStudentAnswer: exec/scan: %w", err)
	}
	return nil
}

func (r *quizRepository) UpdateStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error {
	// Primarily used for grading
	answer.UpdatedAt = time.Now()
	query := r.DB.SQ.Update(studentAnswerTable).
		Set("is_correct", answer.IsCorrect).
		Set("points_awarded", answer.PointsAwarded).
		Set("updated_at", answer.UpdatedAt).
		Where(squirrel.Eq{"id": answer.ID}) // Update by primary key
		// Or update by attempt_id and question_id if ID is not known
		// Where(squirrel.Eq{"quiz_attempt_id": answer.QuizAttemptID, "question_id": answer.QuestionID})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UpdateStudentAnswer: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateStudentAnswer: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateStudentAnswer: not found or no changes (id: %d)", answer.ID)
	}
	return nil
}

func (r *quizRepository) FindStudentAnswersByAttempt(ctx context.Context, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error) {
	answers := []*models.StudentAnswer{}
	query := r.DB.SQ.Select("id", "quiz_attempt_id", "question_id", "selected_option_id", "answer_text", "is_correct", "points_awarded", "created_at", "updated_at").
		From(studentAnswerTable).Where(squirrel.Eq{"quiz_attempt_id": attemptID}).
		OrderBy("question_id ASC").Limit(limit).Offset(offset) // Order by question
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("FindStudentAnswersByAttempt: build query: %w", err)
	}
	err = pgxscan.Select(ctx, r.DB.Pool, &answers, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindStudentAnswersByAttempt: exec/scan: %w", err)
	}
	return answers, nil
}

func (r *quizRepository) GetStudentAnswerForQuestion(ctx context.Context, attemptID int, questionID int) (*models.StudentAnswer, error) {
	answer := &models.StudentAnswer{}
	query := r.DB.SQ.Select("id", "quiz_attempt_id", "question_id", "selected_option_id", "answer_text", "is_correct", "points_awarded", "created_at", "updated_at").
		From(studentAnswerTable).Where(squirrel.Eq{"quiz_attempt_id": attemptID, "question_id": questionID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetStudentAnswerForQuestion: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, answer, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetStudentAnswerForQuestion: not found (attempt: %d, question: %d)", attemptID, questionID)
		}
		return nil, fmt.Errorf("GetStudentAnswerForQuestion: exec/scan: %w", err)
	}
	return answer, nil
}

// Implement Find/Count methods for QuizAttempt similarly to other repositories...
func (r *quizRepository) FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
	attempts := []*models.QuizAttempt{}
	query := r.DB.SQ.Select("id", "student_id", "quiz_id", "college_id", "start_time", "end_time", "score", "status", "created_at", "updated_at").
		From(quizAttemptTable).
		Where(squirrel.Eq{"college_id": collegeID, "student_id": studentID}).
		OrderBy("start_time DESC").Limit(limit).Offset(offset)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("FindQuizAttemptsByStudent: build query: %w", err)
	}
	err = pgxscan.Select(ctx, r.DB.Pool, &attempts, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindQuizAttemptsByStudent: exec/scan: %w", err)
	}
	return attempts, nil
}

func (r *quizRepository) FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
	attempts := []*models.QuizAttempt{}
	query := r.DB.SQ.Select("id", "student_id", "quiz_id", "college_id", "start_time", "end_time", "score", "status", "created_at", "updated_at").
		From(quizAttemptTable).
		Where(squirrel.Eq{"college_id": collegeID, "quiz_id": quizID}).
		OrderBy("student_id ASC", "start_time DESC").Limit(limit).Offset(offset)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("FindQuizAttemptsByQuiz: build query: %w", err)
	}
	err = pgxscan.Select(ctx, r.DB.Pool, &attempts, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindQuizAttemptsByQuiz: exec/scan: %w", err)
	}
	return attempts, nil
}

func (r *quizRepository) countQuizAttempts(ctx context.Context, whereClause squirrel.Sqlizer) (int, error) {
	var count int
	query := r.DB.SQ.Select("COUNT(*)").From(quizAttemptTable).Where(whereClause)
	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("countQuizAttempts: build query: %w", err)
	}
	err = r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("countQuizAttempts: exec/scan: %w", err)
	}
	return count, nil
}

func (r *quizRepository) CountQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int) (int, error) {
	where := squirrel.Eq{"college_id": collegeID, "student_id": studentID}
	return r.countQuizAttempts(ctx, where)
}

func (r *quizRepository) CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
	where := squirrel.Eq{"college_id": collegeID, "quiz_id": quizID}
	return r.countQuizAttempts(ctx, where)
}

func (r *quizRepository) GetStudentAnswerByID(ctx context.Context, answerID int) (*models.StudentAnswer, error) {
	answer := &models.StudentAnswer{}
	query := r.DB.SQ.Select("id", "quiz_attempt_id", "question_id", "selected_option_id", "answer_text", "is_correct", "points_awarded", "created_at", "updated_at").
		From(studentAnswerTable).Where(squirrel.Eq{"id": answerID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetStudentAnswerByID: build query: %w", err)
	}

	err = pgxscan.Get(ctx, r.DB.Pool, answer, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetStudentAnswerByID: not found (id: %d)", answerID)
		}
		return nil, fmt.Errorf("GetStudentAnswerByID: exec/scan: %w", err)
	}

	return answer, nil
}

// GradeQuizAttempt updates a quiz attempt with the final score based on student answers
func (r *quizRepository) GradeQuizAttempt(ctx context.Context, attemptID int) (*models.QuizAttempt, error) {
	// First, get all student answers for this attempt
	// studentAnswers, err := r.FindStudentAnswersByAttempt(ctx, attemptID, 1000, 0) // Assuming we won't have more than 1000 answers per attempt
	// if err != nil {
	// 	return nil, fmt.Errorf("GradeQuizAttempt: finding student answers: %w", err)
	// }

	// // Calculate the total score
	// var totalScore int
	// for _, answer := range studentAnswers {
	// 	if answer.PointsAwarded != nil {
	// 		totalScore += *answer.PointsAwarded
	// 	}
	// }

	// // Get the attempt
	// // First get the quiz attempt to get the collegeID
	// tx, err := r.DB.Pool.Begin(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("GradeQuizAttempt: begin transaction: %w", err)
	// }
	// defer tx.Rollback(ctx)

	// var collegeID int
	// err = tx.QueryRow(ctx, "SELECT q.college_id FROM quizzes q JOIN quiz_attempts qa ON q.id = qa.quiz_id WHERE qa.id = $1", attemptID).Scan(&collegeID)
	// if err != nil {
	// 	return nil, fmt.Errorf("GradeQuizAttempt: get college ID: %w", err)
	// }

	// attempt, err := r.GetQuizAttemptByID(ctx, collegeID, attemptID)
	// if err != nil {
	// 	return nil, fmt.Errorf("GradeQuizAttempt: getting quiz attempt: %w", err)
	// }

	// // Update the attempt with the score and completion status
	// attempt.Score = totalScore
	// attempt.CompletedAt = time.Now()
	// attempt.Status = "completed"

	// // Save the updated attempt
	// err := r.UpdateQuizAttempt(ctx, attempt)
	// if err != nil {
	// 	return nil, fmt.Errorf("GradeQuizAttempt: updating quiz attempt: %w", err)
	// }

	// return updatedAttempt, nil
}

// FindCompletedQuizAttempts finds all completed quiz attempts for a specific quiz
func (r *quizRepository) FindCompletedQuizAttempts(ctx context.Context, collegeID int, quizID int) ([]*models.QuizAttempt, error) {
	attempts := []*models.QuizAttempt{}
	query := r.DB.SQ.Select("id", "student_id", "quiz_id", "status", "score", "started_at", "completed_at", "created_at", "updated_at").
		From(quizAttemptTable).
		Join(fmt.Sprintf("%s ON %s.id = %s.quiz_id", quizTable, quizTable, quizAttemptTable)).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.college_id", quizTable):     collegeID,
			fmt.Sprintf("%s.quiz_id", quizAttemptTable): quizID,
			fmt.Sprintf("%s.status", quizAttemptTable):  "completed",
		}).
		OrderBy(fmt.Sprintf("%s.completed_at DESC", quizAttemptTable))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("FindCompletedQuizAttempts: build query: %w", err)
	}

	err = pgxscan.Select(ctx, r.DB.Pool, &attempts, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindCompletedQuizAttempts: exec/scan: %w", err)
	}

	return attempts, nil
}

// GetQuizStatistics returns statistics for a quiz including average score, attempts, etc.
func (r *quizRepository) GetQuizStatistics(ctx context.Context, collegeID int, quizID int) (*models.QuizStatistics, error) {
	// Get the total number of attempts
	totalAttempts, err := r.CountQuizAttemptsByQuiz(ctx, collegeID, quizID)
	if err != nil {
		return nil, fmt.Errorf("GetQuizStatistics: counting attempts: %w", err)
	}

	// Query for completed attempts
	completedAttempts, err := r.FindCompletedQuizAttempts(ctx, collegeID, quizID)
	if err != nil {
		return nil, fmt.Errorf("GetQuizStatistics: finding completed attempts: %w", err)
	}

	// Calculate statistics
	stats := &models.QuizStatistics{
		QuizID:            quizID,
		TotalAttempts:     totalAttempts,
		CompletedAttempts: len(completedAttempts),
		HighestScore:      0,
		LowestScore:       0,
		AverageScore:      0,
	}

	if len(completedAttempts) > 0 {
		var totalScore int
		stats.HighestScore = *completedAttempts[0].Score
		stats.LowestScore = *completedAttempts[0].Score

		for _, attempt := range completedAttempts {
			totalScore += *attempt.Score
			if *attempt.Score > stats.HighestScore {
				stats.HighestScore = *attempt.Score
			}
			if *attempt.Score < stats.LowestScore {
				stats.LowestScore = *attempt.Score
			}
		}

		stats.AverageScore = totalScore / len(completedAttempts)
	}

	return stats, nil
}

// FindIncompleteQuizAttemptByStudent finds an incomplete quiz attempt for a student
func (r *quizRepository) FindIncompleteQuizAttemptByStudent(ctx context.Context, collegeID int, studentID int, quizID int) (*models.QuizAttempt, error) {
	attempt := &models.QuizAttempt{}
	query := r.DB.SQ.Select(fmt.Sprintf("%s.id", quizAttemptTable),
		fmt.Sprintf("%s.student_id", quizAttemptTable),
		fmt.Sprintf("%s.quiz_id", quizAttemptTable),
		fmt.Sprintf("%s.status", quizAttemptTable),
		fmt.Sprintf("%s.score", quizAttemptTable),
		fmt.Sprintf("%s.started_at", quizAttemptTable),
		fmt.Sprintf("%s.completed_at", quizAttemptTable),
		fmt.Sprintf("%s.created_at", quizAttemptTable),
		fmt.Sprintf("%s.updated_at", quizAttemptTable)).
		From(quizAttemptTable).
		Join(fmt.Sprintf("%s ON %s.id = %s.quiz_id", quizTable, quizTable, quizAttemptTable)).
		Where(squirrel.Eq{
			fmt.Sprintf("%s.college_id", quizTable):        collegeID,
			fmt.Sprintf("%s.student_id", quizAttemptTable): studentID,
			fmt.Sprintf("%s.quiz_id", quizAttemptTable):    quizID,
			fmt.Sprintf("%s.status", quizAttemptTable):     "in_progress",
		}).
		OrderBy(fmt.Sprintf("%s.created_at DESC", quizAttemptTable)).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("FindIncompleteQuizAttemptByStudent: build query: %w", err)
	}

	err = pgxscan.Get(ctx, r.DB.Pool, attempt, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No in-progress attempt found
		}
		return nil, fmt.Errorf("FindIncompleteQuizAttemptByStudent: exec/scan: %w", err)
	}

	return attempt, nil
}

// GetQuestionWithAnswerOptions retrieves a question with all its answer options
func (r *quizRepository) GetQuestionWithAnswerOptions(ctx context.Context, questionID int) (*models.QuestionWithOptions, error) {
	// First get the question
	question, err := r.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("GetQuestionWithAnswerOptions: get question: %w", err)
	}

	// Then get the answer options
	options, err := r.FindAnswerOptionsByQuestion(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("GetQuestionWithAnswerOptions: get answer options: %w", err)
	}

	// Combine into one object
	result := &models.QuestionWithOptions{
		Question:      question,
		AnswerOptions: options,
	}

	return result, nil
}

func (r *quizRepository) GetQuestionWithCorrectAnswers(ctx context.Context, questionID int) (*models.QuestionWithCorrectAnswers, error) {
	question, err := r.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("GetQuestionWithCorrect Answer", err)
	}

	options, err := r.FindAnswerOptionsByQuestion(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("FindAnswerOptions By Question %w", err)
	}
	var correctOptions []*models.AnswerOption
	for _, option := range options {
		if option.IsCorrect {
			correctOptions = append(correctOptions, option)
		}
	}
	return &models.QuestionWithCorrectAnswers{
		Question:       question,
		CorrectOptions: correctOptions,
	}, nil

}
