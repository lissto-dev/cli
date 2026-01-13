package cmdutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/output"
	"github.com/spf13/cobra"
)

// GetAPIClient returns configured API client from environment variables or current context.
// Environment variables (LISSTO_API_KEY, LISSTO_API_URL) take precedence over config file.
// This enables headless CI/CD usage (e.g., GitHub Actions) without requiring login.
func GetAPIClient() (*client.Client, error) {
	// Check for environment variable authentication first (CI/CD mode)
	authOverrides := LoadAuthOverrides()
	if authOverrides.IsConfigured() {
		return client.NewClient(authOverrides.APIURL, authOverrides.APIKey), nil
	}

	// Fall back to config-based authentication
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("no context selected. Run 'lissto login' first, or set %s and %s environment variables", EnvAPIKey, EnvAPIURL)
	}

	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize API client: %w", err)
	}
	return apiClient, nil
}

// GetAPIClientAndEnv returns API client and resolved environment name.
// Environment variables (LISSTO_API_KEY, LISSTO_API_URL) take precedence over config file.
func GetAPIClientAndEnv(cmd *cobra.Command) (*client.Client, string, error) {
	// Get environment from flag first
	envName, _ := cmd.Flags().GetString("env")

	// Check for environment variable authentication first (CI/CD mode)
	authOverrides := LoadAuthOverrides()
	if authOverrides.IsConfigured() {
		if envName == "" {
			return nil, "", fmt.Errorf("--env flag is required when using environment variable authentication")
		}
		return client.NewClient(authOverrides.APIURL, authOverrides.APIKey), envName, nil
	}

	// Fall back to config-based authentication
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config: %w", err)
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return nil, "", fmt.Errorf("no context selected. Run 'lissto login' first, or set %s and %s environment variables", EnvAPIKey, EnvAPIURL)
	}

	// Get environment from config if not provided via flag
	if envName == "" {
		envName = cfg.CurrentEnv
	}

	if envName == "" {
		return nil, "", fmt.Errorf("no environment selected. Use --env flag or 'lissto env use <name>'")
	}

	// Create API client with k8s discovery and validation
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize API client: %w", err)
	}

	return apiClient, envName, nil
}

// GetCurrentEnv returns current environment from config
func GetCurrentEnv() string {
	cfg, err := config.LoadConfig()
	if err != nil {
		return ""
	}
	return cfg.CurrentEnv
}

// GetOutputFormat extracts output format flag from command
func GetOutputFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("output")
	return format
}

// PrintOutput handles JSON/YAML/custom output formatting
// If data is provided and format is json/yaml, it will be serialized
// Otherwise, customFormatter will be called for default formatting
func PrintOutput(cmd *cobra.Command, data interface{}, customFormatter func()) error {
	format := GetOutputFormat(cmd)

	switch format {
	case "json":
		return output.PrintJSON(os.Stdout, data)
	case "yaml":
		return output.PrintYAML(os.Stdout, data)
	default:
		if customFormatter != nil {
			customFormatter()
		}
		return nil
	}
}

// ParseKeyValueArgs parses KEY=value arguments into a map
func ParseKeyValueArgs(args []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format: %s (expected KEY=value)", arg)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}

// GenerateResourceName generates a resource name based on scope, env, and repository
func GenerateResourceName(scope, env, repository string) string {
	switch scope {
	case "global":
		return "global"
	case "repo":
		// Extract repo name from full path
		parts := strings.Split(repository, "/")
		if len(parts) > 0 {
			repoName := parts[len(parts)-1]
			// Remove .git suffix if present
			repoName = strings.TrimSuffix(repoName, ".git")
			return fmt.Sprintf("repo-%s", repoName)
		}
		return "repo"
	case "env":
		fallthrough
	default:
		return env
	}
}

// GetKeysFromMap returns a slice of keys from a map
func GetKeysFromMap(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
