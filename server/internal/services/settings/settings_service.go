package settings

import (
	"context"
	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type SettingsService interface {
	GetSettings(ctx context.Context, userID string) (*models.Settings, error)
	UpdateSettings(ctx context.Context, userID string, req *models.SettingsUpdateRequest) (*models.Settings, error)
}

type settingsService struct {
	settingsRepo repository.SettingsRepository
}

func NewSettingsService(settingsRepo repository.SettingsRepository) SettingsService {
	return &settingsService{
		settingsRepo: settingsRepo,
	}
}

func (s *settingsService) GetSettings(ctx context.Context, userID string) (*models.Settings, error) {
	return s.settingsRepo.GetSettings(ctx, userID)
}

func (s *settingsService) UpdateSettings(ctx context.Context, userID string, req *models.SettingsUpdateRequest) (*models.Settings, error) {
	return s.settingsRepo.UpdateSettings(ctx, userID, req)
}
