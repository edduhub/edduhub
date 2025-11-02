package calendar

import (
	"context"
	"errors"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type CalendarService interface {
	CreateEvent(ctx context.Context, event *models.CalendarBlock) error
	GetEvent(ctx context.Context, collegeID int, eventID int) (*models.CalendarBlock, error)
	GetEvents(ctx context.Context, filter models.CalendarBlockFilter) ([]*models.CalendarBlock, error)
	UpdateEvent(ctx context.Context, collegeID int, eventID int, req *models.UpdateCalendarRequest) error
	DeleteEvent(ctx context.Context, collegeID int, eventID int) error
	GetEventsByCourse(ctx context.Context, collegeID, courseID int) ([]*models.CalendarBlock, error)
	GetUpcomingEvents(ctx context.Context, collegeID int, limit int) ([]*models.CalendarBlock, error)
	SearchEvents(ctx context.Context, collegeID int, query string) ([]*models.CalendarBlock, error)
}

type calendarService struct {
	calendarRepo repository.CalendarRepository
}

func NewCalendarService(calendarRepo repository.CalendarRepository) CalendarService {
	return &calendarService{
		calendarRepo: calendarRepo,
	}
}

func (s *calendarService) CreateEvent(ctx context.Context, event *models.CalendarBlock) error {
	if event == nil {
		return errors.New("event cannot be nil")
	}
	return s.calendarRepo.CreateCalendarBlock(ctx, event)
}

func (s *calendarService) GetEvent(ctx context.Context, collegeID int, eventID int) (*models.CalendarBlock, error) {
	return s.calendarRepo.GetCalendarBlockByID(ctx, eventID, collegeID)
}

func (s *calendarService) GetEvents(ctx context.Context, filter models.CalendarBlockFilter) ([]*models.CalendarBlock, error) {
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.calendarRepo.GetCalendarBlocks(ctx, filter)
}

func (s *calendarService) UpdateEvent(ctx context.Context, collegeID int, eventID int, req *models.UpdateCalendarRequest) error {
	return s.calendarRepo.UpdateCalendarBlockPartial(ctx, collegeID, eventID, req)
}

func (s *calendarService) DeleteEvent(ctx context.Context, collegeID int, eventID int) error {
	return s.calendarRepo.DeleteCalendarBlock(ctx, eventID, collegeID)
}

func (s *calendarService) GetEventsByCourse(ctx context.Context, collegeID, courseID int) ([]*models.CalendarBlock, error) {
	filter := models.CalendarBlockFilter{
		CollegeID: &collegeID,
		CourseID:  &courseID,
		Limit:     100,
	}
	return s.calendarRepo.GetCalendarBlocks(ctx, filter)
}

func (s *calendarService) GetUpcomingEvents(ctx context.Context, collegeID int, limit int) ([]*models.CalendarBlock, error) {
	now := time.Now()
	filter := models.CalendarBlockFilter{
		CollegeID: &collegeID,
		StartDate: &now,
		Limit:     uint64(limit),
	}
	return s.calendarRepo.GetCalendarBlocks(ctx, filter)
}

func (s *calendarService) SearchEvents(ctx context.Context, collegeID int, query string) ([]*models.CalendarBlock, error) {
	filter := models.CalendarBlockFilter{
		CollegeID: &collegeID,
		Search:    &query,
		Limit:     50,
	}
	return s.calendarRepo.GetCalendarBlocks(ctx, filter)
}
