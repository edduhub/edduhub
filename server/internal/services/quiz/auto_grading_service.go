package quiz

import (
	"context"
	"fmt"
	"strconv"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

// AutoGradingService handles automatic grading of quiz attempts
type AutoGradingService interface {
	// AutoGradeAttempt automatically grades all answers in a quiz attempt
	AutoGradeAttempt(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error)

	// AutoGradeAnswer automatically grades a single answer
	AutoGradeAnswer(ctx context.Context, collegeID int, answerID int) error

	// CalculateScore calculates the total score for an attempt
	CalculateScore(ctx context.Context, collegeID int, attemptID int) (int, error)
}

type autoGradingService struct {
	questionRepo      repository.QuestionRepository
	studentAnswerRepo repository.StudentAnswerRepository
	quizAttemptRepo   repository.QuizAttemptRepository
	answerOptionRepo  repository.AnswerOptionRepository
}

// NewAutoGradingService creates a new auto-grading service
func NewAutoGradingService(
	questionRepo repository.QuestionRepository,
	studentAnswerRepo repository.StudentAnswerRepository,
	quizAttemptRepo repository.QuizAttemptRepository,
	answerOptionRepo repository.AnswerOptionRepository,
) AutoGradingService {
	return &autoGradingService{
		questionRepo:      questionRepo,
		studentAnswerRepo: studentAnswerRepo,
		quizAttemptRepo:   quizAttemptRepo,
		answerOptionRepo:  answerOptionRepo,
	}
}

// AutoGradeAttempt automatically grades all answers in a quiz attempt
func (s *autoGradingService) AutoGradeAttempt(ctx context.Context, collegeID int, attemptID int) (*models.QuizAttempt, error) {
	// Get all answers for this attempt
	answers, err := s.studentAnswerRepo.FindStudentAnswersByAttempt(ctx, collegeID, attemptID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get student answers: %w", err)
	}

	// Grade each answer
	for _, answer := range answers {
		if err := s.AutoGradeAnswer(ctx, collegeID, answer.ID); err != nil {
			// Log error but continue grading other answers
			fmt.Printf("Failed to grade answer %d: %v\n", answer.ID, err)
		}
	}

	// Calculate total score
	totalScore, err := s.CalculateScore(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate score: %w", err)
	}

	// Update attempt with graded status
	attempt, err := s.quizAttemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz attempt: %w", err)
	}

	attempt.Score = &totalScore
	attempt.Status = models.QuizAttemptStatusGraded

	if err := s.quizAttemptRepo.UpdateQuizAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("failed to update quiz attempt: %w", err)
	}

	return attempt, nil
}

// AutoGradeAnswer automatically grades a single answer based on question type
func (s *autoGradingService) AutoGradeAnswer(ctx context.Context, collegeID int, answerID int) error {
	// Get the student answer
	answer, err := s.studentAnswerRepo.GetStudentAnswerByID(ctx, collegeID, answerID)
	if err != nil {
		return fmt.Errorf("failed to get student answer: %w", err)
	}

	// Get the question
	question, err := s.questionRepo.GetQuestionByID(ctx, collegeID, answer.QuestionID)
	if err != nil {
		return fmt.Errorf("failed to get question: %w", err)
	}

	// Grade based on question type
	var isCorrect bool
	var pointsAwarded int

	switch question.Type {
	case models.MultipleChoice, models.TrueFalse:
		isCorrect, pointsAwarded = s.gradeMultipleChoice(question, answer)
	case models.ShortAnswer:
		isCorrect, pointsAwarded = s.gradeShortAnswer(question, answer)
	default:
		return fmt.Errorf("unsupported question type: %s", question.Type)
	}

	// Update the answer with grading results
	answer.IsCorrect = &isCorrect
	answer.PointsAwarded = &pointsAwarded

	if err := s.studentAnswerRepo.UpdateStudentAnswer(ctx, collegeID, answer); err != nil {
		return fmt.Errorf("failed to update student answer: %w", err)
	}

	return nil
}

// gradeMultipleChoice grades multiple choice and true/false questions
func (s *autoGradingService) gradeMultipleChoice(question *models.Question, answer *models.StudentAnswer) (bool, int) {
	if answer.SelectedOptionID == nil || len(*answer.SelectedOptionID) == 0 {
		return false, 0
	}

	// Get correct answer options
	correctOptions := s.getCorrectOptions(question)
	if len(correctOptions) == 0 {
		// If no correct options defined, question can't be graded
		return false, 0
	}

	// Check if selected option is correct
	selectedOption := (*answer.SelectedOptionID)[0]
	for _, correctOpt := range correctOptions {
		if correctOptID, err := strconv.Atoi(correctOpt); err == nil && selectedOption == correctOptID {
			return true, question.Points
		}
	}

	return false, 0
}

// gradeShortAnswer grades short answer questions using exact or partial match
func (s *autoGradingService) gradeShortAnswer(question *models.Question, answer *models.StudentAnswer) (bool, int) {
	// Short answer grading is not yet implemented - requires manual grading
	// TODO: Implement short answer grading with correct answer storage
	return false, 0
}

// getCorrectOptions extracts correct option IDs from question
func (s *autoGradingService) getCorrectOptions(question *models.Question) []string {
	correctOptions := []string{}

	if question.Options == nil {
		return correctOptions
	}

	for _, option := range question.Options {
		if option.IsCorrect {
			correctOptions = append(correctOptions, fmt.Sprintf("%d", option.ID))
		}
	}

	return correctOptions
}

// CalculateScore calculates the total score for an attempt
func (s *autoGradingService) CalculateScore(ctx context.Context, collegeID int, attemptID int) (int, error) {
	answers, err := s.studentAnswerRepo.FindStudentAnswersByAttempt(ctx, collegeID, attemptID, 1000, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get student answers: %w", err)
	}

	totalScore := 0
	for _, answer := range answers {
		if answer.PointsAwarded != nil {
			totalScore += *answer.PointsAwarded
		}
	}

	return totalScore, nil
}
