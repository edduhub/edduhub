package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"eduhub/server/internal/models"
)

func TestNewQuestionRepository(t *testing.T) {
	repo := NewQuestionRepository(nil)
	assert.NotNil(t, repo)
}

func TestQuestionRepositoryInterface(t *testing.T) {
	var repo QuestionRepository = NewQuestionRepository(nil)
	assert.NotNil(t, repo)
}

func TestQuestionRepository_MethodsExist(t *testing.T) {
	repo := NewQuestionRepository(nil)
	assert.NotNil(t, repo)
}

func TestQuestionModelValidation(t *testing.T) {
	question := &models.Question{
		ID:          1,
		QuizID:      2,
		Text:        "Test Question",
		Type:        models.MultipleChoice,
		Points:      10,
	}

	assert.Equal(t, 1, question.ID)
	assert.Equal(t, 2, question.QuizID)
	assert.Equal(t, "Test Question", question.Text)
	assert.Equal(t, models.MultipleChoice, question.Type)
	assert.Equal(t, 10, question.Points)
}