package blueprint

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/spf13/cobra"
)

// BlueprintCmd represents the blueprint command
var BlueprintCmd = &cobra.Command{
	Use:   "blueprint",
	Short: "Manage blueprints",
	Long:  `Manage Lissto blueprints. Blueprints are user-specific or global.`,
}

// Helper to get API client
func getAPIClient() (*client.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("no context selected. Run 'lissto login' first")
	}

	// Create API client with k8s discovery and validation
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize API client: %w", err)
	}
	return apiClient, nil
}

func getOutputFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("output")
	return format
}

func init() {
	BlueprintCmd.AddCommand(listCmd)
	BlueprintCmd.AddCommand(getCmd)
	BlueprintCmd.AddCommand(createCmd)
	BlueprintCmd.AddCommand(deleteCmd)
}
