package middleware

import (
	"eduhub/server/internal/services"
)

type Middleware struct {
	Auth *AuthMiddleware
	// other middleware
}

func NewMiddleware(services *services.Services) *Middleware {
	authSvc := services.Auth

	studentService := services.StudentService
	return &Middleware{

		Auth: NewAuthMiddleware(authSvc, studentService),
	}

}
