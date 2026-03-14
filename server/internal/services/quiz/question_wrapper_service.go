package quiz

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

// Simple wrapper for question service
type simpleQuestionService struct {
	questionRepo     repository.QuestionRepository
	answerOptionRepo repository.AnswerOptionRepository
}

func NewSimpleQuestionService(
	questionRepo repository.QuestionRepository,
	answerOptionRepo repository.AnswerOptionRepository,
) QuestionServiceSimple {
	return &simpleQuestionService{
		questionRepo:     questionRepo,
		answerOptionRepo: answerOptionRepo,
	}
}

func (s *simpleQuestionService) CreateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	if err := s.questionRepo.CreateQuestion(ctx, question); err != nil {
		return err
	}

	for _, option := range question.Options {
		option.QuestionID = question.ID
		if err := s.answerOptionRepo.CreateAnswerOption(ctx, option); err != nil {
			return fmt.Errorf("failed to create answer option: %w", err)
		}
	}

	return nil
}

func (s *simpleQuestionService) GetQuestion(ctx context.Context, collegeID, questionID int) (*models.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(ctx, collegeID, questionID)
	if err != nil {
		return nil, err
	}

	options, err := s.answerOptionRepo.FindAnswerOptionsByQuestion(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch answer options: %w", err)
	}
	question.Options = options

	return question, nil
}

func (s *simpleQuestionService) UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error {
	if err := s.questionRepo.UpdateQuestion(ctx, collegeID, question); err != nil {
		return err
	}

	if question.Options == nil {
		return nil
	}

	existingOptions, err := s.answerOptionRepo.FindAnswerOptionsByQuestion(ctx, question.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing answer options: %w", err)
	}

	for _, option := range existingOptions {
		if err := s.answerOptionRepo.DeleteAnswerOption(ctx, collegeID, option.ID); err != nil {
			return fmt.Errorf("failed to delete answer option %d: %w", option.ID, err)
		}
	}

	for _, option := range question.Options {
		option.QuestionID = question.ID
		if err := s.answerOptionRepo.CreateAnswerOption(ctx, option); err != nil {
			return fmt.Errorf("failed to recreate answer option: %w", err)
		}
	}

	return nil
}

func (s *simpleQuestionService) DeleteQuestion(ctx context.Context, collegeID, questionID int) error {
	return s.questionRepo.DeleteQuestion(ctx, collegeID, questionID)
}

func (s *simpleQuestionService) ListQuestionsByQuiz(ctx context.Context, collegeID, quizID int, limit, offset uint64) ([]*models.Question, error) {
	questions, err := s.questionRepo.FindQuestionsByQuiz(ctx, collegeID, quizID, limit, offset)
	if err != nil {
		return nil, err
	}

	for _, question := range questions {
		options, optionErr := s.answerOptionRepo.FindAnswerOptionsByQuestion(ctx, question.ID)
		if optionErr != nil {
			return nil, fmt.Errorf("failed to fetch answer options for question %d: %w", question.ID, optionErr)
		}
		question.Options = options
	}

	return questions, nil
}

// Interface for handler compatibility
type QuestionServiceSimple interface {
	CreateQuestion(ctx context.Context, collegeID int, question *models.Question) error
	GetQuestion(ctx context.Context, collegeID, questionID int) (*models.Question, error)
	UpdateQuestion(ctx context.Context, collegeID int, question *models.Question) error
	DeleteQuestion(ctx context.Context, collegeID, questionID int) error
	ListQuestionsByQuiz(ctx context.Context, collegeID, quizID int, limit, offset uint64) ([]*models.Question, error)
}
