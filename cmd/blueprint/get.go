package blueprint

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <blueprint-name>",
	Short: "Get blueprint details",
	Long:  `Get details of a blueprint by name. Searches both user and global namespaces.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	blueprintName := args[0]

	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	blueprint, err := apiClient.GetBlueprint(blueprintName)
	if err != nil {
		return fmt.Errorf("failed to get blueprint: %w", err)
	}

	return cmdutil.PrintOutput(cmd, blueprint, func() {
		// Human-readable format
		fmt.Printf("ID: %s\n", blueprint.ID)
		fmt.Printf("Title: %s\n", blueprint.Title)

		if len(blueprint.Content.Services) > 0 {
			fmt.Printf("\nServices:\n")
			for _, service := range blueprint.Content.Services {
				fmt.Printf("  - %s\n", service)
			}
		}

		if len(blueprint.Content.Infra) > 0 {
			fmt.Printf("\nInfrastructure:\n")
			for _, infra := range blueprint.Content.Infra {
				fmt.Printf("  - %s\n", infra)
			}
		}
	})
}
