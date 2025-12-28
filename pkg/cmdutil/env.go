package cmdutil

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/interactive"
)

// GetOrCreateDefaultEnv determines or creates an environment for the user
// Priority:
// 1. Use provided envFlag if not empty
// 2. Use first existing environment
// 3. Create default environment with user's name
func GetOrCreateDefaultEnv(apiClient *client.Client, envFlag string, nonInteractive bool) (string, error) {
	// Check flags
	if envFlag != "" {
		return envFlag, nil
	}

	// List existing envs
	envs, err := apiClient.ListEnvs()
	if err != nil {
		return "", fmt.Errorf("failed to list environments: %w", err)
	}

	// Use existing env
	if len(envs) > 0 {
		if nonInteractive {
			// Use first env in non-interactive mode
			return envs[0].Name, nil
		}
		// Interactive env selection
		selectedEnv, err := interactive.SelectEnv(envs)
		if err != nil {
			return "", fmt.Errorf("environment selection cancelled: %w", err)
		}
		return selectedEnv.Name, nil
	}

	// No envs exist, create default
	user, err := apiClient.GetCurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	fmt.Printf("Creating default environment: %s\n", user.Name)
	_, err = apiClient.CreateEnv(user.Name)
	if err != nil {
		return "", fmt.Errorf("failed to create environment: %w", err)
	}

	return user.Name, nil
}



