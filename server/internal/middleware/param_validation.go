package middleware

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ParamValidator provides middleware for validating route parameters
type ParamValidator struct{}

// NewParamValidator creates a new parameter validator middleware
func NewParamValidator() *ParamValidator {
	return &ParamValidator{}
}

// ValidateIDParam validates that a route parameter is a valid positive integer
// Returns 400 Bad Request if validation fails
func (v *ParamValidator) ValidateIDParam(paramName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			paramValue := c.Param(paramName)
			if paramValue == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "Missing required parameter: "+paramName)
			}

			// Try to parse as integer
			id, err := strconv.Atoi(paramValue)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid "+paramName+": must be a valid integer")
			}

			// Check for negative numbers (allow 0 for some cases, but typically IDs start at 1)
			if id < 1 {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid "+paramName+": must be a positive integer")
			}

			return next(c)
		}
	}
}

// ValidateOptionalIDParam validates that an optional route parameter is a valid positive integer (if provided)
// Returns 400 Bad Request if validation fails
func (v *ParamValidator) ValidateOptionalIDParam(paramName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			paramValue := c.Param(paramName)
			
			// If param is empty, allow it (it's optional)
			if paramValue == "" {
				return next(c)
			}

			// Try to parse as integer
			id, err := strconv.Atoi(paramValue)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid "+paramName+": must be a valid integer")
			}

			// Check for negative numbers
			if id < 1 {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid "+paramName+": must be a positive integer")
			}

			return next(c)
		}
	}
}

// ValidateMultipleIDParams validates multiple route parameters as valid positive integers
// Returns 400 Bad Request if any validation fails
func (v *ParamValidator) ValidateMultipleIDParams(paramNames ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, paramName := range paramNames {
				paramValue := c.Param(paramName)
				if paramValue == "" {
					continue // Skip empty params (they might be optional)
				}

				// Try to parse as integer
				id, err := strconv.Atoi(paramValue)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "Invalid "+paramName+": must be a valid integer")
				}

				// Check for negative numbers
				if id < 1 {
					return echo.NewHTTPError(http.StatusBadRequest, "Invalid "+paramName+": must be a positive integer")
				}
			}

			return next(c)
		}
	}
}
