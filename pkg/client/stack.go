package client

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/types"
)

// ListStacks lists all stacks
func (c *Client) ListStacks(env string) ([]types.Stack, error) {
	var stacks []types.Stack

	path := "/api/v1/stacks"
	if env != "" {
		path = fmt.Sprintf("%s?env=%s", path, env)
	}

	if err := c.Do("GET", path, nil, &stacks); err != nil {
		return nil, fmt.Errorf("failed to list stacks: %w", err)
	}

	return stacks, nil
}

// GetStack gets a specific stack (returns identifier)
func (c *Client) GetStack(name, env string) (string, error) {
	var identifier string

	path := fmt.Sprintf("/api/v1/stacks/%s", name)
	if env != "" {
		path = fmt.Sprintf("%s?env=%s", path, env)
	}

	if err := c.Do("GET", path, nil, &identifier); err != nil {
		return "", fmt.Errorf("failed to get stack: %w", err)
	}

	return identifier, nil
}

// CreateStack creates a new stack using a prepared request_id
func (c *Client) CreateStack(blueprint, env, requestID string) (string, error) {
	reqBody := map[string]interface{}{
		"blueprint":  blueprint,
		"env":        env,
		"request_id": requestID,
	}

	var identifier string
	if err := c.Do("POST", "/api/v1/stacks", reqBody, &identifier); err != nil {
		return "", fmt.Errorf("failed to create stack: %w", err)
	}

	return identifier, nil
}

// UpdateStack updates a stack's images
func (c *Client) UpdateStack(name string, images map[string]interface{}) error {
	reqBody := map[string]interface{}{
		"images": images,
	}

	path := fmt.Sprintf("/api/v1/stacks/%s", name)

	if err := c.Do("PUT", path, reqBody, nil); err != nil {
		return fmt.Errorf("failed to update stack: %w", err)
	}

	return nil
}

// DeleteStack deletes a stack
func (c *Client) DeleteStack(name, env string) error {
	path := fmt.Sprintf("/api/v1/stacks/%s", name)
	if env != "" {
		path = fmt.Sprintf("%s?env=%s", path, env)
	}

	if err := c.Do("DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete stack: %w", err)
	}

	return nil
}
