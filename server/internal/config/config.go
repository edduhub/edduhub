// Package config provides centralized configuration management for the EduHub application.
// This file serves as the main configuration manager, integrating all specialized config modules.
package config

import (
	"fmt"

	"eduhub/server/internal/repository"
)

// Config represents the complete application configuration.
// It aggregates all specialized configuration modules for centralized management.
type Config struct {
	// DB holds the database connection pool and related database operations.
	// Loaded via LoadDatabase() from the database configuration module.
	DB *repository.DB

	// DBConfig contains database connection parameters.
	// Loaded via LoadDatabaseConfig() from the database configuration module.
	DBConfig *DBConfig

	// AuthConfig contains authentication service configuration.
	// Loaded via LoadAuthConfig() from the authentication configuration module.
	AuthConfig *AuthConfig

	// AppConfig contains general application settings.
	// Loaded via LoadAppConfig() from the application configuration module.
	AppConfig *AppConfig

	// RedisConfig contains Redis cache configuration.
	// Loaded via LoadRedisConfig() from the cache configuration module.
	RedisConfig *RedisConfig

	// AppPort is the port for the application server (deprecated, use AppConfig.Port).
	// Kept for backward compatibility.
	AppPort string
}

// NewConfig creates and initializes a new Config instance by loading all configuration modules.
// It performs validation on each loaded configuration and ensures all required settings are present.
//
// Returns:
//   - *Config: The fully loaded and validated configuration
//   - error: Any loading or validation errors encountered
//
// This function integrates:
//   - Database configuration (DB and DBConfig)
//   - Authentication configuration (AuthConfig)
//   - General application configuration (AppConfig)
//
// Security Considerations:
//   - All configurations are validated before returning
//   - Secure defaults are applied where appropriate
//   - Environment variables are validated for required fields
func NewConfig() (*Config, error) {
	// Load database configuration
	dbConfig, err := LoadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	// Load database connection
	db := LoadDatabase()

	// Load authentication configuration
	authConfig, err := LoadAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load auth config: %w", err)
	}

	// Load application configuration
	appConfig, err := LoadAppConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	// Load Redis configuration (optional)
	redisConfig, err := LoadRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load redis config: %w", err)
	}

	// Create the main config
	cfg := &Config{
		DB:          db,
		DBConfig:    dbConfig,
		AuthConfig:  authConfig,
		AppConfig:   appConfig,
		RedisConfig: redisConfig,
		AppPort:     appConfig.Port,
	}

	// Perform comprehensive validation
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// LoadConfig is a convenience function that calls NewConfig().
// It provides a simple interface for loading the complete application configuration.
//
// Returns:
//   - *Config: The loaded configuration
//   - error: Any errors during loading or validation
func LoadConfig() (*Config, error) {
	return NewConfig()
}

// Validate performs comprehensive validation on the entire Config.
// It calls Validate() on each sub-configuration module to ensure consistency and correctness.
//
// Returns an error if any validation fails.
func (c *Config) Validate() error {
	if c.DB == nil {
		return fmt.Errorf("Config.DB cannot be nil")
	}
	if c.DBConfig == nil {
		return fmt.Errorf("Config.DBConfig cannot be nil")
	}
	if c.AuthConfig == nil {
		return fmt.Errorf("Config.AuthConfig cannot be nil")
	}
	if c.AppConfig == nil {
		return fmt.Errorf("Config.AppConfig cannot be nil")
	}

	// Validate sub-configurations
	if err := c.DBConfig.Validate(); err != nil {
		return fmt.Errorf("DBConfig validation failed: %w", err)
	}
	if err := c.AuthConfig.Validate(); err != nil {
		return fmt.Errorf("AuthConfig validation failed: %w", err)
	}
	if err := c.AppConfig.Validate(); err != nil {
		return fmt.Errorf("AppConfig validation failed: %w", err)
	}
	if c.RedisConfig != nil {
		if err := c.RedisConfig.Validate(); err != nil {
			return fmt.Errorf("RedisConfig validation failed: %w", err)
	}
	}

	return nil
}
