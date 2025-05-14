package grades

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

type GradeServices interface {
	CreateGrade(ctx context.Context, grade *models.Grade) error
	GetGradeByID(ctx context.Context, gradeId int, collegeID int) (*models.Grade, error)
	UpdateGrade(ctx context.Context, grade *models.Grade) error
	DeleteGrade(ctx context.Context, gradeID int, collegeID int) error
	GetGrades(ctx context.Context, filter models.GradeFilter) ([]*models.Grade, error)
	// CalculateAndStoreStudentGPA(ctx context.Context,collegeID int,RollNo string)error
}

type gradeServices struct {
	gradeRepo      repository.GradeRepository
	studentRepo    repository.StudentRepository
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository

	validate validator.Validate
}

func NewGradeServices(gradeRepo repository.GradeRepository, studentRepo repository.StudentRepository, enrollmentRepo repository.EnrollmentRepository, courseRepo repository.CourseRepository) GradeServices {
	return &gradeServices{
		gradeRepo:      gradeRepo,
		studentRepo:    studentRepo,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		validate:       *validator.New(),
	}
}

func (g *gradeServices) CreateGrade(ctx context.Context, grade *models.Grade) error {
	if err := g.validate.Struct(grade); err != nil {
		return fmt.Errorf("unable to validate %w", err)
	}
	return g.gradeRepo.CreateGrade(ctx, grade)
}

func (g *gradeServices) GetGradeByID(ctx context.Context, gradeID int, collegeID int) (*models.Grade, error) {
	return g.gradeRepo.GetGradeByID(ctx, gradeID, collegeID)
}

func (g *gradeServices) UpdateGrade(ctx context.Context, grade *models.Grade) error {
	if err := g.validate.Struct(grade); err != nil {
		return fmt.Errorf("unable to validate %w", err)
	}
	return g.gradeRepo.UpdateGrade(ctx, grade)
}

func (g *gradeServices) DeleteGrade(ctx context.Context, gradeID int, collegeID int) error {
	return g.gradeRepo.DeleteGrade(ctx, gradeID, collegeID)
}

func (g *gradeServices) GetGrades(ctx context.Context, filters models.GradeFilter) ([]*models.Grade, error) {
	return g.gradeRepo.GetGrades(ctx, filters)
}
