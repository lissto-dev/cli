package cmdutil

import "os"

// Environment variable names for overriding auto-detection
const (
	EnvOverrideRepository  = "LISSTO_REPOSITORY"
	EnvOverrideComposeFile = "LISSTO_COMPOSE_FILE"
)

// Overrides holds environment variable overrides for CLI behavior
type Overrides struct {
	Repository  string // Overrides git repository auto-detection
	ComposeFile string // Overrides compose file auto-detection
}

// LoadOverrides reads all override environment variables
func LoadOverrides() Overrides {
	return Overrides{
		Repository:  os.Getenv(EnvOverrideRepository),
		ComposeFile: os.Getenv(EnvOverrideComposeFile),
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
