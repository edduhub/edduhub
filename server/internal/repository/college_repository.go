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

type CollegeRepository interface {
	CreateCollege(ctx context.Context, college *models.College) error
	GetCollegeByID(ctx context.Context, id int) (*models.College, error)
	GetCollegeByName(ctx context.Context, name string) (*models.College, error)
	GetCollegeByExternalID(ctx context.Context, externalID string) (*models.College, error)
	UpdateCollege(ctx context.Context, college *models.College) error
	UpdateCollegePartial(ctx context.Context, id int, req *models.UpdateCollegeRequest) error
	DeleteCollege(ctx context.Context, id int) error
	ListColleges(ctx context.Context, limit, offset uint64) ([]*models.College, error)
	GetCollegeStats(ctx context.Context, collegeID int) (*models.CollegeStats, error)
}

type collegeRepository struct {
	DB *DB
}

func NewCollegeRepository(DB *DB) CollegeRepository {
	return &collegeRepository{
		DB: DB,
	}
}

// type College struct {
// 	ID        int       `db:"id" json:"id"`
// 	Name      string    `db:"name" json:"name"`
// 	Address   string    `db:"address" json:"address"`
// 	City      string    `db:"city" json:"city"`
// 	State     string    `db:"state" json:"state"`
// 	Country   string    `db:"country" json:"country"`
// 	CreatedAt time.Time `db:"created_at" json:"created_at"`
// 	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

// 	// Relations - not stored in DB
// 	Students []*Student `db:"-" json:"students,omitempty"`
// }

func (c *collegeRepository) CreateCollege(ctx context.Context, college *models.College) error {
	now := time.Now()
	college.CreatedAt = now
	college.UpdatedAt = now

	sql := `INSERT INTO college (name, address, city, state, country, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, name, address, city, state, country, created_at, updated_at`
	var result models.College
	err := pgxscan.Get(ctx, c.DB.Pool, &result, sql, college.Name, college.Address, college.City, college.State, college.Country, college.CreatedAt, college.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateCollege: failed to execute query: %w", err)
	}
	*college = result
	return nil
}

func (c *collegeRepository) GetCollegeByID(ctx context.Context, id int) (*models.College, error) {
	sql := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college WHERE id = $1`
	college := &models.College{}
	findErr := pgxscan.Get(ctx, c.DB.Pool, college, sql, id)
	if findErr != nil {
		if errors.Is(findErr, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetCollegeByID: college with id %d not found", id)
		}
		return nil, fmt.Errorf("GetCollegeByID: failed to execute query: %w", findErr)
	}
	return college, nil
}

func (c *collegeRepository) GetCollegeByName(ctx context.Context, name string) (*models.College, error) {
	sql := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college WHERE name = $1`
	college := &models.College{}
	findErr := pgxscan.Get(ctx, c.DB.Pool, college, sql, name)
	if findErr != nil {
		if errors.Is(findErr, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetCollegeByName: college with name '%s' not found", name)
		}
		return nil, fmt.Errorf("GetCollegeByName: failed to execute query: %w", findErr)
	}
	return college, nil
}

func (c *collegeRepository) GetCollegeByExternalID(ctx context.Context, externalID string) (*models.College, error) {
	sql := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college WHERE external_id = $1`
	college := &models.College{}
	findErr := pgxscan.Get(ctx, c.DB.Pool, college, sql, externalID)
	if findErr != nil {
		if errors.Is(findErr, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetCollegeByExternalID: college with external_id '%s' not found", externalID)
		}
		return nil, fmt.Errorf("GetCollegeByExternalID: failed to execute query: %w", findErr)
	}
	return college, nil
}

// type College struct {
// 	ID        int       `db:"id" json:"id"`
// 	Name      string    `db:"name" json:"name"`
// 	Address   string    `db:"address" json:"address"`
// 	City      string    `db:"city" json:"city"`
// 	State     string    `db:"state" json:"state"`
// 	Country   string    `db:"country" json:"country"`
// 	CreatedAt time.Time `db:"created_at" json:"created_at"`
// 	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

// 	// Relations - not stored in DB
// 	Students []*Student `db:"-" json:"students,omitempty"`
// }

func (c *collegeRepository) UpdateCollege(ctx context.Context, college *models.College) error {
	college.UpdatedAt = time.Now()
	sql := `UPDATE college SET name = $1, address = $2, city = $3, state = $4, country = $5, updated_at = $6 WHERE id = $7`
	commandTag, err := c.DB.Pool.Exec(ctx, sql, college.Name, college.Address, college.City, college.State, college.Country, college.UpdatedAt, college.ID)
	if err != nil {
		return fmt.Errorf("UpdateCollege: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateCollege: no college found with ID %d", college.ID)
	}

	return nil
}

func (c *collegeRepository) DeleteCollege(ctx context.Context, id int) error {
	sql := `DELETE FROM college WHERE id = $1`
	commandTag, err := c.DB.Pool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("DeleteCollege: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteCollege: no college found with ID %d", id)
	}
	return nil
}

func (c *collegeRepository) ListColleges(ctx context.Context, limit, offset uint64) ([]*models.College, error) {
	sql := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college ORDER BY name ASC LIMIT $1 OFFSET $2`
	colleges := []*models.College{}
	err := pgxscan.Select(ctx, c.DB.Pool, &colleges, sql, int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("ListColleges: failed to execute query: %w", err)
	}
	return colleges, nil
}

func (c *collegeRepository) UpdateCollegePartial(ctx context.Context, id int, req *models.UpdateCollegeRequest) error {
	// Build dynamic query based on non-nil fields
	sql := `UPDATE college SET updated_at = NOW()`
	args := []any{}
	argIndex := 1

	if req.Name != nil {
		sql += fmt.Sprintf(`, name = $%d`, argIndex)
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Address != nil {
		sql += fmt.Sprintf(`, address = $%d`, argIndex)
		args = append(args, *req.Address)
		argIndex++
	}
	if req.City != nil {
		sql += fmt.Sprintf(`, city = $%d`, argIndex)
		args = append(args, *req.City)
		argIndex++
	}
	if req.State != nil {
		sql += fmt.Sprintf(`, state = $%d`, argIndex)
		args = append(args, *req.State)
		argIndex++
	}
	if req.Country != nil {
		sql += fmt.Sprintf(`, country = $%d`, argIndex)
		args = append(args, *req.Country)
		argIndex++
	}

	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	sql += fmt.Sprintf(` WHERE id = $%d`, argIndex)
	args = append(args, id)

	commandTag, err := c.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateCollegePartial: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateCollegePartial: college with ID %d not found", id)
	}

	return nil
}

func (c *collegeRepository) GetCollegeStats(ctx context.Context, collegeID int) (*models.CollegeStats, error) {
	stats := &models.CollegeStats{
		CollegeID: collegeID,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Combined query using CTE to get all counts in a single database round trip
	query := `
		WITH student_counts AS (
			SELECT 
				COUNT(*) as total_students,
				COUNT(*) FILTER (WHERE is_active = true) as active_students
			FROM students WHERE college_id = $1
		),
		entity_counts AS (
			SELECT 
				(SELECT COUNT(*) FROM courses WHERE college_id = $1) as total_courses,
				(SELECT COUNT(*) FROM departments WHERE college_id = $1) as total_departments,
				(SELECT COUNT(*) FROM enrollments WHERE college_id = $1) as total_enrollments
		),
		grade_stats AS (
			SELECT COALESCE(AVG(percentage), 0) as average_grade
			FROM grades WHERE college_id = $1
		),
		faculty_count AS (
			SELECT COUNT(*) as total_faculties
			FROM users u 
			JOIN user_roles ur ON u.id = ur.user_id 
			JOIN roles r ON ur.role_id = r.id 
			WHERE u.college_id = $1 AND r.name = 'faculty'
		),
		fee_stats AS (
			SELECT 
				COALESCE(SUM(fa.amount), 0) as total_fee,
				COALESCE((SELECT SUM(fp.amount) FROM fee_payments fp 
					JOIN fee_assignments fa ON fp.fee_assignment_id = fa.id 
					WHERE fa.college_id = $1 AND fp.payment_status = 'completed'), 0) as paid_fee
			FROM fee_assignments fa
			WHERE fa.college_id = $1
		)
		SELECT 
			sc.total_students,
			sc.active_students,
			ec.total_courses,
			ec.total_departments,
			ec.total_enrollments,
			gs.average_grade,
			fc.total_faculties,
			fs.total_fee - fs.paid_fee as pending_fees
		FROM student_counts sc, entity_counts ec, grade_stats gs, faculty_count fc, fee_stats fs`

	err := c.DB.Pool.QueryRow(ctx, query, collegeID).Scan(
		&stats.TotalStudents,
		&stats.ActiveStudents,
		&stats.TotalCourses,
		&stats.TotalDepartments,
		&stats.TotalEnrollments,
		&stats.AverageGrade,
		&stats.TotalFaculties,
		&stats.PendingFees,
	)
	if err != nil {
		return nil, fmt.Errorf("GetCollegeStats: failed to get college stats: %w", err)
	}

	return stats, nil
}
