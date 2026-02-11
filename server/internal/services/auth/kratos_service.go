package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const (
	PUBLIC_URL = "KRATOS_PUBLIC_URL"
	ADMIN_URL  = "KRATOS_ADMIN_URL"
)

type kratosService struct {
	PublicURL  string
	AdminURL   string
	HTTPClient *http.Client
}

type Identity struct {
	ID     string `json:"id"`
	Traits struct {
		Email string `json:"email"`
		Name  struct {
			First string `json:"first"`
			Last  string `json:"last"`
		} `json:"name"`
		College struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"college"`
		Role   string `json:"role"`
		RollNo string `json:"rollNo"`
	} `json:"traits"`
}

type RegistrationRequest struct {
	Method   string `json:"method"`
	Password string `json:"password"`
	Traits   struct {
		Email string `json:"email"`
		Name  struct {
			First string `json:"first"`
			Last  string `json:"last"`
		} `json:"name"`
		College struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"college"`
		Role   string `json:"role"`
		RollNo string `json:"rollNo"`
	} `json:"traits"`
}

func NewKratosService() *kratosService {
	return &kratosService{
		PublicURL:  os.Getenv(PUBLIC_URL),
		AdminURL:   os.Getenv(ADMIN_URL),
		HTTPClient: &http.Client{},
	}
}

// Login authenticates a user with email and password
func (k *kratosService) Login(ctx context.Context, email, password string) (*Identity, error) {
	// Initiate login flow
	url := fmt.Sprintf("%s/self-service/login/api", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create login flow request: %w", err)
	}

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate login flow: %w", err)
	}
	defer resp.Body.Close()

	var flow map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return nil, fmt.Errorf("failed to decode login flow: %w", err)
	}

	flowID, ok := flow["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid flow ID in response")
	}

	// Submit credentials
	loginData := map[string]interface{}{
		"method":     "password",
		"password":   password,
		"identifier": email,
	}
	data, err := json.Marshal(loginData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login data: %w", err)
	}

	submitURL := fmt.Sprintf("%s/self-service/login?flow=%s", k.PublicURL, flowID)
	submitReq, err := http.NewRequestWithContext(ctx, "POST", submitURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create login submit request: %w", err)
	}
	submitReq.Header.Set("Content-Type", "application/json")

	submitResp, err := k.HTTPClient.Do(submitReq)
	if err != nil {
		return nil, fmt.Errorf("failed to submit login: %w", err)
	}
	defer submitResp.Body.Close()

	if submitResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %s", http.StatusText(submitResp.StatusCode))
	}

	var result struct {
		Session struct {
			Identity Identity `json:"identity"`
		} `json:"session"`
	}
	if err := json.NewDecoder(submitResp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}

	return &result.Session.Identity, nil
}

// InitiateRegistrationFlow starts the registration process by calling Ory Kratos.
func (k *kratosService) InitiateRegistrationFlow(ctx context.Context) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/self-service/registration/api", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration request: %w", err)
	}

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute registration request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %w", err)
	}

	return result, nil
}

// CompleteRegistration submits the registration data to complete registration.
func (k *kratosService) CompleteRegistration(ctx context.Context, flowID string, regReq RegistrationRequest) (*Identity, error) {
	data, err := json.Marshal(regReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration data: %w", err)
	}

	url := fmt.Sprintf("%s/self-service/registration?flow=%s", k.PublicURL, flowID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create registration completion request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to complete registration: %w", err)
	}
	defer resp.Body.Close()

	var identity Identity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("failed to decode identity: %w", err)
	}

	return &identity, nil
}

// ValidateSession checks if the session is valid by invoking the Kratos whoami endpoint.
func (k *kratosService) ValidateSession(ctx context.Context, sessionToken string) (*Identity, error) {
	url := fmt.Sprintf("%s/sessions/whoami", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Session-Token", sessionToken)

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid session")
	}

	var result struct {
		Identity Identity `json:"identity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Identity, nil
}

// ValidateJWT validates a JWT token by calling Kratos' JWT introspection endpoint.
func (k *kratosService) ValidateJWT(ctx context.Context, jwtToken string) (*Identity, error) {
	url := fmt.Sprintf("%s/sessions/whoami", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT validation request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate JWT: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid JWT token: %s", http.StatusText(resp.StatusCode))
	}

	var result struct {
		Identity Identity `json:"identity"`
		Active   bool     `json:"active"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode JWT validation response: %w", err)
	}

	if !result.Active {
		return nil, fmt.Errorf("JWT token is not active")
	}

	return &result.Identity, nil
}

// CheckCollegeAccess verifies that the student's college ID matches the provided ID.
func (k *kratosService) CheckCollegeAccess(identity *Identity, collegeID string) bool {
	return identity.Traits.College.ID == collegeID
}

// HasRole verifies if the identity holds the specified role.
func (k *kratosService) HasRole(identity *Identity, role string) bool {
	return identity.Traits.Role == role
}

// GetPublicURL returns the public URL for the Kratos instance.
func (k *kratosService) GetPublicURL() string {
	return k.PublicURL
}

// Logout invalidates the session token
func (k *kratosService) Logout(ctx context.Context, sessionToken string) error {
	endpoint := fmt.Sprintf("%s/self-service/logout/api?session_token=%s", k.PublicURL, url.QueryEscape(sessionToken))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create logout request: %w", err)
	}
	req.Header.Set("X-Session-Token", sessionToken)

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("logout failed with status: %s", http.StatusText(resp.StatusCode))
	}

	return nil
}

// RefreshSession refreshes an existing session and returns a new session token
func (k *kratosService) RefreshSession(ctx context.Context, sessionToken string) (string, error) {
	identity, err := k.ValidateSession(ctx, sessionToken)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid session: %w", err)
	}

	if identity == nil {
		return "", fmt.Errorf("session is not active")
	}

	return sessionToken, nil
}

// InitiatePasswordReset starts the password reset flow
func (k *kratosService) InitiatePasswordReset(ctx context.Context, email string) error {
	url := fmt.Sprintf("%s/self-service/recovery/api", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create password reset request: %w", err)
	}

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initiate password reset: %w", err)
	}
	defer resp.Body.Close()

	var flow map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return fmt.Errorf("failed to decode password reset flow: %w", err)
	}

	flowID, ok := flow["id"].(string)
	if !ok {
		return fmt.Errorf("invalid flow ID in response")
	}

	resetData := map[string]interface{}{
		"method": "link",
		"email":  email,
	}
	data, err := json.Marshal(resetData)
	if err != nil {
		return fmt.Errorf("failed to marshal reset data: %w", err)
	}

	submitURL := fmt.Sprintf("%s/self-service/recovery?flow=%s", k.PublicURL, flowID)
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
		return fmt.Errorf("password reset failed with status: %s", http.StatusText(submitResp.StatusCode))
	}

	return nil
}

// CompletePasswordReset completes the password reset process
func (k *kratosService) CompletePasswordReset(ctx context.Context, flowID string, newPassword string) error {
	resetData := map[string]interface{}{
		"method":   "password",
		"password": newPassword,
	}
	data, err := json.Marshal(resetData)
	if err != nil {
		return fmt.Errorf("failed to marshal reset completion data: %w", err)
	}

	url := fmt.Sprintf("%s/self-service/settings?flow=%s", k.PublicURL, flowID)
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
		return fmt.Errorf("password reset completion failed with status: %s", http.StatusText(resp.StatusCode))
	}

	return nil
}

// VerifyEmail completes email verification
func (k *kratosService) VerifyEmail(ctx context.Context, flowID string, token string) error {
	verifyData := map[string]interface{}{
		"method": "link",
		"token":  token,
	}
	data, err := json.Marshal(verifyData)
	if err != nil {
		return fmt.Errorf("failed to marshal verification data: %w", err)
	}

	url := fmt.Sprintf("%s/self-service/verification?flow=%s", k.PublicURL, flowID)
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
		return fmt.Errorf("email verification failed with status: %s", http.StatusText(resp.StatusCode))
	}

	return nil
}

// InitiateEmailVerification starts email verification flow
func (k *kratosService) InitiateEmailVerification(ctx context.Context, identityID string) (map[string]any, error) {
	url := fmt.Sprintf("%s/self-service/verification/api", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create verification request: %w", err)
	}

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate email verification: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode verification response: %w", err)
	}

	return result, nil
}

// ChangePassword changes the password for a logged-in user
func (k *kratosService) ChangePassword(ctx context.Context, identityID string, oldPassword string, newPassword string) error {
	url := fmt.Sprintf("%s/self-service/settings/api", k.PublicURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create settings flow request: %w", err)
	}

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initiate settings flow: %w", err)
	}
	defer resp.Body.Close()

	var flow map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return fmt.Errorf("failed to decode settings flow: %w", err)
	}

	flowID, ok := flow["id"].(string)
	if !ok {
		return fmt.Errorf("invalid flow ID in response")
	}

	changeData := map[string]interface{}{
		"method":   "password",
		"password": newPassword,
	}
	data, err := json.Marshal(changeData)
	if err != nil {
		return fmt.Errorf("failed to marshal password change data: %w", err)
	}

	submitURL := fmt.Sprintf("%s/self-service/settings?flow=%s", k.PublicURL, flowID)
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
		return fmt.Errorf("password change failed with status: %s", http.StatusText(submitResp.StatusCode))
	}

	return nil
}
