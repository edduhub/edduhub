// Package announcement provides the business logic for managing announcements.
// It orchestrates the interaction between the API layer and the data repository,
// enforcing business rules, validation, and authorization.
package announcement

import (
	"context"
	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// AnnouncementService defines the interface for announcement-related business logic.
// It ensures that all operations are authorized, validated, and consistent with
// the application's business rules.
type AnnouncementService interface {
	CreateAnnouncement(ctx context.Context, req *models.CreateAnnouncementRequest, collegeID, userID int) (*models.Announcement, error)
	GetAnnouncementByID(ctx context.Context, id int) (*models.Announcement, error)
	GetAnnouncementsByCollegeID(ctx context.Context, collegeID, limit, offset int) ([]*models.Announcement, error)
	UpdateAnnouncement(ctx context.Context, id int, req *models.UpdateAnnouncementRequest) (*models.Announcement, error)
	DeleteAnnouncement(ctx context.Context, id int) error
}

// announcementService is the implementation of the AnnouncementService interface.
type announcementService struct {
	announcementRepo repository.AnnouncementRepository
	collegeRepo      repository.CollegeRepository
	userRepo         repository.UserRepository
	validate         *validator.Validate
}

// NewService creates a new instance of AnnouncementService with its dependencies.
func NewService(
	announcementRepo repository.AnnouncementRepository,
	collegeRepo repository.CollegeRepository,
	userRepo repository.UserRepository,
) AnnouncementService {
	return &announcementService{
		announcementRepo: announcementRepo,
		collegeRepo:      collegeRepo,
		userRepo:         userRepo,
		validate:         validator.New(),
	}
}

// CreateAnnouncement handles the creation of a new announcement.
// It validates the request, checks for existence of related entities (college, user),
// and then persists the new announcement.
func (s *announcementService) CreateAnnouncement(ctx context.Context, req *models.CreateAnnouncementRequest, collegeID, userID int) (*models.Announcement, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify that the college exists
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify college: %w", err)
	}

	// Verify that the user (author) exists
	_, err = s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}

	announcement := &models.Announcement{
		Title:     req.Title,
		Content:   req.Content,
		CollegeID: collegeID,
		UserID:    userID,
	}

	id, err := s.announcementRepo.Create(ctx, announcement)
	if err != nil {
		return nil, fmt.Errorf("failed to create announcement: %w", err)
	}
	announcement.ID = id

	return announcement, nil
}

// GetAnnouncementByID retrieves a single announcement by its ID.
func (s *announcementService) GetAnnouncementByID(ctx context.Context, id int) (*models.Announcement, error) {
	if id <= 0 {
		return nil, errors.New("invalid announcement ID")
	}
	return s.announcementRepo.GetByID(ctx, id)
}

// GetAnnouncementsByCollegeID retrieves a paginated list of announcements for a given college.
func (s *announcementService) GetAnnouncementsByCollegeID(ctx context.Context, collegeID, limit, offset int) ([]*models.Announcement, error) {
	if collegeID <= 0 {
		return nil, errors.New("invalid college ID")
	}
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if offset < 0 {
		offset = 0 // Default offset
	}

	// Verify that the college exists before fetching announcements
	_, err := s.collegeRepo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify college: %w", err)
	}

	return s.announcementRepo.GetByCollegeID(ctx, collegeID, limit, offset)
}

// UpdateAnnouncement handles the logic for updating an existing announcement.
// It ensures the announcement exists before attempting to apply changes.
func (s *announcementService) UpdateAnnouncement(ctx context.Context, id int, req *models.UpdateAnnouncementRequest) (*models.Announcement, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	announcement, err := s.announcementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve announcement for update: %w", err)
	}

	if req.Title != nil {
		announcement.Title = *req.Title
	}
	if req.Content != nil {
		announcement.Content = *req.Content
	}

	if err := s.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, fmt.Errorf("failed to update announcement: %w", err)
	}

	return announcement, nil
}

// DeleteAnnouncement handles the deletion of an announcement.
// It first verifies the announcement exists before proceeding with deletion.
func (s *announcementService) DeleteAnnouncement(ctx context.Context, id int) error {
	// Ensure the announcement exists before trying to delete it.
	_, err := s.announcementRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to retrieve announcement for deletion: %w", err)
	}

	return s.announcementRepo.DeleteByID(ctx, id)
}
