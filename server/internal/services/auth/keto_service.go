package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// KetoService defines the interface for authorization
type KetoService interface {
	CheckPermission(ctx context.Context, namespace, subject, action, resource string) (bool, error)
	CreateRelation(ctx context.Context, namespace, object, relation, subject string) error
	DeleteRelation(ctx context.Context, namespace, object, relation, subject string) error
}

// ketoService implements KetoService using Ory Keto
type ketoService struct {
	ReadURL  string
	WriteURL string
	Client   *http.Client
}

func NewKetoService() *ketoService {
	return &ketoService{
		ReadURL:  os.Getenv("KETO_READ_URL"),
		WriteURL: os.Getenv("KETO_WRITE_URL"),
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// CheckPermission checks if a subject has permission to perform an action on a resource
// Uses the Keto Check Permission API
func (k *ketoService) CheckPermission(ctx context.Context, namespace, subject, action, resource string) (bool, error) {
	// Use the OpenAPI check endpoint for better compatibility
	endpoint := fmt.Sprintf("%s/relation-tuples/check/openapi", k.ReadURL)

	// Build query parameters
	query := url.Values{}
	query.Set("namespace", namespace)
	query.Set("object", resource)
	query.Set("relation", action)
	query.Set("subject_id", subject) // Use subject_id for individual subjects

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return false, fmt.Errorf("failed to create check request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := k.Client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		Allowed bool `json:"allowed"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode permission response: %w", err)
	}

	return result.Allowed, nil
}

// CreateRelation creates a new relation tuple in Keto
// namespace: the namespace (e.g., "app")
// object: the resource (e.g., "role:admin")
// relation: the relation (e.g., "member")
// subject: the subject (e.g., "user:123" or just "123")
func (k *ketoService) CreateRelation(ctx context.Context, namespace, object, relation, subject string) error {
	apiURL := fmt.Sprintf("%s/admin/relation-tuples", k.WriteURL)

	// Build the relation tuple
	body := map[string]string{
		"namespace":  namespace,
		"object":     object,
		"relation":   relation,
		"subject_id": subject, // Use subject_id for individual subjects
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal relation: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", apiURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create relation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := k.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create relation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("failed to create relation: %d", resp.StatusCode)
	}

	return nil
}

// DeleteRelation deletes a relation tuple from Keto
func (k *ketoService) DeleteRelation(ctx context.Context, namespace, object, relation, subject string) error {
	apiURL := fmt.Sprintf("%s/admin/relation-tuples", k.WriteURL)

	// Build query parameters for deletion
	query := url.Values{}
	query.Set("namespace", namespace)
	query.Set("object", object)
	query.Set("relation", relation)
	query.Set("subject_id", subject)

	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL+"?"+query.Encode(), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete relation request: %w", err)
	}

	resp, err := k.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete relation: %d", resp.StatusCode)
	}

	return nil
}
