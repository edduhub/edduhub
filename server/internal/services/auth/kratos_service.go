package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	PublicURL = "KRATOS_PUBLIC_URL"
	AdminURL  = "KRATOS_ADMIN_URL"
)

const defaultHTTPTimeout = 30 * time.Second

type kratosService struct {
	PublicURL  string
	AdminURL   string
	HTTPClient *http.Client
}

func NewKratosService() *kratosService {
	return &kratosService{
		PublicURL: os.Getenv(PublicURL),
		AdminURL:  os.Getenv(AdminURL),
		HTTPClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

// GetIdentity retrieves an identity by ID from Kratos admin API
func (k *kratosService) GetIdentity(ctx context.Context, identityID string) (*Identity, error) {
	if identityID == "" {
		return nil, fmt.Errorf("identity ID is required")
	}

	url := fmt.Sprintf("%s/identities/%s", k.AdminURL, identityID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity request: %w", err)
	}

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("identity not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get identity: %d", resp.StatusCode)
	}

	var identity Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("failed to decode identity response: %w", err)
	}

	return &identity, nil
}

// InitiateRegistrationFlow starts the registration process
func (k *kratosService) InitiateRegistrationFlow(ctx context.Context) (map[string]any, error) {
	url := fmt.Sprintf("%s/self-service/registration/flows", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute registration request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	return result, nil
}

// CompleteRegistration submits the registration data
func (k *kratosService) CompleteRegistration(ctx context.Context, flowID string, regReq RegistrationRequest) (*Identity, error) {
	data, err := json.Marshal(regReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration data: %w", err)
	}

	url := fmt.Sprintf("%s/self-service/registration/flows?id=%s", k.PublicURL, flowID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration completion request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := k.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to complete registration: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("registration failed: %s - %s", errResp.Error, errResp.ErrorDescription)
	}

	var result struct {
		Identity Identity `json:"identity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	if result.Identity.ID == "" {
		return nil, fmt.Errorf("registration response missing identity")
	}

	return &result.Identity, nil
}

// InitiatePasswordReset starts the password reset flow
func (k *kratosService) InitiatePasswordReset(ctx context.Context, email string) error {
	url := fmt.Sprintf("%s/self-service/recovery/flows", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create password reset request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initiate password reset: %w", err)
	}
	defer resp.Body.Close()

	var flow map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return fmt.Errorf("failed to decode password reset flow: %w", err)
	}

	flowID, ok := flow["id"].(string)
	if !ok {
		return fmt.Errorf("invalid flow ID in response")
	}

	resetData := map[string]any{
		"method": "link",
		"email":  email,
	}
	data, err := json.Marshal(resetData)
	if err != nil {
		return fmt.Errorf("failed to marshal reset data: %w", err)
	}

	submitURL := fmt.Sprintf("%s/self-service/recovery/flows?id=%s", k.PublicURL, flowID)
	submitReq, err := http.NewRequestWithContext(ctx, "POST", submitURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create reset submit request: %w", err)
	}
	submitReq.Header.Set("Content-Type", "application/json")

	submitResp, err := k.HTTPClient.Do(submitReq)
	if err != nil {
		return fmt.Errorf("failed to submit password reset: %w", err)
	}
	defer submitResp.Body.Close()

	if submitResp.StatusCode != http.StatusOK && submitResp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("password reset failed with status: %d", submitResp.StatusCode)
	}

	return nil
}

// CompletePasswordReset completes the password reset process
func (k *kratosService) CompletePasswordReset(ctx context.Context, flowID string, newPassword string) error {
	resetData := map[string]any{
		"method":   "password",
		"password": newPassword,
	}
	data, err := json.Marshal(resetData)
	if err != nil {
		return fmt.Errorf("failed to marshal reset completion data: %w", err)
	}

	url := fmt.Sprintf("%s/self-service/recovery/flows?flow=%s", k.PublicURL, flowID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create reset completion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to complete password reset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("password reset completion failed with status: %d", resp.StatusCode)
	}

	return nil
}

// VerifyEmail completes email verification
func (k *kratosService) VerifyEmail(ctx context.Context, flowID string, token string) error {
	verifyData := map[string]any{
		"method": "link",
		"token":  token,
	}
	data, err := json.Marshal(verifyData)
	if err != nil {
		return fmt.Errorf("failed to marshal verification data: %w", err)
	}

	url := fmt.Sprintf("%s/self-service/verification/flows?flow=%s", k.PublicURL, flowID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create verification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("email verification failed with status: %d", resp.StatusCode)
	}

	return nil
}

// InitiateEmailVerification starts email verification flow
func (k *kratosService) InitiateEmailVerification(ctx context.Context, identityID string) (map[string]any, error) {
	url := fmt.Sprintf("%s/self-service/verification/flows", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create verification request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate verification: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode verification response: %w", err)
	}

	return result, nil
}

// ChangePassword changes the password for a logged-in user
func (k *kratosService) ChangePassword(ctx context.Context, identityID string, oldPassword string, newPassword string) error {
	url := fmt.Sprintf("%s/self-service/settings/flows", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create settings flow request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initiate settings flow: %w", err)
	}
	defer resp.Body.Close()

	var flow map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return fmt.Errorf("failed to decode settings flow: %w", err)
	}

	flowID, ok := flow["id"].(string)
	if !ok {
		return fmt.Errorf("invalid flow ID in response")
	}

	changeData := map[string]any{
		"method":   "password",
		"password": newPassword,
	}
	data, err := json.Marshal(changeData)
	if err != nil {
		return fmt.Errorf("failed to marshal password change data: %w", err)
	}

	submitURL := fmt.Sprintf("%s/self-service/settings/flows?flow=%s", k.PublicURL, flowID)
	submitReq, err := http.NewRequestWithContext(ctx, "POST", submitURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create password change request: %w", err)
	}
	submitReq.Header.Set("Content-Type", "application/json")

	submitResp, err := k.HTTPClient.Do(submitReq)
	if err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}
	defer submitResp.Body.Close()

	if submitResp.StatusCode != http.StatusOK {
		return fmt.Errorf("password change failed with status: %d", submitResp.StatusCode)
	}

	return nil
}

// ValidateSession validates a Kratos session token (legacy - not used with Hydra)
func (k *kratosService) ValidateSession(ctx context.Context, sessionToken string) (*Identity, error) {
	url := fmt.Sprintf("%s/sessions/whoami", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create session validation request: %w", err)
	}
	// Kratos sessions use X-Session-Token header, not Bearer token
	req.Header.Set("X-Session-Token", sessionToken)
	req.Header.Set("Accept", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid session: %d", resp.StatusCode)
	}

	var result struct {
		Identity Identity `json:"identity"`
		Active   bool     `json:"active"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode session response: %w", err)
	}

	if !result.Active {
		return nil, fmt.Errorf("session is not active")
	}

	return &result.Identity, nil
}

// Logout invalidates a Kratos session (legacy - not used with Hydra)
func (k *kratosService) Logout(ctx context.Context, sessionToken string) error {
	url := fmt.Sprintf("%s/sessions", k.AdminURL)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// RefreshSession refreshes a session (legacy - not used with Hydra)
func (k *kratosService) RefreshSession(ctx context.Context, sessionToken string) (string, error) {
	return "", fmt.Errorf("use Hydra for token refresh instead")
}

// GetPublicURL returns the public URL for the Kratos instance
func (k *kratosService) GetPublicURL() string {
	return k.PublicURL
}
