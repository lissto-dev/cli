package config

import "fmt"

// GetCurrentEnv returns the current active environment
func (c *Config) GetCurrentEnv() (string, error) {
	if c.CurrentEnv == "" {
		return "", fmt.Errorf("no environment selected")
	}
	return c.CurrentEnv, nil
}

// SetCurrentEnv sets the current active environment
func (c *Config) SetCurrentEnv(env string) error {
	c.CurrentEnv = env
	return nil
}
