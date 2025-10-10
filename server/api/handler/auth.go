package handler

import (
	"fmt"
	"net/http"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/auth"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService auth.AuthService
}

func NewAuthHandler(authService auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// InitiateRegistration starts the registration flow
func (h *AuthHandler) InitiateRegistration(c echo.Context) error {
	flow, err := h.authService.InitiateRegistrationFlow(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, flow)
}

// HandleRegistration processes the registration
func (h *AuthHandler) HandleRegistration(c echo.Context) error {
	flowID := c.QueryParam("flow")
	if flowID == "" {
		return helpers.Error(c, "empty flowID", 400)
	}

	var req auth.RegistrationRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "Invalid Registration Request", 400)
	}

	identity, err := h.authService.CompleteRegistration(c.Request().Context(), flowID, req)
	if err != nil {
		helpers.Error(c, "unable to complete registration", http.StatusNotFound)
	}

	return helpers.Success(c, identity, http.StatusOK)
}

// HandleLogin processes login
func (h *AuthHandler) HandleLogin(c echo.Context) error {
	// Will be redirected to Kratos UI
	loginURL := fmt.Sprintf("%s/self-service/login/browser", h.authService.GetPublicURL())
	return c.Redirect(http.StatusTemporaryRedirect, loginURL)
}

// HandleCallback processes the login callback
func (h *AuthHandler) HandleCallback(c echo.Context) error {
	// Extract identity from the context, which should be set by JWT middleware
	identityRaw := c.Get("identity")
	if identityRaw == nil {
		return helpers.Error(c, "authorization required", http.StatusUnauthorized)
	}

	identity, ok := identityRaw.(*auth.Identity)
	if !ok {
		return helpers.Error(c, "invalid identity format", http.StatusInternalServerError)
	}

	return helpers.Success(c, identity, http.StatusOK)
}

// HandleLogout logs out the current user
func (h *AuthHandler) HandleLogout(c echo.Context) error {
	sessionToken := c.Request().Header.Get("X-Session-Token")
	if sessionToken == "" {
		return helpers.Error(c, "no session token provided", http.StatusBadRequest)
	}

	err := h.authService.Logout(c.Request().Context(), sessionToken)
	if err != nil {
		return helpers.Error(c, "failed to logout: "+err.Error(), http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]string{"message": "logout successful"}, http.StatusOK)
}

// RefreshToken refreshes the session token
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	sessionToken := c.Request().Header.Get("X-Session-Token")
	if sessionToken == "" {
		return helpers.Error(c, "no session token provided", http.StatusBadRequest)
	}

	newToken, err := h.authService.RefreshSession(c.Request().Context(), sessionToken)
	if err != nil {
		return helpers.Error(c, "failed to refresh token: "+err.Error(), http.StatusUnauthorized)
	}

	return helpers.Success(c, map[string]string{
		"session_token": newToken,
		"message":       "token refreshed successfully",
	}, http.StatusOK)
}

// RequestPasswordReset initiates password reset flow
func (h *AuthHandler) RequestPasswordReset(c echo.Context) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", http.StatusBadRequest)
	}

	err := h.authService.InitiatePasswordReset(c.Request().Context(), req.Email)
	if err != nil {
		return helpers.Error(c, "failed to initiate password reset: "+err.Error(), http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]string{"message": "password reset email sent"}, http.StatusOK)
}

// CompletePasswordReset completes the password reset process
func (h *AuthHandler) CompletePasswordReset(c echo.Context) error {
	flowID := c.QueryParam("flow")
	if flowID == "" {
		return helpers.Error(c, "flow ID required", http.StatusBadRequest)
	}

	var req struct {
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", http.StatusBadRequest)
	}

	err := h.authService.CompletePasswordReset(c.Request().Context(), flowID, req.Password)
	if err != nil {
		return helpers.Error(c, "failed to reset password: "+err.Error(), http.StatusBadRequest)
	}

	return helpers.Success(c, map[string]string{"message": "password reset successful"}, http.StatusOK)
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	flowID := c.QueryParam("flow")
	token := c.QueryParam("token")

	if flowID == "" || token == "" {
		return helpers.Error(c, "flow and token required", http.StatusBadRequest)
	}

	err := h.authService.VerifyEmail(c.Request().Context(), flowID, token)
	if err != nil {
		return helpers.Error(c, "email verification failed: "+err.Error(), http.StatusBadRequest)
	}

	return helpers.Success(c, map[string]string{"message": "email verified successfully"}, http.StatusOK)
}

// InitiateEmailVerification starts email verification flow
func (h *AuthHandler) InitiateEmailVerification(c echo.Context) error {
	identityRaw := c.Get("identity")
	if identityRaw == nil {
		return helpers.Error(c, "authorization required", http.StatusUnauthorized)
	}

	identity, ok := identityRaw.(*auth.Identity)
	if !ok {
		return helpers.Error(c, "invalid identity format", http.StatusInternalServerError)
	}

	flowData, err := h.authService.InitiateEmailVerification(c.Request().Context(), identity.ID)
	if err != nil {
		return helpers.Error(c, "failed to initiate email verification: "+err.Error(), http.StatusInternalServerError)
	}

	return helpers.Success(c, flowData, http.StatusOK)
}

// ChangePassword allows logged-in users to change their password
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	identityRaw := c.Get("identity")
	if identityRaw == nil {
		return helpers.Error(c, "authorization required", http.StatusUnauthorized)
	}

	identity, ok := identityRaw.(*auth.Identity)
	if !ok {
		return helpers.Error(c, "invalid identity format", http.StatusInternalServerError)
	}

	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", http.StatusBadRequest)
	}

	err := h.authService.ChangePassword(c.Request().Context(), identity.ID, req.OldPassword, req.NewPassword)
	if err != nil {
		return helpers.Error(c, "failed to change password: "+err.Error(), http.StatusBadRequest)
	}

	return helpers.Success(c, map[string]string{"message": "password changed successfully"}, http.StatusOK)
}
