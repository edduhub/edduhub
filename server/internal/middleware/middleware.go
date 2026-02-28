package middleware

import (
	"eduhub/server/internal/services"
)

// Middleware aggregates all middleware instances used by the application.
// The zero value is valid (each field will be a nil-free *AuthMiddleware).
type Middleware struct {
	// Auth is the primary authentication + authorization middleware that
	// handles Hydra token validation, Keto permission checks, college
	// isolation and student profile loading.
	Auth *AuthMiddleware
}

// NewMiddleware creates a Middleware bundle wired to the given services.
// Passing nil is valid and produces a usable (zero-value) middleware bundle,
// which is convenient in tests.
func NewMiddleware(svc *services.Services) *Middleware {
	if svc == nil {
		return &Middleware{Auth: &AuthMiddleware{}}
	}

	authMiddleware := NewAuthMiddleware(
		svc.Auth,            // TokenValidator – the full auth.AuthService satisfies it
		svc.StudentService,  // StudentLoader  – student.StudentService satisfies it
		nil,                 // hydra: already embedded inside svc.Auth
		nil,                 // jwtManager: already embedded inside svc.Auth
	)

	return &Middleware{Auth: authMiddleware}
}
