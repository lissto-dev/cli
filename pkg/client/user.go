package client

import "fmt"

// User represents a user in the system
type User struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// GetCurrentUser fetches the current user info
func (c *Client) GetCurrentUser() (*User, error) {
	var user User

	if err := c.Do("GET", "/api/v1/user/me", nil, &user); err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	if user.Name == "" {
		return nil, fmt.Errorf("invalid response from API: missing user name")
	}

	return &user, nil
}
