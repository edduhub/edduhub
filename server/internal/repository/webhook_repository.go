package repository

import (
	"context"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type WebhookRepository interface {
	CreateWebhook(ctx context.Context, webhook *models.Webhook) error
	GetWebhooksByCollege(ctx context.Context, collegeID int) ([]*models.Webhook, error)
	GetWebhooksByEvent(ctx context.Context, collegeID int, event string) ([]*models.Webhook, error)
	GetWebhookByID(ctx context.Context, collegeID, webhookID int) (*models.Webhook, error)
	UpdateWebhook(ctx context.Context, webhook *models.Webhook) error
	DeleteWebhook(ctx context.Context, collegeID, webhookID int) error
}

type webhookRepository struct {
	DB *DB
}

func NewWebhookRepository(db *DB) WebhookRepository {
	return &webhookRepository{DB: db}
}

func (r *webhookRepository) CreateWebhook(ctx context.Context, webhook *models.Webhook) error {
	now := time.Now()
	webhook.CreatedAt = now
	webhook.UpdatedAt = now

	sql := `INSERT INTO webhooks (college_id, name, description, url, secret, event_types, is_active, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	var id int
	err := r.DB.Pool.QueryRow(ctx, sql,
		webhook.CollegeID,
		webhook.Name,
		webhook.Description,
		webhook.URL,
		webhook.Secret,
		webhook.EventTypes,
		webhook.Active,
		webhook.CreatedBy,
		webhook.CreatedAt,
		webhook.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return err
	}

	webhook.ID = id
	return nil
}

func (r *webhookRepository) GetWebhooksByCollege(ctx context.Context, collegeID int) ([]*models.Webhook, error) {
	sql := `SELECT id, college_id, name, description, url, secret, event_types, is_active, created_by, created_at, updated_at
			FROM webhooks WHERE college_id = $1 ORDER BY created_at DESC`

	var webhooks []*models.Webhook
	err := pgxscan.Select(ctx, r.DB.Pool, &webhooks, sql, collegeID)
	return webhooks, err
}

func (r *webhookRepository) GetWebhooksByEvent(ctx context.Context, collegeID int, event string) ([]*models.Webhook, error) {
	sql := `SELECT id, college_id, name, description, url, secret, event_types, is_active, created_by, created_at, updated_at
			FROM webhooks WHERE college_id = $1 AND event_types @> ARRAY[$2]::varchar[] AND is_active = true`

	var webhooks []*models.Webhook
	err := pgxscan.Select(ctx, r.DB.Pool, &webhooks, sql, collegeID, event)
	return webhooks, err
}

func (r *webhookRepository) GetWebhookByID(ctx context.Context, collegeID, webhookID int) (*models.Webhook, error) {
	sql := `SELECT id, college_id, name, description, url, secret, event_types, is_active, created_by, created_at, updated_at
			FROM webhooks WHERE id = $1 AND college_id = $2`

	var webhook models.Webhook
	err := pgxscan.Get(ctx, r.DB.Pool, &webhook, sql, webhookID, collegeID)
	return &webhook, err
}

func (r *webhookRepository) UpdateWebhook(ctx context.Context, webhook *models.Webhook) error {
	webhook.UpdatedAt = time.Now()

	sql := `UPDATE webhooks
			SET name = $1, description = $2, url = $3, secret = $4, event_types = $5, is_active = $6, updated_at = $7
			WHERE id = $8 AND college_id = $9`

	_, err := r.DB.Pool.Exec(ctx, sql,
		webhook.Name,
		webhook.Description,
		webhook.URL,
		webhook.Secret,
		webhook.EventTypes,
		webhook.Active,
		webhook.UpdatedAt,
		webhook.ID,
		webhook.CollegeID,
	)

	return err
}

func (r *webhookRepository) DeleteWebhook(ctx context.Context, collegeID, webhookID int) error {
	sql := `DELETE FROM webhooks WHERE id = $1 AND college_id = $2`
	_, err := r.DB.Pool.Exec(ctx, sql, webhookID, collegeID)
	return err
}
