package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type DepartmentRepository interface {
	CreateDepartment(ctx context.Context, department *models.Department) error
	GetDepartmentByID(ctx context.Context, collegeID int, departmentID int) (*models.Department, error)
	GetDepartmentByName(ctx context.Context, collegeID int, name string) (*models.Department, error)
	UpdateDepartment(ctx context.Context, department *models.Department) error
	DeleteDepartment(ctx context.Context, collegeID int, departmentID int) error
	ListDepartmentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Department, error)
	CountDepartmentsByCollege(ctx context.Context, collegeID int) (int, error)
}

type departmentRepository struct {
	DB *DB
}

const departmentTable = "departments"

func NewDepartmentRepository(DB *DB) DepartmentRepository {
	return &departmentRepository{
		DB: DB,
	}
}

func (r *departmentRepository) CreateDepartment(ctx context.Context, department *models.Department) error {
	now := time.Now()
	department.CreatedAt = now
	department.UpdatedAt = now

	sql := `INSERT INTO departments (college_id, name, hod, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.DB.Pool.QueryRow(ctx, sql, department.CollegeID, department.Name, department.HOD, department.CreatedAt, department.UpdatedAt).Scan(&department.ID)
	if err != nil {
		return fmt.Errorf("CreateDepartment: failed to execute query or scan ID: %w", err)
	}
	return nil
}

func (r *departmentRepository) GetDepartmentByID(ctx context.Context, collegeID int, departmentID int) (*models.Department, error) {
	department := &models.Department{}
	sql := `SELECT id, college_id, name, hod, created_at, updated_at FROM departments WHERE id = $1 AND college_id = $2`
	err := pgxscan.Get(ctx, r.DB.Pool, department, sql, departmentID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetDepartmentByID: department with ID %d not found for college ID %d", departmentID, collegeID)
		}
		return nil, fmt.Errorf("GetDepartmentByID: failed to execute query or scan: %w", err)
	}
	return department, nil
}

func (r *departmentRepository) GetDepartmentByName(ctx context.Context, collegeID int, name string) (*models.Department, error) {
	department := &models.Department{}
	sql := `SELECT id, college_id, name, hod, created_at, updated_at FROM departments WHERE name = $1 AND college_id = $2`
	err := pgxscan.Get(ctx, r.DB.Pool, department, sql, name, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetDepartmentByName: department with name '%s' not found for college ID %d", name, collegeID)
		}
		return nil, fmt.Errorf("GetDepartmentByName: failed to execute query or scan: %w", err)
	}
	return department, nil
}

func (r *departmentRepository) UpdateDepartment(ctx context.Context, department *models.Department) error {
	department.UpdatedAt = time.Now()
	sql := `UPDATE departments SET name = $1, hod = $2, updated_at = $3 WHERE id = $4 AND college_id = $5`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, department.Name, department.HOD, department.UpdatedAt, department.ID, department.CollegeID)
	if err != nil {
		return fmt.Errorf("UpdateDepartment: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateDepartment: no department found with ID %d for college ID %d, or no changes made", department.ID, department.CollegeID)
	}
	return nil
}

func (r *departmentRepository) DeleteDepartment(ctx context.Context, collegeID int, departmentID int) error {
	sql := `DELETE FROM departments WHERE id = $1 AND college_id = $2`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, departmentID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteDepartment: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteDepartment: no department found with ID %d for college ID %d, or already deleted", departmentID, collegeID)
	}
	return nil
}

func (r *departmentRepository) ListDepartmentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Department, error) {
	departments := []*models.Department{}
	sql := `SELECT id, college_id, name, hod, created_at, updated_at FROM departments WHERE college_id = $1 ORDER BY name ASC LIMIT $2 OFFSET $3`
	err := pgxscan.Select(ctx, r.DB.Pool, &departments, sql, collegeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ListDepartmentsByCollege: failed to execute query or scan: %w", err)
	}
	return departments, nil
}

func (r *departmentRepository) CountDepartmentsByCollege(ctx context.Context, collegeID int) (int, error) {
	var count int
	sql := `SELECT COUNT(*) FROM departments WHERE college_id = $1`
	err := r.DB.Pool.QueryRow(ctx, sql, collegeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountDepartmentsByCollege: exec/scan: %w", err)
	}
	return count, nil
}
