package variable

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a variable config",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	// Use default scope (env) - TODO: add scope flags
	if err := apiClient.DeleteVariable(name, "", "", ""); err != nil {
		return fmt.Errorf("failed to delete variable: %w", err)
	}

	fmt.Printf("Variable '%s' deleted successfully\n", name)
	return nil
}
