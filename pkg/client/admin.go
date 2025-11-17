package client

import "fmt"

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// CreateAPIKeyResponse represents the response from creating an API key
type CreateAPIKeyResponse struct {
	APIKey string `json:"api_key"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

// CreateAPIKey creates a new API key (admin only)
func (c *Client) CreateAPIKey(req CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	var response struct {
		Success bool                  `json:"success"`
		Data    *CreateAPIKeyResponse `json:"data"`
		Message string                `json:"message"`
	}

	if err := c.Do("POST", "/api/v1/_internal/api-keys", req, &response); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	if !response.Success || response.Data == nil {
		return nil, fmt.Errorf("failed to create API key: %s", response.Message)
	}

	return response.Data, nil
}
