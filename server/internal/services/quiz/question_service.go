package quiz

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

// QuestionService defines the interface for question and answer option management operations.
// It handles CRUD operations for questions and their associated answer options with proper
// college-based authorization and business logic validation.
type QuestionService interface {
	// Question Methods
	// CreateQuestion creates a new question for a quiz.
	// Validates question data and ensures the quiz exists within the college context.
	CreateQuestion(ctx context.Context, collegeID int, question *models.Question) error

	// GetQuestionByID retrieves a question by ID with college isolation.
	// Returns an error if the question doesn't exist or doesn't belong to the college.
	GetQuestionByID(ctx context.Context, collegeID int, questionID int) (*models.Question, error)

	// UpdateQuestion updates an existing question's information.
	// Validates input data and ensures the question exists within the college context.
	UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error

	// DeleteQuestion removes a question and all its associated answer options.
	// Ensures the question belongs to the college before deletion.
	DeleteQuestion(ctx context.Context, collegeID int, questionID int) error

	// FindQuestionsByQuiz retrieves questions for a specific quiz with pagination.
	// Optionally includes answer options in the response.
	FindQuestionsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64, withOptions bool) ([]*models.Question, error)

	// CountQuestionsByQuiz returns the total number of questions for a quiz.
	CountQuestionsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error)

	// AnswerOption Methods
	// CreateAnswerOption creates a new answer option for a question.
	// Validates option data and ensures the question exists within the college context.
	CreateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error

	// GetAnswerOptionByID retrieves an answer option by ID with college isolation.
	GetAnswerOptionByID(ctx context.Context, collegeID int, optionID int) (*models.AnswerOption, error)

	// UpdateAnswerOption updates an existing answer option's information.
	UpdateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error

	// DeleteAnswerOption removes an answer option from the database.
	DeleteAnswerOption(ctx context.Context, collegeID int, optionID int) error

	// FindAnswerOptionsByQuestion retrieves all answer options for a specific question.
	FindAnswerOptionsByQuestion(ctx context.Context, collegeID int, questionID int) ([]*models.AnswerOption, error)
}

// questionService implements the QuestionService interface.
type questionService struct {
	questionRepo     repository.QuestionRepository
	answerOptionRepo repository.AnswerOptionRepository
	quizRepo         repository.QuizRepository
	courseRepo       repository.CourseRepository
	collegeRepo      repository.CollegeRepository
	validate         *validator.Validate
}

// NewQuestionService creates a new instance of QuestionService with required dependencies.
// Initializes validator for input validation.
func NewQuestionService(
	questionRepo repository.QuestionRepository,
	answerOptionRepo repository.AnswerOptionRepository,
	quizRepo repository.QuizRepository,
	courseRepo repository.CourseRepository,
	collegeRepo repository.CollegeRepository,
) QuestionService {
	return &questionService{
		questionRepo:     questionRepo,
		answerOptionRepo: answerOptionRepo,
		quizRepo:         quizRepo,
		courseRepo:       courseRepo,
		collegeRepo:      collegeRepo,
		validate:         validator.New(),
	}
}

// CreateQuestion creates a new question for a quiz.
// Validates question data and ensures the quiz exists within the college context.
func (s *questionService) CreateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	// Validate question struct
	if err := s.validate.Struct(question); err != nil {
		return fmt.Errorf("validation failed for question: %w", err)
	}

	// Verify college exists
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return fmt.Errorf("college verification failed: %w", err)
	}

	// Verify quiz exists and belongs to the college
	quiz, err := s.quizRepo.GetQuizByID(ctx, collegeID, question.QuizID)
	if err != nil {
		return fmt.Errorf("failed to get quiz: %w", err)
	}
	if quiz == nil {
		return fmt.Errorf("quiz with ID %d not found in college %d", question.QuizID, collegeID)
	}

	// Verify course exists
	_, err = s.courseRepo.FindCourseByID(ctx, collegeID, quiz.CourseID)
	if err != nil {
		return fmt.Errorf("failed to find course: %w", err)
	}

	return s.questionRepo.CreateQuestion(ctx, question)
}

// GetQuestionByID retrieves a question by ID with college isolation.
func (s *questionService) GetQuestionByID(ctx context.Context, collegeID int, questionID int) (*models.Question, error) {
	return s.questionRepo.GetQuestionByID(ctx, collegeID, questionID)
}

// UpdateQuestion updates an existing question's information.
// Validates input data and ensures the question exists within the college context.
func (s *questionService) UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	// Validate question struct
	if err := s.validate.Struct(question); err != nil {
		return fmt.Errorf("validation failed for question: %w", err)
	}

	// Verify college exists
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return fmt.Errorf("college verification failed: %w", err)
	}

	// Ensure question ID is provided
	if question.ID == 0 {
		return fmt.Errorf("question ID is required for update")
	}

	// Verify quiz exists and belongs to the college
	quiz, err := s.quizRepo.GetQuizByID(ctx, collegeID, question.QuizID)
	if err != nil {
		return fmt.Errorf("failed to fetch quiz: %w", err)
	}
	if quiz == nil {
		return fmt.Errorf("quiz with ID %d not found in college %d", question.QuizID, collegeID)
	}

	return s.questionRepo.UpdateQuestion(ctx, collegeID, question)
}

// DeleteQuestion removes a question and all its associated answer options.
// Ensures the question belongs to the college before deletion.
func (s *questionService) DeleteQuestion(ctx context.Context, collegeID int, questionID int) error {
	// Verify question exists and belongs to the college
	question, err := s.questionRepo.GetQuestionByID(ctx, collegeID, questionID)
	if err != nil {
		return fmt.Errorf("failed to get question: %w", err)
	}
	if question == nil {
		return fmt.Errorf("question with ID %d not found in college %d", questionID, collegeID)
	}

	// Find and delete answer options
	options, err := s.answerOptionRepo.FindAnswerOptionsByQuestion(ctx, questionID)
	if err == nil && options != nil {
		for _, opt := range options {
			if delErr := s.answerOptionRepo.DeleteAnswerOption(ctx, collegeID, opt.ID); delErr != nil {
				// Log warning but continue with deletion
				fmt.Printf("Warning: failed to delete answer option ID %d: %v\n", opt.ID, delErr)
			}
		}
	}

	// Delete the question
	return s.questionRepo.DeleteQuestion(ctx, collegeID, questionID)
}

// FindQuestionsByQuiz retrieves questions for a specific quiz with pagination.
// Optionally includes answer options in the response.
func (s *questionService) FindQuestionsByQuiz(ctx context.Context, collegeID int, quizID int, limit, offset uint64, withOptions bool) ([]*models.Question, error) {
	// Verify quiz exists and belongs to the college
	_, err := s.quizRepo.GetQuizByID(ctx, collegeID, quizID)
	if err != nil {
		return nil, fmt.Errorf("quiz verification failed: %w", err)
	}

	questions, err := s.questionRepo.FindQuestionsByQuiz(ctx, collegeID, quizID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Include answer options if requested
	if withOptions {
		for _, q := range questions {
			options, err := s.answerOptionRepo.FindAnswerOptionsByQuestion(ctx, q.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch options for question ID %d: %w", q.ID, err)
			}
			q.Options = options
		}
	}

	return questions, nil
}

// CountQuestionsByQuiz returns the total number of questions for a quiz.
func (s *questionService) CountQuestionsByQuiz(ctx context.Context, collegeID int, quizID int) (int, error) {
	return s.questionRepo.CountQuestionsByQuiz(ctx, collegeID, quizID)
}

// CreateAnswerOption creates a new answer option for a question.
// Validates option data and ensures the question exists within the college context.
func (s *questionService) CreateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error {
	// Validate answer option struct
	if err := s.validate.Struct(option); err != nil {
		return fmt.Errorf("validation failed for answer option: %w", err)
	}

	// Verify question exists and belongs to the college
	_, err := s.questionRepo.GetQuestionByID(ctx, collegeID, option.QuestionID)
	if err != nil {
		return fmt.Errorf("question verification failed: %w", err)
	}

	return s.answerOptionRepo.CreateAnswerOption(ctx, option)
}

// GetAnswerOptionByID retrieves an answer option by ID with college isolation.
func (s *questionService) GetAnswerOptionByID(ctx context.Context, collegeID int, optionID int) (*models.AnswerOption, error) {
	return s.answerOptionRepo.GetAnswerOptionByID(ctx, collegeID, optionID)
}

// UpdateAnswerOption updates an existing answer option's information.
func (s *questionService) UpdateAnswerOption(ctx context.Context, collegeID int, option *models.AnswerOption) error {
	// Validate answer option struct
	if err := s.validate.Struct(option); err != nil {
		return fmt.Errorf("validation failed for answer option: %w", err)
	}

	// Ensure option ID is provided
	if option.ID == 0 {
		return fmt.Errorf("answer option ID is required for update")
	}

	return s.answerOptionRepo.UpdateAnswerOption(ctx, collegeID, option)
}

// DeleteAnswerOption removes an answer option from the database.
func (s *questionService) DeleteAnswerOption(ctx context.Context, collegeID int, optionID int) error {
	// Verify option exists and belongs to the college
	_, err := s.answerOptionRepo.GetAnswerOptionByID(ctx, collegeID, optionID)
	if err != nil {
		return fmt.Errorf("answer option verification failed: %w", err)
	}

	return s.answerOptionRepo.DeleteAnswerOption(ctx, collegeID, optionID)
}

// FindAnswerOptionsByQuestion retrieves all answer options for a specific question.
// Ensures the question belongs to the college context.
func (s *questionService) FindAnswerOptionsByQuestion(ctx context.Context, collegeID int, questionID int) ([]*models.AnswerOption, error) {
	// Verify question exists and belongs to the college
	_, err := s.questionRepo.GetQuestionByID(ctx, collegeID, questionID)
	if err != nil {
		return nil, fmt.Errorf("question verification failed: %w", err)
	}

	return s.answerOptionRepo.FindAnswerOptionsByQuestion(ctx, questionID)
}