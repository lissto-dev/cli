package cmdutil

import "os"

// Environment variable names for overriding auto-detection
const (
	EnvOverrideRepository  = "LISSTO_REPOSITORY"
	EnvOverrideComposeFile = "LISSTO_COMPOSE_FILE"
)

// Environment variable names for CI/CD authentication (e.g., GitHub Actions)
const (
	EnvAPIKey = "LISSTO_API_KEY"
	EnvAPIURL = "LISSTO_API_URL"
)

// Overrides holds environment variable overrides for CLI behavior
type Overrides struct {
	Repository  string // Overrides git repository auto-detection
	ComposeFile string // Overrides compose file auto-detection
}

// AuthOverrides holds environment variable overrides for API authentication
type AuthOverrides struct {
	APIKey string // Direct API key for CI/CD environments
	APIURL string // Direct API URL for CI/CD environments
}

// LoadOverrides reads all override environment variables
func LoadOverrides() Overrides {
	return Overrides{
		Repository:  os.Getenv(EnvOverrideRepository),
		ComposeFile: os.Getenv(EnvOverrideComposeFile),
	}
}

// LoadAuthOverrides reads authentication override environment variables
func LoadAuthOverrides() AuthOverrides {
	return AuthOverrides{
		APIKey: os.Getenv(EnvAPIKey),
		APIURL: os.Getenv(EnvAPIURL),
	}
}

// HasRepository returns true if repository override is set
func (o Overrides) HasRepository() bool {
	return o.Repository != ""
}

// HasComposeFile returns true if compose file override is set
func (o Overrides) HasComposeFile() bool {
	return o.ComposeFile != ""
}

// IsConfigured returns true if both API key and URL are set
func (a AuthOverrides) IsConfigured() bool {
	return a.APIKey != "" && a.APIURL != ""
}
