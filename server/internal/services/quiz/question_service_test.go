package quiz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuestionService(t *testing.T) {
	service := NewQuestionService(nil, nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestQuestionServiceInterface(t *testing.T) {
	var service QuestionService = NewQuestionService(nil, nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestQuestionService_MethodsExist(t *testing.T) {
	service := NewQuestionService(nil, nil, nil, nil, nil)
	assert.NotNil(t, service)
}