package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lissto-dev/cli/pkg/config"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
	}{
		{
			name:     "newer major version",
			latest:   "v2.0.0",
			current:  "v1.0.0",
			expected: true,
		},
		{
			name:     "newer minor version",
			latest:   "v1.2.0",
			current:  "v1.1.0",
			expected: true,
		},
		{
			name:     "newer patch version",
			latest:   "v1.0.2",
			current:  "v1.0.1",
			expected: true,
		},
		{
			name:     "same version",
			latest:   "v1.0.0",
			current:  "v1.0.0",
			expected: false,
		},
		{
			name:     "older version",
			latest:   "v1.0.0",
			current:  "v2.0.0",
			expected: false,
		},
		{
			name:     "without v prefix",
			latest:   "1.2.0",
			current:  "1.1.0",
			expected: true,
		},
		{
			name:     "mixed v prefix",
			latest:   "v1.2.0",
			current:  "1.1.0",
			expected: true,
		},
		{
			name:     "more parts in latest",
			latest:   "v1.0.0.1",
			current:  "v1.0.0",
			expected: true,
		},
		{
			name:     "more parts in current",
			latest:   "v1.0.0",
			current:  "v1.0.0.1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNewerVersion(tt.latest, tt.current)
			if result != tt.expected {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.latest, tt.current, result, tt.expected)
			}
		})
	}
}

func TestCheckForUpdateSkipsDevVersion(t *testing.T) {
	result, err := CheckForUpdate("dev")
	if err != nil {
		t.Errorf("CheckForUpdate(\"dev\") returned error: %v", err)
	}
	if result != nil {
		t.Errorf("CheckForUpdate(\"dev\") should return nil result, got %+v", result)
	}

	result, err = CheckForUpdate("")
	if err != nil {
		t.Errorf("CheckForUpdate(\"\") returned error: %v", err)
	}
	if result != nil {
		t.Errorf("CheckForUpdate(\"\") should return nil result, got %+v", result)
	}
}

func TestUpdateCacheShouldCheckForUpdate(t *testing.T) {
	tests := []struct {
		name     string
		cache    *config.UpdateCache
		expected bool
	}{
		{
			name: "never checked",
			cache: &config.UpdateCache{
				CheckInterval: config.DefaultUpdateCheckInterval,
			},
			expected: true,
		},
		{
			name: "checked recently",
			cache: &config.UpdateCache{
				LastChecked:   time.Now().Add(-1 * time.Hour),
				CheckInterval: config.DefaultUpdateCheckInterval,
			},
			expected: false,
		},
		{
			name: "checked more than 24h ago",
			cache: &config.UpdateCache{
				LastChecked:   time.Now().Add(-25 * time.Hour),
				CheckInterval: config.DefaultUpdateCheckInterval,
			},
			expected: true,
		},
		{
			name: "custom interval - should check",
			cache: &config.UpdateCache{
				LastChecked:   time.Now().Add(-2 * time.Hour),
				CheckInterval: 3600, // 1 hour
			},
			expected: true,
		},
		{
			name: "custom interval - should not check",
			cache: &config.UpdateCache{
				LastChecked:   time.Now().Add(-30 * time.Minute),
				CheckInterval: 3600, // 1 hour
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cache.ShouldCheckForUpdate()
			if result != tt.expected {
				t.Errorf("ShouldCheckForUpdate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUpdateCacheUpdateLastChecked(t *testing.T) {
	cache := &config.UpdateCache{}

	// Update with version
	cache.UpdateLastChecked("v1.2.3")

	if cache.LatestVersion != "v1.2.3" {
		t.Errorf("LatestVersion = %q, want %q", cache.LatestVersion, "v1.2.3")
	}

	if cache.LastChecked.IsZero() {
		t.Error("LastChecked should not be zero")
	}

	if cache.CheckInterval != config.DefaultUpdateCheckInterval {
		t.Errorf("CheckInterval = %d, want %d", cache.CheckInterval, config.DefaultUpdateCheckInterval)
	}

	// Update without version (should keep existing)
	cache.UpdateLastChecked("")
	if cache.LatestVersion != "v1.2.3" {
		t.Errorf("LatestVersion should remain %q, got %q", "v1.2.3", cache.LatestVersion)
	}
}

func TestUpdateCachePersistence(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "lissto-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set XDG_CACHE_HOME to our temp directory
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Setenv("XDG_CACHE_HOME", oldCacheHome)

	// Create and save a cache
	cache := &config.UpdateCache{
		LastChecked:   time.Now(),
		LatestVersion: "v1.5.0",
		CheckInterval: 7200,
	}

	err = config.SaveUpdateCache(cache)
	if err != nil {
		t.Fatalf("Failed to save update cache: %v", err)
	}

	// Verify file was created
	cachePath := filepath.Join(tmpDir, "lissto", "update.yaml")
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Errorf("Cache file was not created at %s", cachePath)
	}

	// Load the cache back
	loadedCache, err := config.LoadUpdateCache()
	if err != nil {
		t.Fatalf("Failed to load update cache: %v", err)
	}

	if loadedCache.LatestVersion != "v1.5.0" {
		t.Errorf("LatestVersion = %q, want %q", loadedCache.LatestVersion, "v1.5.0")
	}

	if loadedCache.CheckInterval != 7200 {
		t.Errorf("CheckInterval = %d, want %d", loadedCache.CheckInterval, 7200)
	}
}

func TestPrintUpdateMessage(t *testing.T) {
	// Test that PrintUpdateMessage doesn't panic with nil
	PrintUpdateMessage(nil)

	// Test that PrintUpdateMessage doesn't panic with no update available
	PrintUpdateMessage(&CheckResult{
		UpdateAvailable: false,
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.0.0",
	})

	// Test with update available (just verify no panic)
	PrintUpdateMessage(&CheckResult{
		UpdateAvailable: true,
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
		ReleaseURL:      "https://github.com/lissto-dev/cli/releases/tag/v1.1.0",
	})
}

func TestCheckForUpdateDisabled(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "lissto-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set XDG_CONFIG_HOME to our temp directory
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

	// Set XDG_CACHE_HOME to our temp directory
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Setenv("XDG_CACHE_HOME", oldCacheHome)

	// Create a config with update check disabled
	cfg := &config.Config{
		DisableUpdateCheck: true,
	}
	err = config.SaveConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Check that update check returns nil when disabled
	result, err := CheckForUpdate("v1.0.0")
	if err != nil {
		t.Errorf("CheckForUpdate returned error: %v", err)
	}
	if result != nil {
		t.Errorf("CheckForUpdate should return nil when disabled, got %+v", result)
	}
}
