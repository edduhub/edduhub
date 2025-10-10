package quiz

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

// Simple wrapper service for quiz attempts
type simpleQuizAttemptService struct {
	attemptRepo repository.QuizAttemptRepository
	answerRepo  repository.StudentAnswerRepository
	quizRepo    repository.QuizRepository
}

func NewSimpleQuizAttemptService(
	attemptRepo repository.QuizAttemptRepository,
	answerRepo repository.StudentAnswerRepository,
	quizRepo repository.QuizRepository,
) QuizAttemptServiceSimple {
	return &simpleQuizAttemptService{
		attemptRepo: attemptRepo,
		answerRepo:  answerRepo,
		quizRepo:    quizRepo,
	}
}

func (s *simpleQuizAttemptService) StartAttempt(ctx context.Context, collegeID, quizID, studentID int) (*models.QuizAttempt, error) {
	// Verify quiz exists
	_, err := s.quizRepo.GetQuizByID(ctx, collegeID, quizID)
	if err != nil {
		return nil, fmt.Errorf("quiz not found")
	}

	attempt := &models.QuizAttempt{
		QuizID:    quizID,
		StudentID: studentID,
		CollegeID: collegeID,
		Status:    "in_progress",
	}

	err = s.attemptRepo.CreateQuizAttempt(ctx, attempt)
	if err != nil {
		return nil, err
	}

	return attempt, nil
}

func (s *simpleQuizAttemptService) SubmitAttempt(ctx context.Context, collegeID, attemptID, studentID int, answers []models.StudentAnswer) (*models.QuizAttempt, error) {
	// Get attempt
	attempt, err := s.attemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, fmt.Errorf("attempt not found")
	}

	if attempt.StudentID != studentID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Save answers
	for i := range answers {
		answers[i].QuizAttemptID = attemptID
		err = s.answerRepo.CreateStudentAnswer(ctx, &answers[i])
		if err != nil {
			return nil, err
		}
	}

	// Mark as completed
	attempt.Status = "completed"
	// In production, calculate score here

	return attempt, nil
}

func (s *simpleQuizAttemptService) GetAttempt(ctx context.Context, collegeID, attemptID int) (*models.QuizAttempt, error) {
	return s.attemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
}

func (s *simpleQuizAttemptService) GetStudentAttempts(ctx context.Context, collegeID, studentID int) ([]*models.QuizAttempt, error) {
	return s.attemptRepo.FindQuizAttemptsByStudent(ctx, collegeID, studentID, 100, 0)
}

func (s *simpleQuizAttemptService) GetQuizAttempts(ctx context.Context, collegeID, quizID int) ([]*models.QuizAttempt, error) {
	return s.attemptRepo.FindQuizAttemptsByQuiz(ctx, collegeID, quizID, 100, 0)
}

// Interface definition for handler compatibility
type QuizAttemptServiceSimple interface {
	StartAttempt(ctx context.Context, collegeID, quizID, studentID int) (*models.QuizAttempt, error)
	SubmitAttempt(ctx context.Context, collegeID, attemptID, studentID int, answers []models.StudentAnswer) (*models.QuizAttempt, error)
	GetAttempt(ctx context.Context, collegeID, attemptID int) (*models.QuizAttempt, error)
	GetStudentAttempts(ctx context.Context, collegeID, studentID int) ([]*models.QuizAttempt, error)
	GetQuizAttempts(ctx context.Context, collegeID, quizID int) ([]*models.QuizAttempt, error)
}
