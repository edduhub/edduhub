package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	UpdateDepartmentPartial(ctx context.Context, collegeID int, departmentID int, req *models.UpdateDepartmentRequest) error
	DeleteDepartment(ctx context.Context, collegeID int, departmentID int) error
	ListDepartmentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Department, error)
	CountDepartmentsByCollege(ctx context.Context, collegeID int) (int, error)
}

type departmentRepository struct {
	DB *DB
}

func NewDepartmentRepository(db *DB) DepartmentRepository {
	return &departmentRepository{DB: db}
}

const departmentSelect = `
	SELECT
		d.id,
		d.college_id,
		d.name,
		d.code,
		d.description,
		d.head_user_id,
		COALESCE(u.name, d.hod, '') AS hod_name,
		COALESCE(d.hod, u.name, '') AS hod,
		d.is_active,
		0::INT AS student_count,
		CASE WHEN d.head_user_id IS NULL THEN 0 ELSE 1 END AS faculty_count,
		0::INT AS courses_count,
		d.created_at,
		d.updated_at
	FROM departments d
	LEFT JOIN users u ON u.id = d.head_user_id
`

func (r *departmentRepository) CreateDepartment(ctx context.Context, department *models.Department) error {
	now := time.Now()
	department.CreatedAt = now
	department.UpdatedAt = now
	if department.Code == "" {
		department.Code = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(department.Name), " ", ""))
		if len(department.Code) > 12 {
			department.Code = department.Code[:12]
		}
	}

	sql := `
		INSERT INTO departments (college_id, name, code, description, head_user_id, hod, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	hod := department.HOD
	if hod == "" {
		hod = department.HODName
	}

	if err := r.DB.Pool.QueryRow(ctx, sql,
		department.CollegeID,
		department.Name,
		department.Code,
		department.Description,
		department.HeadUserID,
		hod,
		true,
		department.CreatedAt,
		department.UpdatedAt,
	).Scan(&department.ID); err != nil {
		return fmt.Errorf("CreateDepartment: failed to execute query: %w", err)
	}

	fresh, err := r.GetDepartmentByID(ctx, department.CollegeID, department.ID)
	if err != nil {
		return err
	}
	*department = *fresh
	return nil
}

func (r *departmentRepository) GetDepartmentByID(ctx context.Context, collegeID int, departmentID int) (*models.Department, error) {
	department := &models.Department{}
	sql := departmentSelect + ` WHERE d.id = $1 AND d.college_id = $2`
	err := pgxscan.Get(ctx, r.DB.Pool, department, sql, departmentID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetDepartmentByID: department with ID %d not found for college ID %d", departmentID, collegeID)
		}
		return nil, fmt.Errorf("GetDepartmentByID: failed to execute query: %w", err)
	}
	return department, nil
}

func (r *departmentRepository) GetDepartmentByName(ctx context.Context, collegeID int, name string) (*models.Department, error) {
	department := &models.Department{}
	sql := departmentSelect + ` WHERE d.name = $1 AND d.college_id = $2`
	err := pgxscan.Get(ctx, r.DB.Pool, department, sql, name, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetDepartmentByName: department with name '%s' not found for college ID %d", name, collegeID)
		}
		return nil, fmt.Errorf("GetDepartmentByName: failed to execute query: %w", err)
	}
	return department, nil
}

func (r *departmentRepository) UpdateDepartment(ctx context.Context, department *models.Department) error {
	department.UpdatedAt = time.Now()
	sql := `
		UPDATE departments
		SET name = $1, code = $2, description = $3, head_user_id = $4, hod = $5, is_active = $6, updated_at = $7
		WHERE id = $8 AND college_id = $9`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql,
		department.Name,
		department.Code,
		department.Description,
		department.HeadUserID,
		department.HOD,
		department.IsActive,
		department.UpdatedAt,
		department.ID,
		department.CollegeID,
	)
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
	sql := departmentSelect + ` WHERE d.college_id = $1 ORDER BY d.name ASC LIMIT $2 OFFSET $3`
	err := pgxscan.Select(ctx, r.DB.Pool, &departments, sql, collegeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ListDepartmentsByCollege: failed to execute query: %w", err)
	}
	return departments, nil
}

func (r *departmentRepository) UpdateDepartmentPartial(ctx context.Context, collegeID int, departmentID int, req *models.UpdateDepartmentRequest) error {
	setClauses := []string{"updated_at = NOW()"}
	args := []any{}
	argIndex := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Code != nil {
		setClauses = append(setClauses, fmt.Sprintf("code = $%d", argIndex))
		args = append(args, *req.Code)
		argIndex++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.HeadUserID != nil {
		setClauses = append(setClauses, fmt.Sprintf("head_user_id = $%d", argIndex))
		args = append(args, *req.HeadUserID)
		argIndex++
	}
	if req.HOD != nil {
		setClauses = append(setClauses, fmt.Sprintf("hod = $%d", argIndex))
		args = append(args, *req.HOD)
		argIndex++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	args = append(args, departmentID, collegeID)
	sql := fmt.Sprintf(`UPDATE departments SET %s WHERE id = $%d AND college_id = $%d`, strings.Join(setClauses, ", "), argIndex, argIndex+1)
	commandTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateDepartmentPartial: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateDepartmentPartial: no department found with ID %d for college ID %d", departmentID, collegeID)
	}
	return nil
}

func (r *departmentRepository) CountDepartmentsByCollege(ctx context.Context, collegeID int) (int, error) {
	sql := `SELECT COUNT(*) as count FROM departments WHERE college_id = $1`
	var result struct {
		Count int `db:"count"`
	}
	err := pgxscan.Get(ctx, r.DB.Pool, &result, sql, collegeID)
	if err != nil {
		return 0, fmt.Errorf("CountDepartmentsByCollege: exec/scan: %w", err)
	}
	return result.Count, nil
}
