package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// UpdateCache represents the cached update check data
type UpdateCache struct {
	LastChecked   time.Time `yaml:"last-checked"`
	LatestVersion string    `yaml:"latest-version,omitempty"`
	CheckInterval int       `yaml:"check-interval"` // seconds, default 24 hours
}

// DefaultUpdateCheckInterval is 24 hours in seconds
const DefaultUpdateCheckInterval = 86400

// GetUpdateCachePath returns the full path to the update cache file
func GetUpdateCachePath() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "update.yaml"), nil
}

// LoadUpdateCache loads the update cache from disk
func LoadUpdateCache() (*UpdateCache, error) {
	cachePath, err := GetUpdateCachePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get update cache path: %w", err)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty cache if file doesn't exist
			return &UpdateCache{
				CheckInterval: DefaultUpdateCheckInterval,
			}, nil
		}
		return nil, fmt.Errorf("failed to read update cache file: %w", err)
	}

	var cache UpdateCache
	if err := yaml.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse update cache file: %w", err)
	}

	// Ensure default check interval
	if cache.CheckInterval == 0 {
		cache.CheckInterval = DefaultUpdateCheckInterval
	}

	return &cache, nil
}

// SaveUpdateCache saves the update cache to disk
func SaveUpdateCache(cache *UpdateCache) error {
	if err := EnsureCacheDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachePath, err := GetUpdateCachePath()
	if err != nil {
		return fmt.Errorf("failed to get update cache path: %w", err)
	}

	data, err := yaml.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal update cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write update cache file: %w", err)
	}

	return nil
}

// ShouldCheckForUpdate returns true if enough time has passed since the last check
func (c *UpdateCache) ShouldCheckForUpdate() bool {
	if c.LastChecked.IsZero() {
		return true
	}
	interval := time.Duration(c.CheckInterval) * time.Second
	return time.Since(c.LastChecked) > interval
}

// UpdateLastChecked updates the last checked timestamp and optionally the latest version
func (c *UpdateCache) UpdateLastChecked(latestVersion string) {
	c.LastChecked = time.Now()
	if latestVersion != "" {
		c.LatestVersion = latestVersion
	}
	if c.CheckInterval == 0 {
		c.CheckInterval = DefaultUpdateCheckInterval
	}
}
