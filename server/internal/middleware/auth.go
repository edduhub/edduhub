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
	AuthService auth.AuthService
	// StudentRepo repository.StudentRepository
	StudentService student.StudentService
}

// NewAuthMiddleware now accepts an auth.AuthService instance,
// ensuring that the middleware has access to both authentication
// (session validation) and authorization (permission checking) logic.
func NewAuthMiddleware(authSvc auth.AuthService, studentService student.StudentService) *AuthMiddleware {
	return &AuthMiddleware{
		AuthService:    authSvc,
		StudentService: studentService,
	}
}

// ValidateSession checks if the session token provided in the request
// is valid. The AuthService.ValidateSession function should use Ory Kratos
// to validate the session.
func (m *AuthMiddleware) ValidateSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionToken := c.Request().Header.Get("X-Session-Token")
		if sessionToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "No session token provided",
			})
		}

		identity, err := m.AuthService.ValidateSession(c.Request().Context(), sessionToken)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid session",
			})
		}

		// Store identity in context for later use by other middleware handlers.
		c.Set(identityContextKey, identity)
		return next(c)
	}
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

		// Store identity in context for later use by other middleware handlers.
		c.Set(identityContextKey, identity)
		return next(c)
	}
}

// RequireCollege ensures that the authenticated user belongs to the specified college.
// It extracts the collegeID from URL parameters and then calls AuthService.CheckCollegeAccess.
// Under a multitenant setup, this helps isolate college-specific resources.
func (m *AuthMiddleware) RequireCollege(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		identity, ok := c.Get("identity").(*auth.Identity)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Unauthorized",
			})
		}
		userCollegeID := identity.Traits.College.ID
		c.Set("college_id", userCollegeID)

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
// func (m *AuthMiddleware) VerifyStudentOwnership(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		identity, ok := c.Get(identityContextKey).(*auth.Identity)
// 		if !ok || identity == nil {
// 			return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
// 		}
// 		requestedStudentIDStr :=c.Param("studentID")
// 		if requestedStudentIDStr ==  " "{
// 			return helpers.Error(c,"Bad request",400)
// 		}
// 		requestedStudentID,err :=strconv.AToi(requestedStudentIDStr)

// 		// Get the authenticated student's ID from context
// 		authenticatedStudentID := c.Get(studentIDContextKey)
// 		if authenticatedStudentID == nil {
// 			return helpers.Error(c, "Student context not found", http.StatusUnauthorized)
// 		}

// 		// Get the requested student ID from params/query
// 		requestedStudentID, err := helpers.ExtractStudentID(c)
// 		if err != nil {
// 			return helpers.Error(c, "Invalid student ID", http.StatusBadRequest)
// 		}

// 		// Verify if the authenticated student is accessing their own resource
// 		if requestedStudentID != authenticatedStudentID.(int) {
// 			// Check if the user has admin/faculty role that allows them to override
// 			allowed, err := m.AuthService.CheckPermission(c.Request().Context(), identity, strconv.Itoa(requestedStudentID), MarkAction, AttendanceResource)
// 			if err != nil || !allowed {
// 				return helpers.Error(c, "Access denied", http.StatusForbidden)
// 			}
// 		}

//			return next(c)
//		}
//	}
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
				// If Admin or Faculty, allow access based on role.
				// Further checks (e.g., is faculty teaching this student's course?)
				// should be handled by Keto permission checks if needed, either here
				// or in the handler/service.
				// For now, we allow based on role.

				// Example Keto Check (Optional here, could be separate middleware or in handler):
				// subject := identity.ID // User's Kratos ID
				// resource := fmt.Sprintf("%s:%s", StudentResource, requestedStudentIDStr) // e.g., "student_data:123"
				// action := ViewAction // e.g., "view"
				// allowed, ketoErr := m.AuthService.CheckPermission(c.Request().Context(), identity, subject, resource, action)
				// if ketoErr != nil {
				//   return helpers.Error(c, "Error checking admin/faculty permission", http.StatusInternalServerError)
				// }
				// if !allowed {
				//   return helpers.Error(c, "Forbidden - You do not have permission to view this student's data", http.StatusForbidden)
				// }

				return next(c) // Allow admin/faculty
			}

			// If role is none of the above (or empty), deny access.
			return helpers.Error(c, "Forbidden - Invalid role for accessing student data", http.StatusForbidden)
		}
	}
}
