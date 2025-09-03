package course

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

type CourseService interface {
	CreateCourse(ctx context.Context, course *models.Course) error
	FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error) // Added collegeID
	UpdateCourse(ctx context.Context, courseID int, course *models.Course) error
	UpdateCoursePartial(ctx context.Context, collegeID int, courseID int, req *models.UpdateCourseRequest) error
	DeleteCourse(ctx context.Context, collegeID int, courseID int) error

	// Find methods with pagination
	FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error)
	FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error)

	// Count methods
	CountCoursesByCollege(ctx context.Context, collegeID int) (int, error)
	CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error)
}

type courseService struct {
	courseRepo repository.CourseRepository
	validate   *validator.Validate
}

func NewCourseService(courseRepo repository.CourseRepository) CourseService {
	return &courseService{
		courseRepo: courseRepo,
		validate:   validator.New(),
	}
}

func (c *courseService) CreateCourse(ctx context.Context, course *models.Course) error {
	if err := c.validate.Struct(course); err != nil {
		return fmt.Errorf("validation error %w", err)
	}
	return c.courseRepo.CreateCourse(ctx, course)
}

func (c *courseService) FindCourseByID(ctx context.Context, collegeID int, courseID int) (*models.Course, error) {
	return c.courseRepo.FindCourseByID(ctx, collegeID, courseID)
}

func (c *courseService) UpdateCourse(ctx context.Context, courseID int, course *models.Course) error {
	if course.ID != courseID {
		return fmt.Errorf("courseID not matching")
	}

	if err := c.validate.Struct(course); err != nil {
		return fmt.Errorf("validation error %w", err)
	}

	return c.courseRepo.UpdateCourse(ctx, course)
}

func (c *courseService) DeleteCourse(ctx context.Context, collegeID int, courseID int) error {
	return c.courseRepo.DeleteCourse(ctx, collegeID, courseID)
}

// Find methods with pagination
func (c *courseService) FindAllCourses(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Course, error) {
	return c.courseRepo.FindAllCourses(ctx, collegeID, limit, offset)
}

func (c *courseService) FindCoursesByInstructor(ctx context.Context, collegeID int, instructorID int, limit, offset uint64) ([]*models.Course, error) {
	return c.courseRepo.FindCoursesByInstructor(ctx, collegeID, instructorID, limit, offset)
}

// Count methods
func (c *courseService) CountCoursesByCollege(ctx context.Context, collegeID int) (int, error) {
	return c.courseRepo.CountCoursesByCollege(ctx, collegeID)
}

func (c *courseService) CountCoursesByInstructor(ctx context.Context, collegeID int, instructorID int) (int, error) {
	return c.courseRepo.CountCoursesByInstructor(ctx, collegeID, instructorID)
}

func (c *courseService) UpdateCoursePartial(ctx context.Context, collegeID int, courseID int, req *models.UpdateCourseRequest) error {
	if err := c.validate.Struct(req); err != nil {
		return fmt.Errorf("validation failed for course update: %w", err)
	}
	return c.courseRepo.UpdateCoursePartial(ctx, collegeID, courseID, req)
}
