package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/profile"

	"github.com/labstack/echo/v4"
)

type ProfileHandler struct {
	profileService profile.ProfileService
}

func NewProfileHandler(profileService profile.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

// GetUserProfile retrieves the current user's profile
func (h *ProfileHandler) GetUserProfile(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	profileData, err := h.profileService.GetProfileByUserID(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, "profile not found", 404)
	}

	return helpers.Success(c, profileData, 200)
}

// UpdateUserProfile updates the current user's profile
func (h *ProfileHandler) UpdateUserProfile(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	var req models.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.profileService.UpdateProfile(c.Request().Context(), userID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Profile updated successfully", 200)
}

// GetProfile retrieves a specific user's profile (admin only)
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	profileIDStr := c.Param("profileID")
	profileID, err := strconv.Atoi(profileIDStr)
	if err != nil {
		return helpers.Error(c, "invalid profile ID", 400)
	}

	profileData, err := h.profileService.GetProfileByID(c.Request().Context(), profileID)
	if err != nil {
		return helpers.Error(c, "profile not found", 404)
	}

	return helpers.Success(c, profileData, 200)
}
