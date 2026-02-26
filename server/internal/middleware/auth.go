package middleware

import (
	"net/http"
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

const (
	RoleAdmin   = "admin"
	RoleFaculty = "faculty"
	RoleStudent = "student"
	RoleParent  = "parent"

	identityContextKey  = "identity"
	collegeIDContextKey = "college_id"
	studentIDContextKey = "student_id"
	facultyIDContextKey = "faculty_id"

	AttendanceResource = "attendance"
	MarkAction         = "mark"
)

// AuthMiddleware handles authentication using Ory Hydra tokens
// and authorization using Ory Keto
type AuthMiddleware struct {
	AuthService    auth.AuthService
	StudentService student.StudentService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authSvc auth.AuthService, studentService student.StudentService) *AuthMiddleware {
	return &AuthMiddleware{
		AuthService:    authSvc,
		StudentService: studentService,
	}
}

// ValidateToken validates the OAuth2 access token from the Authorization header
// and sets the identity in the context
func (m *AuthMiddleware) ValidateToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "No authorization header provided",
			})
		}

		// Extract token from Bearer header
		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
		}

		accessToken := authHeader[len(bearerPrefix):]
		if accessToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Empty access token",
			})
		}

		// Validate the token using Hydra introspection
		identity, err := m.AuthService.ValidateToken(c.Request().Context(), accessToken)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid access token: " + err.Error(),
			})
		}

		// Set identity in context
		c.Set(identityContextKey, identity)

		// Set user ID if available
		if identity.UserID > 0 {
			c.Set("user_id", identity.UserID)
		}

		return next(c)
	}
}

// RequireCollege ensures that the authenticated user belongs to the specified college
func (m *AuthMiddleware) RequireCollege(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		identity, ok := c.Get("identity").(*auth.Identity)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
		}

		// Get the external college ID from identity
		externalCollegeID := identity.Traits.College.ID
		if externalCollegeID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "No college associated with identity",
			})
		}

		// Try to parse as integer directly
		collegeIDInt, err := strconv.Atoi(externalCollegeID)
		if err != nil {
			// If not a direct integer, it might be an external ID
			// For now, just set the value as-is
			c.Set(collegeIDContextKey, externalCollegeID)
		} else {
			c.Set(collegeIDContextKey, collegeIDInt)
		}

		return next(c)
	}
}

// LoadStudentProfile loads the student profile for student users
func (m *AuthMiddleware) LoadStudentProfile(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		identity, ok := c.Get(identityContextKey).(*auth.Identity)
		if !ok || identity == nil {
			return helpers.Error(c, "Unauthorized", 403)
		}

		// Only load student profile for student role
		if identity.Traits.Role == RoleStudent {
			ctx := c.Request().Context()
			kratosID := identity.ID

			student, err := m.StudentService.FindByKratosID(ctx, kratosID)
			if err != nil {
				return helpers.Error(c, "Student lookup failed", 500)
			}
			if student == nil {
				return helpers.Error(c, "Student not registered", 401)
			}
			if !student.IsActive {
				return helpers.Error(c, "Student account inactive", 401)
			}

			// Set the student ID in context
			c.Set(studentIDContextKey, student.StudentID)
		}

		return next(c)
	}
}

// RequireRole checks if the authenticated user has one of the specified roles
func (m *AuthMiddleware) RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identity, ok := c.Get("identity").(*auth.Identity)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized",
				})
			}

			for _, role := range roles {
				if m.AuthService.HasRole(identity, role) {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Insufficient permissions",
			})
		}
	}
}

// RequirePermission checks if the authenticated user has the specified permission
// using Ory Keto
func (m *AuthMiddleware) RequirePermission(resource, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identity, ok := c.Get("identity").(*auth.Identity)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Unauthorized",
				})
			}

			allowed, err := m.AuthService.CheckPermission(c.Request().Context(), identity, action, resource)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Error checking permissions",
				})
			}

			if !allowed {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions",
				})
			}

			return next(c)
		}
	}
}

// VerifyStudentOwnership ensures a student can only access their own resources
func (m *AuthMiddleware) VerifyStudentOwnership() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identity, ok := c.Get(identityContextKey).(*auth.Identity)
			if !ok || identity == nil {
				return helpers.Error(c, "Unauthorized - Identity required for ownership check", http.StatusUnauthorized)
			}

			// Get the student ID from the URL path parameter
			requestedStudentIDStr := c.Param("studentID")
			if requestedStudentIDStr == "" {
				return helpers.Error(c, "Bad Request - Missing studentID path parameter", http.StatusBadRequest)
			}
			requestedStudentID, err := strconv.Atoi(requestedStudentIDStr)
			if err != nil || requestedStudentID <= 0 {
				return helpers.Error(c, "Bad Request - Invalid studentID path parameter", http.StatusBadRequest)
			}

			// Check based on Role
			userRole := identity.Traits.Role

			if userRole == RoleStudent {
				// Student must access their own record
				authenticatedStudentIDRaw := c.Get(studentIDContextKey)
				if authenticatedStudentIDRaw == nil {
					return helpers.Error(c, "Unauthorized - Student identity not loaded", http.StatusUnauthorized)
				}

				authenticatedStudentID, ok := authenticatedStudentIDRaw.(int)
				if !ok {
					return helpers.Error(c, "Internal error - Invalid student ID format", http.StatusInternalServerError)
				}

				if requestedStudentID != authenticatedStudentID {
					return helpers.Error(c, "Forbidden - Students can only access their own data", http.StatusForbidden)
				}
				return next(c)

			} else if userRole == RoleAdmin || userRole == RoleFaculty {
				// Admin and faculty are allowed based on role
				return next(c)
			}

			return helpers.Error(c, "Forbidden - Invalid role for accessing student data", http.StatusForbidden)
		}
	}
}

// Helper function to get identity from context
func GetIdentity(c echo.Context) (*auth.Identity, bool) {
	identity, ok := c.Get(identityContextKey).(*auth.Identity)
	return identity, ok
}

// Helper function to get student ID from context
func GetStudentID(c echo.Context) (int, bool) {
	studentID, ok := c.Get(studentIDContextKey).(int)
	return studentID, ok
}

// Helper function to get college ID from context
func GetCollegeID(c echo.Context) (any, bool) {
	collegeID := c.Get(collegeIDContextKey)
	if collegeID == nil {
		return nil, false
	}
	return collegeID, true
}
