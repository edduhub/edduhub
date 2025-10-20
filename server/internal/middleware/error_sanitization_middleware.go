package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// ErrorSanitizationMiddleware sanitizes error responses to prevent information leakage
// in production environments while allowing detailed errors in development
type ErrorSanitizationMiddleware struct {
	IsProduction bool
}

// NewErrorSanitizationMiddleware creates a new error sanitization middleware
func NewErrorSanitizationMiddleware() *ErrorSanitizationMiddleware {
	isProduction := os.Getenv("APP_ENV") == "production" || os.Getenv("APP_DEBUG") == "false"
	return &ErrorSanitizationMiddleware{
		IsProduction: isProduction,
	}
}

// Middleware wraps error handling to sanitize responses
func (m *ErrorSanitizationMiddleware) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		// Log the full error for debugging
		log.Error().
			Err(err).
			Str("path", c.Request().URL.Path).
			Str("method", c.Request().Method).
			Msg("Request error")

		// If not production, return the actual error
		if !m.IsProduction {
			return err
		}

		// In production, sanitize errors
		sanitizedErr := m.sanitizeError(err)
		return sanitizedErr
	}
}

// sanitizeError sanitizes error messages to prevent information leakage
func (m *ErrorSanitizationMiddleware) sanitizeError(err error) error {
	if he, ok := err.(*echo.HTTPError); ok {
		// Keep HTTP error codes but sanitize messages
		sanitizedMessage := m.sanitizeMessage(he.Message)
		return echo.NewHTTPError(he.Code, sanitizedMessage)
	}

	// For non-HTTP errors, return generic internal server error
	return echo.NewHTTPError(http.StatusInternalServerError, "An internal error occurred")
}

// sanitizeMessage removes sensitive information from error messages
func (m *ErrorSanitizationMiddleware) sanitizeMessage(message interface{}) interface{} {
	if msg, ok := message.(string); ok {
		// Remove database-specific errors
		if strings.Contains(strings.ToLower(msg), "sql") ||
			strings.Contains(strings.ToLower(msg), "postgres") ||
			strings.Contains(strings.ToLower(msg), "database") ||
			strings.Contains(strings.ToLower(msg), "query") {
			return "A database error occurred"
		}

		// Remove file path information
		if strings.Contains(msg, "/") || strings.Contains(msg, "\\") {
			return "An error occurred while processing your request"
		}

		// Remove stack traces
		if strings.Contains(msg, "goroutine") || strings.Contains(msg, "panic") {
			return "An internal error occurred"
		}

		// Remove connection errors that might reveal infrastructure
		if strings.Contains(strings.ToLower(msg), "connection") ||
			strings.Contains(strings.ToLower(msg), "timeout") ||
			strings.Contains(strings.ToLower(msg), "dial") {
			return "A connectivity error occurred"
		}

		return msg
	}

	// For non-string messages, check if they're map and sanitize
	if msgMap, ok := message.(map[string]interface{}); ok {
		sanitized := make(map[string]interface{})
		for k, v := range msgMap {
			if strVal, ok := v.(string); ok {
				sanitized[k] = m.sanitizeMessage(strVal)
			} else {
				sanitized[k] = v
			}
		}
		return sanitized
	}

	return message
}

// RecoverMiddleware handles panics and prevents application crashes
func (m *ErrorSanitizationMiddleware) RecoverMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = echo.NewHTTPError(http.StatusInternalServerError, r)
				}

				// Log the panic
				log.Error().
					Err(err).
					Interface("panic", r).
					Str("path", c.Request().URL.Path).
					Str("method", c.Request().Method).
					Msg("Panic recovered")

				// Send sanitized error response
				if m.IsProduction {
					c.JSON(http.StatusInternalServerError, map[string]string{
						"error": "An internal error occurred",
					})
				} else {
					c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"error": "Internal server error",
						"panic": r,
					})
				}
			}
		}()
		return next(c)
	}
}
