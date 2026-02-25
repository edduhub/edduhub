package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/college"
	"eduhub/server/internal/services/student"
	"eduhub/server/internal/services/user"

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

// AuthMiddleware uses AuthService to perform authentication (via Kratos)
// and authorization (via Ory Keto) checks.
type AuthMiddleware struct {
	AuthService    auth.AuthService
	StudentService student.StudentService
	CollegeService college.CollegeService
	UserService    user.UserService
}

// NewAuthMiddleware now accepts an auth.AuthService instance,
// ensuring that the middleware has access to both authentication
// (session validation) and authorization (permission checking) logic.
func NewAuthMiddleware(authSvc auth.AuthService, studentService student.StudentService, collegeService college.CollegeService, userService user.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		AuthService:    authSvc,
		StudentService: studentService,
		CollegeService: collegeService,
		UserService:    userService,
	}
}

func (m *AuthMiddleware) setIdentityContext(c echo.Context, identity *auth.Identity) error {
	if identity == nil {
		return fmt.Errorf("identity is nil")
	}

	c.Set(identityContextKey, identity)

	if m.UserService == nil {
		if identity.UserID > 0 {
			c.Set("user_id", identity.UserID)
			return nil
		}
		return fmt.Errorf("user service is not configured")
	}

	userRecord, err := m.UserService.GetUserByKratosID(c.Request().Context(), identity.ID)
	if err != nil {
		return err
	}
	if !userRecord.IsActive {
		return fmt.Errorf("user account is inactive")
	}

	identity.UserID = userRecord.ID
	c.Set("user_id", userRecord.ID)
	return nil
}

// ValidateJWT checks if the JWT token provided in the Authorization header
// is valid. This replaces session-based validation with JWT token validation.
func (m *AuthMiddleware) ValidateJWT(next echo.HandlerFunc) echo.HandlerFunc {
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

		jwtToken := authHeader[len(bearerPrefix):]
		if jwtToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Empty JWT token",
			})
		}

		identity, err := m.AuthService.ValidateJWT(c.Request().Context(), jwtToken)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid JWT token: " + err.Error(),
			})
		}

		if err := m.setIdentityContext(c, identity); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unable to resolve authenticated user",
			})
		}

		return next(c)
	}
}

// RequireCollege ensures that the authenticated user belongs to the specified college.
// It extracts the collegeID from the identity and looks up the integer ID from the database.
// Under a multitenant setup, this helps isolate college-specific resources.
func (m *AuthMiddleware) RequireCollege(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		identity, ok := c.Get("identity").(*auth.Identity)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
		}

		// Get the external college ID from identity (e.g., "COL456")
		externalCollegeID := identity.Traits.College.ID

		// Look up the college by external ID to get the integer ID
		college, err := m.CollegeService.GetCollegeByExternalID(c.Request().Context(), externalCollegeID)
		if err != nil {
			// If not found by external ID, try to parse as integer directly
			collegeIDInt, parseErr := strconv.Atoi(externalCollegeID)
			if parseErr != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "Invalid college ID",
				})
			}
			c.Set("college_id", collegeIDInt)
		} else {
			c.Set("college_id", college.ID)
		}

		return next(c)
	}
}

func (m *AuthMiddleware) LoadStudentProfile(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		identity, ok := c.Get(identityContextKey).(*auth.Identity)
		if !ok || identity == nil {
			return helpers.Error(c, "Unauthorized", 403)
		}

		// Only load student profile for student role - faculty/admin should get student IDs from URL params
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

			// Set the student ID in context for ExtractStudentID helper
			c.Set(studentIDContextKey, student.StudentID)
		}

		return next(c)
	}
}

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

func (m *AuthMiddleware) RequirePermission(subject, resource, action string) echo.MiddlewareFunc {
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

			// --- Check based on Role ---
			userRole := identity.Traits.Role

			if userRole == RoleStudent {
				// If the user is a student, they MUST be accessing their own record.
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
				// Student is accessing their own data - Allow
				return next(c)

			} else if userRole == RoleAdmin || userRole == RoleFaculty {
				// Admin and faculty are allowed based on role. For finer-grained
				// checks, use Keto permission checks in the handler or dedicated middleware.
				return next(c)
			}

			// If role is none of the above (or empty), deny access.
			return helpers.Error(c, "Forbidden - Invalid role for accessing student data", http.StatusForbidden)
		}
	}
}
