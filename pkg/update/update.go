package update

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/lissto-dev/cli/pkg/config"
)

const (
	// Repository is the GitHub repository for lissto CLI
	Repository = "lissto-dev/cli"
)

// CheckResult contains the result of an update check
type CheckResult struct {
	UpdateAvailable bool
	CurrentVersion  string
	LatestVersion   string
	ReleaseURL      string
}

// CheckForUpdate checks if a new version is available
// It respects the 24-hour cache interval and returns nil if no check is needed
func CheckForUpdate(currentVersion string) (*CheckResult, error) {
	// Skip update check for dev builds
	if currentVersion == "dev" || currentVersion == "" {
		return nil, nil
	}

	// Load update cache
	cache, err := config.LoadUpdateCache()
	if err != nil {
		// If we can't load cache, continue with check
		cache = &config.UpdateCache{
			CheckInterval: config.DefaultUpdateCheckInterval,
		}
	}

	// Check if we should perform an update check
	if !cache.ShouldCheckForUpdate() {
		// Return cached result if we have one
		if cache.LatestVersion != "" {
			return &CheckResult{
				UpdateAvailable: isNewerVersion(cache.LatestVersion, currentVersion),
				CurrentVersion:  currentVersion,
				LatestVersion:   cache.LatestVersion,
				ReleaseURL:      fmt.Sprintf("https://github.com/%s/releases/tag/%s", Repository, cache.LatestVersion),
			}, nil
		}
		return nil, nil
	}

	// Perform the update check using go-selfupdate
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(Repository))
	if err != nil {
		// Update cache timestamp even on failure to avoid hammering the API
		cache.UpdateLastChecked("")
		_ = config.SaveUpdateCache(cache)
		return nil, err
	}

	if !found {
		cache.UpdateLastChecked("")
		_ = config.SaveUpdateCache(cache)
		return nil, nil
	}

	latestVersion := latest.Version()
	releaseURL := fmt.Sprintf("https://github.com/%s/releases/tag/v%s", Repository, latestVersion)
	if latest.URL != "" {
		releaseURL = latest.URL
	}

	// Update cache with new information
	cache.UpdateLastChecked("v" + latestVersion)
	_ = config.SaveUpdateCache(cache)

	return &CheckResult{
		UpdateAvailable: latest.GreaterThan(currentVersion),
		CurrentVersion:  currentVersion,
		LatestVersion:   "v" + latestVersion,
		ReleaseURL:      releaseURL,
	}, nil
}

// isNewerVersion compares two version strings and returns true if latest is newer than current
// Used for cached version comparison
func isNewerVersion(latest, current string) bool {
	// Normalize versions by removing 'v' prefix
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// Split into parts
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	// Compare each part
	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		var latestNum, currentNum int
		fmt.Sscanf(latestParts[i], "%d", &latestNum)
		fmt.Sscanf(currentParts[i], "%d", &currentNum)

		if latestNum > currentNum {
			return true
		}
		if latestNum < currentNum {
			return false
		}
	}

	// If all compared parts are equal, the one with more parts is newer
	return len(latestParts) > len(currentParts)
}

// PrintUpdateMessage prints an update notification to stderr if an update is available
func PrintUpdateMessage(result *CheckResult) {
	if result == nil || !result.UpdateAvailable {
		return
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "╭─────────────────────────────────────────────────────────────╮\n")
	fmt.Fprintf(os.Stderr, "│  A new version of lissto is available: %-8s → %-8s │\n",
		truncateVersion(result.CurrentVersion, 8),
		truncateVersion(result.LatestVersion, 8))
	fmt.Fprintf(os.Stderr, "│                                                             │\n")
	fmt.Fprintf(os.Stderr, "│  Homebrew:  brew upgrade lissto                             │\n")
	fmt.Fprintf(os.Stderr, "│  Download:  %-47s │\n", truncateURL(result.ReleaseURL, 47))
	fmt.Fprintf(os.Stderr, "╰─────────────────────────────────────────────────────────────╯\n")
}

// truncateVersion truncates a version string to a max width
func truncateVersion(v string, maxWidth int) string {
	if len(v) > maxWidth {
		return v[:maxWidth]
	}
	return v
}

// truncateURL truncates a URL string to a max width
func truncateURL(url string, maxWidth int) string {
	if len(url) > maxWidth {
		return url[:maxWidth-3] + "..."
	}
	return url
}
