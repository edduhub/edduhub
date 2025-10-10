package calendar

import (
	"context"
	"errors"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type CalendarService interface {
	CreateEvent(ctx context.Context, event *models.CalendarBlock) error
	GetEvent(ctx context.Context, collegeID int, eventID int) (*models.CalendarBlock, error)
	GetEvents(ctx context.Context, filter models.CalendarBlockFilter) ([]*models.CalendarBlock, error)
	UpdateEvent(ctx context.Context, collegeID int, eventID int, req *models.UpdateCalendarRequest) error
	DeleteEvent(ctx context.Context, collegeID int, eventID int) error
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
