package admin

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var (
	apikeyName string
	apikeyRole string
)

// apikeyCmd represents the apikey command
var apikeyCmd = &cobra.Command{
	Use:   "apikey create",
	Short: "Create a new API key (admin only)",
	Long:  `Create a new API key for a user. Requires admin privileges.`,
	RunE:  runCreateAPIKey,
}

func init() {
	apikeyCmd.Flags().StringVar(&apikeyName, "name", "", "User name for the API key (required)")
	apikeyCmd.Flags().StringVar(&apikeyRole, "role", "user", "Role for the API key (user, deploy)")
	_ = apikeyCmd.MarkFlagRequired("name")
}

func runCreateAPIKey(cmd *cobra.Command, args []string) error {
	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	req := client.CreateAPIKeyRequest{
		Name: apikeyName,
		Role: apikeyRole,
	}

	result, err := apiClient.CreateAPIKey(req)
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	fmt.Printf("API key created successfully\n")
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Role: %s\n", result.Role)
	fmt.Printf("API Key: %s\n", result.APIKey)
	fmt.Println("\nIMPORTANT: Save this API key securely. It cannot be retrieved later.")

	return nil
}


