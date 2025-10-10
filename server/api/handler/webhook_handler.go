package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/webhook"

	"github.com/labstack/echo/v4"
)

type WebhookHandler struct {
	webhookService webhook.WebhookService
}

func NewWebhookHandler(webhookService webhook.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// CreateWebhook creates a new webhook subscription
func (h *WebhookHandler) CreateWebhook(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.Webhook
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	req.CollegeID = collegeID

	err = h.webhookService.CreateWebhook(c.Request().Context(), &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, req, 201)
}

// ListWebhooks retrieves all webhooks for the college
func (h *WebhookHandler) ListWebhooks(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	webhooks, err := h.webhookService.ListWebhooks(c.Request().Context(), collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, webhooks, 200)
}

// GetWebhook retrieves a specific webhook
func (h *WebhookHandler) GetWebhook(c echo.Context) error {
	webhookIDStr := c.Param("webhookID")
	webhookID, err := strconv.Atoi(webhookIDStr)
	if err != nil {
		return helpers.Error(c, "invalid webhook ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	webhook, err := h.webhookService.GetWebhook(c.Request().Context(), collegeID, webhookID)
	if err != nil {
		return helpers.Error(c, "webhook not found", 404)
	}

	return helpers.Success(c, webhook, 200)
}

// UpdateWebhook updates a webhook
func (h *WebhookHandler) UpdateWebhook(c echo.Context) error {
	webhookIDStr := c.Param("webhookID")
	webhookID, err := strconv.Atoi(webhookIDStr)
	if err != nil {
		return helpers.Error(c, "invalid webhook ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.Webhook
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	req.ID = webhookID
	req.CollegeID = collegeID

	err = h.webhookService.UpdateWebhook(c.Request().Context(), &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Webhook updated successfully", 200)
}

// DeleteWebhook deletes a webhook
func (h *WebhookHandler) DeleteWebhook(c echo.Context) error {
	webhookIDStr := c.Param("webhookID")
	webhookID, err := strconv.Atoi(webhookIDStr)
	if err != nil {
		return helpers.Error(c, "invalid webhook ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.webhookService.DeleteWebhook(c.Request().Context(), collegeID, webhookID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Webhook deleted successfully", 200)
}

// TestWebhook sends a test event to the webhook
func (h *WebhookHandler) TestWebhook(c echo.Context) error {
	webhookIDStr := c.Param("webhookID")
	webhookID, err := strconv.Atoi(webhookIDStr)
	if err != nil {
		return helpers.Error(c, "invalid webhook ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.webhookService.TestWebhook(c.Request().Context(), collegeID, webhookID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Test event sent successfully", 200)
}
