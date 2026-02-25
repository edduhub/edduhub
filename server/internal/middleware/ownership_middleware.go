package middleware

import (
	"net/http"
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

// OwnershipMiddleware handles resource ownership verification for authenticated users.
// It ensures that users can only access resources they own or have permission to access,
// with special handling for student, admin, and faculty roles.
type OwnershipMiddleware struct {
	// AuthService provides authentication and authorization services
	AuthService auth.AuthService
	// StudentService provides student-related services for profile verification
	StudentService student.StudentService
}

// NewOwnershipMiddleware creates a new instance of OwnershipMiddleware with the provided services.
// Both AuthService and StudentService are required for ownership verification operations.
func NewOwnershipMiddleware(authSvc auth.AuthService, studentSvc student.StudentService) *OwnershipMiddleware {
	return &OwnershipMiddleware{
		AuthService:    authSvc,
		StudentService: studentSvc,
	}
}

// VerifyStudentOwnership is a middleware function that ensures a student can only access their own resources.
// It performs role-based checks:
// - Students can only access their own data
// - Admins and faculty have broader access but may still have restrictions
//
// This middleware expects the identity to be set in the context by previous authentication middleware,
// and for students, the student profile to be loaded by LoadStudentProfile middleware.
//
// The middleware extracts the requested student ID from the URL path parameter "studentID"
// and compares it with the authenticated user's student ID (for students) or allows access
// based on role permissions.
//
// Error responses:
// - 401 Unauthorized: When identity is missing, student context is broken, or user is not authenticated
// - 400 Bad Request: When studentID path parameter is missing or invalid
// - 403 Forbidden: When access is denied due to ownership or permission restrictions
func (m *OwnershipMiddleware) VerifyStudentOwnership() echo.MiddlewareFunc {
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
				authenticatedStudentID, err := helpers.ExtractStudentID(c) // Get logged-in student's DB ID from context
				if err != nil {
					// This means LoadStudentProfile failed or context is broken
					return helpers.Error(c, "Unauthorized - Could not verify student identity", http.StatusUnauthorized)
				}

				if requestedStudentID != authenticatedStudentID {
					return helpers.Error(c, "Forbidden - Students can only access their own data", http.StatusForbidden)
				}
				// Student is accessing their own data - Allow
				return next(c)

			} else if userRole == RoleAdmin || userRole == RoleFaculty {
				// Admin and faculty are allowed based on role. For finer-grained
				// checks (e.g., faculty teaching this student's course), use
				// Keto permission checks in the handler or a dedicated middleware.
				return next(c)
			}

			// If role is none of the above (or empty), deny access.
			return helpers.Error(c, "Forbidden - Invalid role for accessing student data", http.StatusForbidden)
		}
	}
}