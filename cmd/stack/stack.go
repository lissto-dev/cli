package stack

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/spf13/cobra"
)

// StackCmd represents the stack command
var StackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Manage stacks",
	Long:  `Manage Lissto stacks. Stacks are deployed instances of blueprints in environments.`,
}

// Helper to get API client and resolve environment
func getAPIClientAndEnv(cmd *cobra.Command) (*client.Client, string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config: %w", err)
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return nil, "", fmt.Errorf("no context selected. Run 'lissto login' first")
	}

	// Get environment (from flag or config)
	envName, _ := cmd.Flags().GetString("env")
	if envName == "" {
		envName = cfg.CurrentEnv
	}

	if envName == "" {
		return nil, "", fmt.Errorf("no environment selected. Use --env flag or 'lissto env use <name>'")
	}

	// Create API client with k8s discovery and validation
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize API client: %w", err)
	}

	return apiClient, envName, nil
}

func getOutputFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("output")
	return format
}

func init() {
	StackCmd.AddCommand(listCmd)
	StackCmd.AddCommand(getCmd)
	StackCmd.AddCommand(createCmd)
	StackCmd.AddCommand(deleteCmd)
}
