package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
