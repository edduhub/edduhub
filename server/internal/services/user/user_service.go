package user

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

type UserService interface {
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
	UpdateUserPartial(ctx context.Context, userID int, req *models.UpdateUserRequest) error
	GetUserByKratosID(ctx context.Context, kratosID string) (*models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	validate *validator.Validate
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
		validate: validator.New(),
	}
}

func (u *userService) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	return u.userRepo.GetUserByID(ctx, userID)
}

func (u *userService) GetUserByKratosID(ctx context.Context, kratosID string) (*models.User, error) {
	return u.userRepo.GetUserByKratosID(ctx, kratosID)
}

func (u *userService) UpdateUserPartial(ctx context.Context, userID int, req *models.UpdateUserRequest) error {
	if err := u.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed for user update: %w", err)
	}
	return u.userRepo.UpdateUserPartial(ctx, userID, req)
}