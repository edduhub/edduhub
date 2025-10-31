// Package config provides configuration management for the EduHub application.
// This file contains the general application configuration module.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// AppConfig holds general application configuration settings.
// It includes settings that are not specific to database or authentication.
type AppConfig struct {
	// Port specifies the port on which the application server will listen.
	// This is loaded from the APP_PORT environment variable.
	// Default: "8080" if not set.
	Port string

	// Debug indicates whether the application is running in debug mode.
	// This enables additional logging and error details.
	// Loaded from APP_DEBUG environment variable (true/false).
	// Default: false
	Debug bool

	// LogLevel specifies the logging level (e.g., "info", "debug", "error").
	// Loaded from APP_LOG_LEVEL environment variable.
	// Default: "info"
	LogLevel string

	// CORSOrigins specifies allowed origins for CORS requests.
	// Loaded from CORS_ORIGINS environment variable (comma-separated).
	// Default: ["http://localhost:3000"] for development
	CORSOrigins []string
}

// LoadAppConfig loads the general application configuration from environment variables.
// It performs validation on the loaded values and sets secure defaults where appropriate.
//
// Returns:
//   - *AppConfig: The loaded application configuration
//   - error: Any validation or loading errors
//
// Environment Variables:
//   - APP_PORT: The port for the application server (default: "8080")
//   - APP_DEBUG: Enable debug mode (default: false)
//   - APP_LOG_LEVEL: Logging level (default: "info")
//
// Security Considerations:
//   - Port is validated to be a valid integer between 1 and 65535
//   - Debug mode defaults to false for production safety
//   - Log level is restricted to known safe values
func LoadAppConfig() (*AppConfig, error) {
	config := &AppConfig{}

	// Load port with secure default
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080" // Secure default
	}

	// Validate port
	portInt, err := strconv.Atoi(port)
	if err != nil || portInt < 1 || portInt > 65535 {
		return nil, fmt.Errorf("invalid APP_PORT: must be a valid port number (1-65535), got %s", port)
	}
	config.Port = port

	// Load debug mode with secure default (false)
	debugStr := os.Getenv("APP_DEBUG")
	config.Debug = debugStr == "true" // Defaults to false if not "true"

	// Load log level with secure default
	logLevel := os.Getenv("APP_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // Secure default
	}

	// Validate log level (restrict to known safe values)
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[logLevel] {
		return nil, fmt.Errorf("invalid APP_LOG_LEVEL: must be one of 'debug', 'info', 'warn', 'error', got %s", logLevel)
	}
	config.LogLevel = logLevel

	// Load CORS origins with secure default
	corsOriginsStr := os.Getenv("CORS_ORIGINS")
	if corsOriginsStr == "" {
		// Secure default: only allow localhost in development
		config.CORSOrigins = []string{"http://localhost:3000"}
	} else {
		// Parse comma-separated origins
		origins := strings.Split(corsOriginsStr, ",")
		config.CORSOrigins = make([]string, 0, len(origins))
		for _, origin := range origins {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				config.CORSOrigins = append(config.CORSOrigins, trimmed)
			}
		}
		if len(config.CORSOrigins) == 0 {
			return nil, fmt.Errorf("CORS_ORIGINS is set but contains no valid origins")
		}
	}

	return config, nil
}

// Validate performs additional validation on the AppConfig.
// This method can be called after loading to ensure all settings are consistent.
//
// Returns an error if validation fails.
func (c *AppConfig) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("AppConfig.Port cannot be empty")
	}
	if c.LogLevel == "" {
		return fmt.Errorf("AppConfig.LogLevel cannot be empty")
	}
	return nil
}