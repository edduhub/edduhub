package middleware

import (
	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

// CollegeMiddleware handles college-based isolation and student profile loading.
// It ensures users belong to the correct college and loads student profiles for
// student users, setting appropriate context values for downstream middleware.
type CollegeMiddleware struct {
	// AuthService provides authentication services
	AuthService auth.AuthService
	// StudentService provides student-related services for profile loading
	StudentService student.StudentService
}

// NewCollegeMiddleware creates a new instance of CollegeMiddleware with the provided services.
// Both AuthService and StudentService are required for college isolation and profile operations.
func NewCollegeMiddleware(authSvc auth.AuthService, studentSvc student.StudentService) *CollegeMiddleware {
	return &CollegeMiddleware{
		AuthService:    authSvc,
		StudentService: studentSvc,
	}
}

// RequireCollege is a middleware function that ensures the authenticated user belongs to the specified college.
// It extracts the college ID from the user's identity traits and stores it in the request context
// for use by other middleware and handlers. This helps isolate college-specific resources
// in a multitenant setup.
//
// This middleware should be used after authentication middleware to ensure the identity is available.
//
// The college ID is stored in the context using the collegeIDContextKey for later retrieval.
//
// Error responses:
// - 401 Unauthorized: When no identity is found in the context
func (m *CollegeMiddleware) RequireCollege(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		identity, ok := c.Get(identityContextKey).(*auth.Identity)
		if !ok {
			return c.JSON(401, map[string]string{
				"error": "Unauthorized",
			})
		}
		userCollegeID := identity.Traits.College.ID
		c.Set(collegeIDContextKey, userCollegeID)

		return next(c)
	}
}

// LoadStudentProfile is a middleware function that loads the student profile for authenticated student users.
// It checks the user's role and if they are a student, retrieves their profile from the database
// using their Kratos ID. The student ID is then stored in the request context for use by
// other middleware and handlers.
//
// This middleware should be used after authentication middleware and is particularly important
// for ownership verification middleware that needs to compare student IDs.
//
// For non-student users (admin, faculty), this middleware simply passes through without action.
//
// Error responses:
// - 403 Unauthorized: When identity is missing, student profile cannot be found,
//   student is not registered, or student account is inactive
func (m *CollegeMiddleware) LoadStudentProfile(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		identity, ok := c.Get(identityContextKey).(*auth.Identity)
		if !ok || identity == nil {
			return helpers.Error(c, "Unauthorized", 403)
		}
		ctx := c.Request().Context()
		kratosID := identity.ID
		if identity.Traits.Role == RoleStudent {
			student, err := m.StudentService.FindByKratosID(ctx, kratosID)
			if err != nil {
				return helpers.Error(c, "Unauthorized", 403)
			}
			if student == nil {
				return helpers.Error(c, "Not registered", 401)
			}
			if !student.IsActive {
				return helpers.Error(c, "Inactive", 401)
			}
			c.Set(studentIDContextKey, student.StudentID)
		}
		return next(c)
	}
}