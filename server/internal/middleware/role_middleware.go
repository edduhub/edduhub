package middleware

import (
	"net/http"

	"eduhub/server/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// RoleMiddleware handles role-based access control for authenticated users.
// It provides middleware functions to check if users have specific roles or permissions
// before allowing access to protected resources.
type RoleMiddleware struct {
	// AuthService provides authentication and authorization services
	AuthService auth.AuthService
}

// NewRoleMiddleware creates a new instance of RoleMiddleware with the provided AuthService.
// The AuthService is required for role and permission checking operations.
func NewRoleMiddleware(authSvc auth.AuthService) *RoleMiddleware {
	return &RoleMiddleware{
		AuthService: authSvc,
	}
}

// RequireRole is a middleware function that checks if the authenticated user has one of the specified roles.
// It retrieves the user's identity from the request context (set by previous authentication middleware)
// and verifies if the user has any of the required roles using the AuthService.
//
// This middleware should be used after authentication middleware (like ValidateJWT) to ensure
// the identity is available in the context.
//
// Parameters:
// - roles: A variadic list of role strings to check against the user's roles
//
// Error responses:
// - 401 Unauthorized: When no identity is found in the context
// - 403 Forbidden: When the user does not have any of the required roles
func (m *RoleMiddleware) RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identity, ok := c.Get(identityContextKey).(*auth.Identity)
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

// RequirePermission is a middleware function that checks if the authenticated user has a specific permission
// for a given resource and action. It uses Ory Keto (via AuthService) to perform fine-grained
// authorization checks beyond simple role-based access.
//
// This middleware should be used after authentication middleware to ensure the identity is available.
// It's suitable for more complex authorization scenarios where role-based checks are insufficient.
//
// Parameters:
// - subject: The subject identifier (typically the user's ID)
// - resource: The resource being accessed
// - action: The action being performed on the resource
//
// Error responses:
// - 401 Unauthorized: When no identity is found in the context
// - 500 Internal Server Error: When there's an error checking permissions
// - 403 Forbidden: When the user does not have the required permission
func (m *RoleMiddleware) RequirePermission(subject, resource, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identity, ok := c.Get(identityContextKey).(*auth.Identity)
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