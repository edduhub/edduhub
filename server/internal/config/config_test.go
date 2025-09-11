package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	// This test ensures the NewConfig function can be called
	// In a real environment, environment variables would need to be set
	_, err := NewConfig()
	// We expect this to fail in test environment due to missing env vars
	assert.Error(t, err)
}

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{}

	// Test with nil fields
	err := cfg.Validate()
	assert.Error(t, err)

	// Test with empty config
	cfg = &Config{
		DB:         nil,
		DBConfig:   &DBConfig{},
		AuthConfig: &AuthConfig{},
		AppConfig:  &AppConfig{},
	}
	err = cfg.Validate()
	assert.Error(t, err)
}
