package profile

import (
	"context"
	"strconv"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type ProfileService interface {
	GetProfileByUserID(ctx context.Context, userID int) (*models.Profile, error)
	GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error)
	UpdateProfile(ctx context.Context, userID int, req *models.UpdateProfileRequest) error
	CreateProfile(ctx context.Context, profile *models.Profile) error
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
	return s.profileRepo.GetProfileByUserID(ctx, strconv.Itoa(userID))
}

func (s *profileService) GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error) {
	return s.profileRepo.GetProfileByID(ctx, profileID)
}

func (s *profileService) UpdateProfile(ctx context.Context, userID int, req *models.UpdateProfileRequest) error {
	// Get existing profile first
	profile, err := s.profileRepo.GetProfileByUserID(ctx, strconv.Itoa(userID))
	if err != nil {
		return err
	}
	
	return s.profileRepo.UpdateProfilePartial(ctx, strconv.Itoa(profile.ID), req)
}

func (s *profileService) CreateProfile(ctx context.Context, profile *models.Profile) error {
	return s.profileRepo.CreateProfile(ctx, profile)
}
