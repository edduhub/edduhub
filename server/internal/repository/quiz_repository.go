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
	GetQuestionByID(ctx context.Context, collegeID int, questionID int) (*models.Question, error)
	UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error
	DeleteQuestion(ctx context.Context, collegeID int, questionID int) error
	FindQuestionsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.Question, error)
	CountQuestionsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)

	// AnswerOption Methods
	CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error
	GetAnswerOptionByID(ctx context.Context, collegeID int, optionID int) (*models.AnswerOption, error)
	UpdateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error
	DeleteAnswerOption(ctx context.Context, collegeID int, optionID int) error
	FindAnswerOptionsByQuestion(ctx context.Context, questionID int) ([]*models.AnswerOption, error) // No pagination usually needed here

	// QuizAttempt Methods
	CreateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error
	GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)
	UpdateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error // Updated to return the attempt
	FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error)

	// StudentAnswer Methods
	CreateStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error
	UpdateStudentAnswer(ctx context.Context, collegeID int, answer *models.StudentAnswer) error // For grading, add collegeID
	FindStudentAnswersByAttempt(ctx context.Context, collegeID int, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error)
	GetStudentAnswerForQuestion(ctx context.Context, collegeID int, attemptID int, questionID int) (*models.StudentAnswer, error)
	GetStudentAnswerByID(ctx context.Context, collegeID int, answerID int) (*models.StudentAnswer, error) // Fixed return type

	// Advanced Quiz Methods
	GradeQuizAttempt(ctx context.Context, collegeID, attemptID int) (*models.QuizAttempt, error)
	FindCompletedQuizAttempts(ctx context.Context, collegeID int, quizID int) ([]*models.QuizAttempt, error)
	GetQuizStatistics(ctx context.Context, collegeID int, quizID int) (*models.QuizStatistics, error)
	FindIncompleteQuizAttemptByStudent(ctx context.Context, collegeID int, studentID int, quizID int) (*models.QuizAttempt, error)
	GetQuestionWithAnswerOptions(ctx context.Context, collegeID, questionID int) (*models.QuestionWithOptions, error)
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

func (r *quizRepository) GetQuestionByID(ctx context.Context, collegeID int, questionID int) (*models.Question, error) {
	question := &models.Question{}
	// Join with quizzes table to filter by college_id
	query := r.DB.SQ.Select("q.id", "q.quiz_id", "q.text", "q.type", "q.points", "q.created_at", "q.updated_at").
		From(questionTable + " AS q").
		Join(quizTable + " AS qz ON q.quiz_id = qz.id").
		Where(squirrel.Eq{"q.id": questionID, "qz.college_id": collegeID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetQuestionByID: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, question, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return a more specific error indicating it wasn't found *for this college*
			return nil, fmt.Errorf("GetQuestionByID: not found (id: %d for college %d)", questionID, collegeID)
		}
		return nil, fmt.Errorf("GetQuestionByID: exec/scan: %w", err)
	}
	// Ensure the QuizID is populated if needed by the service layer
	// question.QuizID = ... // This would require selecting q.quiz_id
	return question, nil
}

func (r *quizRepository) UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	question.UpdatedAt = time.Now()
	query := r.DB.SQ.Update(questionTable).
		Set("text", question.Text).Set("type", question.Type).Set("points", question.Points).
		Set("updated_at", question.UpdatedAt).
		Where(squirrel.Eq{"id": question.ID}).
		// Ensure the question belongs to the correct college by joining or subquery
		Where(fmt.Sprintf("quiz_id IN (SELECT id FROM %s WHERE college_id = ?)", quizTable), collegeID)
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UpdateQuestion: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateQuestion: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateQuestion: not found or no changes (id: %d for college %d)", question.ID, collegeID)
	}
	return nil
}

// Corrected DeleteQuestion implementation
func (r *quizRepository) DeleteQuestion(ctx context.Context, collegeID int, questionID int) error {
	// Ensure the question belongs to the given college by checking its quiz
	query := r.DB.SQ.Delete(questionTable).
		Where(squirrel.Eq{"id": questionID}).
		Where(fmt.Sprintf("quiz_id IN (SELECT id FROM %s WHERE college_id = ?)", quizTable), collegeID)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("DeleteQuestion: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteQuestion: exec for college %d: %w", collegeID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteQuestion: not found (id: %d for college %d)", questionID, collegeID)
	}
	return nil
}

func (r *quizRepository) FindQuestionsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.Question, error) {
	questions := []*models.Question{}
	// Join with quizzes table to filter by college_id
	query := r.DB.SQ.Select("q.id", "q.quiz_id", "q.text", "q.type", "q.points", "q.created_at", "q.updated_at").
		From(questionTable + " AS q").
		Join(quizTable + " AS qz ON q.quiz_id = qz.id").Where(squirrel.Eq{"q.quiz_id": quizID, "qz.college_id": collegeID}).
		OrderBy("created_at ASC").Limit(limit).Offset(offset)
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

func (r *quizRepository) CountQuestionsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
	var count int
	// Join with quizzes table to filter by college_id
	query := r.DB.SQ.Select("COUNT(q.id)").
		From(questionTable + " AS q").
		Join(quizTable + " AS qz ON q.quiz_id = qz.id").Where(squirrel.Eq{"q.quiz_id": quizID, "qz.college_id": collegeID})
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

func (r *quizRepository) GetAnswerOptionByID(ctx context.Context, collegeID int, optionID int) (*models.AnswerOption, error) {
	option := &models.AnswerOption{}
	// Join through questions and quizzes to filter by college_id
	query := r.DB.SQ.Select("ao.id", "ao.question_id", "ao.text", "ao.is_correct", "ao.created_at", "ao.updated_at").
		From(answerOptionTable + " AS ao").
		Join(questionTable + " AS q ON ao.question_id = q.id").
		Join(quizTable + " AS qz ON q.quiz_id = qz.id").Where(squirrel.Eq{"ao.id": optionID, "qz.college_id": collegeID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetAnswerOptionByID: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, option, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetAnswerOptionByID: not found (id: %d for college %d)", optionID, collegeID)
		}
		return nil, fmt.Errorf("GetAnswerOptionByID: exec/scan: %w", err)
	}
	// Ensure QuestionID is populated
	// option.QuestionID = ... // Needs to be selected in the query

	return option, nil
}

func (r *quizRepository) UpdateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error {
	option.UpdatedAt = time.Now()
	query := r.DB.SQ.Update(answerOptionTable).
		Set("text", option.Text).Set("is_correct", option.IsCorrect).
		Set("updated_at", option.UpdatedAt).
		Where(squirrel.Eq{"id": option.ID}).
		// Ensure the option belongs to the correct college by joining or subquery
		Where(fmt.Sprintf("question_id IN (SELECT q.id FROM %s q JOIN %s qz ON q.quiz_id = qz.id WHERE qz.college_id = ?)", questionTable, quizTable), collegeID)
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UpdateAnswerOption: build query: %w", err)
	}

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateAnswerOption: exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateAnswerOption: not found or no changes (id: %d for college %d)", option.ID, collegeID)
	}
	return nil
}

// Corrected DeleteAnswerOption implementation
func (r *quizRepository) DeleteAnswerOption(ctx context.Context, collegeID int, optionID int) error {
	// Ensure the option belongs to the given college by checking its question and quiz
	query := r.DB.SQ.Delete(answerOptionTable).
		Where(squirrel.Eq{"id": optionID}).
		Where(fmt.Sprintf("question_id IN (SELECT q.id FROM %s q JOIN %s qz ON q.quiz_id = qz.id WHERE qz.college_id = ?)", questionTable, quizTable), collegeID)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("DeleteAnswerOption: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteAnswerOption: exec for college %d: %w", collegeID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no options deleted")
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

func (r *quizRepository) FindStudentAnswersByAttempt(ctx context.Context, collegeID int, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error) {
	answers := []*models.StudentAnswer{}
	// Join with quiz_attempts table to filter by college_id
	query := r.DB.SQ.Select("id", "quiz_attempt_id", "question_id", "selected_option_id", "answer_text", "is_correct", "points_awarded", "created_at", "updated_at").
		From(studentAnswerTable + " AS sa").
		Join(quizAttemptTable + " AS qa ON sa.quiz_attempt_id = qa.id").
		Where(squirrel.Eq{"sa.quiz_attempt_id": attemptID, "qa.college_id": collegeID}).
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

func (r *quizRepository) GetStudentAnswerForQuestion(ctx context.Context, collegeID int, attemptID int, questionID int) (*models.StudentAnswer, error) {
	answer := &models.StudentAnswer{}
	// Join with quiz_attempts table to filter by college_id
	query := r.DB.SQ.Select("id", "quiz_attempt_id", "question_id", "selected_option_id", "answer_text", "is_correct", "points_awarded", "created_at", "updated_at").
		From(studentAnswerTable + " AS sa").
		Join(quizAttemptTable + " AS qa ON sa.quiz_attempt_id = qa.id").
		Where(squirrel.Eq{"sa.quiz_attempt_id": attemptID, "sa.question_id": questionID, "qa.college_id": collegeID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetStudentAnswerForQuestion: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, answer, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetStudentAnswerForQuestion: not found (attempt: %d, question: %d for college %d)", attemptID, questionID, collegeID)
		}
		return nil, fmt.Errorf("GetStudentAnswerForQuestion: exec/scan: %w", err)
	}
	return answer, nil
}

func (r *quizRepository) GetStudentAnswerByID(ctx context.Context, collegeID int, answerID int) (*models.StudentAnswer, error) {
	answer := &models.StudentAnswer{}
	// Join with quiz_attempts table to filter by college_id
	query := r.DB.SQ.Select("sa.id", "sa.quiz_attempt_id", "sa.question_id", "sa.selected_option_id", "sa.answer_text", "sa.is_correct", "sa.points_awarded", "sa.created_at", "sa.updated_at").
		From(studentAnswerTable + " AS sa").
		Join(quizAttemptTable + " AS qa ON sa.quiz_attempt_id = qa.id").
		Where(squirrel.Eq{"sa.id": answerID, "qa.college_id": collegeID})
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetStudentAnswerByID: build query: %w", err)
	}
	err = pgxscan.Get(ctx, r.DB.Pool, answer, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetStudentAnswerByID: not found (id: %d for college %d)", answerID, collegeID)
		}
		return nil, fmt.Errorf("GetStudentAnswerByID: exec/scan: %w", err)
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

// Corrected UpdateStudentAnswer implementation
func (r *quizRepository) UpdateStudentAnswer(ctx context.Context, collegeID int, answer *models.StudentAnswer) error {
	existingAnswer, err := r.GetStudentAnswerByID(ctx, collegeID, answer.ID)
	if err != nil {
		return fmt.Errorf("UpdateStudentAnswer Error : student answer update error %w", err)

	}
	if existingAnswer.QuizAttemptID != answer.QuizAttemptID {
		return fmt.Errorf("UpdateStudentAnswer: attempt ID mismatch for answer ID %d", answer.ID)
	}

	answer.UpdatedAt = time.Now()
	query := r.DB.SQ.Update(studentAnswerTable).
		Set("is_correct", answer.IsCorrect).
		Set("points_awarded", answer.PointsAwarded).
		Set("updated_at", answer.UpdatedAt).Where(squirrel.Eq{"college_id": collegeID, "id": answer.ID})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("UpdateStudentAnswer: build query: %w", err)
	}
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateStudentAnswer: exec for college %d: %w", collegeID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateStudentAnswer: not found or no changes (id: %d for college %d)", answer.ID, collegeID)
	}
	return nil
}

func (r *quizRepository) GetStudentAnswersForAttempt(ctx context.Context, collegeID int, attemptID int) ([]*models.QuestionWithStudentAnswer, error) {
	query := r.DB.SQ.
		Select(
			"q.id AS question_id", "q.quiz_id", "q.text", "q.type", "q.points", "q.created_at", "q.updated_at",
			"sa.id AS student_answer_id", "sa.quiz_attempt_id", "sa.question_id AS sa_question_id", "sa.selected_option_id", "sa.answer_text", "sa.is_correct", "sa.points_awarded", "sa.created_at", "sa.updated_at",
		).
		From(questionTable+" AS q").
		LeftJoin(studentAnswerTable+" AS sa ON q.id = sa.question_id AND sa.quiz_attempt_id = ?", attemptID).
		Join(quizTable + " AS qu ON q.quiz_id = qu.id").
		Where(squirrel.Eq{"qu.college_id": collegeID}).
		OrderBy("q.id")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("GetStudentAnswersForAttempt: build query: %w", err)
	}

	rows, err := r.DB.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("GetStudentAnswersForAttempt: exec query: %w", err)
	}
	defer rows.Close()

	questionMap := make(map[int]*models.QuestionWithStudentAnswer)

	for rows.Next() {
		var question models.Question
		var studentAnswer models.StudentAnswer
		var studentAnswerID *int // Use a pointer to handle NULL values

		err := rows.Scan(
			&question.ID, &question.QuizID, &question.Text, &question.Type, &question.Points, &question.CreatedAt, &question.UpdatedAt,
			&studentAnswerID, &studentAnswer.QuizAttemptID, &studentAnswer.QuestionID, &studentAnswer.SelectedOptionID, &studentAnswer.AnswerText, &studentAnswer.IsCorrect, &studentAnswer.PointsAwarded, &studentAnswer.CreatedAt, &studentAnswer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetStudentAnswersForAttempt: scan row: %w", err)
		}

		if _, ok := questionMap[question.ID]; !ok {
			questionMap[question.ID] = &models.QuestionWithStudentAnswer{Question: &question}
		}
		if studentAnswerID != nil {
			studentAnswer.ID = *studentAnswerID // Dereference the pointer if it's not nil
			questionMap[question.ID].StudentAnswer = append(questionMap[question.ID].StudentAnswer, &studentAnswer)
		}
	}

	var result []*models.QuestionWithStudentAnswer
	for _, q := range questionMap {
		result = append(result, q)
	}
	return result, nil
}

// GradeQuizAttempt updates a quiz attempt with the final score based on student answers
func (r *quizRepository) GradeQuizAttempt(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error) {
	// 1. Fetch the quiz attempt to get QuizID, StudentID, and verify CollegeID.
	attempt, err := r.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("GradeQuizAttempt: failed to get quiz attempt %d for college %d: %w", attemptID, collegeID, err)
	}
	if attempt == nil {
		return nil, fmt.Errorf("GradeQuizAttempt: quiz attempt %d not found for college %d", attemptID, collegeID)
	}

	// 2. Fetch all questions for the quiz.
	questions, err := r.FindQuestionsByQuiz(ctx, collegeID, attempt.QuizID, 0, 0) // 0, 0 for no limit/offset
	if err != nil {
		return nil, fmt.Errorf("GradeQuizAttempt: failed to find questions for quiz %d: %w", attempt.QuizID, err)
	}

	// 3. Fetch all student answers for this attempt.
	allStudentAnswersForAttempt, err := r.FindStudentAnswersByAttempt(ctx, collegeID, attemptID, 0, 0) // 0, 0 for no limit/offset
	if err != nil {
		// If error is not pgx.ErrNoRows, then it's a problem. ErrNoRows is acceptable (student might not have answered anything).
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GradeQuizAttempt: failed to find student answers for attempt %d: %w", attemptID, err)
		}
		// If ErrNoRows, allStudentAnswersForAttempt will be empty or nil, which is fine.
	}

	// Create a map for quick lookup of student answers by question ID.
	studentAnswersMap := make(map[int]*models.StudentAnswer)
	for _, sa := range allStudentAnswersForAttempt {
		studentAnswersMap[sa.QuestionID] = sa
	}

	var totalScoreAchieved int = 0

	// 4. Grade each question.
	for _, question := range questions {
		studentAnswer, studentDidAnswer := studentAnswersMap[question.ID]

		var pointsAwardedForThisQuestion int = 0
		var isThisAnswerCorrect bool = false // Default to false

		// Check if the answer was already manually graded (PointsAwarded is not nil).
		if studentDidAnswer && studentAnswer.PointsAwarded != nil {
			studentAnswer.PointsAwarded = 1

			pointsAwardedForThisQuestion = *studentAnswer.PointsAwarded
			if studentAnswer.IsCorrect != nil {
				isThisAnswerCorrect = *studentAnswer.IsCorrect
			} else {
				// If points are awarded but IsCorrect is nil, assume correct if points > 0.
				isThisAnswerCorrect = pointsAwardedForThisQuestion > 0
			}
		} else if studentDidAnswer { // Auto-grade if student answered and not manually graded.
			// Fetch correct answer options for this question to perform auto-grading.
			// GetQuestionWithCorrectAnswers returns the question model and a slice of *only* its correct options.
			questionWithCorrectOptions, errCorrectOpt := r.GetQuestionWithCorrectAnswers(ctx, collegeID, question.ID)
			if errCorrectOpt != nil {
				// Log warning, but continue grading other questions. Points for this question will be 0.
				fmt.Printf("Warning: GradeQuizAttempt: failed to get correct answer options for question %d: %v\n", question.ID, errCorrectOpt)
			} else {
				// Perform auto-grading based on question type
				if question.Type == models.MultipleChoice || question.Type == models.TrueFalse {
					// gradeMultipleChoice handles *[]int SelectedOptionID and compares against correct options.
					// For True/False, it works if correct options are set up like MC (e.g., one correct option).
					isThisAnswerCorrect, pointsAwardedForThisQuestion = gradeMultipleChoice(studentAnswer, question, questionWithCorrectOptions.CorrectOptions)
				} else if question.Type == "ShortAnswer" { // Assuming "ShortAnswer" is a valid type from models.QuizType
					// Auto-grading short answers is complex. Default to 0 points unless manually graded.
					// You might implement keyword matching or other heuristics here if desired.
					isThisAnswerCorrect = false
					pointsAwardedForThisQuestion = 0
				}
				// Add other question types (e.g., FillInTheBlanks) as needed.
			}
		}
		// If !studentDidAnswer (studentDidAnswer is false), pointsAwardedForThisQuestion remains 0, isThisAnswerCorrect remains false.

		totalScoreAchieved += pointsAwardedForThisQuestion

		// 5. Upsert the StudentAnswer record with grading results.
		// CreateStudentAnswer handles insert or update based on (quiz_attempt_id, question_id) conflict.
		saToUpsert := models.StudentAnswer{
			QuizAttemptID: attempt.ID,
			QuestionID:    question.ID,
			IsCorrect:     &isThisAnswerCorrect,
			PointsAwarded: &pointsAwardedForThisQuestion,
		}

		if studentDidAnswer {
			// If student answered, preserve their original answer details (ID, selected options, text).
			// CreateStudentAnswer's ON CONFLICT clause will use these if updating.
			saToUpsert.ID = studentAnswer.ID // Important if upsert logic might rely on ID for existing records.
			saToUpsert.SelectedOptionID = studentAnswer.SelectedOptionID
			saToUpsert.AnswerText = studentAnswer.AnswerText
		} else {
			// If student did not answer, SelectedOptionID and AnswerText will be nil/empty by default.
			// CreateStudentAnswer will insert a new record.
		}

		if err := r.CreateStudentAnswer(ctx, &saToUpsert); err != nil {
			return nil, fmt.Errorf("GradeQuizAttempt: failed to upsert student answer for question %d, attempt %d: %w", question.ID, attempt.ID, err)
		}
	}

	// 6. Update the QuizAttempt record.
	attempt.Score = &totalScoreAchieved
	attempt.Status = models.QuizAttemptStatusGraded // Mark as Graded

	// EndTime should ideally be set when the student submits or when the attempt auto-concludes.
	// Only set it here if it's not already set (e.g., for an attempt that was never formally submitted but is being graded).
	if attempt.EndTime.IsZero() {
		attempt.EndTime = time.Now()
	}

	if err := r.UpdateQuizAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("GradeQuizAttempt: failed to update quiz attempt %d: %w", attempt.ID, err)
	}

	// 7. Return the updated QuizAttempt.
	return attempt, nil
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
	// totalAttempts, err := r.CountQuizAttemptsByQuiz(ctx, collegeID, quizID)
	// if err != nil {
	// 	return nil, fmt.Errorf("GetQuizStatistics: counting attempts: %w", err)
	// }

	// Query for completed attempts
	completedAttempts, err := r.FindCompletedQuizAttempts(ctx, collegeID, quizID)
	if err != nil {
		return nil, fmt.Errorf("GetQuizStatistics: finding completed attempts: %w", err)
	}

	// Calculate statistics
	stats := &models.QuizStatistics{
		QuizID:            quizID,
		TotalAttempts:     0,
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

// gradeMultipleChoice compares student's selected options against correct options.
// Handles both single and multiple correct options based on the correct options defined for the question.
// `studentAnswer` is the student's submitted answer for this question.
// `question` is the question model, primarily for getting `question.Points`.
// `correctAnswerOptions` is a slice containing *only* the AnswerOption models that are marked as correct for this question.
// Returns (isCorrect bool, pointsAwarded int).
func gradeMultipleChoice(studentAnswer *models.StudentAnswer, question *models.Question, correctAnswerOptions []*models.AnswerOption) (bool, int) {
	// Build a map of actual correct option IDs for the question for quick lookup.
	actualCorrectOptionIDsMap := make(map[int]bool)
	for _, opt := range correctAnswerOptions {
		// No need to check opt.IsCorrect here, as correctAnswerOptions should only contain correct ones.
		actualCorrectOptionIDsMap[opt.ID] = true
	}

	// If no correct options are defined for the question in the database, it's ungradable automatically.
	if len(actualCorrectOptionIDsMap) == 0 {
		return false, 0
	}

	// Get student's selected option IDs.
	// studentAnswer.SelectedOptionID is *[]int (a pointer to a slice of ints).
	var studentSelectedIDs []int
	if studentAnswer != nil && studentAnswer.SelectedOptionID != nil && *studentAnswer.SelectedOptionID != nil {
		studentSelectedIDs = *studentAnswer.SelectedOptionID // Dereference the pointer to get the slice.
	}

	// Build a map of student's selected option IDs for efficient comparison.
	studentSelectedIDsMap := make(map[int]bool)
	for _, id := range studentSelectedIDs {
		studentSelectedIDsMap[id] = true
	}

	// To be considered correct, the student must select *all* the correct options and *no* incorrect options.
	// This means the set of student's selected IDs must be identical to the set of actual correct option IDs.
	if len(studentSelectedIDsMap) != len(actualCorrectOptionIDsMap) {
		return false, 0 // Incorrect number of options selected.
	}

	for id := range studentSelectedIDsMap { // Check if every option selected by student is in the set of correct options.
		if !actualCorrectOptionIDsMap[id] { // If a student-selected ID is not in the map of correct IDs.
			return false, 0 // Student selected an option that is not actually correct.
		}
	}

	// If all checks pass, the answer is correct.
	return true, question.Points
}

// GetQuestionWithAnswerOptions retrieves a question with all its answer options
func (r *quizRepository) GetQuestionWithAnswerOptions(ctx context.Context, collegeID int, questionID int) (*models.QuestionWithOptions, error) {
	// First get the question
	question, err := r.GetQuestionByID(ctx, collegeID, questionID)
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

func (r *quizRepository) GetQuestionWithCorrectAnswers(ctx context.Context, collegeID, questionID int) (*models.QuestionWithCorrectAnswers, error) {
	question, err := r.GetQuestionByID(ctx, collegeID, questionID)
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

func (r *quizRepository) FindQuizAttemptByStudentAndQuiz(ctx context.Context, collegeID int, studentID int, quizID int) (*models.QuizAttempt, error) {
	attempt := &models.QuizAttempt{}
	query := r.DB.SQ.Select("id", "student_id", "quiz_id", "college_id",
		"start_time", "end_time", "score", "status").
		From(quizAttemptTable).
		Where(squirrel.Eq{
			"college_id": collegeID,
			"student_id": studentID,
			"quiz_id":    quizID,
		})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	err = pgxscan.Get(ctx, r.DB.Pool, attempt, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No attempt exists
		}
		return nil, fmt.Errorf("query execution: %w", err)
	}

	return attempt, nil
}
