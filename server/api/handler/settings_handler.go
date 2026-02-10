package handler

import (
	"net/http"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/settings"

	"github.com/labstack/echo/v4"
)

type SettingsHandler struct {
	settingsService settings.SettingsService
}

func NewSettingsHandler(settingsService settings.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
	}
}

// GetSettings godoc
// @Summary Get user settings
// @Description Returns the current user's settings/preferences
// @Tags Settings
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/settings [get]
func (h *SettingsHandler) GetSettings(c echo.Context) error {
	userID, err := helpers.GetKratosID(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	settings, err := h.settingsService.GetSettings(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, "Failed to fetch settings", http.StatusInternalServerError)
	}

	return helpers.Success(c, settings, http.StatusOK)
}

func (h *SettingsHandler) UpdateSettings(c echo.Context) error {
	userID, err := helpers.GetKratosID(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	var req models.SettingsUpdateRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	updatedSettings, err := h.settingsService.UpdateSettings(c.Request().Context(), userID, &req)
	if err != nil {
		return helpers.Error(c, "Failed to update settings", http.StatusInternalServerError)
	}

	return helpers.Success(c, updatedSettings, http.StatusOK)
}
