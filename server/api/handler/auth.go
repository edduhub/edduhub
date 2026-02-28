package handler

import (
	"net/http"
	"strings"

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

	// Complete registration
	_, identity, err := h.authService.CompleteRegistration(c.Request().Context(), flowID, kratosReq)
	if err != nil {
		return helpers.Error(c, "unable to complete registration: "+err.Error(), http.StatusBadRequest)
	}

	userIDValue := any(identity.ID)
	if identity.UserID > 0 {
		userIDValue = identity.UserID
	}

	// After registration, user needs to go through OAuth2 flow to get tokens
	// Return user info - frontend should redirect to login to get OAuth2 tokens
	return helpers.Success(c, map[string]any{
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
		"message": "registration successful. please login to get access token",
	}, http.StatusCreated)
}

// InitiateLogin starts the OAuth2 authorization code flow
// Returns the Hydra authorization URL to redirect the user to
func (h *AuthHandler) InitiateLogin(c echo.Context) error {
	var req struct {
		RedirectURI string `json:"redirectUri"`
	}

	if err := c.Bind(&req); err != nil {
		// Use default redirect URI if not provided
		req.RedirectURI = ""
	}

	// Get the OAuth2 authorization URL from Hydra
	authURL, state, err := h.authService.InitiateLogin(c.Request().Context(), req.RedirectURI)
	if err != nil {
		return helpers.Error(c, "failed to initiate login: "+err.Error(), http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]string{
		"authorizationUrl": authURL,
		"state":            state,
	}, http.StatusOK)
}

// HandleLogin processes login callback from Hydra
// This handles the OAuth2 authorization code flow callback
func (h *AuthHandler) HandleLogin(c echo.Context) error {
	var req struct {
		Code        string `json:"code"`
		RedirectURI string `json:"redirectUri"`
		State       string `json:"state"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", http.StatusBadRequest)
	}

	if req.Code == "" {
		return helpers.Error(c, "authorization code is required", http.StatusBadRequest)
	}

	// Complete the OAuth2 flow - exchange code for tokens
	oauthToken, identity, err := h.authService.CompleteLogin(c.Request().Context(), req.Code, req.RedirectURI, req.State)
	if err != nil {
		return helpers.Error(c, "authentication failed: "+err.Error(), http.StatusUnauthorized)
	}

	userIDValue := any(identity.ID)
	if identity.UserID > 0 {
		userIDValue = identity.UserID
	}

	// Return OAuth2 tokens and user info
	return helpers.Success(c, map[string]any{
		"accessToken":  oauthToken.AccessToken,
		"refreshToken": oauthToken.RefreshToken,
		"tokenType":    oauthToken.TokenType,
		"expiresIn":    oauthToken.ExpiresIn,
		"idToken":      oauthToken.IDToken,
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
	}, http.StatusOK)
}

// HandleCallback processes the login callback (alias for HandleLogin for compatibility)
func (h *AuthHandler) HandleCallback(c echo.Context) error {
	// Extract identity from the context, which should be set by ValidateToken middleware
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

// HandleLogout logs out the current user by revoking the OAuth2 token
func (h *AuthHandler) HandleLogout(c echo.Context) error {
	accessToken := extractBearerToken(c)
	if accessToken == "" {
		return helpers.Error(c, "no access token provided", http.StatusBadRequest)
	}

	// Revoke the OAuth2 token via Hydra
	err := h.authService.RevokeAccessToken(c.Request().Context(), accessToken)
	if err != nil {
		return helpers.Error(c, "failed to revoke token: "+err.Error(), http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]string{"message": "logout successful"}, http.StatusOK)
}

// RefreshToken refreshes the OAuth2 access token using refresh token
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", http.StatusBadRequest)
	}

	if req.RefreshToken == "" {
		return helpers.Error(c, "refresh token is required", http.StatusBadRequest)
	}

	// Refresh the OAuth2 token
	oauthToken, err := h.authService.RefreshAccessToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return helpers.Error(c, "failed to refresh token: "+err.Error(), http.StatusUnauthorized)
	}

	return helpers.Success(c, map[string]any{
		"accessToken":  oauthToken.AccessToken,
		"refreshToken": oauthToken.RefreshToken,
		"tokenType":    oauthToken.TokenType,
		"expiresIn":    oauthToken.ExpiresIn,
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
