package quiz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuizService(t *testing.T) {
	service := NewQuizService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestQuizServiceInterface(t *testing.T) {
	var service QuizService = NewQuizService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestQuizService_MethodsExist(t *testing.T) {
	service := NewQuizService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}