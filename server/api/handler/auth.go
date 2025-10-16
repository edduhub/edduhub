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
	var req struct {
		Email       string `json:"email" validate:"required,email"`
		Password    string `json:"password" validate:"required,min=8"`
		FirstName   string `json:"firstName" validate:"required"`
		LastName    string `json:"lastName" validate:"required"`
		Role        string `json:"role" validate:"required"`
		CollegeId   string `json:"collegeId" validate:"required"`
		CollegeName string `json:"collegeName" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "Invalid Registration Request", http.StatusBadRequest)
	}

	// Create Kratos registration request
	var kratosReq auth.RegistrationRequest
	kratosReq.Method = "password"
	kratosReq.Password = req.Password
	kratosReq.Traits.Email = req.Email
	kratosReq.Traits.Name.First = req.FirstName
	kratosReq.Traits.Name.Last = req.LastName
	kratosReq.Traits.Role = req.Role
	kratosReq.Traits.College.ID = req.CollegeId
	kratosReq.Traits.College.Name = req.CollegeName

	// Initiate flow first
	flow, err := h.authService.InitiateRegistrationFlow(c.Request().Context())
	if err != nil {
		return helpers.Error(c, "unable to initiate registration: "+err.Error(), http.StatusInternalServerError)
	}

	flowID, ok := flow["id"].(string)
	if !ok {
		return helpers.Error(c, "invalid flow response", http.StatusInternalServerError)
	}

	// Complete registration
	identity, err := h.authService.CompleteRegistration(c.Request().Context(), flowID, kratosReq)
	if err != nil {
		return helpers.Error(c, "unable to complete registration: "+err.Error(), http.StatusBadRequest)
	}

	// Login the user automatically after registration
	token, _, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		// Registration succeeded but auto-login failed
		return helpers.Success(c, map[string]interface{}{
			"message":  "Registration successful. Please login.",
			"identity": identity,
		}, http.StatusCreated)
	}

	return helpers.Success(c, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":          identity.ID,
			"email":       identity.Traits.Email,
			"firstName":   identity.Traits.Name.First,
			"lastName":    identity.Traits.Name.Last,
			"role":        identity.Traits.Role,
			"collegeId":   identity.Traits.College.ID,
			"collegeName": identity.Traits.College.Name,
		},
		"expiresAt": fmt.Sprintf("%v", c.Request().Context().Value("token_expiry")),
	}, http.StatusCreated)
}

// HandleLogin processes login
func (h *AuthHandler) HandleLogin(c echo.Context) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", http.StatusBadRequest)
	}

	// Authenticate and generate JWT
	token, identity, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return helpers.Error(c, "authentication failed: "+err.Error(), http.StatusUnauthorized)
	}

	// Return token and user info
	return helpers.Success(c, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":          identity.ID,
			"email":       identity.Traits.Email,
			"firstName":   identity.Traits.Name.First,
			"lastName":    identity.Traits.Name.Last,
			"role":        identity.Traits.Role,
			"collegeId":   identity.Traits.College.ID,
			"collegeName": identity.Traits.College.Name,
		},
		"expiresAt": fmt.Sprintf("%v", c.Request().Context().Value("token_expiry")),
	}, http.StatusOK)
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
