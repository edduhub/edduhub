package announcement

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type AnnouncementService interface {
	CreateAnnouncement(ctx context.Context, announcement *models.Announcement) error
	GetAnnouncement(ctx context.Context, collegeID, announcementID int) (*models.Announcement, error)
	GetAnnouncements(ctx context.Context, filter models.AnnouncementFilter) ([]*models.Announcement, error)
	UpdateAnnouncement(ctx context.Context, collegeID, announcementID int, req *models.UpdateAnnouncementRequest) error
	DeleteAnnouncement(ctx context.Context, collegeID, announcementID int) error
}

type announcementService struct {
	announcementRepo repository.AnnouncementRepository
}

func NewAnnouncementService(announcementRepo repository.AnnouncementRepository) AnnouncementService {
	return &announcementService{
		announcementRepo: announcementRepo,
	}
}

func (s *announcementService) CreateAnnouncement(ctx context.Context, announcement *models.Announcement) error {
	if announcement.Title == "" {
		return fmt.Errorf("announcement title is required")
	}
	if announcement.Content == "" {
		return fmt.Errorf("announcement content is required")
	}
	return s.announcementRepo.CreateAnnouncement(ctx, announcement)
}

func (s *announcementService) GetAnnouncement(ctx context.Context, collegeID, announcementID int) (*models.Announcement, error) {
	return s.announcementRepo.GetAnnouncementByID(ctx, collegeID, announcementID)
}

func (s *announcementService) GetAnnouncements(ctx context.Context, filter models.AnnouncementFilter) ([]*models.Announcement, error) {
	return s.announcementRepo.GetAnnouncements(ctx, filter)
}

func (s *announcementService) UpdateAnnouncement(ctx context.Context, collegeID, announcementID int, req *models.UpdateAnnouncementRequest) error {
	return s.announcementRepo.UpdateAnnouncementPartial(ctx, collegeID, announcementID, req)
}

func (s *announcementService) DeleteAnnouncement(ctx context.Context, collegeID, announcementID int) error {
	return s.announcementRepo.DeleteAnnouncement(ctx, collegeID, announcementID)
}
