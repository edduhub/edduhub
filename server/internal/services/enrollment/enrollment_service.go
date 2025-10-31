package enrollment

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

type EnrollmentService interface {
	CreateEnrollment(ctx context.Context, enrollment *models.Enrollment) error
	IsStudentEnrolled(ctx context.Context, collegeID, studentID, courseID int) (bool, error)
	UpdateEnrollment(ctx context.Context, enrollment *models.Enrollment) error
	UpdateEnrollmentStatus(ctx context.Context, collegeID, enrollmentID int, Newstatus string) error
	DeleteEnrollment(ctx context.Context, collegeID int, enrollmentID int) error
	FindEnrollmentsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.Enrollment, error)
	FindEnrollmentsByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Enrollment, error)
	GetEnrollmentByID(ctx context.Context, collegeID int, enrollmentID int) (*models.Enrollment, error)
}

type enrollmentService struct {
	enrollmentRepo repository.EnrollmentRepository
	validate       *validator.Validate
}

func NewEnrollmentService(enrollmentRepo repository.EnrollmentRepository) EnrollmentService {
	return &enrollmentService{
		enrollmentRepo: enrollmentRepo,
		validate:       validator.New(),
	}
}

func (e *enrollmentService) CreateEnrollment(ctx context.Context, enrollment *models.Enrollment) error {
	if err := e.validate.Struct(enrollment); err != nil {
		return fmt.Errorf("struct validaton failed %w", err)
	}
	return e.enrollmentRepo.CreateEnrollment(ctx, enrollment)
}

func (e *enrollmentService) IsStudentEnrolled(ctx context.Context, collegeID, studentID, courseID int) (bool, error) {
	return e.enrollmentRepo.IsStudentEnrolled(ctx, collegeID, studentID, courseID)
}

func (e *enrollmentService) UpdateEnrollment(ctx context.Context, enrollment *models.Enrollment) error {
	if err := e.validate.Struct(enrollment); err != nil {
		return fmt.Errorf("validation failed Error :%w", err)
	}

	return e.enrollmentRepo.UpdateEnrollment(ctx, enrollment)
}

func (e *enrollmentService) UpdateEnrollmentStatus(ctx context.Context, collegeID int, enrollmentID int, NewStatus string) error {
	if NewStatus != models.Active && NewStatus != models.Inactive && NewStatus != models.Completed {
		return fmt.Errorf("cannot change to %s status", NewStatus)
	}
	return e.enrollmentRepo.UpdateEnrollmentStatus(ctx, collegeID, enrollmentID, NewStatus)
}

func (e *enrollmentService) DeleteEnrollment(ctx context.Context, collegeID int, enrollmentID int) error {
	return e.enrollmentRepo.DeleteEnrollment(ctx, collegeID, enrollmentID)
}

func (e *enrollmentService) FindEnrollmentsByStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.Enrollment, error) {
	return e.enrollmentRepo.FindEnrollmentsByStudent(ctx, collegeID, studentID, limit, offset)
}

func (e *enrollmentService) FindEnrollmentsByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Enrollment, error) {
	return e.enrollmentRepo.FindEnrollmentsByCourse(ctx, collegeID, courseID, limit, offset)
}

func (e *enrollmentService) GetEnrollmentByID(ctx context.Context, collegeID, enrollmentID int) (*models.Enrollment, error) {
	enrollments, err := e.enrollmentRepo.GetEnrollmentByID(ctx, collegeID, enrollmentID)
	if err != nil {
		return nil, err
	}
	return enrollments, nil
}
