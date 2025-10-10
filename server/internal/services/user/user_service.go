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
	ListUsers(ctx context.Context, limit, offset uint64) ([]*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, userID int) error
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

func (u *userService) ListUsers(ctx context.Context, limit, offset uint64) ([]*models.User, error) {
	return u.userRepo.FindAllUsers(ctx, limit, offset)
}

func (u *userService) CreateUser(ctx context.Context, user *models.User) error {
	if err := u.validate.Struct(user); err != nil {
		return fmt.Errorf("validation failed for user creation: %w", err)
	}
	return u.userRepo.CreateUser(ctx, user)
}

func (u *userService) DeleteUser(ctx context.Context, userID int) error {
	return u.userRepo.DeleteUserByID(ctx, userID)
}