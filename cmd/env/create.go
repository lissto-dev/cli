package env

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <env-name>",
	Short: "Create a new environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
	envName := args[0]

	apiClient, err := getAPIClient()
	if err != nil {
		return err
	}

	identifier, err := apiClient.CreateEnv(envName)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	fmt.Printf("Environment '%s' created successfully\n", envName)
	fmt.Printf("ID: %s\n", identifier)

	return nil
}

