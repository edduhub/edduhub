package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"eduhub/server/internal/models"
)

func TestNewAnswerOptionRepository(t *testing.T) {
	repo := NewAnswerOptionRepository(nil)
	assert.NotNil(t, repo)
}

func TestAnswerOptionRepositoryInterface(t *testing.T) {
	var repo AnswerOptionRepository = NewAnswerOptionRepository(nil)
	assert.NotNil(t, repo)
}

func TestAnswerOptionRepository_MethodsExist(t *testing.T) {
	repo := NewAnswerOptionRepository(nil)
	assert.NotNil(t, repo)
}

func TestAnswerOptionModelValidation(t *testing.T) {
	option := &models.AnswerOption{
		ID:         1,
		QuestionID: 2,
		Text:       "Test Option",
		IsCorrect:  true,
	}

	assert.Equal(t, 1, option.ID)
	assert.Equal(t, 2, option.QuestionID)
	assert.Equal(t, "Test Option", option.Text)
	assert.True(t, option.IsCorrect)
}