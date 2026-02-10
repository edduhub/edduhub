package repository

import (
	"context"
	"errors"

	"eduhub/server/internal/models"

	"github.com/jackc/pgx/v5"
)

type SettingsRepository interface {
	GetSettings(ctx context.Context, userID string) (*models.Settings, error)
	UpdateSettings(ctx context.Context, userID string, req *models.SettingsUpdateRequest) (*models.Settings, error)
}

type settingsRepository struct {
	db PoolIface
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *DB) SettingsRepository {
	return &settingsRepository{
		db: db.Pool,
	}
}

func (r *settingsRepository) GetSettings(ctx context.Context, userID string) (*models.Settings, error) {
	query := `
		SELECT user_id, email_notifications, push_notifications, theme, language, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1`

	var s models.Settings
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&s.UserID, &s.EmailNotifications, &s.PushNotifications, &s.Theme, &s.Language, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		// Return default settings if not found
		return &models.Settings{
			UserID:             userID,
			EmailNotifications: true,
			PushNotifications:  true,
			Theme:              "system",
			Language:           "en",
		}, nil
	}
	return &s, nil
}

func (r *settingsRepository) UpdateSettings(ctx context.Context, userID string, req *models.SettingsUpdateRequest) (*models.Settings, error) {
	existing, err := r.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.EmailNotifications != nil {
		existing.EmailNotifications = *req.EmailNotifications
	}
	if req.PushNotifications != nil {
		existing.PushNotifications = *req.PushNotifications
	}
	if req.Theme != nil {
		existing.Theme = *req.Theme
	}
	if req.Language != nil {
		existing.Language = *req.Language
	}
	query := `
		INSERT INTO user_settings (user_id, email_notifications, push_notifications, theme, language, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			email_notifications = EXCLUDED.email_notifications,
			push_notifications = EXCLUDED.push_notifications,
			theme = EXCLUDED.theme,
			language = EXCLUDED.language,
			updated_at = NOW()
		RETURNING created_at, updated_at`

	err = r.db.QueryRow(ctx, query, userID, existing.EmailNotifications, existing.PushNotifications, existing.Theme, existing.Language).
		Scan(&existing.CreatedAt, &existing.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return existing, nil
}
