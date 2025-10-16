package quiz

import (
	"context"
	"fmt"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

// QuizAttemptService defines the interface for quiz attempt orchestration and grading operations.
// It handles the lifecycle of quiz attempts, including starting, submitting, and grading
// with proper college-based authorization and business logic validation.
type QuizAttemptService interface {
	// StartQuizAttempt creates a new quiz attempt for a student.
	// Validates that the student hasn't already attempted the quiz and sets initial state.
	StartQuizAttempt(ctx context.Context, collegeID int, attempt *models.QuizAttempt) error

	// GetQuizAttemptByID retrieves a quiz attempt by ID with college isolation.
	GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)

	// SubmitQuizAttempt marks a quiz attempt as completed and calculates the final score.
	// Aggregates scores from all student answers for the attempt.
	SubmitQuizAttempt(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)

	// GradeQuizAttempt manually grades a completed quiz attempt with a specific score.
	// Validates that the attempt is in an appropriate state for grading.
	GradeQuizAttempt(ctx context.Context, collegeID int, attemptID int, score int) (*models.QuizAttempt, error)

	// FindQuizAttemptsByStudent retrieves quiz attempts for a specific student with pagination.
	FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error)

	// FindQuizAttemptsByQuiz retrieves quiz attempts for a specific quiz with pagination.
	FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error)

	// CountQuizAttemptsByQuiz returns the total number of attempts for a quiz.
	CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)
}

// quizAttemptService implements the QuizAttemptService interface.
type quizAttemptService struct {
	quizAttemptRepo repository.QuizAttemptRepository
	studentAnswerRepo repository.StudentAnswerRepository
	quizRepo         repository.QuizRepository
	collegeRepo      repository.CollegeRepository
	validate         *validator.Validate
}

// NewQuizAttemptService creates a new instance of QuizAttemptService with required dependencies.
// Initializes validator for input validation.
func NewQuizAttemptService(
	quizAttemptRepo repository.QuizAttemptRepository,
	studentAnswerRepo repository.StudentAnswerRepository,
	quizRepo repository.QuizRepository,
	collegeRepo repository.CollegeRepository,
) QuizAttemptService {
	return &quizAttemptService{
		quizAttemptRepo:  quizAttemptRepo,
		studentAnswerRepo: studentAnswerRepo,
		quizRepo:         quizRepo,
		collegeRepo:      collegeRepo,
		validate:         validator.New(),
	}
}

// StartQuizAttempt creates a new quiz attempt for a student.
// Validates that the student hasn't already attempted the quiz and sets initial state.
func (s *quizAttemptService) StartQuizAttempt(ctx context.Context, collegeID int, attempt *models.QuizAttempt) error {
	// Validate attempt struct
	if err := s.validate.Struct(attempt); err != nil {
		return fmt.Errorf("validation failed for quiz attempt: %w", err)
	}

	// Verify college exists
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return fmt.Errorf("college verification failed: %w", err)
	}

	// Check if student has already attempted this quiz
	existingAttempts, err := s.quizAttemptRepo.FindQuizAttemptsByStudent(ctx, collegeID, attempt.StudentID, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to check existing attempts: %w", err)
	}

	// Check if any existing attempt is for this quiz
	for _, existing := range existingAttempts {
		if existing.QuizID == attempt.QuizID {
			return fmt.Errorf("student has already attempted this quiz")
		}
	}

	// Verify quiz exists and belongs to the college
	quiz, err := s.quizRepo.GetQuizByID(ctx, collegeID, attempt.QuizID)
	if err != nil {
		return fmt.Errorf("failed to get quiz: %w", err)
	}
	if quiz == nil {
		return fmt.Errorf("quiz with ID %d not found in college %d", attempt.QuizID, collegeID)
	}

	// Set attempt properties
	attempt.CollegeID = collegeID
	attempt.CourseID = quiz.CourseID
	attempt.StartTime = time.Now()
	attempt.Status = models.QuizAttemptStatusInProgress

	return s.quizAttemptRepo.CreateQuizAttempt(ctx, attempt)
}

// GetQuizAttemptByID retrieves a quiz attempt by ID with college isolation.
func (s *quizAttemptService) GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error) {
	return s.quizAttemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
}

// SubmitQuizAttempt marks a quiz attempt as completed and calculates the final score.
// Aggregates scores from all student answers for the attempt.
func (s *quizAttemptService) SubmitQuizAttempt(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error) {
	// Get the attempt
	attempt, err := s.quizAttemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz attempt: %w", err)
	}
	if attempt == nil {
		return nil, fmt.Errorf("quiz attempt with ID %d not found", attemptID)
	}

	// Check if attempt is in progress
	if attempt.Status != models.QuizAttemptStatusInProgress {
		return nil, fmt.Errorf("quiz attempt is not in progress, current status: %s", attempt.Status)
	}

	// Get all answers for this attempt
	answers, err := s.studentAnswerRepo.FindStudentAnswersByAttempt(ctx, collegeID, attemptID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get student answers: %w", err)
	}

	// Calculate total score
	var totalScore int
	for _, answer := range answers {
		if answer.PointsAwarded != nil {
			totalScore += *answer.PointsAwarded
		}
	}

	// Update attempt
	attempt.EndTime = time.Now()
	attempt.Status = models.QuizAttemptStatusCompleted
	attempt.Score = &totalScore

	if err := s.quizAttemptRepo.UpdateQuizAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to update quiz attempt: %w", err)
	}

	return attempt, nil
}

// GradeQuizAttempt manually grades a completed quiz attempt with a specific score.
// Validates that the attempt is in an appropriate state for grading.
func (s *quizAttemptService) GradeQuizAttempt(ctx context.Context, collegeID int, attemptID int, score int) (*models.QuizAttempt, error) {
	// Get the attempt
	attempt, err := s.quizAttemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz attempt: %w", err)
	}
	if attempt == nil {
		return nil, fmt.Errorf("quiz attempt with ID %d not found", attemptID)
	}

	// Check if attempt can be graded
	if attempt.Status != models.QuizAttemptStatusCompleted && attempt.Status != models.QuizAttemptStatusGraded {
		return nil, fmt.Errorf("quiz attempt must be completed or already graded, current status: %s", attempt.Status)
	}

	// Update score and status
	attempt.Score = &score
	attempt.Status = models.QuizAttemptStatusGraded

	if err := s.quizAttemptRepo.UpdateQuizAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to update quiz attempt: %w", err)
	}

	return attempt, nil
}

// FindQuizAttemptsByStudent retrieves quiz attempts for a specific student with pagination.
func (s *quizAttemptService) FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
	return s.quizAttemptRepo.FindQuizAttemptsByStudent(ctx, collegeID, studentID, limit, offset)
}

// FindQuizAttemptsByQuiz retrieves quiz attempts for a specific quiz with pagination.
func (s *quizAttemptService) FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
	return s.quizAttemptRepo.FindQuizAttemptsByQuiz(ctx, collegeID, quizID, limit, offset)
}

// CountQuizAttemptsByQuiz returns the total number of attempts for a quiz.
// Used for pagination calculations.
func (s *quizAttemptService) CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
	// Verify college exists
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return 0, fmt.Errorf("college verification failed: %w", err)
	}

	// Verify quiz exists
	_, err = s.quizRepo.GetQuizByID(ctx, collegeID, quizID)
	if err != nil {
		return 0, fmt.Errorf("quiz verification failed: %w", err)
	}

	return s.quizAttemptRepo.CountQuizAttemptsByQuiz(ctx, collegeID, quizID)
}
