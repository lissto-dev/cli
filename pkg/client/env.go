package client

import (
	"fmt"
)

// EnvResponse represents an environment from the API
type EnvResponse struct {
	ID   string `json:"id"`   // Scoped identifier: namespace/envname
	Name string `json:"name"` // Env name
}

// ListEnvs lists all environments
func (c *Client) ListEnvs() ([]EnvResponse, error) {
	var envs []EnvResponse

	if err := c.Do("GET", "/api/v1/envs", nil, &envs); err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	return envs, nil
}

// GetEnv gets a specific environment
func (c *Client) GetEnv(name string) (*EnvResponse, error) {
	var env EnvResponse

	path := fmt.Sprintf("/api/v1/envs/%s", name)
	if err := c.Do("GET", path, nil, &env); err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}

	return &env, nil
}

// CreateEnv creates a new environment
func (c *Client) CreateEnv(name string) (string, error) {
	reqBody := map[string]interface{}{
		"name": name,
	}

	var identifier string
	if err := c.Do("POST", "/api/v1/envs", reqBody, &identifier); err != nil {
		return "", fmt.Errorf("failed to create environment: %w", err)
	}

	return identifier, nil
}
