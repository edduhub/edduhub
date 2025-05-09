package lecture

import (
	"context"
	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type LectureService interface {
	CreateLecture(ctx context.Context, lecture *models.Lecture) error
	GetLectureByID(ctx context.Context, collegeID int, lectureID int) (*models.Lecture, error)
	UpdateLecture(ctx context.Context, lecture *models.Lecture) error
	DeleteLecture(ctx context.Context, collegeID int, lectureID int) error

	// Finder methods
	FindLecturesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Lecture, error)
	CountLecturesByCourse(ctx context.Context, collegeID int, courseID int) (int, error)
}

type lectureService struct {
	lectureRepo repository.LectureRepository
	validate    validator.Validate
}

func NewLectureService(lectureRepo repository.LectureRepository) LectureService {
	return &lectureService{
		lectureRepo: lectureRepo,
		validate:    *validator.New(),
	}
}

func (l *lectureService) CreateLecture(ctx context.Context, lecture *models.Lecture) error {
	if err := l.validate.Struct(lecture); err != nil {
		return fmt.Errorf("validation failed %w", err)
	}
	return l.CreateLecture(ctx, lecture)
}

func (l *lectureService) GetLectureByID(ctx context.Context, collegeID int, lectureID int) (*models.Lecture, error) {
	return l.lectureRepo.GetLectureByID(ctx, collegeID, lectureID)
}

func (l *lectureService) UpdateLecture(ctx context.Context, lecture *models.Lecture) error {
	if err := l.validate.Struct(lecture); err != nil {
		return fmt.Errorf("validation failed %w", err)
	}
	return l.lectureRepo.UpdateLecture(ctx, lecture)
}

func (l *lectureService) DeleteLecture(ctx context.Context, collegeID int, lectureID int) error {
	return l.lectureRepo.DeleteLecture(ctx, collegeID, lectureID)
}

func (l *lectureService) FindLecturesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Lecture, error) {
	return l.lectureRepo.FindLecturesByCourse(ctx, collegeID, courseID, limit, offset)
}

func (l *lectureService) CountLecturesByCourse(ctx context.Context, collegeID int, courseID int) (int, error) {
	return l.lectureRepo.CountLecturesByCourse(ctx, collegeID, courseID)
}
