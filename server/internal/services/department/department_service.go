package department

import (
	"context"
	"errors"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type DepartmentService interface {
	CreateDepartment(ctx context.Context, department *models.Department) error
	GetDepartment(ctx context.Context, collegeID int, departmentID int) (*models.Department, error)
	GetDepartments(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Department, error)
	UpdateDepartment(ctx context.Context, collegeID int, departmentID int, req *models.UpdateDepartmentRequest) error
	DeleteDepartment(ctx context.Context, collegeID int, departmentID int) error
}

type departmentService struct {
	departmentRepo repository.DepartmentRepository
}

func NewDepartmentService(departmentRepo repository.DepartmentRepository) DepartmentService {
	return &departmentService{
		departmentRepo: departmentRepo,
	}
}

func (s *departmentService) CreateDepartment(ctx context.Context, department *models.Department) error {
	if department == nil {
		return errors.New("department cannot be nil")
	}
	return s.departmentRepo.CreateDepartment(ctx, department)
}

func (s *departmentService) GetDepartment(ctx context.Context, collegeID int, departmentID int) (*models.Department, error) {
	return s.departmentRepo.GetDepartmentByID(ctx, departmentID, collegeID)
}

func (s *departmentService) GetDepartments(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Department, error) {
	if limit > 100 {
		limit = 100
	}
	return s.departmentRepo.ListDepartmentsByCollege(ctx, collegeID, limit, offset)
}

func (s *departmentService) UpdateDepartment(ctx context.Context, collegeID int, departmentID int, req *models.UpdateDepartmentRequest) error {
	return s.departmentRepo.UpdateDepartmentPartial(ctx, req)
}

func (s *departmentService) DeleteDepartment(ctx context.Context, collegeID int, departmentID int) error {
	return s.departmentRepo.DeleteDepartment(ctx, departmentID, collegeID)
}
