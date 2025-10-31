package quiz

import (
	"context"
	"fmt"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

// Simple wrapper service for quiz attempts
type simpleQuizAttemptService struct {
	attemptRepo repository.QuizAttemptRepository
	answerRepo  repository.StudentAnswerRepository
	quizRepo    repository.QuizRepository
	questionRepo     repository.QuestionRepository
	answerOptionRepo repository.AnswerOptionRepository
    autoGrader       AutoGradingService
}

func NewSimpleQuizAttemptService(
	attemptRepo repository.QuizAttemptRepository,
	answerRepo repository.StudentAnswerRepository,
	quizRepo repository.QuizRepository,
	questionRepo repository.QuestionRepository,
    answerOptionRepo repository.AnswerOptionRepository,
    autoGrader AutoGradingService,
) QuizAttemptServiceSimple {
	return &simpleQuizAttemptService{
		attemptRepo: attemptRepo,
		answerRepo:  answerRepo,
		quizRepo:    quizRepo,
		questionRepo:     questionRepo,
        answerOptionRepo: answerOptionRepo,
        autoGrader:       autoGrader,
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
        Status:    models.QuizAttemptStatusInProgress,
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

    // Auto-grade and finalize attempt status/score
    gradedAttempt, err := s.autoGrader.AutoGradeAttempt(ctx, collegeID, attemptID)
    if err != nil {
        // If auto-grading fails, still mark attempt completed without score
        attempt.Status = models.QuizAttemptStatusCompleted
        attempt.EndTime = time.Now()
        _ = s.attemptRepo.UpdateQuizAttempt(ctx, attempt)
        return nil, fmt.Errorf("failed to auto-grade attempt: %w", err)
    }

    // Ensure end time is set post grading
    if gradedAttempt.EndTime.IsZero() {
        gradedAttempt.EndTime = time.Now()
        _ = s.attemptRepo.UpdateQuizAttempt(ctx, gradedAttempt)
    }

    return gradedAttempt, nil
}

func (s *simpleQuizAttemptService) GetAttempt(ctx context.Context, collegeID, attemptID int) (*models.QuizAttempt, error) {
	attempt, err := s.attemptRepo.GetQuizAttemptByID(ctx, collegeID, attemptID)
	if err != nil {
		return nil, err
	}

	quiz, err := s.quizRepo.GetQuizByID(ctx, collegeID, attempt.QuizID)
	if err == nil {
		attempt.Quiz = quiz
	}

	questions, err := s.questionRepo.FindQuestionsByQuiz(ctx, collegeID, attempt.QuizID, 1000, 0)
	if err == nil {
		for _, q := range questions {
			options, optErr := s.answerOptionRepo.FindAnswerOptionsByQuestion(ctx, q.ID)
			if optErr == nil {
				q.Options = options
			}
		}
		attempt.Quiz.Questions = questions
	}

    // Load student answers for this attempt
    answers, err := s.answerRepo.FindStudentAnswersByAttempt(ctx, collegeID, attemptID, 1000, 0)
    if err == nil {
        attempt.Answers = answers
    }

	return attempt, nil
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
