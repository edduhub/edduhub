package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"eduhub/server/internal/models"

	"github.com/labstack/echo/v4"
)

// AuditMiddleware creates middleware that logs API operations
func AuditMiddleware(auditService AuditService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip audit logging for certain endpoints
			if shouldSkipAudit(c) {
				return next(c)
			}

			// Extract context information
			userID, _ := extractUserID(c)
			collegeID, _ := extractCollegeID(c)
			ipAddress := c.RealIP()
			userAgent := c.Request().UserAgent()

			// Capture request body for POST/PUT/PATCH
			var requestBody map[string]interface{}
			if shouldCaptureBody(c) {
				if body, err := captureRequestBody(c); err == nil {
					requestBody = body
				}
			}

			// Create audit log entry
			auditLog := &models.AuditLog{
				CollegeID:  collegeID,
				UserID:     userID,
				Action:     getActionFromMethod(c.Request().Method),
				EntityType: getEntityTypeFromPath(c.Request().URL.Path),
				EntityID:   getEntityIDFromPath(c.Request().URL.Path),
				IPAddress:  ipAddress,
				UserAgent:  userAgent,
			}

			// Add request data to changes
			if requestBody != nil {
				auditLog.Changes = map[string]interface{}{
					"request_body": requestBody,
					"method":       c.Request().Method,
					"path":         c.Request().URL.Path,
					"query":        c.Request().URL.RawQuery,
				}
			}

			// Log the action asynchronously to not block the request
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := auditService.LogAction(ctx, auditLog); err != nil {
					// Log the error but don't fail the request
					c.Logger().Error("Failed to log audit event:", err)
				}
			}()

			// Continue with the request
			return next(c)
		}
	}
}

// shouldSkipAudit determines if the request should be excluded from audit logging
func shouldSkipAudit(c echo.Context) bool {
	path := c.Request().URL.Path

	// Skip health checks, static files, and auth endpoints
	skipPaths := []string{
		"/health",
		"/ready",
		"/alive",
		"/swagger",
		"/docs",
		"/auth/login",
		"/auth/register",
		"/auth/callback",
		"/auth/logout",
		"/auth/password-reset",
		"/auth/verify-email",
		"/api/notifications/ws", // WebSocket connections
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// shouldCaptureBody determines if request body should be captured
func shouldCaptureBody(c echo.Context) bool {
	method := c.Request().Method
	return method == "POST" || method == "PUT" || method == "PATCH"
}

// captureRequestBody captures the request body for audit logging
func captureRequestBody(c echo.Context) (map[string]interface{}, error) {
	// Read the body
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}

	// Restore the body for the next handler
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Try to parse as JSON
	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		// If not JSON, store as string
		body = map[string]interface{}{
			"raw_body": string(bodyBytes),
		}
	}

	// Remove sensitive fields
	sanitizeBody(body)

	return body, nil
}

// sanitizeBody removes sensitive information from the request body
func sanitizeBody(body map[string]interface{}) {
	sensitiveFields := []string{"password", "token", "secret", "key", "api_key", "access_token"}

	for _, field := range sensitiveFields {
		if _, exists := body[field]; exists {
			body[field] = "***REDACTED***"
		}
	}

	// Recursively sanitize nested objects
	for _, value := range body {
		if nested, ok := value.(map[string]interface{}); ok {
			sanitizeBody(nested)
		}
	}
}

// getActionFromMethod converts HTTP method to audit action
func getActionFromMethod(method string) string {
	switch method {
	case "GET":
		return "READ"
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// getEntityTypeFromPath extracts entity type from URL path
func getEntityTypeFromPath(path string) string {
	// Extract entity type from API paths like /api/users, /api/courses/123, etc.
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 2 || parts[0] != "api" {
		return "unknown"
	}

	// Handle different path patterns
	switch {
	case len(parts) >= 2 && parts[1] == "users":
		return "user"
	case len(parts) >= 2 && parts[1] == "students":
		return "student"
	case len(parts) >= 2 && parts[1] == "courses":
		return "course"
	case len(parts) >= 2 && parts[1] == "lectures":
		return "lecture"
	case len(parts) >= 2 && parts[1] == "attendance":
		return "attendance"
	case len(parts) >= 2 && parts[1] == "grades":
		return "grade"
	case len(parts) >= 2 && parts[1] == "assignments":
		return "assignment"
	case len(parts) >= 2 && parts[1] == "quizzes":
		return "quiz"
	case len(parts) >= 2 && parts[1] == "announcements":
		return "announcement"
	case len(parts) >= 2 && parts[1] == "notifications":
		return "notification"
	case len(parts) >= 2 && parts[1] == "profile":
		return "profile"
	case len(parts) >= 2 && parts[1] == "departments":
		return "department"
	case len(parts) >= 2 && parts[1] == "college":
		return "college"
	case len(parts) >= 2 && parts[1] == "analytics":
		return "analytics"
	case len(parts) >= 2 && parts[1] == "reports":
		return "report"
	case len(parts) >= 2 && parts[1] == "webhooks":
		return "webhook"
	case len(parts) >= 2 && parts[1] == "audit":
		return "audit"
	case len(parts) >= 2 && parts[1] == "files":
		return "file"
	case len(parts) >= 2 && parts[1] == "file-management":
		return "file"
	case len(parts) >= 2 && parts[1] == "folders":
		return "folder"
	default:
		return "unknown"
	}
}

// getEntityIDFromPath extracts entity ID from URL path
func getEntityIDFromPath(path string) int {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Look for numeric IDs in the path
	for _, part := range parts {
		if id, err := parseIntSafe(part); err == nil && id > 0 {
			return id
		}
	}

	return 0
}

// parseIntSafe safely parses string to int
func parseIntSafe(s string) (int, error) {
	// Simple implementation - in real code you'd use strconv.Atoi
	if s == "" {
		return 0, nil
	}

	// Check if all characters are digits
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, echo.ErrBadRequest
		}
	}

	// Convert to int (simplified)
	result := 0
	for _, r := range s {
		result = result*10 + int(r-'0')
	}

	return result, nil
}

// extractUserID extracts user ID from context (simplified version)
func extractUserID(c echo.Context) (int, error) {
	if userID := c.Get("user_id"); userID != nil {
		if id, ok := userID.(int); ok {
			return id, nil
		}
	}
	if userID := c.Get("userID"); userID != nil {
		if id, ok := userID.(int); ok {
			return id, nil
		}
	}
	return 0, echo.ErrUnauthorized
}

// extractCollegeID extracts college ID from context (simplified version)
func extractCollegeID(c echo.Context) (int, error) {
	if collegeID := c.Get("college_id"); collegeID != nil {
		if id, ok := collegeID.(int); ok {
			return id, nil
		}
	}
	if collegeID := c.Get("collegeID"); collegeID != nil {
		if id, ok := collegeID.(int); ok {
			return id, nil
		}
	}
	return 0, echo.ErrBadRequest
}
