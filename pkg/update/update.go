package update

import (
	"bytes"
	"context"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/lissto-dev/cli/pkg/cache"
	"github.com/lissto-dev/cli/pkg/config"
)

const (
	// Repository is the GitHub repository for lissto CLI
	Repository = "lissto-dev/cli"

	// CacheKey is the cache key for update check results
	CacheKey = "update"

	// CacheTTL is the time-to-live for cached update check results
	CacheTTL = 24 * time.Hour
)

// CachedRelease stores release information from go-selfupdate library
type CachedRelease struct {
	Version    string `yaml:"version"`
	URL        string `yaml:"url"`
	ReleaseURL string `yaml:"release-url"`
}

// CheckResult contains the result of an update check
type CheckResult struct {
	UpdateAvailable bool
	CurrentVersion  string
	LatestVersion   string
	ReleaseURL      string
}

// isNewerVersion compares two semver strings using the same library as go-selfupdate
func isNewerVersion(latest, current string) bool {
	// Normalize by removing 'v' prefix
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		return false
	}
	currentVer, err := semver.NewVersion(current)
	if err != nil {
		return false
	}
	return latestVer.GreaterThan(currentVer)
}

// CheckForUpdate checks if a new version is available
// It respects the 24-hour cache interval and returns nil if no check is needed
func CheckForUpdate(currentVersion string) (*CheckResult, error) {
	// Skip update check for dev builds
	if currentVersion == "dev" || currentVersion == "" {
		return nil, nil
	}

	// Check if update check is enabled in config
	cfg, err := config.LoadConfig()
	if err == nil && !cfg.Settings.UpdateCheck {
		return nil, nil
	}

	// Get cache instance
	c, err := cache.Default()
	if err != nil {
		return nil, err
	}

	// Try to get cached release info
	var cached CachedRelease
	found, err := c.Get(CacheKey, &cached)
	if err == nil && found && cached.Version != "" {
		// Use cached data - compare using semver library
		return &CheckResult{
			UpdateAvailable: isNewerVersion(cached.Version, currentVersion),
			CurrentVersion:  currentVersion,
			LatestVersion:   cached.Version,
			ReleaseURL:      cached.ReleaseURL,
		}, nil
	}

	// Perform the update check using go-selfupdate
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(Repository))
	if err != nil {
		// Cache empty result on failure to avoid hammering the API
		_ = c.Set(CacheKey, CachedRelease{}, CacheTTL)
		return nil, err
	}

	if !found {
		_ = c.Set(CacheKey, CachedRelease{}, CacheTTL)
		return nil, nil
	}

	// Cache the release info from library
	cachedRelease := CachedRelease{
		Version:    latest.Version(),
		URL:        latest.AssetURL,
		ReleaseURL: latest.URL,
	}
	_ = c.Set(CacheKey, cachedRelease, CacheTTL)

	return &CheckResult{
		UpdateAvailable: latest.GreaterThan(currentVersion),
		CurrentVersion:  currentVersion,
		LatestVersion:   latest.Version(),
		ReleaseURL:      latest.URL,
	}, nil
}

// Update message template
const updateMessageTemplate = `
╭─────────────────────────────────────────────────────────────╮
│  A new version of lissto is available: {{printf "%-8s" .CurrentVersion}} → {{printf "%-8s" .LatestVersion}} │
│                                                             │
│  Homebrew:  brew upgrade lissto                             │
│  Download:  {{printf "%-47s" (truncate .ReleaseURL 47)}} │
╰─────────────────────────────────────────────────────────────╯
`

// PrintUpdateMessage prints an update notification to stderr if an update is available
func PrintUpdateMessage(result *CheckResult) {
	if result == nil || !result.UpdateAvailable {
		return
	}

	funcMap := template.FuncMap{
		"truncate": func(s string, maxLen int) string {
			if len(s) > maxLen {
				return s[:maxLen-3] + "..."
			}
			return s
		},
	}

	tmpl, err := template.New("update").Funcs(funcMap).Parse(updateMessageTemplate)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, result); err != nil {
		return
	}

	os.Stderr.Write(buf.Bytes())
}
