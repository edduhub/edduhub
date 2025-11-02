package timetable

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type TimetableService interface {
	CreateTimeTableBlock(ctx context.Context, block *models.TimeTableBlock) error
	GetTimeTableBlock(ctx context.Context, blockID int, collegeID int) (*models.TimeTableBlock, error)
	UpdateTimeTableBlock(ctx context.Context, block *models.TimeTableBlock) error
	DeleteTimeTableBlock(ctx context.Context, blockID int, collegeID int) error
	GetTimeTableBlocks(ctx context.Context, filter models.TimeTableBlockFilter) ([]*models.TimeTableBlock, error)
	GetStudentTimetable(ctx context.Context, studentID int) ([]*models.TimeTableBlock, error)
	GetFacultyTimetable(ctx context.Context, facultyID string, collegeID int) ([]*models.TimeTableBlock, error)
}

type timetableService struct {
	timetableRepo repository.TimeTableRepository
	studentRepo   repository.StudentRepository
}

func NewTimetableService(timetableRepo repository.TimeTableRepository, studentRepo repository.StudentRepository) TimetableService {
	return &timetableService{
		timetableRepo: timetableRepo,
		studentRepo:   studentRepo,
	}
}

func (s *timetableService) CreateTimeTableBlock(ctx context.Context, block *models.TimeTableBlock) error {
	return s.timetableRepo.CreateTimeTableBlock(ctx, block)
}

func (s *timetableService) GetTimeTableBlock(ctx context.Context, blockID int, collegeID int) (*models.TimeTableBlock, error) {
	return s.timetableRepo.GetTimeTableBlockByID(ctx, blockID, collegeID)
}

func (s *timetableService) UpdateTimeTableBlock(ctx context.Context, block *models.TimeTableBlock) error {
	return s.timetableRepo.UpdateTimeTableBlock(ctx, block)
}

func (s *timetableService) DeleteTimeTableBlock(ctx context.Context, blockID int, collegeID int) error {
	return s.timetableRepo.DeleteTimeTableBlock(ctx, blockID, collegeID)
}

func (s *timetableService) GetTimeTableBlocks(ctx context.Context, filter models.TimeTableBlockFilter) ([]*models.TimeTableBlock, error) {
	return s.timetableRepo.GetTimeTableBlocks(ctx, filter)
}

func (s *timetableService) GetStudentTimetable(ctx context.Context, studentID int) ([]*models.TimeTableBlock, error) {
	// Get student to find their college
	student, err := s.studentRepo.GetStudentByID(ctx, studentID, 0) // College ID will be fetched from student record
	if err != nil {
		return nil, fmt.Errorf("student not found: %w", err)
	}

	// Get timetable blocks for student's college
	filter := models.TimeTableBlockFilter{
		CollegeID: student.CollegeID,
	}

	return s.timetableRepo.GetTimeTableBlocks(ctx, filter)
}

func (s *timetableService) GetFacultyTimetable(ctx context.Context, facultyID string, collegeID int) ([]*models.TimeTableBlock, error) {
	filter := models.TimeTableBlockFilter{
		CollegeID:    collegeID,
		InstructorID: &facultyID,
	}

	return s.timetableRepo.GetTimeTableBlocks(ctx, filter)
}
