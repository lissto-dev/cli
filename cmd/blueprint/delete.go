package blueprint

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <blueprint-name>",
	Short: "Delete a blueprint",
	Long:  `Delete a blueprint by name. Will search both user and global namespaces.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	blueprintName := args[0]

	apiClient, err := getAPIClient()
	if err != nil {
		return err
	}

	if err := apiClient.DeleteBlueprint(blueprintName); err != nil {
		return fmt.Errorf("failed to delete blueprint: %w", err)
	}

	fmt.Printf("Blueprint '%s' deleted successfully\n", blueprintName)

	return nil
}
