package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"eduhub/server/internal/models"
)

func TestNewStudentAnswerRepository(t *testing.T) {
	repo := NewStudentAnswerRepository(nil)
	assert.NotNil(t, repo)
}

func TestStudentAnswerRepositoryInterface(t *testing.T) {
	var repo StudentAnswerRepository = NewStudentAnswerRepository(nil)
	assert.NotNil(t, repo)
}

func TestStudentAnswerRepository_MethodsExist(t *testing.T) {
	repo := NewStudentAnswerRepository(nil)
	assert.NotNil(t, repo)
}

func TestStudentAnswerModelValidation(t *testing.T) {
	answer := &models.StudentAnswer{
		ID:             1,
		QuizAttemptID:  2,
		QuestionID:     3,
		AnswerText:     "Test Answer",
	}

	assert.Equal(t, 1, answer.ID)
	assert.Equal(t, 2, answer.QuizAttemptID)
	assert.Equal(t, 3, answer.QuestionID)
	assert.Equal(t, "Test Answer", answer.AnswerText)
}