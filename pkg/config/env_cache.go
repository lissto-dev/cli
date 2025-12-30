package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// EnvCache represents the cached environment data
type EnvCache struct {
	LastUpdated time.Time `yaml:"last-updated"`
	TTL         int       `yaml:"ttl"` // seconds
	Envs        []EnvInfo `yaml:"envs"`
}

// EnvInfo represents cached environment information
type EnvInfo struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// LoadEnvCache loads the environment cache from disk
func LoadEnvCache() (*EnvCache, error) {
	cachePath, err := GetEnvCachePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache path: %w", err)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty cache if file doesn't exist
			return &EnvCache{
				TTL:  300, // Default 5 minutes
				Envs: []EnvInfo{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache EnvCache
	if err := yaml.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &cache, nil
}

// SaveEnvCache saves the environment cache to disk
func SaveEnvCache(cache *EnvCache) error {
	if err := EnsureCacheDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachePath, err := GetEnvCachePath()
	if err != nil {
		return fmt.Errorf("failed to get cache path: %w", err)
	}

	data, err := yaml.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// IsStale checks if the cache is stale
func (c *EnvCache) IsStale() bool {
	if c.LastUpdated.IsZero() {
		return true
	}
	ttlDuration := time.Duration(c.TTL) * time.Second
	return time.Since(c.LastUpdated) > ttlDuration
}

// UpdateEnvs updates the cached environments
func (c *EnvCache) UpdateEnvs(envs []EnvInfo) {
	c.Envs = envs
	c.LastUpdated = time.Now()
	if c.TTL == 0 {
		c.TTL = 300 // Default 5 minutes
	}
}

// GetEnv returns an environment by name
func (c *EnvCache) GetEnv(name string) (*EnvInfo, error) {
	for _, env := range c.Envs {
		if env.Name == name {
			return &env, nil
		}
	}
	return nil, fmt.Errorf("environment '%s' not found in cache", name)
}
