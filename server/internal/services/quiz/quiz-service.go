package quiz

import (
	"context"
	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// QuizService defines the interface for quiz-related business logic.
type QuizService interface {
	// Quiz Methods
	CreateQuiz(ctx context.Context, quiz *models.Quiz) error
	GetQuizByID(ctx context.Context, collegeID int, quizID int) (*models.Quiz, error)
	UpdateQuiz(ctx context.Context, quiz *models.Quiz) error
	DeleteQuiz(ctx context.Context, collegeID int, quizID int) error
	FindQuizzesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Quiz, error)
	CountQuizzesByCourse(ctx context.Context, collegeID int, courseID int) (int, error)

	// Question Methods
	CreateQuestion(ctx context.Context, question *models.Question) error
	GetQuestionByID(ctx context.Context, questionID int) (*models.Question, error)
	UpdateQuestion(ctx context.Context, question *models.Question) error
	DeleteQuestion(ctx context.Context, questionID int) error
	FindQuestionsByQuiz(ctx context.Context, quizID int, limit, offset uint64, withOptions bool) ([]*models.Question, error)
	CountQuestionsByQuiz(ctx context.Context, quizID int) (int, error)

	// AnswerOption Methods
	CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error
	GetAnswerOptionByID(ctx context.Context, optionID int) (*models.AnswerOption, error)
	UpdateAnswerOption(ctx context.Context, option *models.AnswerOption) error
	DeleteAnswerOption(ctx context.Context, optionID int) error
	FindAnswerOptionsByQuestion(ctx context.Context, questionID int) ([]*models.AnswerOption, error)

	// QuizAttempt Methods
	StartQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error
	GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)
	SubmitQuizAttempt(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)
	GradeQuizAttempt(ctx context.Context, collegeID int, attemptID int, score int) (*models.QuizAttempt, error)
	FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error)
	FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error)
	CountQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int) (int, error)
	CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)

	// StudentAnswer Methods
	SubmitStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error
	GradeStudentAnswer(ctx context.Context, answerID int, isCorrect *bool, pointsAwarded *int) (*models.StudentAnswer, error)
	FindStudentAnswersByAttempt(ctx context.Context, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error)
	GetStudentAnswerForQuestion(ctx context.Context, attemptID int, questionID int) (*models.StudentAnswer, error)
}

type quizService struct {
	quizRepo repository.QuizRepository
	// For more complex business logic, you might inject other repositories or services:
	courseRepo     repository.CourseRepository
	enrollmentRepo repository.EnrollmentRepository
	validate       *validator.Validate
}

// NewQuizService creates a new QuizService.
func NewQuizService(quizRepo repository.QuizRepository, courseRepo repository.CourseRepository, enrollmentRepo repository.EnrollmentRepository) QuizService {
	return &quizService{
		quizRepo:       quizRepo,
		courseRepo:     courseRepo,
		enrollmentRepo: enrollmentRepo,
		validate:       validator.New(),
	}
}

// --- Quiz Methods ---

func (s *quizService) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	if err := s.validate.Struct(quiz); err != nil {
		return fmt.Errorf("validation failed for quiz: %w", err)
	}
	// Business logic: e.g., check if quiz.CourseID exists via courseRepo if injected.
	// check course exists or not
	course, err := s.courseRepo.FindCourseByID(ctx, quiz.CollegeID, quiz.CourseID)
	if course != nil {
		return s.quizRepo.CreateQuiz(ctx, quiz)
	}
	return err

}

func (s *quizService) GetQuizByID(ctx context.Context, collegeID int, quizID int) (*models.Quiz, error) {
	return s.quizRepo.GetQuizByID(ctx, collegeID, quizID)
}

func (s *quizService) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	if err := s.validate.Struct(quiz); err != nil {
		return fmt.Errorf("validation failed for quiz: %w", err)
	}
	if quiz.ID == 0 {
		return fmt.Errorf("quiz ID is required for update")
	}
	return s.quizRepo.UpdateQuiz(ctx, quiz)
}

func (s *quizService) DeleteQuiz(ctx context.Context, collegeID int, quizID int) error {
	// Business logic: Consider if there are active attempts or if questions should be cascade deleted.
	// For now, direct deletion. The repository handles deleting the quiz itself.
	// If questions/options need to be deleted, fetch them first and delete them.

	return s.quizRepo.DeleteQuiz(ctx, collegeID, quizID)
}

func (s *quizService) FindQuizzesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Quiz, error) {
	return s.quizRepo.FindQuizzesByCourse(ctx, collegeID, courseID, limit, offset)
}

func (s *quizService) CountQuizzesByCourse(ctx context.Context, collegeID int, courseID int) (int, error) {
	return s.quizRepo.CountQuizzesByCourse(ctx, collegeID, courseID)
}

// --- Question Methods ---

func (s *quizService) CreateQuestion(ctx context.Context, question *models.Question) error {
	if err := s.validate.Struct(question); err != nil {
		return fmt.Errorf("validation failed for question: %w", err)
	}

	return s.quizRepo.CreateQuestion(ctx, question)
}

func (s *quizService) GetQuestionByID(ctx context.Context, questionID int) (*models.Question, error) {
	return s.quizRepo.GetQuestionByID(ctx, questionID)
}

func (s *quizService) UpdateQuestion(ctx context.Context, question *models.Question) error {
	if err := s.validate.Struct(question); err != nil {
		return fmt.Errorf("validation failed for question: %w", err)
	}
	if question.ID == 0 {
		return fmt.Errorf("question ID is required for update")
	}
	return s.quizRepo.UpdateQuestion(ctx, question)
}

func (s *quizService) DeleteQuestion(ctx context.Context, questionID int) error {
	// Business logic: Delete associated answer options first.
	options, err := s.quizRepo.FindAnswerOptionsByQuestion(ctx, questionID)
	if err == nil && options != nil { // If no error and options exist
		for _, opt := range options {
			if delErr := s.quizRepo.DeleteAnswerOption(ctx, opt.ID); delErr != nil {
				// Log or handle error deleting option, but attempt to continue
				fmt.Printf("Warning: failed to delete answer option ID %d for question ID %d: %v\n", opt.ID, questionID, delErr)
			}
		}
	} else if err != nil {
		fmt.Printf("Warning: failed to find answer options for question ID %d before deletion: %v\n", questionID, err)
	}
	return s.quizRepo.DeleteQuestion(ctx, questionID)
}

func (s *quizService) FindQuestionsByQuiz(ctx context.Context, quizID int, limit, offset uint64, withOptions bool) ([]*models.Question, error) {
	questions, err := s.quizRepo.FindQuestionsByQuiz(ctx, quizID, limit, offset)
	if err != nil {
		return nil, err
	}
	if withOptions {
		for _, q := range questions {
			options, err := s.quizRepo.FindAnswerOptionsByQuestion(ctx, q.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch options for question ID %d: %w", q.ID, err)
			}
			q.Options = options
		}
	}
	return questions, nil
}

func (s *quizService) CountQuestionsByQuiz(ctx context.Context, quizID int) (int, error) {
	return s.quizRepo.CountQuestionsByQuiz(ctx, quizID)
}

// --- AnswerOption Methods ---

func (s *quizService) CreateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	if err := s.validate.Struct(option); err != nil {
		return fmt.Errorf("validation failed for answer option: %w", err)
	}
	questionID, err := s.quizRepo.GetQuestionByID(ctx, option.QuestionID)
	if err != nil {
		return fmt.Errorf("invalid questionID", questionID)
	}
	return s.quizRepo.CreateAnswerOption(ctx, option)

}

func (s *quizService) GetAnswerOptionByID(ctx context.Context, optionID int) (*models.AnswerOption, error) {
	return s.quizRepo.GetAnswerOptionByID(ctx, optionID)
}

func (s *quizService) UpdateAnswerOption(ctx context.Context, option *models.AnswerOption) error {
	if err := s.validate.Struct(option); err != nil {
		return fmt.Errorf("validation failed for answer option: %w", err)
	}
	if option.ID == 0 {
		return fmt.Errorf("answer option ID is required for update")
	}
	return s.quizRepo.UpdateAnswerOption(ctx, option)
}

func (s *quizService) DeleteAnswerOption(ctx context.Context, optionID int) error {
	// check if answer actually exists or not
	_, err := s.quizRepo.GetAnswerOptionByID(ctx, optionID)
	if err != nil {
		return fmt.Errorf("error getting option ID", err)
	}
	return s.quizRepo.DeleteAnswerOption(ctx, optionID)

}

func (s *quizService) FindAnswerOptionsByQuestion(ctx context.Context, questionID int) ([]*models.AnswerOption, error) {
	// check if question id exists or not
	_, err := s.quizRepo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	return s.quizRepo.FindAnswerOptionsByQuestion(ctx, questionID)
}

// --- QuizAttempt Methods ---

func (s *quizService) StartQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error {
	if err := s.validate.Struct(attempt); err != nil { // Basic validation for IDs
		return fmt.Errorf("validation failed for quiz attempt: %w", err)
	}

	quiz, err := s.quizRepo.GetQuizByID(ctx, attempt.CollegeID, attempt.QuizID)
	if err != nil {
		return fmt.Errorf("failed to get quiz ID %d for attempt: %w", attempt.QuizID, err)
	}
	if quiz == nil {
		return fmt.Errorf("quiz with ID %d not found", attempt.QuizID)
	}
	if !quiz.DueDate.IsZero() && time.Now().After(quiz.DueDate) {
		return fmt.Errorf("quiz ID %d is past its due date (%v)", attempt.QuizID, quiz.DueDate)
	}

	// Further business logic: check student enrollment, existing attempts, etc.
	ok, err := s.enrollmentRepo.IsStudentEnrolled(ctx, attempt.CollegeID, attempt.StudentID, attempt.CourseID)

	if ok {

		attempt.StartTime = time.Now()
		attempt.Status = models.QuizAttemptStatusInProgress
		return s.quizRepo.CreateQuizAttempt(ctx, attempt)
	}
	return err
}
func (s *quizService) GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error) {
	return s.quizRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
}

func (s *quizService) SubmitQuizAttempt(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error) {
	attempt, err := s.quizRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz attempt ID %d: %w", attemptID, err)
	}
	if attempt == nil {
		return nil, fmt.Errorf("quiz attempt with ID %d not found", attemptID)
	}

	if attempt.Status != models.QuizAttemptStatusInProgress {
		return nil, fmt.Errorf("quiz attempt ID %d is not in progress, current status: %s", attemptID, attempt.Status)
	}

	attempt.EndTime = time.Now()
	attempt.Status = models.QuizAttemptStatusCompleted

	if err := s.quizRepo.UpdateQuizAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to update quiz attempt ID %d on submission: %w", attemptID, err)
	}
	// use autograding here
	//
	return attempt, nil
}

func (s *quizService) GradeQuizAttempt(ctx context.Context, collegeID int, attemptID int, score int) (*models.QuizAttempt, error) {
	attempt, err := s.quizRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz attempt ID %d for grading: %w", attemptID, err)
	}
	if attempt == nil {
		return nil, fmt.Errorf("quiz attempt with ID %d not found for grading", attemptID)
	}

	if attempt.Status != models.QuizAttemptStatusCompleted && attempt.Status != models.QuizAttemptStatusGraded {
		return nil, fmt.Errorf("quiz attempt ID %d must be completed or already graded to update grade, current status: %s", attemptID, attempt.Status)
	}
	// Business logic: validate score against quiz's max possible score.

	attempt.Score = &score
	attempt.Status = models.QuizAttemptStatusGraded

	if err := s.quizRepo.UpdateQuizAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to update quiz attempt ID %d with grade: %w", attemptID, err)
	}
	return attempt, nil
}

func (s *quizService) FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
	return s.quizRepo.FindQuizAttemptsByStudent(ctx, collegeID, studentID, limit, offset)
}

func (s *quizService) FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
	return s.quizRepo.FindQuizAttemptsByQuiz(ctx, collegeID, quizID, limit, offset)
}

func (s *quizService) CountQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int) (int, error) {
	return s.quizRepo.CountQuizAttemptsByStudent(ctx, collegeID, studentID)
}

func (s *quizService) CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
	return s.quizRepo.CountQuizAttemptsByQuiz(ctx, collegeID, quizID)
}

// --- StudentAnswer Methods ---

func (s *quizService) SubmitStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error {
	if err := s.validate.Struct(answer); err != nil { // Basic validation for IDs
		return fmt.Errorf("validation failed for student answer: %w", err)
	}
	// Business logic: Check if the quiz attempt is still in progress.
	// This would require fetching the attempt, which needs CollegeID.
	// For simplicity, we rely on the repository's upsert.
	return s.quizRepo.CreateStudentAnswer(ctx, answer)
}

func (s *quizService) GradeStudentAnswer(ctx context.Context, answerID int, isCorrect *bool, pointsAwarded *int) (*models.StudentAnswer, error) {
	// For full functionality, this requires GetStudentAnswerByID in the QuizRepository.
	// Assuming such a method exists or will be added:

	sa, err := s.quizRepo.GetStudentAnswerByID(ctx, answerID) // Assumed method
	if err != nil {
		return nil, fmt.Errorf("could not retrieve student answer %d for grading: %w", answerID, err)
	}
	if sa == nil {
		return nil, fmt.Errorf("student answer with ID %d not found for grading", answerID)
	}

	sa.IsCorrect = isCorrect
	sa.PointsAwarded = pointsAwarded

	if err := s.quizRepo.UpdateStudentAnswer(ctx, sa); err != nil {
		return nil, fmt.Errorf("failed to update grade for student answer ID %d: %w", answerID, err)
	}
	return sa, nil

}

func (s *quizService) FindStudentAnswersByAttempt(ctx context.Context, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error) {
	return s.quizRepo.FindStudentAnswersByAttempt(ctx, attemptID, limit, offset)
}

func (s *quizService) GetStudentAnswerForQuestion(ctx context.Context, attemptID int, questionID int) (*models.StudentAnswer, error) {
	return s.quizRepo.GetStudentAnswerForQuestion(ctx, attemptID, questionID)
}
