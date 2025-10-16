package quiz

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

// StudentAnswerService defines the interface for student answer submission and validation operations.
// It handles CRUD operations for student answers with proper college-based authorization
// and business logic validation for quiz attempts.
type StudentAnswerService interface {
	// SubmitStudentAnswer creates or updates a student's answer for a specific question.
	// Uses UPSERT to handle duplicate submissions for the same question in the same attempt.
	SubmitStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error

	// GradeStudentAnswer updates the correctness and points awarded for a student answer.
	// Validates that the answer exists within the college context.
	GradeStudentAnswer(ctx context.Context, collegeID int, answerID int, isCorrect *bool, pointsAwarded *int) (*models.StudentAnswer, error)

	// FindStudentAnswersByAttempt retrieves student answers for a specific quiz attempt with pagination.
	FindStudentAnswersByAttempt(ctx context.Context, collegeID int, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error)

	// GetStudentAnswerForQuestion retrieves a student's answer for a specific question in an attempt.
	GetStudentAnswerForQuestion(ctx context.Context, collegeID int, attemptID int, questionID int) (*models.StudentAnswer, error)

	// GetStudentAnswerByID retrieves a student answer by ID with college isolation.
	GetStudentAnswerByID(ctx context.Context, collegeID int, answerID int) (*models.StudentAnswer, error)
}

// studentAnswerService implements the StudentAnswerService interface.
type studentAnswerService struct {
	studentAnswerRepo repository.StudentAnswerRepository
	quizAttemptRepo   repository.QuizAttemptRepository
	questionRepo      repository.QuestionRepository
	collegeRepo       repository.CollegeRepository
	validate          *validator.Validate
}

// NewStudentAnswerService creates a new instance of StudentAnswerService with required dependencies.
// Initializes validator for input validation.
func NewStudentAnswerService(
	studentAnswerRepo repository.StudentAnswerRepository,
	quizAttemptRepo repository.QuizAttemptRepository,
	questionRepo repository.QuestionRepository,
	collegeRepo repository.CollegeRepository,
) StudentAnswerService {
	return &studentAnswerService{
		studentAnswerRepo: studentAnswerRepo,
		quizAttemptRepo:   quizAttemptRepo,
		questionRepo:      questionRepo,
		collegeRepo:       collegeRepo,
		validate:          validator.New(),
	}
}

// SubmitStudentAnswer creates or updates a student's answer for a specific question.
// Uses UPSERT to handle duplicate submissions for the same question in the same attempt.
func (s *studentAnswerService) SubmitStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error {
	// Validate answer struct
	if err := s.validate.Struct(answer); err != nil {
		return fmt.Errorf("validation failed for student answer: %w", err)
	}

	// Basic validation for required fields
	if answer.QuizAttemptID == 0 {
		return fmt.Errorf("quiz attempt ID is required")
	}
	if answer.QuestionID == 0 {
		return fmt.Errorf("question ID is required")
	}

	// Business logic validation: Check if the quiz attempt is still in progress
	attempt, err := s.quizAttemptRepo.GetQuizAttemptByID(ctx, 0, answer.QuizAttemptID) // college ID will be validated below
	if err != nil {
		return fmt.Errorf("failed to get quiz attempt: %w", err)
	}
	if attempt == nil {
		return fmt.Errorf("quiz attempt with ID %d not found", answer.QuizAttemptID)
	}
	if attempt.Status != models.QuizAttemptStatusInProgress {
		return fmt.Errorf("cannot submit answer: quiz attempt is %s", attempt.Status)
	}

	// Business logic validation: Verify the question belongs to the quiz being attempted
	question, err := s.questionRepo.GetQuestionByID(ctx, attempt.CollegeID, answer.QuestionID)
	if err != nil {
		return fmt.Errorf("failed to validate question: %w", err)
	}
	if question == nil || question.QuizID != attempt.QuizID {
		return fmt.Errorf("question with ID %d does not belong to attempt's quiz", answer.QuestionID)
	}

	// Business logic validation: Validate answer format based on question type
	switch question.Type {
	case models.MultipleChoice, models.TrueFalse:
		if answer.SelectedOptionID == nil || len(*answer.SelectedOptionID) == 0 {
			return fmt.Errorf("answer option selection is required for %s questions", question.Type)
		}
		if answer.AnswerText != "" {
			return fmt.Errorf("text answers not allowed for %s questions", question.Type)
		}
	case models.ShortAnswer:
		if answer.AnswerText == "" {
			return fmt.Errorf("text answer is required for %s questions", question.Type)
		}
		if answer.SelectedOptionID != nil && len(*answer.SelectedOptionID) > 0 {
			return fmt.Errorf("selected options not allowed for %s questions", question.Type)
		}
	}

	return s.studentAnswerRepo.CreateStudentAnswer(ctx, answer)
}

// GradeStudentAnswer updates the correctness and points awarded for a student answer.
// Validates that the answer exists within the college context.
func (s *studentAnswerService) GradeStudentAnswer(ctx context.Context, collegeID int, answerID int, isCorrect *bool, pointsAwarded *int) (*models.StudentAnswer, error) {
	// Verify college exists
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("college verification failed: %w", err)
	}

	// Retrieve the student answer
	answer, err := s.studentAnswerRepo.GetStudentAnswerByID(ctx, collegeID, answerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve student answer: %w", err)
	}
	if answer == nil {
		return nil, fmt.Errorf("student answer with ID %d not found in college %d", answerID, collegeID)
	}

	// Update the answer
	answer.IsCorrect = isCorrect
	answer.PointsAwarded = pointsAwarded

	if err := s.studentAnswerRepo.UpdateStudentAnswer(ctx, collegeID, answer); err != nil {
		return nil, fmt.Errorf("failed to update student answer: %w", err)
	}

	return answer, nil
}

// FindStudentAnswersByAttempt retrieves student answers for a specific quiz attempt with pagination.
// Ensures the attempt belongs to the college context.
func (s *studentAnswerService) FindStudentAnswersByAttempt(ctx context.Context, collegeID int, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error) {
	// Verify attempt exists and belongs to the college
	_, err := s.quizAttemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("quiz attempt verification failed: %w", err)
	}

	return s.studentAnswerRepo.FindStudentAnswersByAttempt(ctx, collegeID, attemptID, limit, offset)
}

// GetStudentAnswerForQuestion retrieves a student's answer for a specific question in an attempt.
// Ensures both the attempt and question belong to the college context.
func (s *studentAnswerService) GetStudentAnswerForQuestion(ctx context.Context, collegeID int, attemptID int, questionID int) (*models.StudentAnswer, error) {
	// Verify attempt exists and belongs to the college
	_, err := s.quizAttemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("quiz attempt verification failed: %w", err)
	}

	// Verify question exists and belongs to the college
	_, err = s.questionRepo.GetQuestionByID(ctx, collegeID, questionID)
	if err != nil {
		return nil, fmt.Errorf("question verification failed: %w", err)
	}

	return s.studentAnswerRepo.GetStudentAnswerForQuestion(ctx, collegeID, attemptID, questionID)
}

// GetStudentAnswerByID retrieves a student answer by ID with college isolation.
func (s *studentAnswerService) GetStudentAnswerByID(ctx context.Context, collegeID int, answerID int) (*models.StudentAnswer, error) {
	return s.studentAnswerRepo.GetStudentAnswerByID(ctx, collegeID, answerID)
}
