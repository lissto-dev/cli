package env

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <env-name>",
	Short: "Get environment details",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	envName := args[0]

	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	env, err := apiClient.GetEnv(envName)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}

	return cmdutil.PrintOutput(cmd, env, func() {
		// Human-readable format
		fmt.Printf("Name: %s\n", env.Name)
		fmt.Printf("ID: %s\n", env.ID)
	})
}
