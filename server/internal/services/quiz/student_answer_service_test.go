package quiz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStudentAnswerService(t *testing.T) {
	service := NewStudentAnswerService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestStudentAnswerServiceInterface(t *testing.T) {
	var service StudentAnswerService = NewStudentAnswerService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}

func TestStudentAnswerService_MethodsExist(t *testing.T) {
	service := NewStudentAnswerService(nil, nil, nil, nil)
	assert.NotNil(t, service)
}