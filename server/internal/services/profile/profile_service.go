package profile

import (
	"context"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type ProfileService interface {
	GetProfileByUserID(ctx context.Context, userID int) (*models.Profile, error)
	GetProfileByKratosID(ctx context.Context, kratosID string) (*models.Profile, error)
	GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error)
	UpdateProfile(ctx context.Context, userID int, req *models.UpdateProfileRequest) error
	CreateProfile(ctx context.Context, profile *models.Profile) error
	GetProfileHistory(ctx context.Context, profileID int, limit, offset int) ([]*models.ProfileHistory, error)
	CreateProfileHistory(ctx context.Context, history *models.ProfileHistory) error
}

type profileService struct {
	profileRepo repository.ProfileRepository
}

func NewProfileService(profileRepo repository.ProfileRepository) ProfileService {
	return &profileService{
		profileRepo: profileRepo,
	}
}

func (s *profileService) GetProfileByUserID(ctx context.Context, userID int) (*models.Profile, error) {
	return s.profileRepo.GetProfileByUserID(ctx, userID)
}

func (s *profileService) GetProfileByKratosID(ctx context.Context, kratosID string) (*models.Profile, error) {
	return s.profileRepo.GetProfileByKratosID(ctx, kratosID)
}

func (s *profileService) GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error) {
	return s.profileRepo.GetProfileByID(ctx, profileID)
}

func (s *profileService) UpdateProfile(ctx context.Context, userID int, req *models.UpdateProfileRequest) error {
	profile, err := s.profileRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return s.profileRepo.UpdateProfilePartial(ctx, profile.ID, req)
}

func (s *profileService) CreateProfile(ctx context.Context, profile *models.Profile) error {
	return s.profileRepo.CreateProfile(ctx, profile)
}

func (s *profileService) GetProfileHistory(ctx context.Context, profileID int, limit, offset int) ([]*models.ProfileHistory, error) {
	return s.profileRepo.GetProfileHistory(ctx, profileID, limit, offset)
}

func (s *profileService) CreateProfileHistory(ctx context.Context, history *models.ProfileHistory) error {
	return s.profileRepo.CreateProfileHistory(ctx, history)
}
