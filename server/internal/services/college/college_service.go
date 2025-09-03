package college

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

type CollegeService interface {
	CreateCollege(ctx context.Context, college *models.College) error
	GetCollegeByID(ctx context.Context, id int) (*models.College, error)
	GetCollegeByName(ctx context.Context, name string) (*models.College, error)
	UpdateCollege(ctx context.Context, college *models.College) error
	UpdateCollegePartial(ctx context.Context, id int, req *models.UpdateCollegeRequest) error
	DeleteCollege(ctx context.Context, id int) error
	ListColleges(ctx context.Context, limit, offset uint64) ([]*models.College, error)
}

type collegeService struct {
	collegeRepo repository.CollegeRepository
	validate    *validator.Validate
}

func NewCollegeService(collegeRepo repository.CollegeRepository) CollegeService {
	return &collegeService{
		collegeRepo: collegeRepo,
		validate:    validator.New(),
	}
}

func (c *collegeService) CreateCollege(ctx context.Context, college *models.College) error {
	if err := c.validate.Struct(college); err != nil {
		return fmt.Errorf("validation failed for college %w", err)
	}
	return c.collegeRepo.CreateCollege(ctx, college)
}

func (c *collegeService) UpdateCollege(ctx context.Context, college *models.College) error {
	if err := c.validate.Struct(college); err != nil {
		return fmt.Errorf("validation failed for college %w", err)
	}
	return c.collegeRepo.UpdateCollege(ctx, college)
}

func (c *collegeService) GetCollegeByID(ctx context.Context, id int) (*models.College, error) {
	return c.collegeRepo.GetCollegeByID(ctx, id)
}

func (c *collegeService) GetCollegeByName(ctx context.Context, name string) (*models.College, error) {
	return c.collegeRepo.GetCollegeByName(ctx, name)
}

func (c *collegeService) DeleteCollege(ctx context.Context, id int) error {
	return c.collegeRepo.DeleteCollege(ctx, id)
}

func (c *collegeService) ListColleges(ctx context.Context, limit, offset uint64) ([]*models.College, error) {
	return c.collegeRepo.ListColleges(ctx, limit, offset)
}

func (c *collegeService) UpdateCollegePartial(ctx context.Context, id int, req *models.UpdateCollegeRequest) error {
	if err := c.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed for college update: %w", err)
	}
	return c.collegeRepo.UpdateCollegePartial(ctx, id, req)
}
