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

type PlacementRepository interface {
	CreatePlacement(ctx context.Context, placement *models.Placement) error
	GetPlacementByID(ctx context.Context, collegeID int, placementID int) (*models.Placement, error)
	UpdatePlacement(ctx context.Context, placement *models.Placement) error
	DeletePlacement(ctx context.Context, collegeID int, placementID int) error

	// Find methods with pagination
	FindPlacementsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.Placement, error)
	FindPlacementsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Placement, error)
	FindPlacementsByCompany(ctx context.Context, collegeID int, companyName string, limit, offset uint64) ([]*models.Placement, error)

	// Count methods
	CountPlacementsByStudent(ctx context.Context, collegeID int, studentID int) (int, error)
	CountPlacementsByCollege(ctx context.Context, collegeID int) (int, error)
}

type placementRepository struct {
	DB *DB
}

func NewPlacementRepository(db *DB) PlacementRepository {
	return &placementRepository{DB: db}
}

const placementTable = "placements"

func (r *placementRepository) CreatePlacement(ctx context.Context, placement *models.Placement) error {
	now := time.Now()
	placement.CreatedAt = now
	placement.UpdatedAt = now

	sql := `INSERT INTO placements (college_id, student_id, company_name, job_title, package, placement_date, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	err := r.DB.Pool.QueryRow(ctx, sql, placement.CollegeID, placement.StudentID, placement.CompanyName, placement.JobTitle, placement.Package, placement.PlacementDate, placement.Status, placement.CreatedAt, placement.UpdatedAt).Scan(&placement.ID)
	if err != nil {
		return fmt.Errorf("CreatePlacement: failed to execute query or scan ID: %w", err)
	}
	return nil
}

func (r *placementRepository) GetPlacementByID(ctx context.Context, collegeID int, placementID int) (*models.Placement, error) {
	sql := `SELECT id, college_id, student_id, company_name, job_title, package, placement_date, status, created_at, updated_at FROM placements WHERE id = $1 AND college_id = $2`
	placement := &models.Placement{}
	err := pgxscan.Get(ctx, r.DB.Pool, placement, sql, placementID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetPlacementByID: placement with ID %d not found for college ID %d", placementID, collegeID)
		}
		return nil, fmt.Errorf("GetPlacementByID: failed to execute query or scan: %w", err)
	}
	return placement, nil
}

func (r *placementRepository) UpdatePlacement(ctx context.Context, placement *models.Placement) error {
	placement.UpdatedAt = time.Now()

	sql := `UPDATE placements SET company_name = $1, job_title = $2, package = $3, placement_date = $4, status = $5, updated_at = $6 WHERE id = $7 AND college_id = $8`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, placement.CompanyName, placement.JobTitle, placement.Package, placement.PlacementDate, placement.Status, placement.UpdatedAt, placement.ID, placement.CollegeID)
	if err != nil {
		return fmt.Errorf("UpdatePlacement: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdatePlacement: no placement found with ID %d for college ID %d, or no changes made", placement.ID, placement.CollegeID)
	}
	return nil
}

func (r *placementRepository) DeletePlacement(ctx context.Context, collegeID int, placementID int) error {
	sql := `DELETE FROM placements WHERE id = $1 AND college_id = $2`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, placementID, collegeID)
	if err != nil {
		return fmt.Errorf("DeletePlacement: failed to execute query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeletePlacement: no placement found with ID %d for college ID %d, or already deleted", placementID, collegeID)
	}
	return nil
}

func (r *placementRepository) findPlacements(ctx context.Context, sql string, args []interface{}, limit, offset uint64) ([]*models.Placement, error) {
	fullSQL := fmt.Sprintf("%s ORDER BY placement_date DESC, student_id ASC LIMIT $%d OFFSET $%d", sql, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	placements := []*models.Placement{}
	err := pgxscan.Select(ctx, r.DB.Pool, &placements, fullSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("findPlacements: failed to execute query or scan: %w", err)
	}
	return placements, nil
}

func (r *placementRepository) FindPlacementsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.Placement, error) {
	sql := `SELECT id, college_id, student_id, company_name, job_title, package, placement_date, status, created_at, updated_at FROM placements WHERE college_id = $1 AND student_id = $2`
	args := []interface{}{collegeID, studentID}
	return r.findPlacements(ctx, sql, args, limit, offset)
}

func (r *placementRepository) FindPlacementsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Placement, error) {
	sql := `SELECT id, college_id, student_id, company_name, job_title, package, placement_date, status, created_at, updated_at FROM placements WHERE college_id = $1`
	args := []interface{}{collegeID}
	return r.findPlacements(ctx, sql, args, limit, offset)
}

func (r *placementRepository) FindPlacementsByCompany(ctx context.Context, collegeID int, companyName string, limit, offset uint64) ([]*models.Placement, error) {
	// Use ILIKE for case-insensitive search, adjust if case-sensitive is needed
	sql := `SELECT id, college_id, student_id, company_name, job_title, package, placement_date, status, created_at, updated_at FROM placements WHERE college_id = $1 AND company_name ILIKE '%' || $2 || '%'`
	args := []interface{}{collegeID, companyName}
	return r.findPlacements(ctx, sql, args, limit, offset)
}

func (r *placementRepository) countPlacements(ctx context.Context, sql string, args []interface{}) (int, error) {
	var count int
	err := r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("countPlacements: failed to execute query or scan: %w", err)
	}
	return count, nil
}

func (r *placementRepository) CountPlacementsByStudent(ctx context.Context, collegeID int, studentID int) (int, error) {
	sql := `SELECT COUNT(*) FROM placements WHERE college_id = $1 AND student_id = $2`
	args := []interface{}{collegeID, studentID}
	return r.countPlacements(ctx, sql, args)
}

func (r *placementRepository) CountPlacementsByCollege(ctx context.Context, collegeID int) (int, error) {
	sql := `SELECT COUNT(*) FROM placements WHERE college_id = $1`
	args := []interface{}{collegeID}
	return r.countPlacements(ctx, sql, args)
}
