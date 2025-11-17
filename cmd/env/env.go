package env

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/spf13/cobra"
)

// EnvCmd represents the env command
var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environments",
	Long:  `Manage Lissto environments. Environments provide isolated namespaces for your resources.`,
}

// Helper to get API client from current context
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

// Helper to get output format from global flag
func getOutputFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("output")
	return format
}

func init() {
	EnvCmd.AddCommand(listCmd)
	EnvCmd.AddCommand(getCmd)
	EnvCmd.AddCommand(createCmd)
	EnvCmd.AddCommand(useCmd)
	EnvCmd.AddCommand(currentCmd)
}
