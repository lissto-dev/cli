package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	CurrentContext string    `yaml:"current-context"`
	Contexts       []Context `yaml:"contexts"`
	CurrentEnv     string    `yaml:"current-env,omitempty"`
	Kubeconfig     string    `yaml:"kubeconfig,omitempty"`
}

// Context represents an API connection context
type Context struct {
	Name             string `yaml:"name"`
	KubeContext      string `yaml:"kube-context"`
	ServiceName      string `yaml:"service-name"`
	ServiceNamespace string `yaml:"service-namespace"`
	APIKey           string `yaml:"api-key"`
	APIUrl           string `yaml:"api-url,omitempty"`
	APIID            string `yaml:"api-id,omitempty"`
}

// LoadConfig loads the configuration from disk
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{
				Contexts: []Context{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetCurrentContext returns the current active context
func (c *Config) GetCurrentContext() (*Context, error) {
	if c.CurrentContext == "" {
		return nil, fmt.Errorf("no context selected")
	}

	for _, ctx := range c.Contexts {
		if ctx.Name == c.CurrentContext {
			return &ctx, nil
		}
	}

	return nil, fmt.Errorf("current context '%s' not found", c.CurrentContext)
}

// AddOrUpdateContext adds a new context or updates an existing one
func (c *Config) AddOrUpdateContext(ctx Context) {
	for i, existingCtx := range c.Contexts {
		if existingCtx.Name == ctx.Name {
			c.Contexts[i] = ctx
			return
		}
	}
	c.Contexts = append(c.Contexts, ctx)
}

// DeleteContext removes a context by name
func (c *Config) DeleteContext(name string) error {
	for i, ctx := range c.Contexts {
		if ctx.Name == name {
			c.Contexts = append(c.Contexts[:i], c.Contexts[i+1:]...)
			// If we're deleting the current context, clear it
			if c.CurrentContext == name {
				c.CurrentContext = ""
			}
			return nil
		}
	}
	return fmt.Errorf("context '%s' not found", name)
}

// SetCurrentContext sets the current context
func (c *Config) SetCurrentContext(name string) error {
	for _, ctx := range c.Contexts {
		if ctx.Name == name {
			c.CurrentContext = name
			return nil
		}
	}
	return fmt.Errorf("context '%s' not found", name)
}

// GetContext returns a context by name
func (c *Config) GetContext(name string) (*Context, error) {
	for _, ctx := range c.Contexts {
		if ctx.Name == name {
			return &ctx, nil
		}
	}
	return nil, fmt.Errorf("context '%s' not found", name)
}
