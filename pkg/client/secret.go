package client

import (
	"fmt"
)

// SecretResponse represents a secret config from the API (keys only, no values)
type SecretResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Scope      string   `json:"scope"`
	Env        string   `json:"env,omitempty"`
	Repository string   `json:"repository,omitempty"`
	Keys       []string `json:"keys"`
}

// CreateSecretRequest represents a request to create a secret config
type CreateSecretRequest struct {
	Name       string            `json:"name"`
	Scope      string            `json:"scope,omitempty"`
	Env        string            `json:"env,omitempty"`
	Repository string            `json:"repository,omitempty"`
	Secrets    map[string]string `json:"secrets,omitempty"`
}

// SetSecretRequest represents a request to set/update secret values
type SetSecretRequest struct {
	Secrets map[string]string `json:"secrets"`
}

// ListSecrets lists all secrets (keys only)
func (c *Client) ListSecrets() ([]SecretResponse, error) {
	var secrets []SecretResponse

	if err := c.Do("GET", "/api/v1/secrets", nil, &secrets); err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secrets, nil
}

// GetSecret gets a specific secret (keys only)
func (c *Client) GetSecret(id, scope, env, repository string) (*SecretResponse, error) {
	var secret SecretResponse
	path := buildResourcePath("/api/v1/secrets", id, scope, env, repository)

	if err := c.Do("GET", path, nil, &secret); err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return &secret, nil
}

// CreateSecret creates a new secret config
func (c *Client) CreateSecret(req *CreateSecretRequest) (*SecretResponse, error) {
	var secret SecretResponse

	if err := c.Do("POST", "/api/v1/secrets", req, &secret); err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	return &secret, nil
}

// UpdateSecret updates/sets secret values
func (c *Client) UpdateSecret(id, scope, env, repository string, req *SetSecretRequest) (*SecretResponse, error) {
	var secret SecretResponse
	path := buildResourcePath("/api/v1/secrets", id, scope, env, repository)

	if err := c.Do("PUT", path, req, &secret); err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}

	return &secret, nil
}

// DeleteSecret deletes a secret config
func (c *Client) DeleteSecret(id, scope, env, repository string) error {
	path := buildResourcePath("/api/v1/secrets", id, scope, env, repository)

	if err := c.Do("DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}
