package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Entry represents a cached item with metadata
type Entry[T any] struct {
	Data      T         `yaml:"data"`
	CachedAt  time.Time `yaml:"cached-at"`
	ExpiresAt time.Time `yaml:"expires-at"`
}

// IsExpired returns true if the cache entry has expired
func (e *Entry[T]) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Age returns how long ago the entry was cached
func (e *Entry[T]) Age() time.Duration {
	return time.Since(e.CachedAt)
}

// TTL returns the remaining time until expiration
func (e *Entry[T]) TTL() time.Duration {
	return time.Until(e.ExpiresAt)
}

// Cache provides YAML-based file caching with TTL support
type Cache struct {
	dir string
}

// New creates a new Cache instance using the specified directory
func New(dir string) *Cache {
	return &Cache{dir: dir}
}

// Default creates a new Cache instance using the default cache directory
func Default() (*Cache, error) {
	dir, err := GetCacheDir()
	if err != nil {
		return nil, err
	}
	return New(dir), nil
}

// GetCacheDir returns the cache directory path
func GetCacheDir() (string, error) {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".cache")
	}
	return filepath.Join(cacheDir, "lissto"), nil
}

// EnsureDir ensures the cache directory exists
func (c *Cache) EnsureDir() error {
	return os.MkdirAll(c.dir, 0755)
}

// path returns the full path for a cache key
func (c *Cache) path(key string) string {
	return filepath.Join(c.dir, key+".yaml")
}

// Set stores data in the cache with the specified TTL
func (c *Cache) Set(key string, data any, ttl time.Duration) error {
	if err := c.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	now := time.Now()
	entry := struct {
		Data      any       `yaml:"data"`
		CachedAt  time.Time `yaml:"cached-at"`
		ExpiresAt time.Time `yaml:"expires-at"`
	}{
		Data:      data,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
	}

	content, err := yaml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	if err := os.WriteFile(c.path(key), content, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Get retrieves data from the cache. Returns false if not found or expired.
func (c *Cache) Get(key string, dest any) (bool, error) {
	content, err := os.ReadFile(c.path(key))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read cache file: %w", err)
	}

	// First unmarshal to check expiration
	var meta struct {
		ExpiresAt time.Time `yaml:"expires-at"`
	}
	if err := yaml.Unmarshal(content, &meta); err != nil {
		return false, fmt.Errorf("failed to parse cache metadata: %w", err)
	}

	if time.Now().After(meta.ExpiresAt) {
		return false, nil
	}

	// Unmarshal the full entry with data
	var entry struct {
		Data yaml.Node `yaml:"data"`
	}
	if err := yaml.Unmarshal(content, &entry); err != nil {
		return false, fmt.Errorf("failed to parse cache entry: %w", err)
	}

	if err := entry.Data.Decode(dest); err != nil {
		return false, fmt.Errorf("failed to decode cache data: %w", err)
	}

	return true, nil
}

// GetWithMeta retrieves data and metadata from the cache
func GetWithMeta[T any](c *Cache, key string) (*Entry[T], bool, error) {
	content, err := os.ReadFile(c.path(key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry Entry[T]
	if err := yaml.Unmarshal(content, &entry); err != nil {
		return nil, false, fmt.Errorf("failed to parse cache entry: %w", err)
	}

	if entry.IsExpired() {
		return nil, false, nil
	}

	return &entry, true, nil
}

// Delete removes an entry from the cache
func (c *Cache) Delete(key string) error {
	err := os.Remove(c.path(key))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}
	return nil
}

// Clear removes all entries from the cache
func (c *Cache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".yaml" {
			if err := os.Remove(filepath.Join(c.dir, entry.Name())); err != nil {
				return fmt.Errorf("failed to remove cache file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}
