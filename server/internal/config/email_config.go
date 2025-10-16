// Package config provides email configuration for the EduHub application.
// This file implements SMTP email configuration management.
package config

import (
	"fmt"
	"os"
)

// EmailConfig holds SMTP email configuration parameters.
// It supports connection to SMTP servers for sending emails.
type EmailConfig struct {
	// Host is the SMTP server hostname (e.g., "smtp.gmail.com")
	Host string

	// Port is the SMTP server port (e.g., "587" for TLS, "465" for SSL)
	Port string

	// Username is the SMTP authentication username
	Username string

	// Password is the SMTP authentication password
	Password string

	// FromAddress is the email address used as the "From" field
	FromAddress string

	// EnableStartTLS indicates whether to use STARTTLS for encryption
	EnableStartTLS bool
}

// LoadEmailConfig loads email configuration from environment variables.
// It supports both required and optional SMTP settings with sensible defaults.
//
// Environment variables:
//   - SMTP_HOST: SMTP server hostname (required)
//   - SMTP_PORT: SMTP server port (default: "587")
//   - SMTP_USERNAME: SMTP authentication username (required)
//   - SMTP_PASSWORD: SMTP authentication password (required)
//   - EMAIL_FROM: Email address for "From" field (required)
//   - SMTP_STARTTLS: Enable STARTTLS (default: "true")
//
// Returns:
//   - *EmailConfig: The loaded SMTP configuration
//   - error: Any validation errors
func LoadEmailConfig() (*EmailConfig, error) {
	config := &EmailConfig{
		Host:         os.Getenv("SMTP_HOST"),
		Port:         getEnvOrDefault("SMTP_PORT", "587"),
		Username:     os.Getenv("SMTP_USERNAME"),
		Password:     os.Getenv("SMTP_PASSWORD"),
		FromAddress:  os.Getenv("EMAIL_FROM"),
		EnableStartTLS: getEnvOrDefault("SMTP_STARTTLS", "true") == "true",
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("email config validation failed: %w", err)
	}

	return config, nil
}

// Validate performs validation on the EmailConfig.
// It ensures all required parameters are present and valid.
//
// Returns an error if validation fails.
func (c *EmailConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}
	if c.Port == "" {
		return fmt.Errorf("SMTP_PORT is required")
	}
	if c.Username == "" {
		return fmt.Errorf("SMTP_USERNAME is required")
	}
	if c.Password == "" {
		return fmt.Errorf("SMTP_PASSWORD is required")
	}
	if c.FromAddress == "" {
		return fmt.Errorf("EMAIL_FROM is required")
	}

	return nil
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
