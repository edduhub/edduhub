package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

func extractBearerToken(c echo.Context) string {
	const bearerPrefix = "Bearer "
	authHeader := c.Request().Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
}

// InitiateRegistration starts the registration flow
func (h *AuthHandler) InitiateRegistration(c echo.Context) error {
	flow, err := h.authService.InitiateRegistrationFlow(c.Request().Context())
	if err != nil {
		return helpers.Error(c, "unable to initiate registration flow", http.StatusInternalServerError)
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
		RollNo      string `json:"rollNo" validate:"required"`
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
	kratosReq.Traits.RollNo = req.RollNo

	// Initiate flow first
	flow, err := h.authService.InitiateRegistrationFlow(c.Request().Context())
	if err != nil {
		return helpers.Error(c, "unable to initiate registration: "+err.Error(), http.StatusInternalServerError)
	}

	flowID, ok := flow["id"].(string)
	if !ok {
		return helpers.Error(c, "invalid flow response", http.StatusInternalServerError)
	}

	// Complete registration and generate token
	token, identity, err := h.authService.CompleteRegistration(c.Request().Context(), flowID, kratosReq)
	if err != nil {
		return helpers.Error(c, "unable to complete registration: "+err.Error(), http.StatusBadRequest)
	}

	userIDValue := any(identity.ID)
	if identity.UserID > 0 {
		userIDValue = identity.UserID
	}

	return helpers.Success(c, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":          userIDValue,
			"kratosId":    identity.ID,
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

	// Calculate expiration time (24 hours from now as per JWT manager)
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	userIDValue := any(identity.ID)
	if identity.UserID > 0 {
		userIDValue = identity.UserID
	}

	// Return token and user info
	return helpers.Success(c, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":          userIDValue,
			"kratosId":    identity.ID,
			"email":       identity.Traits.Email,
			"firstName":   identity.Traits.Name.First,
			"lastName":    identity.Traits.Name.Last,
			"role":        identity.Traits.Role,
			"collegeId":   identity.Traits.College.ID,
			"collegeName": identity.Traits.College.Name,
		},
		"expiresAt": expiresAt,
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

	// JWT mode: no server-side session invalidation is required.
	jwtToken := extractBearerToken(c)
	if jwtToken == "" {
		return helpers.Error(c, "no session token or bearer token provided", http.StatusBadRequest)
	}

	_, err := h.authService.ValidateJWT(c.Request().Context(), jwtToken)
	if err != nil {
		return helpers.Error(c, "failed to validate token: "+err.Error(), http.StatusUnauthorized)
	}

	return helpers.Success(c, map[string]string{"message": "logout successful"}, http.StatusOK)
}

// RefreshToken refreshes the JWT Token
func (h *AuthHandler) RefreshToken(c echo.Context) error {

	jwtToken := extractBearerToken(c)
	if jwtToken == "" {
		return helpers.Error(c, "no session token or bearer token provided", http.StatusBadRequest)
	}

	newToken, err := h.authService.RefreshToken(c.Request().Context(), jwtToken)
	if err != nil {
		return helpers.Error(c, "failed to refresh token: "+err.Error(), http.StatusUnauthorized)
	}

	return helpers.Success(c, map[string]string{
		"token":   newToken,
		"message": "token refreshed successfully",
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
