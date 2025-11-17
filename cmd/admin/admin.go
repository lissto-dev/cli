package admin

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/spf13/cobra"
)

// AdminCmd represents the admin command
var AdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin commands",
	Long:  `Admin-only commands for managing Lissto resources and API keys.`,
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

func init() {
	AdminCmd.AddCommand(apikeyCmd)
}
