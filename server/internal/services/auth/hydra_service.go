package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	ory "github.com/ory/hydra-client-go"
)

const (
	HydraPublicURL    = "HYDRA_PUBLIC_URL"
	HydraAdminURL     = "HYDRA_ADMIN_URL"
	HydraClientID     = "HYDRA_CLIENT_ID"
	HydraClientSecret = "HYDRA_CLIENT_SECRET"
)

type HydraService interface {
	// OAuth2 Flow
	InitiateLogin(ctx context.Context, clientID, redirectURI, state, nonce string) (string, error)
	ExchangeCode(ctx context.Context, code, redirectURI string) (*OAuth2Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*OAuth2Token, error)
	RevokeToken(ctx context.Context, token string) error

	// Token Validation
	IntrospectToken(ctx context.Context, accessToken string) (*IntrospectedToken, error)
	ValidateAccessToken(ctx context.Context, accessToken string) (*IntrospectedToken, error)

	// User Info
	GetUserInfo(ctx context.Context, accessToken string) (map[string]any, error)

	// URL Helpers
	GetAuthorizationURL(clientID, redirectURI, state, nonce, scope string) string
}

type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type IntrospectedToken struct {
	Active    bool
	Sub       string
	Aud       []string
	ClientID  string
	Exp       int64
	Iat       int64
	Iss       string
	Scope     string
	TokenType string
	TokenUse  string
	Ext       map[string]any
}

type hydraService struct {
	PublicURL    string
	AdminURL     string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
	Scopes       []string
}

func NewHydraService() *hydraService {
	return &hydraService{
		PublicURL:    os.Getenv(HydraPublicURL),
		AdminURL:     os.Getenv(HydraAdminURL),
		ClientID:     os.Getenv(HydraClientID),
		ClientSecret: os.Getenv(HydraClientSecret),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Scopes: []string{"openid", "profile", "email", "offline_access"},
	}
}

// GetAuthorizationURL builds the OAuth2 authorization URL
func (h *hydraService) GetAuthorizationURL(clientID, redirectURI, state, nonce, scope string) string {
	if scope == "" {
		scope = strings.Join(h.Scopes, " ")
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", scope)
	params.Set("state", state)
	if nonce != "" {
		params.Set("nonce", nonce)
	}

	return fmt.Sprintf("%s/oauth2/auth?%s", h.PublicURL, params.Encode())
}

// InitiateLogin initiates the OAuth2 authorization code flow
// Returns the authorization URL to redirect the user to
func (h *hydraService) InitiateLogin(ctx context.Context, clientID, redirectURI, state, nonce string) (string, error) {
	// Use provided client_id or fall back to configured client
	if clientID == "" {
		clientID = h.ClientID
	}

	// Validate required parameters
	if redirectURI == "" {
		return "", fmt.Errorf("redirect_uri is required")
	}

	authURL := h.GetAuthorizationURL(clientID, redirectURI, state, nonce, "")
	return authURL, nil
}

// ExchangeCode exchanges an authorization code for access and refresh tokens
func (h *hydraService) ExchangeCode(ctx context.Context, code, redirectURI string) (*OAuth2Token, error) {
	if code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}

	tokenURL := fmt.Sprintf("%s/oauth2/token", h.PublicURL)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", h.ClientID)
	data.Set("client_secret", h.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("token exchange failed: %s - %s", errResp.Error, errResp.ErrorDescription)
	}

	var token OAuth2Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &token, nil
}

// RefreshToken exchanges a refresh token for new access and refresh tokens
func (h *hydraService) RefreshToken(ctx context.Context, refreshToken string) (*OAuth2Token, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	tokenURL := fmt.Sprintf("%s/oauth2/token", h.PublicURL)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", h.ClientID)
	data.Set("client_secret", h.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("token refresh failed: %s - %s", errResp.Error, errResp.ErrorDescription)
	}

	var token OAuth2Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	return &token, nil
}

// RevokeToken revokes an access or refresh token
func (h *hydraService) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	revokeURL := fmt.Sprintf("%s/oauth2/revoke", h.AdminURL)

	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", h.ClientID)
	data.Set("client_secret", h.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	// 200 OK is success, 401/404 are also considered success (token already invalid)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("token revocation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// IntrospectToken introspects an OAuth2 token to validate it
func (h *hydraService) IntrospectToken(ctx context.Context, accessToken string) (*IntrospectedToken, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	introspectURL := fmt.Sprintf("%s/oauth2/token/introspect", h.AdminURL)

	data := url.Values{}
	data.Set("token", accessToken)
	data.Set("client_id", h.ClientID)
	data.Set("client_secret", h.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create introspect request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to introspect token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token introspection failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Active    bool           `json:"active"`
		Sub       string         `json:"sub"`
		Aud       []string       `json:"aud"`
		ClientID  string         `json:"client_id"`
		Exp       int64          `json:"exp"`
		Iat       int64          `json:"iat"`
		Iss       string         `json:"iss"`
		Scope     string         `json:"scope"`
		TokenType string         `json:"token_type"`
		TokenUse  string         `json:"token_use"`
		Ext       map[string]any `json:"ext"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode introspect response: %w", err)
	}

	return &IntrospectedToken{
		Active:    result.Active,
		Sub:       result.Sub,
		Aud:       result.Aud,
		ClientID:  result.ClientID,
		Exp:       result.Exp,
		Iat:       result.Iat,
		Iss:       result.Iss,
		Scope:     result.Scope,
		TokenType: result.TokenType,
		TokenUse:  result.TokenUse,
		Ext:       result.Ext,
	}, nil
}

// ValidateAccessToken validates an access token and returns the introspected token
// This is a convenience method that checks if the token is active
func (h *hydraService) ValidateAccessToken(ctx context.Context, accessToken string) (*IntrospectedToken, error) {
	introspected, err := h.IntrospectToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	if !introspected.Active {
		return nil, fmt.Errorf("token is inactive or expired")
	}

	// Check token type - must be access_token
	if introspected.TokenUse != "access_token" {
		return nil, fmt.Errorf("invalid token type: expected access_token, got %s", introspected.TokenUse)
	}

	return introspected, nil
}

// GetUserInfo retrieves user information from the UserInfo endpoint
func (h *hydraService) GetUserInfo(ctx context.Context, accessToken string) (map[string]any, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	userInfoURL := fmt.Sprintf("%s/userinfo", h.PublicURL)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var userInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return userInfo, nil
}

// GetClient returns the Ory Hydra client for advanced operations
func (h *hydraService) GetClient() *ory.APIClient {
	config := ory.NewConfiguration()
	config.Servers = ory.ServerConfigurations{
		{URL: h.PublicURL},
		{URL: h.AdminURL},
	}
	return ory.NewAPIClient(config)
}
