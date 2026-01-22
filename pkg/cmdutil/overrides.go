package cmdutil

import "os"

// Environment variable names
const (
	// CI/CD authentication
	EnvAPIKey = "LISSTO_API_KEY"
	EnvAPIURL = "LISSTO_API_URL"

	// Behavior overrides
	EnvOverrideRepository  = "LISSTO_REPOSITORY"
	EnvOverrideComposeFile = "LISSTO_COMPOSE_FILE"
)

// EnvOverrides holds all environment variable overrides for CLI behavior.
// This includes both authentication overrides (for CI/CD) and
// behavior overrides (repository, compose file).
type EnvOverrides struct {
	// Authentication (CI/CD mode)
	APIKey string // LISSTO_API_KEY - Direct API key for headless CI/CD
	APIURL string // LISSTO_API_URL - Direct API URL for headless CI/CD

	// Behavior overrides
	Repository  string // LISSTO_REPOSITORY - Override git repo auto-detection
	ComposeFile string // LISSTO_COMPOSE_FILE - Override compose file path
}

// LoadEnvOverrides reads all environment variable overrides
func LoadEnvOverrides() EnvOverrides {
	return EnvOverrides{
		APIKey:      os.Getenv(EnvAPIKey),
		APIURL:      os.Getenv(EnvAPIURL),
		Repository:  os.Getenv(EnvOverrideRepository),
		ComposeFile: os.Getenv(EnvOverrideComposeFile),
	}
}

// IsCICDMode returns true if both API key and URL are set (headless CI/CD mode)
func (o EnvOverrides) IsCICDMode() bool {
	return o.APIKey != "" && o.APIURL != ""
}

// HasRepository returns true if repository override is set
func (o EnvOverrides) HasRepository() bool {
	return o.Repository != ""
}

// HasComposeFile returns true if compose file override is set
func (o EnvOverrides) HasComposeFile() bool {
	return o.ComposeFile != ""
}
