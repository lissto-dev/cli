package client

import (
	"fmt"
)

// ServiceMetadata represents service metadata from the API
type ServiceMetadata struct {
	Services []string `json:"services"` // List of service names
	Infra    []string `json:"infra"`    // List of infrastructure service names
}

// BlueprintResponse represents a blueprint from the API
type BlueprintResponse struct {
	ID      string          `json:"id"`      // Scoped identifier
	Title   string          `json:"title"`   // Title from annotations
	Content ServiceMetadata `json:"content"` // Service metadata
}

// ListBlueprints lists all blueprints (user and optionally global)
func (c *Client) ListBlueprints(includeGlobal bool) ([]BlueprintResponse, error) {
	var blueprints []BlueprintResponse

	path := "/api/v1/blueprints"
	if includeGlobal {
		path += "?global=true"
	}

	if err := c.Do("GET", path, nil, &blueprints); err != nil {
		return nil, fmt.Errorf("failed to list blueprints: %w", err)
	}

	return blueprints, nil
}

// GetBlueprint gets a specific blueprint by name
func (c *Client) GetBlueprint(name string) (*BlueprintResponse, error) {
	var blueprint BlueprintResponse

	path := fmt.Sprintf("/api/v1/blueprints/%s", name)

	if err := c.Do("GET", path, nil, &blueprint); err != nil {
		return nil, fmt.Errorf("failed to get blueprint: %w", err)
	}

	return &blueprint, nil
}

// CreateBlueprintRequest represents the request to create a blueprint
type CreateBlueprintRequest struct {
	Compose    string
	Branch     string
	Author     string
	Repository string
}

// CreateBlueprint creates a new blueprint
func (c *Client) CreateBlueprint(req CreateBlueprintRequest) (string, error) {
	reqBody := map[string]interface{}{
		"compose": req.Compose,
	}

	// Add optional fields if provided
	if req.Branch != "" {
		reqBody["branch"] = req.Branch
	}
	if req.Author != "" {
		reqBody["author"] = req.Author
	}
	if req.Repository != "" {
		reqBody["repository"] = req.Repository
	}

	var identifier string
	if err := c.Do("POST", "/api/v1/blueprints", reqBody, &identifier); err != nil {
		return "", fmt.Errorf("failed to create blueprint: %w", err)
	}

	return identifier, nil
}

// DeleteBlueprint deletes a blueprint
func (c *Client) DeleteBlueprint(name string) error {
	path := fmt.Sprintf("/api/v1/blueprints/%s", name)

	if err := c.Do("DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete blueprint: %w", err)
	}

	return nil
}
