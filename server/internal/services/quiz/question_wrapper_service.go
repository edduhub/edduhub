package quiz

import (
	"context"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

// Simple wrapper for question service
type simpleQuestionService struct {
	questionRepo repository.QuestionRepository
}

func NewSimpleQuestionService(questionRepo repository.QuestionRepository) QuestionServiceSimple {
	return &simpleQuestionService{
		questionRepo: questionRepo,
	}
}

func (s *simpleQuestionService) CreateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	return s.questionRepo.CreateQuestion(ctx, question)
}

func (s *simpleQuestionService) GetQuestion(ctx context.Context, collegeID, questionID int) (*models.Question, error) {
	return s.questionRepo.GetQuestionByID(ctx, collegeID, questionID)
}

func (s *simpleQuestionService) UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	return s.questionRepo.UpdateQuestion(ctx, collegeID, question)
}

func (s *simpleQuestionService) DeleteQuestion(ctx context.Context, collegeID, questionID int) error {
	return s.questionRepo.DeleteQuestion(ctx, collegeID, questionID)
}

func (s *simpleQuestionService) ListQuestionsByQuiz(ctx context.Context, collegeID, quizID int, limit, offset uint64) ([]*models.Question, error) {
	return s.questionRepo.FindQuestionsByQuiz(ctx, collegeID, quizID, limit, offset)
}

// Interface for handler compatibility
type QuestionServiceSimple interface {
	CreateQuestion(ctx context.Context, collegeID int, question *models.Question) error
	GetQuestion(ctx context.Context, collegeID, questionID int) (*models.Question, error)
	UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error
	DeleteQuestion(ctx context.Context, collegeID, questionID int) error
	ListQuestionsByQuiz(ctx context.Context, collegeID, quizID int, limit, offset uint64) ([]*models.Question, error)
}
