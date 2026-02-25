package middleware

import (
	"eduhub/server/internal/services"
)

type Middleware struct {
	Auth *AuthMiddleware
	// other middleware
}

func NewMiddleware(services *services.Services) *Middleware {
	if services == nil {
		return &Middleware{
			Auth: &AuthMiddleware{},
		}
	}

	authSvc := services.Auth

	studentService := services.StudentService
	collegeService := services.CollegeService
	userService := services.UserService
	return &Middleware{
		Auth: NewAuthMiddleware(authSvc, studentService, collegeService, userService),
	}
}
