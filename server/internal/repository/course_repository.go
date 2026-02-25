// Package repository provides data access layer for course-related operations.
// It encapsulates database interactions for course management, ensuring proper
// error handling, query optimization, and data consistency.
//
// The repository layer abstracts the underlying database implementation and
// provides a clean interface for the service layer to interact with course data.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eduhub/server/internal/models"
)

// CourseRepository defines the interface for course data access operations.
// It provides methods for CRUD operations on courses with proper college scoping.
type CourseRepository interface {
	// CreateCourse creates a new course in the database
	CreateCourse(ctx context.Context, course *models.Course) error

	// FindCourseByID retrieves a course by ID within a specific college
	FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error)

	// UpdateCourse updates an entire course record
	UpdateCourse(ctx context.Context, course *models.Course) error

	// UpdateCoursePartial updates specific fields of a course
	UpdateCoursePartial(ctx context.Context, collegeID int, courseID int, req *models.UpdateCourseRequest) error

	// DeleteCourse removes a course from the database
	DeleteCourse(ctx context.Context, collegeID int, courseID int) error

	// FindAllCourses retrieves all courses for a college with pagination
	FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error)

	// FindCoursesByInstructor retrieves courses taught by a specific instructor
	FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error)

	// CountCoursesByCollege returns the total number of courses in a college
	CountCoursesByCollege(ctx context.Context, collegeID int) (int, error)

	// CountCoursesByInstructor returns the number of courses taught by an instructor
	CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error)

	// CheckCourseNameExists checks if a course with the given name already exists in the college
	CheckCourseNameExists(ctx context.Context, collegeID int, courseName string, excludeCourseID *int) (bool, error)
}

// courseRepository implements the CourseRepository interface
type courseRepository struct {
	Pool *pgxpool.Pool
}

// NewCourseRepository creates a new instance of CourseRepository
func NewCourseRepository(database *DB) CourseRepository {
	return &courseRepository{
		Pool: database.Pool,
	}
}

// Constants for table and column names
const (
	courseTable        = "courses"
	courseSelectFields = "id, name, description, credits, instructor_id, college_id, created_at, updated_at"
)

// setCourseTimestamps sets created_at and updated_at timestamps for a course
func setCourseTimestamps(course *models.Course, isNew bool) {
	now := time.Now()
	if isNew && course.CreatedAt.IsZero() {
		course.CreatedAt = now
	}
	if course.UpdatedAt.IsZero() {
		course.UpdatedAt = now
	}
}

// CreateCourse creates a new course in the database with proper timestamp handling
func (c *courseRepository) CreateCourse(ctx context.Context, course *models.Course) error {
	if course == nil {
		return errors.New("course cannot be nil")
	}

	// Set timestamps
	setCourseTimestamps(course, true)

	sql := fmt.Sprintf(`INSERT INTO %s (
		name, description, credits, instructor_id, college_id, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING %s`, courseTable, courseSelectFields)

	var result models.Course
	err := pgxscan.Get(ctx, c.Pool, &result, sql,
		course.Name, course.Description, int32(course.Credits),
		int32(course.InstructorID), int32(course.CollegeID),
		course.CreatedAt, course.UpdatedAt)

	if err != nil {
		return fmt.Errorf("CreateCourse: failed to execute query: %w", err)
	}

	// Update the course struct with the returned values
	*course = result
	return nil
}

// FindCourseByID retrieves a course by ID within a specific college
func (c *courseRepository) FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE id = $1 AND college_id = $2`,
		courseSelectFields, courseTable)

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

// UpdateCourse updates an entire course record with timestamp handling
func (c *courseRepository) UpdateCourse(ctx context.Context, course *models.Course) error {
	if course == nil {
		return errors.New("course cannot be nil")
	}

	// Set updated timestamp
	setCourseTimestamps(course, false)

	sql := fmt.Sprintf(`UPDATE %s
		SET name = $1, description = $2, credits = $3, instructor_id = $4, updated_at = $5
		WHERE id = $6 AND college_id = $7`, courseTable)

	commandTag, err := c.Pool.Exec(ctx, sql,
		course.Name, course.Description, int32(course.Credits),
		int32(course.InstructorID), course.UpdatedAt,
		int32(course.ID), int32(course.CollegeID))

	if err != nil {
		return fmt.Errorf("UpdateCourse: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateCourse: course with ID %d not found in college %d", course.ID, course.CollegeID)
	}

	return nil
}

// CheckCourseNameExists checks if a course with the given name already exists in the college
func (c *courseRepository) CheckCourseNameExists(ctx context.Context, collegeID int, courseName string, excludeCourseID *int) (bool, error) {
	sql := fmt.Sprintf(`SELECT COUNT(*) as count FROM %s WHERE college_id = $1 AND LOWER(name) = LOWER($2)`, courseTable)
	args := []any{int32(collegeID), courseName}
	argIndex := 3

	if excludeCourseID != nil {
		sql += fmt.Sprintf(` AND id != $%d`, argIndex)
		args = append(args, int32(*excludeCourseID))
	}

	var result struct {
		Count int64 `db:"count"`
	}
	err := pgxscan.Get(ctx, c.Pool, &result, sql, args...)
	if err != nil {
		return false, fmt.Errorf("CheckCourseNameExists: failed to execute query: %w", err)
	}

	return result.Count > 0, nil
}

// DeleteCourse removes a course from the database with proper error handling
func (c *courseRepository) DeleteCourse(ctx context.Context, collegeID int, courseID int) error {
	sql := fmt.Sprintf(`DELETE FROM %s WHERE id = $1 AND college_id = $2`, courseTable)

	commandTag, err := c.Pool.Exec(ctx, sql, int32(courseID), int32(collegeID))
	if err != nil {
		return fmt.Errorf("DeleteCourse: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteCourse: course with ID %d not found in college %d", courseID, collegeID)
	}

	return nil
}

// FindAllCourses retrieves all courses for a college with pagination
func (c *courseRepository) FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE college_id = $1 ORDER BY name ASC LIMIT $2 OFFSET $3`,
		courseSelectFields, courseTable)

	courses := make([]*models.Course, 0)
	err := pgxscan.Select(ctx, c.Pool, &courses, sql, int32(collegeID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("FindAllCourses: failed to scan: %w", err)
	}

	return courses, nil
}

// FindCoursesByInstructor retrieves courses taught by a specific instructor with pagination
func (c *courseRepository) FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE college_id = $1 AND instructor_id = $2 ORDER BY name ASC LIMIT $3 OFFSET $4`,
		courseSelectFields, courseTable)

	courses := make([]*models.Course, 0)
	err := pgxscan.Select(ctx, c.Pool, &courses, sql, int32(collegeID), int32(instructorID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("FindCoursesByInstructor: failed to scan: %w", err)
	}

	return courses, nil
}

// CountCoursesByCollege returns the total number of courses in a college
func (c *courseRepository) CountCoursesByCollege(ctx context.Context, collegeID int) (int, error) {
	sql := fmt.Sprintf(`SELECT COUNT(*) as count FROM %s WHERE college_id = $1`, courseTable)

	var result struct {
		Count int64 `db:"count"`
	}
	err := pgxscan.Get(ctx, c.Pool, &result, sql, int32(collegeID))
	if err != nil {
		return 0, fmt.Errorf("CountCoursesByCollege: failed to execute query: %w", err)
	}

	return int(result.Count), nil
}

// CountCoursesByInstructor returns the number of courses taught by an instructor
func (c *courseRepository) CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error) {
	sql := fmt.Sprintf(`SELECT COUNT(*) as count FROM %s WHERE college_id = $1 AND instructor_id = $2`,
		courseTable)

	var result struct {
		Count int64 `db:"count"`
	}
	err := pgxscan.Get(ctx, c.Pool, &result, sql, int32(collegeID), int32(instructorID))
	if err != nil {
		return 0, fmt.Errorf("CountCoursesByInstructor: failed to execute query: %w", err)
	}

	return int(result.Count), nil
}

// UpdateCoursePartial updates specific fields of a course with dynamic query building
func (c *courseRepository) UpdateCoursePartial(ctx context.Context, collegeID int, courseID int, req *models.UpdateCourseRequest) error {
	if req == nil {
		return errors.New("update request cannot be nil")
	}

	// Build dynamic query based on non-nil fields
	sql := fmt.Sprintf(`UPDATE %s SET updated_at = NOW()`, courseTable)
	args := []any{}
	argIndex := 1

	// Add fields to update based on what's provided
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
		return fmt.Errorf("UpdateCoursePartial: no fields to update")
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
