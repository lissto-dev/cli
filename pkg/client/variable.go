package client

import (
	"fmt"
)

// VariableResponse represents a variable config from the API
type VariableResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Scope        string            `json:"scope"`
	Env          string            `json:"env,omitempty"`
	Repository   string            `json:"repository,omitempty"`
	Data         map[string]string `json:"data"`
	CreatedAt    string            `json:"created_at,omitempty"`
	KeyUpdatedAt map[string]int64  `json:"key_updated_at,omitempty"` // Unix timestamps per key
}

// CreateVariableRequest represents a request to create a variable config
type CreateVariableRequest struct {
	Name       string            `json:"name"`
	Scope      string            `json:"scope,omitempty"`
	Env        string            `json:"env,omitempty"`
	Repository string            `json:"repository,omitempty"`
	Data       map[string]string `json:"data"`
}

// UpdateVariableRequest represents a request to update a variable config
type UpdateVariableRequest struct {
	Data map[string]string `json:"data"`
}

// ListVariables lists all variables
func (c *Client) ListVariables() ([]VariableResponse, error) {
	var variables []VariableResponse

	if err := c.Do("GET", "/api/v1/variables", nil, &variables); err != nil {
		return nil, fmt.Errorf("failed to list variables: %w", err)
	}

	return variables, nil
}

// GetVariable gets a specific variable
func (c *Client) GetVariable(id, scope, env, repository string) (*VariableResponse, error) {
	var variable VariableResponse
	path := buildResourcePath("/api/v1/variables", id, scope, env, repository)

	if err := c.Do("GET", path, nil, &variable); err != nil {
		return nil, fmt.Errorf("failed to get variable: %w", err)
	}

	return &variable, nil
}

// CreateVariable creates a new variable config
func (c *Client) CreateVariable(req *CreateVariableRequest) (*VariableResponse, error) {
	var variable VariableResponse

	if err := c.Do("POST", "/api/v1/variables", req, &variable); err != nil {
		return nil, fmt.Errorf("failed to create variable: %w", err)
	}

	return &variable, nil
}

// UpdateVariable updates an existing variable config
func (c *Client) UpdateVariable(id, scope, env, repository string, req *UpdateVariableRequest) (*VariableResponse, error) {
	var variable VariableResponse
	path := buildResourcePath("/api/v1/variables", id, scope, env, repository)

	if err := c.Do("PUT", path, req, &variable); err != nil {
		return nil, fmt.Errorf("failed to update variable: %w", err)
	}

	return &variable, nil
}

// DeleteVariable deletes a variable config
func (c *Client) DeleteVariable(id, scope, env, repository string) error {
	path := buildResourcePath("/api/v1/variables", id, scope, env, repository)

	if err := c.Do("DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete variable: %w", err)
	}

	return nil
}
