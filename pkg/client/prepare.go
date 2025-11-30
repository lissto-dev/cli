package client

import (
	"fmt"
)

// ImageCandidate represents a single image candidate that was tried
type ImageCandidate struct {
	ImageURL string `json:"image_url"`
	Tag      string `json:"tag"`
	Source   string `json:"source"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Digest   string `json:"digest,omitempty"`
}

// DetailedImageResolutionInfo contains detailed info about image resolution
type DetailedImageResolutionInfo struct {
	Service    string           `json:"service"`
	Digest     string           `json:"digest"`
	Image      string           `json:"image,omitempty"`
	Method     string           `json:"method"`
	Registry   string           `json:"registry,omitempty"`
	ImageName  string           `json:"image_name,omitempty"`
	Candidates []ImageCandidate `json:"candidates,omitempty"`
	Exposed    bool             `json:"exposed,omitempty"`
	URL        string           `json:"url,omitempty"`
}

// ExposedServiceInfo contains information about an exposed service
type ExposedServiceInfo struct {
	Service string `json:"service"`
	URL     string `json:"url"`
}

// PrepareStackResponse contains the result of stack preparation
type PrepareStackResponse struct {
	RequestID string                        `json:"request_id"`
	Blueprint string                        `json:"blueprint"`
	Images    []DetailedImageResolutionInfo `json:"images"`
	Exposed   []ExposedServiceInfo          `json:"exposed,omitempty"`
}

// PrepareStack prepares a stack by resolving images
func (c *Client) PrepareStack(blueprint, env, commit, branch, tag string, detailed bool) (*PrepareStackResponse, error) {
	reqBody := map[string]interface{}{
		"blueprint": blueprint,
		"env":       env,
		"detailed":  detailed,
	}

	if commit != "" {
		reqBody["commit"] = commit
	}
	if branch != "" {
		reqBody["branch"] = branch
	}
	if tag != "" {
		reqBody["tag"] = tag
	}

	var response PrepareStackResponse
	if err := c.Do("POST", "/api/v1/prepare", reqBody, &response); err != nil {
		return nil, fmt.Errorf("failed to prepare stack: %w", err)
	}

	return &response, nil
}





