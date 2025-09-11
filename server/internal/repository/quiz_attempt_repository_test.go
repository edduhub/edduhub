package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"eduhub/server/internal/models"
)

func TestNewQuizAttemptRepository(t *testing.T) {
	repo := NewQuizAttemptRepository(nil)
	assert.NotNil(t, repo)
}

func TestQuizAttemptRepositoryInterface(t *testing.T) {
	var repo QuizAttemptRepository = NewQuizAttemptRepository(nil)
	assert.NotNil(t, repo)
}

func TestQuizAttemptRepository_MethodsExist(t *testing.T) {
	repo := NewQuizAttemptRepository(nil)
	assert.NotNil(t, repo)
}

func TestQuizAttemptModelValidation(t *testing.T) {
	attempt := &models.QuizAttempt{
		ID:        1,
		StudentID: 101,
		QuizID:    2,
		CollegeID: 1,
		CourseID:  3,
		Status:    models.QuizAttemptStatusInProgress,
	}

	assert.Equal(t, 1, attempt.ID)
	assert.Equal(t, 101, attempt.StudentID)
	assert.Equal(t, 2, attempt.QuizID)
	assert.Equal(t, 1, attempt.CollegeID)
	assert.Equal(t, 3, attempt.CourseID)
	assert.Equal(t, models.QuizAttemptStatusInProgress, attempt.Status)
}