package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type WebhookService interface {
	CreateWebhook(ctx context.Context, webhook *models.Webhook) error
	ListWebhooks(ctx context.Context, collegeID int) ([]*models.Webhook, error)
	GetWebhook(ctx context.Context, collegeID, webhookID int) (*models.Webhook, error)
	UpdateWebhook(ctx context.Context, webhook *models.Webhook) error
	DeleteWebhook(ctx context.Context, collegeID, webhookID int) error
	TriggerEvent(ctx context.Context, collegeID int, event string, payload interface{}) error
	TestWebhook(ctx context.Context, collegeID, webhookID int) error
}

type webhookService struct {
	webhookRepo repository.WebhookRepository
	httpClient  *http.Client
}

func NewWebhookService(webhookRepo repository.WebhookRepository) WebhookService {
	return &webhookService{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *webhookService) CreateWebhook(ctx context.Context, webhook *models.Webhook) error {
	if webhook.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if webhook.Event == "" {
		return fmt.Errorf("webhook event is required")
	}

	webhook.Active = true
	return s.webhookRepo.CreateWebhook(ctx, webhook)
}

func (s *webhookService) ListWebhooks(ctx context.Context, collegeID int) ([]*models.Webhook, error) {
	return s.webhookRepo.GetWebhooksByCollege(ctx, collegeID)
}

func (s *webhookService) GetWebhook(ctx context.Context, collegeID, webhookID int) (*models.Webhook, error) {
	return s.webhookRepo.GetWebhookByID(ctx, collegeID, webhookID)
}

func (s *webhookService) UpdateWebhook(ctx context.Context, webhook *models.Webhook) error {
	return s.webhookRepo.UpdateWebhook(ctx, webhook)
}

func (s *webhookService) DeleteWebhook(ctx context.Context, collegeID, webhookID int) error {
	return s.webhookRepo.DeleteWebhook(ctx, collegeID, webhookID)
}

func (s *webhookService) TriggerEvent(ctx context.Context, collegeID int, event string, payload interface{}) error {
	// Get all active webhooks for this event
	webhooks, err := s.webhookRepo.GetWebhooksByEvent(ctx, collegeID, event)
	if err != nil {
		return err
	}

	// Send webhook to each endpoint
	for _, webhook := range webhooks {
		if !webhook.Active {
			continue
		}

		go s.sendWebhook(webhook, payload)
	}

	return nil
}

func (s *webhookService) TestWebhook(ctx context.Context, collegeID, webhookID int) error {
	webhook, err := s.webhookRepo.GetWebhookByID(ctx, collegeID, webhookID)
	if err != nil {
		return err
	}

	testPayload := map[string]interface{}{
		"event": "test",
		"data": map[string]string{
			"message": "This is a test webhook event",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return s.sendWebhook(webhook, testPayload)
}

func (s *webhookService) sendWebhook(webhook *models.Webhook, payload interface{}) error {
	// Marshal payload
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "EduHub-Webhook/1.0")

	// Add signature if secret is provided
	if webhook.Secret != "" {
		signature := s.generateSignature(data, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook request failed with status %d", resp.StatusCode)
	}

	return nil
}

func (s *webhookService) generateSignature(data []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
