package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lissto-dev/cli/pkg/config"
)

const (
	// GitHubReleasesURL is the URL to fetch the latest release from GitHub
	GitHubReleasesURL = "https://api.github.com/repos/lissto-dev/cli/releases/latest"
)

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

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
				ReleaseURL:      fmt.Sprintf("https://github.com/lissto-dev/cli/releases/tag/%s", cache.LatestVersion),
			}, nil
		}
		return nil, nil
	}

	// Perform the update check
	release, err := fetchLatestRelease()
	if err != nil {
		// Update cache timestamp even on failure to avoid hammering the API
		cache.UpdateLastChecked("")
		_ = config.SaveUpdateCache(cache)
		return nil, err
	}

	// Update cache with new information
	cache.UpdateLastChecked(release.TagName)
	_ = config.SaveUpdateCache(cache)

	return &CheckResult{
		UpdateAvailable: isNewerVersion(release.TagName, currentVersion),
		CurrentVersion:  currentVersion,
		LatestVersion:   release.TagName,
		ReleaseURL:      release.HTMLURL,
	}, nil
}

// fetchLatestRelease fetches the latest release from GitHub
func fetchLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", GitHubReleasesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "lissto-cli")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

// isNewerVersion compares two version strings and returns true if latest is newer than current
// Versions are expected to be in format "vX.Y.Z" or "X.Y.Z"
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
	fmt.Fprintf(os.Stderr, "│  A new version of lissto is available: %s → %s  │\n",
		padVersion(result.CurrentVersion, 8),
		padVersion(result.LatestVersion, 8))
	fmt.Fprintf(os.Stderr, "│  Run: brew upgrade lissto                                   │\n")
	fmt.Fprintf(os.Stderr, "│  Or download from: %s  │\n", padURL(result.ReleaseURL, 39))
	fmt.Fprintf(os.Stderr, "╰─────────────────────────────────────────────────────────────╯\n")
}

// padVersion pads a version string to a fixed width
func padVersion(v string, width int) string {
	if len(v) >= width {
		return v[:width]
	}
	return v + strings.Repeat(" ", width-len(v))
}

// padURL pads a URL string to a fixed width
func padURL(url string, width int) string {
	if len(url) >= width {
		return url[:width]
	}
	return url + strings.Repeat(" ", width-len(url))
}
