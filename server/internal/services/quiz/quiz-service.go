package quiz

import (
	"context"
	"fmt"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

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
	CreateQuestion(ctx context.Context, collegeID int, question *models.Question) error
	GetQuestionByID(ctx context.Context, collegeID int, questionID int) (*models.Question, error)
	UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error
	DeleteQuestion(ctx context.Context, collegeID int, questionID int) error
	FindQuestionsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64, withOptions bool) ([]*models.Question, error)
	CountQuestionsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)

	// AnswerOption Methods
	CreateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error
	GetAnswerOptionByID(ctx context.Context, collegeID int, optionID int) (*models.AnswerOption, error)
	UpdateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error
	DeleteAnswerOption(ctx context.Context, collegeID int, optionID int) error
	FindAnswerOptionsByQuestion(ctx context.Context, collegeID, questionID int) ([]*models.AnswerOption, error)
	// Note: FindAnswerOptionsByQuestion relies on questionID being valid within the college context,
	// which should be ensured by the caller (e.g., after fetching the question via GetQuestionByID).
	// QuizAttempt Methods
	StartQuizAttempt(ctx context.Context, collegeID int, attempt *models.QuizAttempt) error
	GetQuizAttemptByID(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)
	SubmitQuizAttempt(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)
	GradeQuizAttempt(ctx context.Context, collegeID int, attemptID int, score int) (*models.QuizAttempt, error)
	FindQuizAttemptsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.QuizAttempt, error)
	// FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error)
	// CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)

	// StudentAnswer Methods
	SubmitStudentAnswer(ctx context.Context, answer *models.StudentAnswer) error
	GradeStudentAnswer(ctx context.Context, collegeID int, answerID int, isCorrect *bool, pointsAwarded *int) (*models.StudentAnswer, error)
	FindStudentAnswersByAttempt(ctx context.Context, collegeID int, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error)
	GetStudentAnswerForQuestion(ctx context.Context, collegeID int, attemptID int, questionID int) (*models.StudentAnswer, error)
}

type quizService struct {
	quizRepo repository.QuizRepository
	// For more complex business logic, you might inject other repositories or services:
	courseRepo     repository.CourseRepository
	collegeRepo    repository.CollegeRepository
	enrollmentRepo repository.EnrollmentRepository
	validate       *validator.Validate
}

// NewQuizService creates a new QuizService.
func NewQuizService(quizRepo repository.QuizRepository, courseRepo repository.CourseRepository, collegeRepo repository.CollegeRepository, enrollmentRepo repository.EnrollmentRepository) QuizService {
	return &quizService{
		quizRepo:       quizRepo,
		courseRepo:     courseRepo,
		collegeRepo:    collegeRepo,
		enrollmentRepo: enrollmentRepo,
		validate:       validator.New(),
	}
}

// --- Quiz Methods ---

func (s *quizService) CreateQuiz(ctx context.Context, quiz *models.Quiz) error {
	if err := s.validate.Struct(quiz); err != nil {
		return fmt.Errorf("validation failed for quiz: %w", err)
	}

	// check if quiz.CollegeID exists via collegeRepo
	_, appErr := s.collegeRepo.GetCollegeByID(ctx, quiz.CollegeID)
	if appErr == nil {
		course, err := s.courseRepo.FindCourseByID(ctx, quiz.CollegeID, quiz.CourseID)
		if course == nil {
			return err
		}
		if course != nil {
			return s.quizRepo.CreateQuiz(ctx, quiz)
		}
		if err != nil {
			return err
		}
	}
	// Business logic: e.g., check if quiz.CourseID exists via courseRepo if injected.
	// check course exists or not

	return nil
}

func (s *quizService) GetQuizByID(ctx context.Context, collegeID int, quizID int) (*models.Quiz, error) {
	return s.quizRepo.GetQuizByID(ctx, collegeID, quizID)
}

func (s *quizService) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	if quiz == nil {
		return fmt.Errorf("quiz is nil")
	}
	if err := s.validate.Struct(quiz); err != nil {
		return fmt.Errorf("validation failed for quiz: %w", err)
	}
	if quiz.ID == 0 {
		return fmt.Errorf("quiz ID is required for update")
	}

	if err := s.quizRepo.UpdateQuiz(ctx, quiz); err != nil {
		return fmt.Errorf("failed to update quiz: %w", err)
	}
	return nil
}

func (s *quizService) DeleteQuiz(ctx context.Context, collegeID int, quizID int) error {
	// Check for active quiz attempts
	attempts, err := s.quizRepo.FindCompletedQuizAttempts(ctx, collegeID, quizID)
	if err != nil {
		return fmt.Errorf("DeleteQuiz: failed to check quiz attempts: %w", err)
	}
	if len(attempts) > 0 {
		return fmt.Errorf("DeleteQuiz: cannot delete quiz with active attempts")
	}

	// Fetch and delete questions associated with the quiz
	questions, err := s.quizRepo.FindQuestionsByQuiz(ctx, collegeID, quizID, 0, 0)
	if err != nil {
		return fmt.Errorf("DeleteQuiz: failed to fetch questions: %w", err)
	}

	for _, question := range questions {
		if err := s.quizRepo.DeleteQuestion(ctx, collegeID, question.ID); err != nil {
			return fmt.Errorf("DeleteQuiz: failed to delete question (id: %d): %w", question.ID, err)
		}
	}

	// Delete the quiz
	if err := s.quizRepo.DeleteQuiz(ctx, collegeID, quizID); err != nil {
		return fmt.Errorf("DeleteQuiz: failed to delete quiz: %w", err)
	}

	return nil
}

func (s *quizService) FindQuizzesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Quiz, error) {
	// Validate input parameters
	if collegeID <= 0 {
		return nil, fmt.Errorf("invalid college ID")
	}
	if courseID <= 0 {
		return nil, fmt.Errorf("invalid course ID")
	}

	// Verify college exists
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("college verification failed: %w", err)
	}

	// Verify course exists and belongs to college
	course, err := s.courseRepo.FindCourseByID(ctx, collegeID, courseID)
	if err != nil {
		return nil, fmt.Errorf("course verification failed: %w", err)
	}
	if course == nil {
		return nil, fmt.Errorf("course not found in college")
	}

	// Apply reasonable limits
	if limit > 100 {
		limit = 100 // Prevent excessive queries
	}

	return s.quizRepo.FindQuizzesByCourse(ctx, collegeID, courseID, limit, offset)
}

func (s *quizService) CountQuizzesByCourse(ctx context.Context, collegeID int, courseID int) (int, error) {
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return 0, fmt.Errorf("college verfication failed %w", err)
	}
	course, err := s.courseRepo.FindCourseByID(ctx, collegeID, courseID)
	if course == nil {
		return 0, fmt.Errorf("no course found with course ID %d", courseID)
	}
	if err != nil {
		return 0, fmt.Errorf("course verification failed %w", err)
	}

	return s.quizRepo.CountQuizzesByCourse(ctx, collegeID, courseID)
}

// --- Question Methods ---

func (s *quizService) CreateQuestion(ctx context.Context, collegeID int, question *models.Question) error {

	if err := s.validate.Struct(question); err != nil {
		return fmt.Errorf("validation failed for question: %w", err)
	}
	// check if quiz exists
	quiz, err := s.quizRepo.GetQuizByID(ctx, collegeID, question.QuizID)
	if quiz == nil {
		return fmt.Errorf("Cannot Create Question into empty quiz ")
	}
	if err != nil {
		return fmt.Errorf("failed to get quiz ")
	}
	_, err = s.courseRepo.FindCourseByID(ctx, collegeID, quiz.CourseID)
	if err != nil {
		return fmt.Errorf("failed to find course by ID %d", quiz.CourseID)
	}
	// check if course exists or not

	return s.quizRepo.CreateQuestion(ctx, question)
}

func (s *quizService) GetQuestionByID(ctx context.Context, collegeID int, questionID int) (*models.Question, error) {
	// The repository method now handles the college scope check.
	return s.quizRepo.GetQuestionByID(ctx, collegeID, questionID)
}

func (s *quizService) UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	if err := s.validate.Struct(question); err != nil {
		return fmt.Errorf("validation failed for question: %w", err)
	}
	_, collegeErr := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if collegeErr != nil {
		return fmt.Errorf("failed to get college with ID %d", collegeID)
	}
	if question.ID == 0 {
		return fmt.Errorf("question ID is required for update")
	}
	// check if quiz exists or not
	quiz, err := s.quizRepo.GetQuizByID(ctx, collegeID, question.QuizID)
	if err != nil {
		return fmt.Errorf("failed to fetch quiz with ID %d", question.QuizID)
	}
	if quiz == nil {
		return fmt.Errorf("cannot update question to non existent quiz")
	}

	// The repository method now handles the college scope check using collegeID and question.ID.
	return s.quizRepo.UpdateQuestion(ctx, collegeID, question)
}

func (s *quizService) DeleteQuestion(ctx context.Context, collegeID int, questionID int) error {
	// Before deleting options or the question, verify the question exists and belongs to the college.
	// The repository's GetQuestionByID now includes the college check.
	question, err := s.quizRepo.GetQuestionByID(ctx, collegeID, questionID)
	if err != nil {
		// This error already includes "not found" or other repo errors, scoped by college.
		return fmt.Errorf("cannot delete question %d: %w", questionID, err)
	}

	// Now find and delete options. FindAnswerOptionsByQuestion is not college-scoped itself,
	// but we've already validated the parent questionID belongs to the college.
	options, err := s.quizRepo.FindAnswerOptionsByQuestion(ctx, question.ID) // Use question.ID for clarity
	if err == nil && options != nil {                                        // If no error and options exist
		for _, opt := range options {
			// Delete each option, ensuring it's also scoped by college.
			// The repository's DeleteAnswerOption should handle this.
			if delErr := s.quizRepo.DeleteAnswerOption(ctx, collegeID, opt.ID); delErr != nil {
				// Log or handle error deleting option, but attempt to continue
				fmt.Printf("Warning: failed to delete answer option ID %d for question ID %d: %v\n", opt.ID, questionID, delErr)
			}
		}
	} else if err != nil {
		// Handle error finding options, but still attempt to delete the question.
		fmt.Printf("Warning: failed to find answer options for question ID %d before deletion: %v\n", questionID, err)
	}

	// Finally, delete the question itself, scoped by college.
	return s.quizRepo.DeleteQuestion(ctx, collegeID, questionID)
}

func (s *quizService) FindQuestionsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64, withOptions bool) ([]*models.Question, error) {
	questions, err := s.quizRepo.FindQuestionsByQuiz(ctx, collegeID, quizID, limit, offset)
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

func (s *quizService) CountQuestionsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
	return s.quizRepo.CountQuestionsByQuiz(ctx, collegeID, quizID)
}

// --- AnswerOption Methods ---

func (s *quizService) CreateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error {
	if err := s.validate.Struct(option); err != nil {
		return fmt.Errorf("validation failed for answer option: %w", err)
	}
	questionID, err := s.quizRepo.GetQuestionByID(ctx, collegeID, option.QuestionID)
	if err != nil {
		return fmt.Errorf("invalid questionID", questionID)
	}
	return s.quizRepo.CreateAnswerOption(ctx, option)
}

func (s *quizService) GetAnswerOptionByID(ctx context.Context, collegeID int, optionID int) (*models.AnswerOption, error) {
	// The repository method now handles the college scope check.
	return s.quizRepo.GetAnswerOptionByID(ctx, collegeID, optionID)
}

func (s *quizService) UpdateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error {
	if err := s.validate.Struct(option); err != nil {
		return fmt.Errorf("validation failed for answer option: %w", err)
	}
	if option.ID == 0 {
		return fmt.Errorf("answer option ID is required for update")
	}
	// The repository method now handles the college scope check using collegeID and option.ID.
	return s.quizRepo.UpdateAnswerOption(ctx, collegeID, option)
}

func (s *quizService) DeleteAnswerOption(ctx context.Context, collegeID int, optionID int) error {
	// check if answer actually exists or not
	// The repository's GetAnswerOptionByID now includes the college check.
	_, err := s.quizRepo.GetAnswerOptionByID(ctx, collegeID, optionID)
	if err != nil {
		// This error already includes "not found" or other repo errors, scoped by college.
		return fmt.Errorf("error getting option ID %d: %w", optionID, err)
	}
	// The repository's DeleteAnswerOption now includes the college check.
	return s.quizRepo.DeleteAnswerOption(ctx, collegeID, optionID)
}

func (s *quizService) FindAnswerOptionsByQuestion(ctx context.Context, collegeID int, questionID int) ([]*models.AnswerOption, error) {
	// This method is typically called after fetching a Question.

	_, err := s.quizRepo.GetQuestionByID(ctx, collegeID, questionID) // Placeholder 0 for collegeID - needs actual collegeID
	if err != nil {
		return nil, err
	}
	return s.quizRepo.FindAnswerOptionsByQuestion(ctx, questionID)
}

// --- QuizAttempt Methods ---

func (s *quizService) StartQuizAttempt(ctx context.Context, collegeID int, attempt *models.QuizAttempt) error {
	// Validate input
	if err := s.validate.Struct(attempt); err != nil {
		return fmt.Errorf("validation failed for quiz attempt: %w", err)
	}

	// Business logic: Check if the student has already attempted this quiz.
	_, err := s.quizRepo.GetQuizAttemptByID(ctx, collegeID, attempt.ID)
	if err != nil {
		return fmt.Errorf("quiz alredy attempted with ID %d", err)
	}

	// Rest of your existing validation logic
	quiz, err := s.quizRepo.GetQuizByID(ctx, attempt.CollegeID, attempt.QuizID)
	if err != nil {
		return fmt.Errorf("failed to get quiz ID %d for attempt: %w", attempt.QuizID, err)
	}
	if quiz == nil {
		return fmt.Errorf("quiz with ID %d not found", attempt.QuizID)
	}

	// Set initial attempt state
	attempt.StartTime = time.Now()
	attempt.Status = models.QuizAttemptStatusInProgress

	return s.quizRepo.CreateQuizAttempt(ctx, attempt)
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

	// Get all answers for this attempt
	answers, err := s.quizRepo.FindStudentAnswersByAttempt(ctx, collegeID, attemptID, 1000, 0) // Pass collegeID
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

	attempt.EndTime = time.Now()
	attempt.Status = models.QuizAttemptStatusCompleted
	attempt.Score = &totalScore

	if err := s.quizRepo.UpdateQuizAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to update quiz attempt ID %d on submission: %w", attemptID, err)
	}

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

// func (s *quizService) FindQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64) ([]*models.QuizAttempt, error) {
// 	return s.quizRepo.FindQuizAttemptsByQuiz(ctx, collegeID, quizID, limit, offset)
// }

// func (s *quizService) CountQuizAttemptsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
// 	return s.quizRepo.CountQuizAttemptsByQuiz(ctx, collegeID, quizID)
// }

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

func (s *quizService) GradeStudentAnswer(ctx context.Context, collegeID int, answerID int, isCorrect *bool, pointsAwarded *int) (*models.StudentAnswer, error) {
	// Retrieve the student answer, scoped by college.
	sa, err := s.quizRepo.GetStudentAnswerByID(ctx, collegeID, answerID) // Pass collegeID
	if err != nil {
		return nil, fmt.Errorf("could not retrieve student answer %d for grading: %w", answerID, err)
	}
	if sa == nil {
		// The repository error already indicates not found for the specific college.
		return nil, fmt.Errorf("student answer with ID %d not found for grading in college %d", answerID, collegeID)
	}

	sa.IsCorrect = isCorrect
	sa.PointsAwarded = pointsAwarded

	// Update the student answer. The repository method UpdateStudentAnswer
	// should ideally also check college scope, but it currently updates by ID.
	if err := s.quizRepo.UpdateStudentAnswer(ctx, sa); err != nil { // UpdateStudentAnswer needs collegeID too
		return nil, fmt.Errorf("failed to update grade for student answer ID %d: %w", answerID, err)
	}
	return sa, nil
}

func (s *quizService) FindStudentAnswersByAttempt(ctx context.Context, collegeID, attemptID int, limit, offset uint64) ([]*models.StudentAnswer, error) {
	// This method needs collegeID to scope the attempt.
	return s.quizRepo.FindStudentAnswersByAttempt(ctx, collegeID, attemptID, limit, offset) // HACK: Needs collegeID
}

func (s *quizService) GetStudentAnswerForQuestion(ctx context.Context, collegeID, attemptID int, questionID int) (*models.StudentAnswer, error) {
	// This method needs collegeID to scope the attempt.
	return s.quizRepo.GetStudentAnswerForQuestion(ctx, collegeID, attemptID, questionID) // HACK: Needs collegeID
}
