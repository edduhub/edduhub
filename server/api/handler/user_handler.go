package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/user"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService user.UserService
}

func NewUserHandler(userService user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	user, err := h.userService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, "user not found", 404)
	}

	return helpers.Success(c, user, 200)
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.userService.UpdateUserPartial(c.Request().Context(), userID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Profile updated successfully", 200)
}

// ListUsers returns a list of users (admin only)
func (h *UserHandler) ListUsers(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := uint64(20)
	offset := uint64(0)

	if limitStr != "" {
		l, err := strconv.ParseUint(limitStr, 10, 64)
		if err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		o, err := strconv.ParseUint(offsetStr, 10, 64)
		if err == nil {
			offset = o
		}
	}

	users, err := h.userService.ListUsers(c.Request().Context(), limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, users, 200)
}

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(c echo.Context) error {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err := h.userService.CreateUser(c.Request().Context(), &user)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, user, 201)
}

// GetUser retrieves a specific user (admin only)
func (h *UserHandler) GetUser(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return helpers.Error(c, "invalid user ID", 400)
	}

	user, err := h.userService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, "user not found", 404)
	}

	return helpers.Success(c, user, 200)
}

// UpdateUser updates a specific user (admin only)
func (h *UserHandler) UpdateUser(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return helpers.Error(c, "invalid user ID", 400)
	}

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.userService.UpdateUserPartial(c.Request().Context(), userID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "User updated successfully", 200)
}

// DeleteUser deletes a specific user (admin only)
func (h *UserHandler) DeleteUser(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return helpers.Error(c, "invalid user ID", 400)
	}

	err = h.userService.DeleteUser(c.Request().Context(), userID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "User deleted successfully", 200)
}

// UpdateUserRole updates a user's role (admin only)
func (h *UserHandler) UpdateUserRole(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return helpers.Error(c, "invalid user ID", 400)
	}

	var req struct {
		Role string `json:"role" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	updateReq := models.UpdateUserRequest{
		Role: &req.Role,
	}

	err = h.userService.UpdateUserPartial(c.Request().Context(), userID, &updateReq)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "User role updated successfully", 200)
}

// UpdateUserStatus updates a user's active status (admin only)
func (h *UserHandler) UpdateUserStatus(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return helpers.Error(c, "invalid user ID", 400)
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	updateReq := models.UpdateUserRequest{
		IsActive: &req.IsActive,
	}

	err = h.userService.UpdateUserPartial(c.Request().Context(), userID, &updateReq)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "User status updated successfully", 200)
}

// ChangePassword allows user to change their own password
func (h *UserHandler) ChangePassword(c echo.Context) error {
	_, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "user ID required", 401)
	}

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	// Get services from context
	services := c.Get("services")

	// Type assert to get our services
	allServices, ok := services.(*interface{})
	if !ok {
		return helpers.Error(c, "internal server error", 500)
	}

	servicesMap, ok := (*allServices).(map[string]interface{})
	if !ok {
		return helpers.Error(c, "internal server error", 500)
	}

	authService, ok := servicesMap["Auth"].(auth.AuthService)
	if !ok {
		return helpers.Error(c, "internal server error", 500)
	}

	// Get user's identity from context to extract identity ID
	identity, ok := c.Get("identity").(*auth.Identity)
	if !ok {
		return helpers.Error(c, "authentication required", 401)
	}

	// Use the auth service's ChangePassword method
	err = authService.ChangePassword(c.Request().Context(), identity.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		return helpers.Error(c, "Password change failed: "+err.Error(), 400)
	}

	return helpers.Success(c, "Password changed successfully", 200)
}
