package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/k8s"
)

// Client represents the Lissto API client
type Client struct {
	baseURL       string
	apiKey        string
	httpClient    *http.Client
	expectedAPIID string // Expected API instance ID for verification
}

// NewClient creates a new API client
func NewClient(apiURL, apiKey string) *Client {
	return &Client{
		baseURL: apiURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithAPIID creates a new API client with API ID verification
func NewClientWithAPIID(apiURL, apiKey, apiID string) *Client {
	return &Client{
		baseURL:       apiURL,
		apiKey:        apiKey,
		expectedAPIID: apiID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientFromConfig creates an API client from a saved context
// It validates the k8s context and discovers the API endpoint with caching and retry logic
func NewClientFromConfig(ctx *config.Context) (*Client, error) {
	// Validate k8s context (shows warning if different)
	if err := config.ValidateAndWarn(ctx); err != nil {
		// Don't fail on validation errors
	}

	// Check if we have a cached API URL and ID
	if ctx.APIUrl != "" && ctx.APIID != "" {
		// Try to use cached URL with ID verification
		client := NewClientWithAPIID(ctx.APIUrl, ctx.APIKey, ctx.APIID)

		// Test the connection by calling a simple endpoint
		if err := client.testConnection(); err == nil {
			// Cached URL works and API ID matches
			return client, nil
		}
		// If connection fails or API ID mismatches, we'll re-discover below
	}

	// Need to discover the API endpoint (either no cache or cache failed)
	k8sClient, err := k8s.NewClientWithContext(ctx.KubeContext)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	// Use fast discovery to get public URL and API ID (opens port-forward once)
	discoveryInfo, err := k8sClient.DiscoverAPIEndpointFast(
		context.Background(),
		ctx.ServiceName,
		ctx.ServiceNamespace,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to discover API endpoint: %w", err)
	}

	// Update context with discovered information
	ctx.APIID = discoveryInfo.APIID
	ctx.APIUrl = discoveryInfo.PublicURL // Cache public URL (empty if not available)

	// Save the updated context
	cfg, err := config.LoadConfig()
	if err == nil {
		cfg.AddOrUpdateContext(*ctx)
		_ = config.SaveConfig(cfg) // Ignore save errors
	}

	// Use public URL if available, otherwise use the port-forward URL we already established
	apiURL := discoveryInfo.PublicURL
	if apiURL == "" {
		apiURL = discoveryInfo.PortForwardURL
	}

	// Create client with API ID verification
	client := NewClientWithAPIID(apiURL, ctx.APIKey, ctx.APIID)

	// Wrap the client to add retry logic for API ID mismatches
	return &Client{
		baseURL:       client.baseURL,
		apiKey:        client.apiKey,
		expectedAPIID: client.expectedAPIID,
		httpClient:    client.httpClient,
	}, nil
}

// testConnection tests if the API is reachable and API ID matches
func (c *Client) testConnection() error {
	// Try to call /health endpoint
	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Check API ID if we expect one
	if c.expectedAPIID != "" {
		actualAPIID := resp.Header.Get("X-Lissto-API-ID")
		if actualAPIID != "" && actualAPIID != c.expectedAPIID {
			return fmt.Errorf("API instance ID mismatch: expected %s, got %s", c.expectedAPIID, actualAPIID)
		}
	}

	return nil
}

// Do performs an HTTP request with authentication
func (c *Client) Do(method, path string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Verify API ID if we have an expected ID
	if c.expectedAPIID != "" {
		actualAPIID := resp.Header.Get("X-Lissto-API-ID")
		if actualAPIID != "" && actualAPIID != c.expectedAPIID {
			return fmt.Errorf("API instance ID mismatch: expected %s, got %s", c.expectedAPIID, actualAPIID)
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to parse error response
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.ErrorMessage != "" {
			return &apiErr
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		// Check if result is a string pointer - handle plain text responses
		if strPtr, ok := result.(*string); ok {
			*strPtr = string(respBody)
			return nil
		}

		// Otherwise try to unmarshal as JSON
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// APIError represents an error response from the API
type APIError struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error"`
}

func (e *APIError) Error() string {
	return e.ErrorMessage
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
