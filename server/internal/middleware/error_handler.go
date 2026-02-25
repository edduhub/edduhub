package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string         `json:"error"`
	Message string         `json:"message"`
	Code    string         `json:"code,omitempty"`
	Details map[string]any `json:"details,omitempty"`
	Status  int            `json:"status"`
}

// AppError represents an application error with additional context
type AppError struct {
	Status  int
	Code    string
	Message string
	Details map[string]any
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(status int, code, message string, err error) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]any) *AppError {
	e.Details = details
	return e
}

// Common error constructors
func BadRequestError(message string, err error) *AppError {
	return NewAppError(http.StatusBadRequest, "BAD_REQUEST", message, err)
}

func UnauthorizedError(message string, err error) *AppError {
	return NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", message, err)
}

func ForbiddenError(message string, err error) *AppError {
	return NewAppError(http.StatusForbidden, "FORBIDDEN", message, err)
}

func NotFoundError(message string, err error) *AppError {
	return NewAppError(http.StatusNotFound, "NOT_FOUND", message, err)
}

func ConflictError(message string, err error) *AppError {
	return NewAppError(http.StatusConflict, "CONFLICT", message, err)
}

func ValidationError(message string, details map[string]any) *AppError {
	return NewAppError(http.StatusUnprocessableEntity, "VALIDATION_ERROR", message, nil).WithDetails(details)
}

func InternalServerError(message string, err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, err)
}

func ServiceUnavailableError(message string, err error) *AppError {
	return NewAppError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, err)
}

// ErrorHandlerMiddleware provides centralized error handling
func ErrorHandlerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			// Check if response has already been sent
			if c.Response().Committed {
				return err
			}

			// Handle AppError
			if appErr, ok := err.(*AppError); ok {
				return c.JSON(appErr.Status, ErrorResponse{
					Error:   appErr.Code,
					Message: appErr.Message,
					Code:    appErr.Code,
					Details: appErr.Details,
					Status:  appErr.Status,
				})
			}

			// Handle Echo HTTPError
			if he, ok := err.(*echo.HTTPError); ok {
				status := he.Code
				message := fmt.Sprintf("%v", he.Message)

				// Determine error code based on status
				code := getErrorCode(status)

				return c.JSON(status, ErrorResponse{
					Error:   code,
					Message: message,
					Code:    code,
					Status:  status,
				})
			}

			// Handle JSON parsing errors
			if _, ok := err.(*json.UnmarshalTypeError); ok {
				return c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "BAD_REQUEST",
					Message: "Invalid JSON format in request body",
					Code:    "INVALID_JSON",
					Status:  http.StatusBadRequest,
				})
			}

			// Default to internal server error for unknown errors
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "INTERNAL_SERVER_ERROR",
				Message: "An unexpected error occurred. Please try again later.",
				Code:    "INTERNAL_ERROR",
				Status:  http.StatusInternalServerError,
			})
		}
	}
}

// getErrorCode returns an error code based on HTTP status
func getErrorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "VALIDATION_ERROR"
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	default:
		return "UNKNOWN_ERROR"
	}
}

// RecoverMiddleware recovers from panics and returns a proper error response
func RecoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					// Log the panic
					c.Logger().Error("Panic recovered: ", err)

					// Return internal server error
					c.JSON(http.StatusInternalServerError, ErrorResponse{
						Error:   "INTERNAL_SERVER_ERROR",
						Message: "An unexpected error occurred. Please try again later.",
						Code:    "PANIC_RECOVERED",
						Status:  http.StatusInternalServerError,
					})
				}
			}()
			return next(c)
		}
	}
}
