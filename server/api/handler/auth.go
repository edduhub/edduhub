package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/auth"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService auth.AuthService
}

const (
	sessionTTL           = 24 * time.Hour
	sessionCookieTTL     = 24 * time.Hour
	refreshTokenTTL      = 30 * 24 * time.Hour
	authAccessTokenName  = "edduhub_access_token"
	authRefreshTokenName = "edduhub_refresh_token"
	authSessionTokenName = "edduhub_session_token"
)

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

func readTokenFromCookie(c echo.Context, name string) string {
	cookie, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func setAuthCookie(c echo.Context, name string, token string, maxAge int) {
	req := c.Request()
	secure := req.TLS != nil || strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https")

	cookie := &http.Cookie{
		Name:     name,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}

	if strings.TrimSpace(token) == "" || maxAge <= 0 {
		cookie.MaxAge = -1
		cookie.Value = ""
		cookie.Expires = time.Unix(0, 0).UTC()
	} else {
		cookie.Value = token
		cookie.MaxAge = maxAge
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second).UTC()
	}

	c.SetCookie(cookie)
}

func setAuthCookies(c echo.Context, accessToken, refreshToken, sessionToken string, accessTTL int) {
	sessionTTLSeconds := int(sessionTTL / time.Second)
	refreshTTLSeconds := int(refreshTokenTTL / time.Second)

	setAuthCookie(c, authAccessTokenName, accessToken, accessTTL)
	setAuthCookie(c, authRefreshTokenName, refreshToken, refreshTTLSeconds)
	setAuthCookie(c, authSessionTokenName, sessionToken, int(sessionCookieTTL/time.Second))
}

func clearAuthCookies(c echo.Context) {
	setAuthCookie(c, authAccessTokenName, "", -1)
	setAuthCookie(c, authRefreshTokenName, "", -1)
	setAuthCookie(c, authSessionTokenName, "", -1)
}

func identityUserPayload(identity *auth.Identity) map[string]any {
	if identity == nil {
		return map[string]any{}
	}

	userIDValue := any(identity.ID)
	if identity.UserID > 0 {
		userIDValue = identity.UserID
	}

	return map[string]any{
		"id":          userIDValue,
		"kratosId":    identity.ID,
		"email":       identity.Traits.Email,
		"firstName":   identity.Traits.Name.First,
		"lastName":    identity.Traits.Name.Last,
		"role":        identity.Traits.Role,
		"collegeId":   identity.Traits.College.ID,
		"collegeName": identity.Traits.College.Name,
		"verified":    true,
	}
}

func sessionPayload(identity *auth.Identity, token string) map[string]any {
	payload := map[string]any{
		"user":      identityUserPayload(identity),
		"expiresAt": time.Now().Add(sessionTTL).UTC().Format(time.RFC3339),
	}

	if token != "" {
		payload["token"] = token
	}

	return payload
}

func getSessionTokenForResponse(c echo.Context) string {
	if token := extractBearerToken(c); token != "" {
		return token
	}
	if token := readTokenFromCookie(c, authAccessTokenName); token != "" {
		return token
	}
	return readTokenFromCookie(c, authSessionTokenName)
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
	identity, err := h.authService.CompleteRegistration(c.Request().Context(), flowID, kratosReq)
	if err != nil {
		return helpers.Error(c, "unable to complete registration: "+err.Error(), http.StatusBadRequest)
	}

	// After registration, user may use OAuth2 or direct login.
	// Return user info and keep the flow API-compatible.
	return helpers.Success(c, map[string]any{
		"user":    identityUserPayload(identity),
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

// HandleLogin processes login callback from Hydra.
// This handles the OAuth2 authorization code flow callback.
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

	// Complete the OAuth2 flow - exchange code for tokens.
	oauthToken, identity, err := h.authService.CompleteLogin(c.Request().Context(), req.Code, req.RedirectURI, req.State)
	if err != nil {
		return helpers.Error(c, "authentication failed: "+err.Error(), http.StatusUnauthorized)
	}
	if oauthToken.ExpiresIn <= 0 {
		oauthToken.ExpiresIn = int(sessionTTL.Seconds())
	}

	setAuthCookies(c, oauthToken.AccessToken, oauthToken.RefreshToken, "", oauthToken.ExpiresIn)

	// Return OAuth2 tokens and user info
	return helpers.Success(c, map[string]any{
		"token":        oauthToken.AccessToken,
		"accessToken":  oauthToken.AccessToken,
		"refreshToken": oauthToken.RefreshToken,
		"tokenType":    oauthToken.TokenType,
		"expiresIn":    oauthToken.ExpiresIn,
		"idToken":      oauthToken.IDToken,
		"user":         identityUserPayload(identity),
	}, http.StatusOK)
}

// DirectLogin handles direct email+password login and stores a Kratos session cookie.
func (h *AuthHandler) DirectLogin(c echo.Context) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", http.StatusBadRequest)
	}

	if req.Email == "" || req.Password == "" {
		return helpers.Error(c, "email and password are required", http.StatusBadRequest)
	}

	identity, sessionToken, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return helpers.Error(c, "invalid email or password", http.StatusUnauthorized)
	}
	setAuthCookies(c, "", "", sessionToken, int(sessionCookieTTL.Seconds()))

	return helpers.Success(c, sessionPayload(identity, sessionToken), http.StatusOK)
}

// HandleSession returns the current authenticated session identity.
func (h *AuthHandler) HandleSession(c echo.Context) error {
	// Extract identity from the context, which should be set by ValidateToken middleware
	identityRaw := c.Get("identity")
	if identityRaw == nil {
		return helpers.Error(c, "authorization required", http.StatusUnauthorized)
	}

	identity, ok := identityRaw.(*auth.Identity)
	if !ok {
		return helpers.Error(c, "invalid identity format", http.StatusInternalServerError)
	}

	return helpers.Success(c, sessionPayload(identity, getSessionTokenForResponse(c)), http.StatusOK)
}

// HandleCallback processes the login callback (alias for HandleSession for compatibility).
func (h *AuthHandler) HandleCallback(c echo.Context) error {
	return h.HandleSession(c)
}

// HandleLogout revokes active auth artifacts and clears auth cookies.
func (h *AuthHandler) HandleLogout(c echo.Context) error {
	accessToken := extractBearerToken(c)
	sessionToken := readTokenFromCookie(c, authSessionTokenName)
	if accessToken == "" && sessionToken == "" {
		accessToken = readTokenFromCookie(c, authAccessTokenName)
	}

	if accessToken == "" && sessionToken == "" {
		return helpers.Error(c, "no access token provided", http.StatusBadRequest)
	}

	if accessToken != "" {
		if err := h.authService.RevokeAccessToken(c.Request().Context(), accessToken); err != nil {
			// Log the error but don't fail request, as session-only flows may not have revocable access tokens.
			c.Logger().Warnf("Failed to revoke access token via Hydra: %v", err)
		}
	}
	if sessionToken != "" {
		if err := h.authService.Logout(c.Request().Context(), sessionToken); err != nil {
			c.Logger().Warnf("Failed to invalidate Kratos session: %v", err)
		}
	}

	clearAuthCookies(c)

	return helpers.Success(c, map[string]string{"message": "logout successful"}, http.StatusOK)
}

// RefreshToken refreshes the OAuth2 access token using refresh token
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if c.Request().ContentLength > 0 {
		if err := c.Bind(&req); err != nil && !errors.Is(err, io.EOF) {
			return helpers.Error(c, "invalid request body", http.StatusBadRequest)
		}
	}

	if req.RefreshToken == "" {
		req.RefreshToken = readTokenFromCookie(c, authRefreshTokenName)
	}
	if req.RefreshToken == "" {
		return helpers.Error(c, "refresh token is required", http.StatusBadRequest)
	}

	// Refresh the OAuth2 token
	oauthToken, err := h.authService.RefreshAccessToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return helpers.Error(c, "failed to refresh token: "+err.Error(), http.StatusUnauthorized)
	}
	if oauthToken.ExpiresIn <= 0 {
		oauthToken.ExpiresIn = int(sessionTTL.Seconds())
	}
	setAuthCookies(c, oauthToken.AccessToken, oauthToken.RefreshToken, "", oauthToken.ExpiresIn)

	return helpers.Success(c, map[string]any{
		"token":        oauthToken.AccessToken,
		"accessToken":  oauthToken.AccessToken,
		"refreshToken": oauthToken.RefreshToken,
		"tokenType":    oauthToken.TokenType,
		"expiresIn":    oauthToken.ExpiresIn,
		"expiresAt":    time.Now().Add(sessionTTL).UTC().Format(time.RFC3339),
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
