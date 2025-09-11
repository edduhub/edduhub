// Package course provides business logic for course management operations.
// It handles course creation, updates, retrieval, and deletion with proper
// validation, business rule enforcement, and error handling.
//
// The service layer acts as an intermediary between the API handlers and
// the data repository, ensuring data integrity and business logic consistency.
package course

import (
	"context"
	"errors"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

// CourseService defines the interface for course business logic operations.
// It provides methods for creating, retrieving, updating, and deleting courses
// with proper validation and business rule enforcement.
type CourseService interface {
	// CreateCourse creates a new course with validation and business rule checks
	CreateCourse(ctx context.Context, course *models.Course) error

	// FindCourseByID retrieves a course by ID within a specific college
	FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error)

	// UpdateCourse updates an entire course with full validation
	UpdateCourse(ctx context.Context, courseID int, course *models.Course) error

	// UpdateCoursePartial updates specific fields of a course with partial validation
	UpdateCoursePartial(ctx context.Context, collegeID int, courseID int, req *models.UpdateCourseRequest) error

	// DeleteCourse removes a course with cascade consideration
	DeleteCourse(ctx context.Context, collegeID int, courseID int) error

	// FindAllCourses retrieves all courses for a college with pagination
	FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error)

	// FindCoursesByInstructor retrieves courses taught by a specific instructor
	FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error)

	// CountCoursesByCollege returns the total number of courses in a college
	CountCoursesByCollege(ctx context.Context, collegeID int) (int, error)

	// CountCoursesByInstructor returns the number of courses taught by an instructor
	CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error)
}

// courseService implements the CourseService interface
type courseService struct {
	courseRepo  repository.CourseRepository
	collegeRepo repository.CollegeRepository
	userRepo    repository.UserRepository
	validate    *validator.Validate
}

// NewCourseService creates a new instance of CourseService with required dependencies
func NewCourseService(
	courseRepo repository.CourseRepository,
	collegeRepo repository.CollegeRepository,
	userRepo repository.UserRepository,
) CourseService {
	return &courseService{
		courseRepo:  courseRepo,
		collegeRepo: collegeRepo,
		userRepo:    userRepo,
		validate:    validator.New(),
	}
}

func (c *courseService) CreateCourse(ctx context.Context, course *models.Course) error {
	// Input validation
	if course == nil {
		return errors.New("course cannot be nil")
	}

	if err := c.validate.Struct(course); err != nil {
		return fmt.Errorf("course validation failed: %w", err)
	}

	// Business logic validation: Verify college exists
	college, err := c.collegeRepo.GetCollegeByID(ctx, course.CollegeID)
	if err != nil {
		return fmt.Errorf("failed to verify college existence: %w", err)
	}
	if college == nil {
		return fmt.Errorf("college with ID %d does not exist", course.CollegeID)
	}

	// Business logic validation: Verify instructor exists and is a valid user
	// Note: This assumes instructor validation is handled at the user level
	// Additional instructor role validation could be added here if needed

	// Business logic validation: Check for duplicate course names within the college
	exists, err := c.courseRepo.CheckCourseNameExists(ctx, course.CollegeID, course.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to check course name uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("course with name '%s' already exists in college %d", course.Name, course.CollegeID)
	}

	return c.courseRepo.CreateCourse(ctx, course)
}

func (c *courseService) FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error) {
	// Input validation
	if collegeID <= 0 {
		return nil, errors.New("invalid college ID")
	}
	if courseID <= 0 {
		return nil, errors.New("invalid course ID")
	}

	// Business logic validation: Verify college exists
	_, err := c.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify college existence: %w", err)
	}

	return c.courseRepo.FindCourseByID(ctx, collegeID, courseID)
}

func (c *courseService) UpdateCourse(ctx context.Context, courseID int, course *models.Course) error {
	// Input validation
	if course == nil {
		return errors.New("course cannot be nil")
	}

	if courseID <= 0 {
		return errors.New("invalid course ID")
	}

	if course.ID != courseID {
		return fmt.Errorf("course ID mismatch: provided %d, expected %d", course.ID, courseID)
	}

	if err := c.validate.Struct(course); err != nil {
		return fmt.Errorf("course validation failed: %w", err)
	}

	// Business logic validation: Verify course exists
	existingCourse, err := c.courseRepo.FindCourseByID(ctx, course.CollegeID, courseID)
	if err != nil {
		return fmt.Errorf("failed to verify course existence: %w", err)
	}
	if existingCourse == nil {
		return fmt.Errorf("course with ID %d not found in college %d", courseID, course.CollegeID)
	}

	// Business logic validation: Verify college exists (if college ID is being changed)
	if course.CollegeID != existingCourse.CollegeID {
		college, err := c.collegeRepo.GetCollegeByID(ctx, course.CollegeID)
		if err != nil {
			return fmt.Errorf("failed to verify new college existence: %w", err)
		}
		if college == nil {
			return fmt.Errorf("new college with ID %d does not exist", course.CollegeID)
		}
	}

	// Business logic validation: Check for duplicate course names (if name is being changed)
	if course.Name != existingCourse.Name {
		exists, err := c.courseRepo.CheckCourseNameExists(ctx, course.CollegeID, course.Name, &course.ID)
		if err != nil {
			return fmt.Errorf("failed to check course name uniqueness: %w", err)
		}
		if exists {
			return fmt.Errorf("course with name '%s' already exists in college %d", course.Name, course.CollegeID)
		}
	}

	return c.courseRepo.UpdateCourse(ctx, course)
}

func (c *courseService) DeleteCourse(ctx context.Context, collegeID int, courseID int) error {
	// Input validation
	if collegeID <= 0 {
		return errors.New("invalid college ID")
	}
	if courseID <= 0 {
		return errors.New("invalid course ID")
	}

	// Business logic validation: Verify course exists before deletion
	course, err := c.courseRepo.FindCourseByID(ctx, collegeID, courseID)
	if err != nil {
		return fmt.Errorf("failed to verify course existence: %w", err)
	}
	if course == nil {
		return fmt.Errorf("course with ID %d not found in college %d", courseID, collegeID)
	}

	// Business logic validation: Check for dependent resources
	// This could include enrollments, lectures, assignments, etc.
	// For now, we'll let the database handle foreign key constraints

	return c.courseRepo.DeleteCourse(ctx, collegeID, courseID)
}

// FindAllCourses retrieves all courses for a college with pagination and validation
func (c *courseService) FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error) {
	// Input validation
	if collegeID <= 0 {
		return nil, errors.New("invalid college ID")
	}

	// Apply reasonable limits to prevent excessive queries
	if limit > 100 {
		limit = 100
	}

	// Business logic validation: Verify college exists
	_, err := c.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify college existence: %w", err)
	}

	return c.courseRepo.FindAllCourses(ctx, collegeID, limit, offset)
}

// FindCoursesByInstructor retrieves courses taught by a specific instructor with validation
func (c *courseService) FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error) {
	// Input validation
	if collegeID <= 0 {
		return nil, errors.New("invalid college ID")
	}
	if instructorID <= 0 {
		return nil, errors.New("invalid instructor ID")
	}

	// Apply reasonable limits to prevent excessive queries
	if limit > 100 {
		limit = 100
	}

	// Business logic validation: Verify college exists
	_, err := c.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify college existence: %w", err)
	}

	return c.courseRepo.FindCoursesByInstructor(ctx, collegeID, instructorID, limit, offset)
}

// CountCoursesByCollege returns the total number of courses in a college with validation
func (c *courseService) CountCoursesByCollege(ctx context.Context, collegeID int) (int, error) {
	// Input validation
	if collegeID <= 0 {
		return 0, errors.New("invalid college ID")
	}

	// Business logic validation: Verify college exists
	_, err := c.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify college existence: %w", err)
	}

	return c.courseRepo.CountCoursesByCollege(ctx, collegeID)
}

// CountCoursesByInstructor returns the number of courses taught by an instructor with validation
func (c *courseService) CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error) {
	// Input validation
	if collegeID <= 0 {
		return 0, errors.New("invalid college ID")
	}
	if instructorID <= 0 {
		return 0, errors.New("invalid instructor ID")
	}

	// Business logic validation: Verify college exists
	_, err := c.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return 0, fmt.Errorf("failed to verify college existence: %w", err)
	}

	return c.courseRepo.CountCoursesByInstructor(ctx, collegeID, instructorID)
}

// UpdateCoursePartial updates specific fields of a course with partial validation
func (c *courseService) UpdateCoursePartial(ctx context.Context, collegeID int, courseID int, req *models.UpdateCourseRequest) error {
	// Input validation
	if collegeID <= 0 {
		return errors.New("invalid college ID")
	}
	if courseID <= 0 {
		return errors.New("invalid course ID")
	}
	if req == nil {
		return errors.New("update request cannot be nil")
	}

	// Validate the request struct
	if err := c.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed for course update: %w", err)
	}

	// Business logic validation: Verify course exists
	course, err := c.courseRepo.FindCourseByID(ctx, collegeID, courseID)
	if err != nil {
		return fmt.Errorf("failed to verify course existence: %w", err)
	}
	if course == nil {
		return fmt.Errorf("course with ID %d not found in college %d", courseID, collegeID)
	}

	// Business logic validation: Verify college exists (if college ID is being changed)
	if req.CollegeID != nil && *req.CollegeID != collegeID {
		newCollege, err := c.collegeRepo.GetCollegeByID(ctx, *req.CollegeID)
		if err != nil {
			return fmt.Errorf("failed to verify new college existence: %w", err)
		}
		if newCollege == nil {
			return fmt.Errorf("new college with ID %d does not exist", *req.CollegeID)
		}
	}

	// Business logic validation: Check for duplicate course names (if name is being changed)
	if req.Name != nil {
		targetCollegeID := collegeID
		if req.CollegeID != nil {
			targetCollegeID = *req.CollegeID
		}

		exists, err := c.courseRepo.CheckCourseNameExists(ctx, targetCollegeID, *req.Name, &courseID)
		if err != nil {
			return fmt.Errorf("failed to check course name uniqueness: %w", err)
		}
		if exists {
			return fmt.Errorf("course with name '%s' already exists in college %d", *req.Name, targetCollegeID)
		}
	}

	return c.courseRepo.UpdateCoursePartial(ctx, collegeID, courseID, req)
}
