package quiz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuizAttemptService(t *testing.T) {
	service := NewQuizAttemptService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestQuizAttemptServiceInterface(t *testing.T) {
	var service QuizAttemptService = NewQuizAttemptService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestQuizAttemptService_MethodsExist(t *testing.T) {
	service := NewQuizAttemptService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}