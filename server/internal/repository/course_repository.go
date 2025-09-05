package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v5/pgxpool"

	"eduhub/server/internal/models"
)

type CourseRepository interface {
	CreateCourse(ctx context.Context, course *models.Course) error
	FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error)
	UpdateCourse(ctx context.Context, course *models.Course) error
	UpdateCoursePartial(ctx context.Context, collegeID int, courseID int, req *models.UpdateCourseRequest) error
	DeleteCourse(ctx context.Context, collegeID int, courseID int) error

	// Find methods with pagination
	FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error)
	FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error)

	// Count methods
	CountCoursesByCollege(ctx context.Context, collegeID int) (int, error)
	CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error)
}

type courseRepository struct {
	Pool *pgxpool.Pool
}

func NewCourseRepository(database *DB) CourseRepository {
	return &courseRepository{
		Pool: database.Pool,
	}
}

const courseTable = "courses" // Define table name constant

func (c *courseRepository) CreateCourse(ctx context.Context, course *models.Course) error {
	// Set timestamps
	now := time.Now()
	if course.CreatedAt.IsZero() {
		course.CreatedAt = now
	}
	if course.UpdatedAt.IsZero() {
		course.UpdatedAt = now
	}

	sql := `INSERT INTO courses (
    name,
    description,
    credits,
    instructor_id,
    college_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, name, description, credits, instructor_id, college_id, created_at, updated_at`

	var result models.Course

	err := pgxscan.Get(ctx, c.Pool, &result, sql, course.Name, course.Description, int32(course.Credits), int32(course.InstructorID), int32(course.CollegeID), course.CreatedAt, course.UpdatedAt)

	if err != nil {
		return fmt.Errorf("CreateCourse: failed to execute query: %w", err)
	}

	// Update the course struct with the returned values
	*course = result

	return nil
}

func (c *courseRepository) FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error) {
	sql := `SELECT id, name, description, credits, instructor_id, college_id, created_at, updated_at
FROM courses
WHERE id = $1 AND college_id = $2`

	var course models.Course

	err := pgxscan.Get(ctx, c.Pool, &course, sql, int32(courseID), int32(collegeID))

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("FindCourseByID: course with ID %d not found for college ID %d", courseID, collegeID)
		}
		return nil, fmt.Errorf("FindCourseByID: failed to execute query: %w", err)
	}

	return &course, nil
}

func (c *courseRepository) UpdateCourse(ctx context.Context, course *models.Course) error {
	course.UpdatedAt = time.Now()

	sql := `UPDATE courses
SET name = $1,
    description = $2,
    credits = $3,
    instructor_id = $4,
    updated_at = $5
WHERE id = $6 AND college_id = $7`

	_, err := c.Pool.Exec(ctx, sql,
		course.Name,
		course.Description,
		int32(course.Credits),
		int32(course.InstructorID),
		course.UpdatedAt,
		int32(course.ID),
		int32(course.CollegeID),
	)

	if err != nil {
		return fmt.Errorf("UpdateCourse: failed to execute query: %w", err)
	}

	return nil
}

func (c *courseRepository) DeleteCourse(ctx context.Context, collegeID int, courseID int) error {
	sql := `DELETE FROM courses
WHERE id = $1 AND college_id = $2`

	_, err := c.Pool.Exec(ctx, sql, int32(courseID), int32(collegeID))
	if err != nil {
		// Consider foreign key constraint errors
		return fmt.Errorf("DeleteCourse: failed to execute query: %w", err)
	}

	return nil
}

func (c *courseRepository) FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error) {
	sql := `SELECT id, name, description, credits, instructor_id, college_id, created_at, updated_at
FROM courses
WHERE college_id = $1
ORDER BY name ASC
LIMIT $2 OFFSET $3`

	courses := make([]*models.Course, 0)
	err := pgxscan.Select(ctx, c.Pool, &courses, sql, int32(collegeID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("FindAllCourses: failed to scan: %w", err)
	}

	return courses, nil
}

func (c *courseRepository) FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error) {
	sql := `SELECT id, name, description, credits, instructor_id, college_id, created_at, updated_at
FROM courses
WHERE college_id = $1 AND instructor_id = $2
ORDER BY name ASC
LIMIT $3 OFFSET $4`

	courses := make([]*models.Course, 0)
	err := pgxscan.Select(ctx, c.Pool, &courses, sql, int32(collegeID), int32(instructorID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("FindCoursesByInstructor: failed to scan: %w", err)
	}

	return courses, nil
}

func (c *courseRepository) CountCoursesByCollege(ctx context.Context, collegeID int) (int, error) {
	sql := `SELECT COUNT(*) as count
FROM courses
WHERE college_id = $1`

	var result struct {
		Count int64 `db:"count"`
	}
	err := pgxscan.Get(ctx, c.Pool, &result, sql, int32(collegeID))
	if err != nil {
		return 0, fmt.Errorf("CountCoursesByCollege: failed to execute query: %w", err)
	}

	return int(result.Count), nil
}

func (c *courseRepository) CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error) {
	sql := `SELECT COUNT(*) as count
FROM courses
WHERE college_id = $1 AND instructor_id = $2`

	var result struct {
		Count int64 `db:"count"`
	}
	err := pgxscan.Get(ctx, c.Pool, &result, sql, int32(collegeID), int32(instructorID))
	if err != nil {
		return 0, fmt.Errorf("CountCoursesByInstructor: failed to execute query: %w", err)
	}

	return int(result.Count), nil
}

func (c *courseRepository) UpdateCoursePartial(ctx context.Context, collegeID int, courseID int, req *models.UpdateCourseRequest) error {
	// Build dynamic query based on non-nil fields
	sql := `UPDATE courses SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		sql += fmt.Sprintf(`, name = $%d`, argIndex)
		args = append(args, *req.Name)
		argIndex++
	}
	if req.CollegeID != nil {
		sql += fmt.Sprintf(`, college_id = $%d`, argIndex)
		args = append(args, int32(*req.CollegeID))
		argIndex++
	}
	if req.Description != nil {
		sql += fmt.Sprintf(`, description = $%d`, argIndex)
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Credits != nil {
		sql += fmt.Sprintf(`, credits = $%d`, argIndex)
		args = append(args, int32(*req.Credits))
		argIndex++
	}
	if req.InstructorID != nil {
		sql += fmt.Sprintf(`, instructor_id = $%d`, argIndex)
		args = append(args, int32(*req.InstructorID))
		argIndex++
	}

	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	sql += fmt.Sprintf(` WHERE id = $%d AND college_id = $%d`, argIndex, argIndex+1)
	args = append(args, int32(courseID), int32(collegeID))

	commandTag, err := c.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateCoursePartial: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateCoursePartial: course with ID %d not found in college %d", courseID, collegeID)
	}

	return nil
}