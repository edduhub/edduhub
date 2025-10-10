package grades

import (
	"context"
	"fmt"
	"math"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/go-playground/validator/v10"
)

type GradeServices interface {
	CreateGrade(ctx context.Context, grade *models.Grade) error
	GetGradeByID(ctx context.Context, gradeId int, collegeID int) (*models.Grade, error)
	UpdateGrade(ctx context.Context, grade *models.Grade) error
	UpdateGradePartial(ctx context.Context, collegeID int, gradeID int, req *models.UpdateGradeRequest) error
	DeleteGrade(ctx context.Context, gradeID int, collegeID int) error
	GetGrades(ctx context.Context, filter models.GradeFilter) ([]*models.Grade, error)
	GetGradesByCourse(ctx context.Context, collegeID int, courseID int) ([]*models.Grade, error)
	GetGradesByStudent(ctx context.Context, collegeID int, studentID int) ([]*models.Grade, error)
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
	if grade.TotalMarks <= 0 {
		return fmt.Errorf("total marks must be greater than zero")
	}

	if grade.ObtainedMarks < 0 || grade.ObtainedMarks > grade.TotalMarks {
		return fmt.Errorf("obtained marks must be between 0 and total marks")
	}

	if grade.Percentage == 0 {
		grade.Percentage = math.Round(float64(grade.ObtainedMarks)/float64(grade.TotalMarks)*10000) / 100
	}

	if err := g.validate.Struct(grade); err != nil {
		return fmt.Errorf("unable to validate grade: %w", err)
	}

	return g.gradeRepo.CreateGrade(ctx, grade)
}

func (g *gradeServices) GetGradeByID(ctx context.Context, gradeID int, collegeID int) (*models.Grade, error) {
	return g.gradeRepo.GetGradeByID(ctx, gradeID, collegeID)
}

func (g *gradeServices) UpdateGrade(ctx context.Context, grade *models.Grade) error {
	if grade.TotalMarks <= 0 {
		return fmt.Errorf("total marks must be greater than zero")
	}

	if grade.ObtainedMarks < 0 || grade.ObtainedMarks > grade.TotalMarks {
		return fmt.Errorf("obtained marks must be between 0 and total marks")
	}

	if grade.Percentage == 0 {
		grade.Percentage = math.Round(float64(grade.ObtainedMarks)/float64(grade.TotalMarks)*10000) / 100
	}

	if err := g.validate.Struct(grade); err != nil {
		return fmt.Errorf("unable to validate grade: %w", err)
	}

	return g.gradeRepo.UpdateGrade(ctx, grade)
}

func (g *gradeServices) DeleteGrade(ctx context.Context, gradeID int, collegeID int) error {
	return g.gradeRepo.DeleteGrade(ctx, gradeID, collegeID)
}

func (g *gradeServices) GetGrades(ctx context.Context, filters models.GradeFilter) ([]*models.Grade, error) {
	return g.gradeRepo.GetGrades(ctx, filters)
}

func (g *gradeServices) UpdateGradePartial(ctx context.Context, collegeID int, gradeID int, req *models.UpdateGradeRequest) error {
	if req == nil {
		return fmt.Errorf("update request cannot be nil")
	}

	if req.TotalMarks != nil && *req.TotalMarks <= 0 {
		return fmt.Errorf("total marks must be greater than zero")
	}

	if req.ObtainedMarks != nil && req.TotalMarks != nil {
		if *req.ObtainedMarks < 0 || *req.ObtainedMarks > *req.TotalMarks {
			return fmt.Errorf("obtained marks must be between 0 and total marks")
		}
	}

	if req.ObtainedMarks != nil && (req.TotalMarks != nil || req.Percentage == nil) {
		// calculate derived percentage when total marks provided or existing percentage is not supplied
		total := req.TotalMarks
		if total == nil {
			// fetch existing grade to use its total marks
			existing, err := g.gradeRepo.GetGradeByID(ctx, gradeID, collegeID)
			if err != nil {
				return err
			}
			val := existing.TotalMarks
			total = &val
		}
		if *total <= 0 {
			return fmt.Errorf("total marks must be greater than zero")
		}
		percentage := math.Round(float64(*req.ObtainedMarks)/float64(*total)*10000) / 100
		req.Percentage = &percentage
	}

	return g.gradeRepo.UpdateGradePartial(ctx, collegeID, gradeID, req)
}

func (g *gradeServices) GetGradesByCourse(ctx context.Context, collegeID int, courseID int) ([]*models.Grade, error) {
	filter := models.GradeFilter{
		CollegeID: &collegeID,
		CourseID:  &courseID,
	}
	return g.gradeRepo.GetGrades(ctx, filter)
}

func (g *gradeServices) GetGradesByStudent(ctx context.Context, collegeID int, studentID int) ([]*models.Grade, error) {
	return g.gradeRepo.GetGradesByStudent(ctx, collegeID, studentID)
}
