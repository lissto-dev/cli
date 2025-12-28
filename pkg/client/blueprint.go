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

// DetailedMetadata represents normalized k8s object metadata
type DetailedMetadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"` // Normalized: "global" or developer name
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   string            `json:"createdAt,omitempty"`
}

// BlueprintDetailedResponse represents full blueprint with all annotations
type BlueprintDetailedResponse struct {
	Metadata DetailedMetadata `json:"metadata"`
	Spec     struct {
		DockerCompose string  `json:"dockerCompose"`
		Hash          string  `json:"hash"`
		Data          *string `json:"data,omitempty"`
	} `json:"spec"`
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

// GetBlueprintDetailed gets the complete blueprint object including all annotations
func (c *Client) GetBlueprintDetailed(name string) (*BlueprintDetailedResponse, error) {
	var blueprint BlueprintDetailedResponse

	path := fmt.Sprintf("/api/v1/blueprints/%s?format=detailed", name)

	if err := c.Do("GET", path, nil, &blueprint); err != nil {
		return nil, fmt.Errorf("failed to get blueprint details: %w", err)
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

// FindBlueprintsByRepository finds all blueprints matching a normalized repository URL
// Returns blueprints sorted by ID descending (newest first)
func (c *Client) FindBlueprintsByRepository(normalizedRepo string) ([]BlueprintResponse, error) {
	allBlueprints, err := c.ListBlueprints(true)
	if err != nil {
		return nil, err
	}

	var matching []BlueprintResponse
	for _, bp := range allBlueprints {
		// Get detailed info to access repository annotation
		detailed, err := c.GetBlueprintDetailed(bp.ID)
		if err != nil {
			continue // Skip if can't get details
		}

		// Check repository annotation
		if repo, ok := detailed.Metadata.Annotations["lissto.dev/repository"]; ok && repo == normalizedRepo {
			matching = append(matching, bp)
		}
	}

	// Sort by ID descending (newest first)
	// Blueprint IDs have format: scope/YYYYMMDD-HHMMSS-hash
	// Lexicographic sort works due to timestamp format
	for i := 0; i < len(matching)-1; i++ {
		for j := i + 1; j < len(matching); j++ {
			if matching[i].ID < matching[j].ID {
				matching[i], matching[j] = matching[j], matching[i]
			}
		}
	}

	return matching, nil
}
