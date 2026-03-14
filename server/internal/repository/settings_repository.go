package repository

import (
	"context"
	"errors"
	"fmt"

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

func (r *settingsRepository) ensureSchema(ctx context.Context) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS user_settings (
		user_id VARCHAR(255) PRIMARY KEY,
		email_notifications BOOLEAN DEFAULT true,
		push_notifications BOOLEAN DEFAULT true,
		theme VARCHAR(20) DEFAULT 'system',
		language VARCHAR(10) DEFAULT 'en',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := r.db.Exec(ctx, createTableSQL); err != nil {
		return fmt.Errorf("ensure settings table: %w", err)
	}

	alterTableSQL := `ALTER TABLE user_settings
		ADD COLUMN IF NOT EXISTS assignment_reminders BOOLEAN DEFAULT true,
		ADD COLUMN IF NOT EXISTS grade_updates BOOLEAN DEFAULT true,
		ADD COLUMN IF NOT EXISTS announcement_alerts BOOLEAN DEFAULT true,
		ADD COLUMN IF NOT EXISTS timezone VARCHAR(50) DEFAULT 'UTC',
		ADD COLUMN IF NOT EXISTS session_timeout INTEGER DEFAULT 30`
	if _, err := r.db.Exec(ctx, alterTableSQL); err != nil {
		return fmt.Errorf("ensure settings columns: %w", err)
	}

	return nil
}

func (r *settingsRepository) GetSettings(ctx context.Context, userID string) (*models.Settings, error) {
	if err := r.ensureSchema(ctx); err != nil {
		return nil, err
	}

	query := `
		SELECT user_id, email_notifications, push_notifications, assignment_reminders, 
		       grade_updates, announcement_alerts, theme, language, timezone, session_timeout,
		       created_at, updated_at
		FROM user_settings
		WHERE user_id = $1`

	var s models.Settings
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&s.UserID, &s.EmailNotifications, &s.PushNotifications, &s.AssignmentReminders,
		&s.GradeUpdates, &s.AnnouncementAlerts, &s.Theme, &s.Language, &s.Timezone,
		&s.SessionTimeout, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return &models.Settings{
			UserID:              userID,
			EmailNotifications:  true,
			PushNotifications:   true,
			AssignmentReminders: true,
			GradeUpdates:        true,
			AnnouncementAlerts:  true,
			Theme:               "system",
			Language:            "en",
			Timezone:            "UTC",
			SessionTimeout:      30,
		}, nil
	}
	return &s, nil
}

func (r *settingsRepository) UpdateSettings(ctx context.Context, userID string, req *models.SettingsUpdateRequest) (*models.Settings, error) {
	if err := r.ensureSchema(ctx); err != nil {
		return nil, err
	}

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
	if req.AssignmentReminders != nil {
		existing.AssignmentReminders = *req.AssignmentReminders
	}
	if req.GradeUpdates != nil {
		existing.GradeUpdates = *req.GradeUpdates
	}
	if req.AnnouncementAlerts != nil {
		existing.AnnouncementAlerts = *req.AnnouncementAlerts
	}
	if req.Theme != nil {
		existing.Theme = *req.Theme
	}
	if req.Language != nil {
		existing.Language = *req.Language
	}
	if req.Timezone != nil {
		existing.Timezone = *req.Timezone
	}
	if req.SessionTimeout != nil {
		existing.SessionTimeout = *req.SessionTimeout
	}
	query := `
		INSERT INTO user_settings (user_id, email_notifications, push_notifications, 
			assignment_reminders, grade_updates, announcement_alerts, theme, language, 
			timezone, session_timeout, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			email_notifications = EXCLUDED.email_notifications,
			push_notifications = EXCLUDED.push_notifications,
			assignment_reminders = EXCLUDED.assignment_reminders,
			grade_updates = EXCLUDED.grade_updates,
			announcement_alerts = EXCLUDED.announcement_alerts,
			theme = EXCLUDED.theme,
			language = EXCLUDED.language,
			timezone = EXCLUDED.timezone,
			session_timeout = EXCLUDED.session_timeout,
			updated_at = NOW()
		RETURNING created_at, updated_at`

	err = r.db.QueryRow(ctx, query, userID, existing.EmailNotifications, existing.PushNotifications,
		existing.AssignmentReminders, existing.GradeUpdates, existing.AnnouncementAlerts,
		existing.Theme, existing.Language, existing.Timezone, existing.SessionTimeout).
		Scan(&existing.CreatedAt, &existing.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return existing, nil
}
