package repository

import (
	"context"
	"errors"
	"fmt" // Keep fmt for error wrapping
	"time"

	"eduhub/server/internal/models"

	"database/sql"

	"eduhub/server/internal/repository/db"

	"github.com/jackc/pgx/v4"              // Use v4 for pgx.ErrNoRows
)

type CourseRepository interface {
	CreateCourse(ctx context.Context, course *models.Course) error
	FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error)
	UpdateCourse(ctx context.Context, course *models.Course) error
	DeleteCourse(ctx context.Context, collegeID int, courseID int) error

	// Find methods with pagination
	FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error)
	FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error)

	// Count methods
	CountCoursesByCollege(ctx context.Context, collegeID int) (int, error)
	CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error)
}

type courseRepository struct {
	DB *DB
	*db.Queries
}

func NewCourseRepository(database *DB) CourseRepository {
	return &courseRepository{
		DB: database,
		Queries: db.New(database.Pool),
	}
}

const courseTable = "course" // Define table name constant

func (c *courseRepository) CreateCourse(ctx context.Context, course *models.Course) error {
	// Set timestamps
	now := time.Now()
	if course.CreatedAt.IsZero() {
		course.CreatedAt = now
	}
	if course.UpdatedAt.IsZero() {
		course.UpdatedAt = now
	}

	// Use sqlc generated code
	params := db.CreateCourseParams{
		Name:         course.Name,
		Description:  sql.NullString{String: course.Description, Valid: course.Description != ""},
		Credits:      int32(course.Credits),
		InstructorID: sql.NullInt32{Int32: int32(course.InstructorID), Valid: course.InstructorID > 0},
		CreatedAt:    course.CreatedAt,
		UpdatedAt:    course.UpdatedAt,
	}

	result, err := c.Queries.CreateCourse(ctx, params)
	if err != nil {
		return fmt.Errorf("CreateCourse: failed to execute query: %w", err)
	}

	// Update the course struct with the returned values
	course.ID = int(result.ID)
	course.Name = result.Name
	course.Description = result.Description.String
	course.Credits = int(result.Credits)
	course.InstructorID = int(result.InstructorID.Int32)
	course.CreatedAt = result.CreatedAt
	course.UpdatedAt = result.UpdatedAt

	return nil
}

func (c *courseRepository) FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error) {
	// Use sqlc generated code
	params := db.FindCourseByIDParams{
		ID:       int32(courseID),
		CollegeID: sql.NullInt32{Int32: int32(collegeID), Valid: collegeID > 0},
	}
	result, err := c.Queries.FindCourseByID(ctx, params)
	if err != nil {
		// It's better to check for specific errors like "no rows"
		if errors.Is(err, pgx.ErrNoRows) { // Use errors.Is for checking pgx.ErrNoRows
			return nil, fmt.Errorf("FindCourseByID: course with ID %d not found for college ID %d", courseID, collegeID) // Or a custom ErrNotFound
		}
		return nil, fmt.Errorf("unable to find course: %w", err) // Wrap the original error
	}

	// Convert db.FindCourseByIDRow to models.Course
	course := &models.Course{
		ID:          int(result.ID),
		Name:        result.Name,
		Description: result.Description.String,
		Credits:     int(result.Credits),
		InstructorID: int(result.InstructorID.Int32),
		CollegeID:   int(result.CollegeID.Int32),
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}

	return course, nil
}

func (c *courseRepository) UpdateCourse(ctx context.Context, course *models.Course) error {
	course.UpdatedAt = time.Now()

	// Use sqlc generated code
	params := db.UpdateCourseParams{
		Name:         course.Name,
		Description:  sql.NullString{String: course.Description, Valid: course.Description != ""},
		Credits:      int32(course.Credits),
		InstructorID: sql.NullInt32{Int32: int32(course.InstructorID), Valid: course.InstructorID > 0},
		CollegeID:    sql.NullInt32{Int32: int32(course.CollegeID), Valid: course.CollegeID > 0},
		UpdatedAt:    course.UpdatedAt,
		ID:           int32(course.ID),
	}

	err := c.Queries.UpdateCourse(ctx, params)
	if err != nil {
		return fmt.Errorf("UpdateCourse: failed to execute query: %w", err)
	}

	return nil
}

func (c *courseRepository) DeleteCourse(ctx context.Context, collegeID int, courseID int) error {
	// Use sqlc generated code
	params := db.DeleteCourseParams{
		ID:       int32(courseID),
		CollegeID: sql.NullInt32{Int32: int32(collegeID), Valid: collegeID > 0},
	}
	err := c.Queries.DeleteCourse(ctx, params)
	if err != nil {
		// Consider foreign key constraint errors
		return fmt.Errorf("DeleteCourse: failed to execute query: %w", err)
	}

	return nil
}

func (c *courseRepository) FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error) {
	// Use sqlc generated code
	params := db.FindAllCoursesParams{
		CollegeID: sql.NullInt32{Int32: int32(collegeID), Valid: collegeID > 0},
		Limit:     int32(limit),
		Offset:    int32(offset),
	}

	results, err := c.Queries.FindAllCourses(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("FindAllCourses: failed to execute query or scan: %w", err)
	}

	// Convert []db.FindAllCoursesRow to []*models.Course
	courses := make([]*models.Course, len(results))
	for i, result := range results {
		courses[i] = &models.Course{
			ID:          int(result.ID),
			Name:        result.Name,
			Description: result.Description.String,
			Credits:     int(result.Credits),
			InstructorID: int(result.InstructorID.Int32),
			CollegeID:   int(result.CollegeID.Int32),
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		}
	}

	return courses, nil
}

func (c *courseRepository) FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error) {
	// Use sqlc generated code
	params := db.FindCoursesByInstructorParams{
		CollegeID:    sql.NullInt32{Int32: int32(collegeID), Valid: collegeID > 0},
		InstructorID: sql.NullInt32{Int32: int32(instructorID), Valid: instructorID > 0},
		Limit:        int32(limit),
		Offset:       int32(offset),
	}

	results, err := c.Queries.FindCoursesByInstructor(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("FindCoursesByInstructor: failed to execute query or scan: %w", err)
	}

	// Convert []db.Course to []*models.Course
	courses := make([]*models.Course, len(results))
	for i, result := range results {
		courses[i] = &models.Course{
			ID:          int(result.ID),
			Name:        result.Name,
			Description: result.Description.String,
			Credits:     int(result.Credits),
			InstructorID: int(result.InstructorID.Int32),
			CollegeID:   int(result.CollegeID.Int32),
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		}
	}

	return courses, nil
}

func (c *courseRepository) CountCoursesByCollege(ctx context.Context, collegeID int) (int, error) {
	// Use sqlc generated code
	params := sql.NullInt32{Int32: int32(collegeID), Valid: collegeID > 0}
	count, err := c.Queries.CountCoursesByCollege(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("CountCoursesByCollege: failed to execute query or scan: %w", err)
	}

	return int(count), nil
}

func (c *courseRepository) CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error) {
	// Use sqlc generated code
	params := db.CountCoursesByInstructorParams{
		CollegeID:    sql.NullInt32{Int32: int32(collegeID), Valid: collegeID > 0},
		InstructorID: sql.NullInt32{Int32: int32(instructorID), Valid: instructorID > 0},
	}
	count, err := c.Queries.CountCoursesByInstructor(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("CountCoursesByInstructor: failed to execute query or scan: %w", err)
	}

	return int(count), nil
}

