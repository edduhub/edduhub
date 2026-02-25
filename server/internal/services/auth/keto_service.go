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

// subject (who wants to do something)
// object (on whom they want to act on )
// relation (like edges in a graph)

type KetoService interface {
	CheckPermission(ctx context.Context, namespace, subject, action, resource string) (bool, error)
	CreateRelation(ctx context.Context, namespace, object, relation, subject string) error
	DeleteRelation(ctx context.Context, namespace, object, relation, subject string) error
}

type ketoService struct {
	readURL  string
	writeURL string
	client   *http.Client
}

func NewKetoService() *ketoService {
	return &ketoService{
		readURL:  os.Getenv("KETO_READ_URL"),
		writeURL: os.Getenv("KETO_WRITE_URL"),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (k *ketoService) CheckPermission(ctx context.Context, namespace, subject, action, resource string) (bool, error) {
	endpoint := fmt.Sprintf("%s/relation-tuples/check", k.readURL)
	query := map[string]string{
		"namespace": namespace,
		"object":    resource,
		"relation":  action,
		"subject":   subject,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return false, err
	}

	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := k.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Allowed, nil
}

func (k *ketoService) CreateRelation(ctx context.Context, namespace, object, relation, subject string) error {
	url := fmt.Sprintf("%s/admin/relation-tuples", k.writeURL)
	body := map[string]string{
		"namespace": namespace,
		"object":    object,
		"relation":  relation,
		"subject":   subject,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create relation: %d", resp.StatusCode)
	}

	return nil
}

func (k *ketoService) DeleteRelation(ctx context.Context, namespace, object, relation, subject string) error {
	url := fmt.Sprintf("%s/admin/relation-tuples", k.writeURL)
	query := map[string]string{
		"namespace": namespace,
		"object":    object,
		"relation":  relation,
		"subject":   subject,
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := k.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete relation: %d", resp.StatusCode)
	}

	return nil
}
