package stack

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <stack-name>",
	Short: "Get stack details",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	stackName := args[0]

	apiClient, envName, err := cmdutil.GetAPIClientAndEnv(cmd)
	if err != nil {
		return err
	}

	identifier, err := apiClient.GetStack(stackName, envName)
	if err != nil {
		return fmt.Errorf("failed to get stack: %w", err)
	}

	return cmdutil.PrintOutput(cmd, map[string]string{"id": identifier}, func() {
		// Human-readable format
		fmt.Printf("Stack ID: %s\n", identifier)
	})
}

