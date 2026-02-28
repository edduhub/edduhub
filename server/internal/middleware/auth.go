package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/auth"

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

// TokenValidator is the minimal interface that AuthMiddleware requires from the
// auth service. It is satisfied by auth.AuthService but is kept small so that
// test mocks need only implement the three methods that the middleware actually calls.
type TokenValidator interface {
	// ValidateJWT validates a locally-signed JWT and returns the Identity.
	ValidateJWT(ctx context.Context, token string) (*auth.Identity, error)
	// ValidateToken validates a Hydra / OAuth2 access token and returns the Identity.
	ValidateToken(ctx context.Context, accessToken string) (*auth.Identity, error)
	// HasRole returns true when the identity carries the given role.
	HasRole(identity *auth.Identity, role string) bool
	// CheckPermission calls Keto to verify a subject–action–resource tuple.
	CheckPermission(ctx context.Context, identity *auth.Identity, action, resource string) (bool, error)
}

// StudentLoader is the minimal interface required to load a student profile.
// Satisfied by student.StudentService.
type StudentLoader interface {
	FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error)
}

// AuthMiddleware handles authentication (Hydra/JWT), multi-tenant college isolation,
// role-based access control (Keto), and student profile loading.
type AuthMiddleware struct {
	// AuthService provides token validation, role checks, and Keto permission checks.
	AuthService TokenValidator
	// StudentService provides student profile look-ups.
	StudentService StudentLoader
	// hydraService is the optional Hydra client used by ValidateToken.
	hydraService auth.HydraService
	// jwtManager is the optional local JWT manager used by ValidateJWT.
	jwtManager auth.JWTManager
}

// NewAuthMiddleware creates a new AuthMiddleware.
//
// hydra and jwtMgr are optional; pass nil when not needed (e.g. in tests).
func NewAuthMiddleware(authSvc TokenValidator, studentSvc StudentLoader, hydra auth.HydraService, jwtMgr auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		AuthService:    authSvc,
		StudentService: studentSvc,
		hydraService:   hydra,
		jwtManager:     jwtMgr,
	}
}

// extractBearer strips the "Bearer " prefix from an Authorization header value.
func extractBearer(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

// ValidateToken validates the Hydra OAuth2 access token from the Authorization header
// and sets the identity in the context. This is the primary middleware for all API routes.
func (m *AuthMiddleware) ValidateToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := extractBearer(c.Request().Header.Get("Authorization"))
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid Authorization header")
		}

		identity, err := m.AuthService.ValidateToken(c.Request().Context(), token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid access token: "+err.Error())
		}

		c.Set(identityContextKey, identity)
		if identity.UserID > 0 {
			c.Set("user_id", identity.UserID)
		}

		return next(c)
	}
}

// ValidateJWT validates a locally-signed JWT Bearer token.
// Use this middleware for routes that bypass Hydra introspection.
func (m *AuthMiddleware) ValidateJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := extractBearer(c.Request().Header.Get("Authorization"))
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid Authorization header")
		}

		identity, err := m.AuthService.ValidateJWT(c.Request().Context(), token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid JWT token: "+err.Error())
		}

		c.Set(identityContextKey, identity)
		if identity.UserID > 0 {
			c.Set("user_id", identity.UserID)
		}

		return next(c)
	}
}

// RequireCollege ensures that the authenticated user's college is set in the context.
// It parses the College.ID string from the identity and stores the integer value under
// the "college_id" context key so downstream handlers can use it for tenant isolation.
func (m *AuthMiddleware) RequireCollege(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		identity, ok := c.Get("identity").(*auth.Identity)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		// Get the external college ID from identity
		externalCollegeID := identity.Traits.College.ID
		if externalCollegeID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "no college associated with identity")
		}

		// Try to parse as integer directly
		collegeIDInt, err := strconv.Atoi(externalCollegeID)
		if err != nil {
			// Not a numeric ID – store as-is (e.g. Kratos external UUID)
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
			return echo.NewHTTPError(http.StatusForbidden, "unauthorized")
		}

		// Only load student profile for student role
		if identity.Traits.Role == RoleStudent {
			ctx := c.Request().Context()
			kratosID := identity.ID

			student, err := m.StudentService.FindByKratosID(ctx, kratosID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "student lookup failed")
			}
			if student == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "student not registered")
			}
			if !student.IsActive {
				return echo.NewHTTPError(http.StatusUnauthorized, "student account inactive")
			}

			// Set the student ID in context
			c.Set(studentIDContextKey, student.StudentID)
		}

		return next(c)
	}
}

// RequireRole checks if the authenticated user has one of the specified roles.
// Uses Kratos role claim from the Identity – no Keto call needed for role checks.
func (m *AuthMiddleware) RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identity, ok := c.Get("identity").(*auth.Identity)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			for _, role := range roles {
				if m.AuthService.HasRole(identity, role) {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		}
	}
}

// RequirePermission checks if the authenticated user has the specified permission
// using Ory Keto's relationship-based access control.
func (m *AuthMiddleware) RequirePermission(subject, resource, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identity, ok := c.Get("identity").(*auth.Identity)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}

			allowed, err := m.AuthService.CheckPermission(c.Request().Context(), identity, action, resource)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "error checking permissions"})
			}

			if !allowed {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
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
