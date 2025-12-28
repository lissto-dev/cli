package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIInfo represents the information returned from the API's health endpoint
type APIInfo struct {
	PublicURL string `json:"public_url"`
	APIID     string `json:"api_id"`
}

// GetAPIInfo fetches API information from the health endpoint
// This endpoint works without authentication for initial discovery
func (c *Client) GetAPIInfo() (*APIInfo, error) {
	url := c.baseURL + "/health?info=true"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var info APIInfo
	if err := json.Unmarshal(respBody, &info); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &info, nil
}
