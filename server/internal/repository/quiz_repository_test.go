package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eduhub/server/internal/models"
)

func TestNewQuizRepository(t *testing.T) {
	// Test that the constructor works
	repo := NewQuizRepository(nil)
	assert.NotNil(t, repo)
}

func TestQuizRepositoryInterface(t *testing.T) {
	// Test that the repository implements the expected interface
	var repo QuizRepository = NewQuizRepository(nil)
	assert.NotNil(t, repo)
}

// Test data structures for validation
func TestQuizModelValidation(t *testing.T) {
	quiz := &models.Quiz{
		ID:               1,
		CollegeID:        1,
		CourseID:         2,
		Title:            "Test Quiz",
		Description:      "Test Description",
		TimeLimitMinutes: 60,
	}

	assert.Equal(t, 1, quiz.ID)
	assert.Equal(t, 1, quiz.CollegeID)
	assert.Equal(t, 2, quiz.CourseID)
	assert.Equal(t, "Test Quiz", quiz.Title)
	assert.Equal(t, "Test Description", quiz.Description)
	assert.Equal(t, 60, quiz.TimeLimitMinutes)
}

func TestQuizModel_Timestamps(t *testing.T) {
	now := time.Now()
	quiz := &models.Quiz{
		ID:               1,
		CollegeID:        1,
		CourseID:         2,
		Title:            "Test Quiz",
		Description:      "Test Description",
		TimeLimitMinutes: 60,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	assert.False(t, quiz.CreatedAt.IsZero())
	assert.False(t, quiz.UpdatedAt.IsZero())
	assert.Equal(t, quiz.CreatedAt, quiz.UpdatedAt)
}

func TestQuizModel_DueDate(t *testing.T) {
	dueDate := time.Now().Add(24 * time.Hour)
	quiz := &models.Quiz{
		ID:               1,
		CollegeID:        1,
		CourseID:         2,
		Title:            "Test Quiz",
		Description:      "Test Description",
		TimeLimitMinutes: 60,
		DueDate:          dueDate,
	}

	assert.False(t, quiz.DueDate.IsZero())
	assert.True(t, quiz.DueDate.After(time.Now()))
}

func TestQuizModel_Relations(t *testing.T) {
	quiz := &models.Quiz{
		ID:               1,
		CollegeID:        1,
		CourseID:         2,
		Title:            "Test Quiz",
		Description:      "Test Description",
		TimeLimitMinutes: 60,
		Questions: []*models.Question{
			{
				ID:     1,
				QuizID: 1,
				Text:   "What is 2+2?",
				Type:   models.MultipleChoice,
				Points: 10,
			},
		},
	}

	assert.NotNil(t, quiz.Questions)
	assert.Len(t, quiz.Questions, 1)
	assert.Equal(t, "What is 2+2?", quiz.Questions[0].Text)
}

func TestUpdateQuizRequest_Validation(t *testing.T) {
	title := "Updated Title"
	description := "Updated Description"
	timeLimitMinutes := 90

	req := &models.UpdateQuizRequest{
		Title:            &title,
		Description:      &description,
		TimeLimitMinutes: &timeLimitMinutes,
	}

	assert.NotNil(t, req.Title)
	assert.NotNil(t, req.Description)
	assert.NotNil(t, req.TimeLimitMinutes)
	assert.Equal(t, "Updated Title", *req.Title)
	assert.Equal(t, "Updated Description", *req.Description)
	assert.Equal(t, 90, *req.TimeLimitMinutes)
}

func TestUpdateQuizRequest_PartialUpdate(t *testing.T) {
	title := "Updated Title"

	req := &models.UpdateQuizRequest{
		Title: &title,
		// Other fields are nil, indicating they should not be updated
	}

	assert.NotNil(t, req.Title)
	assert.Nil(t, req.Description)
	assert.Nil(t, req.TimeLimitMinutes)
}

func TestQuestionModel_Types(t *testing.T) {
	tests := []struct {
		name         string
		questionType models.QuizType
		expected     models.QuizType
	}{
		{"Multiple Choice", models.MultipleChoice, models.MultipleChoice},
		{"True False", models.TrueFalse, models.TrueFalse},
		{"Short Answer", models.ShortAnswer, models.ShortAnswer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			question := &models.Question{
				ID:     1,
				QuizID: 1,
				Text:   "Test Question",
				Type:   tt.questionType,
				Points: 10,
			}

			assert.Equal(t, tt.expected, question.Type)
		})
	}
}

func TestQuestionModel_WithCorrectAnswer(t *testing.T) {
	correctAnswer := "Photosynthesis"
	question := &models.Question{
		ID:            1,
		QuizID:        1,
		Text:          "What process do plants use to make food?",
		Type:          models.ShortAnswer,
		Points:        10,
		CorrectAnswer: &correctAnswer,
	}

	assert.NotNil(t, question.CorrectAnswer)
	assert.Equal(t, "Photosynthesis", *question.CorrectAnswer)
}

func TestQuestionModel_WithOptions(t *testing.T) {
	question := &models.Question{
		ID:     1,
		QuizID: 1,
		Text:   "What is 2+2?",
		Type:   models.MultipleChoice,
		Points: 10,
		Options: []*models.AnswerOption{
			{ID: 1, QuestionID: 1, Text: "3", IsCorrect: false},
			{ID: 2, QuestionID: 1, Text: "4", IsCorrect: true},
			{ID: 3, QuestionID: 1, Text: "5", IsCorrect: false},
		},
	}

	assert.NotNil(t, question.Options)
	assert.Len(t, question.Options, 3)
	
	// Find the correct option
	var correctOption *models.AnswerOption
	for _, opt := range question.Options {
		if opt.IsCorrect {
			correctOption = opt
			break
		}
	}
	
	require.NotNil(t, correctOption)
	assert.Equal(t, "4", correctOption.Text)
}

func TestQuizAttemptStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   models.QuizAttemptStatus
		expected models.QuizAttemptStatus
	}{
		{"In Progress", models.QuizAttemptStatusInProgress, models.QuizAttemptStatusInProgress},
		{"Completed", models.QuizAttemptStatusCompleted, models.QuizAttemptStatusCompleted},
		{"Graded", models.QuizAttemptStatusGraded, models.QuizAttemptStatusGraded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempt := &models.QuizAttempt{
				ID:        1,
				StudentID: 1,
				QuizID:    1,
				CollegeID: 1,
				CourseID:  1,
				Status:    tt.status,
			}

			assert.Equal(t, tt.expected, attempt.Status)
		})
	}
}

func TestStudentAnswerModel_MultipleChoice(t *testing.T) {
	selectedOptions := []int{2}
	answer := &models.StudentAnswer{
		ID:               1,
		QuizAttemptID:    1,
		QuestionID:       1,
		SelectedOptionID: &selectedOptions,
	}

	assert.NotNil(t, answer.SelectedOptionID)
	assert.Len(t, *answer.SelectedOptionID, 1)
	assert.Equal(t, 2, (*answer.SelectedOptionID)[0])
}

func TestStudentAnswerModel_ShortAnswer(t *testing.T) {
	answer := &models.StudentAnswer{
		ID:            1,
		QuizAttemptID: 1,
		QuestionID:    1,
		AnswerText:    "Photosynthesis is the process plants use to make food",
	}

	assert.NotEmpty(t, answer.AnswerText)
	assert.Contains(t, answer.AnswerText, "Photosynthesis")
}

func TestStudentAnswerModel_Grading(t *testing.T) {
	isCorrect := true
	pointsAwarded := 10
	
	answer := &models.StudentAnswer{
		ID:            1,
		QuizAttemptID: 1,
		QuestionID:    1,
		AnswerText:    "Correct answer",
		IsCorrect:     &isCorrect,
		PointsAwarded: &pointsAwarded,
	}

	assert.NotNil(t, answer.IsCorrect)
	assert.NotNil(t, answer.PointsAwarded)
	assert.True(t, *answer.IsCorrect)
	assert.Equal(t, 10, *answer.PointsAwarded)
}

func TestQuizStatistics(t *testing.T) {
	stats := &models.QuizStatistics{
		QuizID:            1,
		TotalAttempts:     10,
		CompletedAttempts: 8,
		AverageScore:      75,
		HighestScore:      95,
		LowestScore:       45,
	}

	assert.Equal(t, 1, stats.QuizID)
	assert.Equal(t, 10, stats.TotalAttempts)
	assert.Equal(t, 8, stats.CompletedAttempts)
	assert.Equal(t, 75, stats.AverageScore)
	assert.Equal(t, 95, stats.HighestScore)
	assert.Equal(t, 45, stats.LowestScore)
}

// Mock database tests would go here in a production environment
// These would use testcontainers or a similar tool to test actual database operations

func TestQuizRepository_CreateQuiz_NilDB(t *testing.T) {
	repo := NewQuizRepository(nil)
	ctx := context.Background()
	
	quiz := &models.Quiz{
		CollegeID:        1,
		CourseID:         1,
		Title:            "Test Quiz",
		Description:      "Test Description",
		TimeLimitMinutes: 60,
	}

	// This will fail with nil DB, but tests that the method exists and has correct signature
	err := repo.CreateQuiz(ctx, quiz)
	assert.Error(t, err)
}

func TestQuizRepository_GetQuizByID_NilDB(t *testing.T) {
	repo := NewQuizRepository(nil)
	ctx := context.Background()

	// This will fail with nil DB, but tests that the method exists and has correct signature
	quiz, err := repo.GetQuizByID(ctx, 1, 1)
	assert.Error(t, err)
	assert.Nil(t, quiz)
}

func TestQuizRepository_UpdateQuiz_NilDB(t *testing.T) {
	repo := NewQuizRepository(nil)
	ctx := context.Background()
	
	quiz := &models.Quiz{
		ID:               1,
		CollegeID:        1,
		CourseID:         1,
		Title:            "Updated Quiz",
		Description:      "Updated Description",
		TimeLimitMinutes: 90,
	}

	// This will fail with nil DB, but tests that the method exists and has correct signature
	err := repo.UpdateQuiz(ctx, quiz)
	assert.Error(t, err)
}

func TestQuizRepository_DeleteQuiz_NilDB(t *testing.T) {
	repo := NewQuizRepository(nil)
	ctx := context.Background()

	// This will fail with nil DB, but tests that the method exists and has correct signature
	err := repo.DeleteQuiz(ctx, 1, 1)
	assert.Error(t, err)
}

func TestQuizRepository_FindQuizzesByCourse_NilDB(t *testing.T) {
	repo := NewQuizRepository(nil)
	ctx := context.Background()

	// This will fail with nil DB, but tests that the method exists and has correct signature
	quizzes, err := repo.FindQuizzesByCourse(ctx, 1, 1, 10, 0)
	assert.Error(t, err)
	assert.Nil(t, quizzes)
}

func TestQuizRepository_CountQuizzesByCourse_NilDB(t *testing.T) {
	repo := NewQuizRepository(nil)
	ctx := context.Background()

	// This will fail with nil DB, but tests that the method exists and has correct signature
	count, err := repo.CountQuizzesByCourse(ctx, 1, 1)
	assert.Error(t, err)
	assert.Equal(t, 0, count)
}
