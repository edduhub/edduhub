package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMiddleware(t *testing.T) {
	middleware := NewMiddleware(nil)
	assert.NotNil(t, middleware)
	assert.NotNil(t, middleware.Auth)
}

func TestMiddleware_Struct(t *testing.T) {
	m := &Middleware{}
	assert.NotNil(t, m)
}