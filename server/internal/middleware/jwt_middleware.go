package middleware

import (
	"net/http"

	"eduhub/server/internal/services/auth"

	"github.com/labstack/echo/v4"
)

// JWTMiddleware handles JWT token validation for authentication.
// It provides middleware functions to validate JWT tokens from the Authorization header
// and store the authenticated identity in the request context for use by subsequent middleware.
type JWTMiddleware struct {
	// AuthService provides authentication services including JWT validation
	AuthService auth.AuthService
}

// NewJWTMiddleware creates a new instance of JWTMiddleware with the provided AuthService.
// The AuthService is required for JWT token validation operations.
func NewJWTMiddleware(authSvc auth.AuthService) *JWTMiddleware {
	return &JWTMiddleware{
		AuthService: authSvc,
	}
}

// ValidateJWT is a middleware function that validates JWT tokens provided in the Authorization header.
// It expects the header to be in the format "Bearer <token>" and validates the token using the AuthService.
// Upon successful validation, it stores the identity in the request context using the identityContextKey.
//
// This middleware should be used early in the middleware chain to authenticate requests before
// other middleware that depend on the authenticated identity.
//
// Error responses:
// - 401 Unauthorized: When no authorization header is provided, header format is invalid,
//   token is empty, or token validation fails
func (m *JWTMiddleware) ValidateJWT(next echo.HandlerFunc) echo.HandlerFunc {
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