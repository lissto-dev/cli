package stack

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <stack-name>",
	Short: "Delete a stack",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	stackName := args[0]

	apiClient, envName, err := cmdutil.GetAPIClientAndEnv(cmd)
	if err != nil {
		return err
	}

	if err := apiClient.DeleteStack(stackName, envName); err != nil {
		return fmt.Errorf("failed to delete stack: %w", err)
	}

	fmt.Printf("Stack '%s' deleted successfully\n", stackName)

	return nil
}
