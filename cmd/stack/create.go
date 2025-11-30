package stack

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <blueprint-name>",
	Short: "Create a new stack from a blueprint",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
	blueprintName := args[0]

	apiClient, envName, err := cmdutil.GetAPIClientAndEnv(cmd)
	if err != nil {
		return err
	}

	// First, prepare the stack to get request_id
	fmt.Println("Preparing stack...")
	prepareResp, err := apiClient.PrepareStack(blueprintName, envName, "", "", "", true)
	if err != nil {
		return fmt.Errorf("failed to prepare stack: %w", err)
	}

	// Check for missing images
	hasMissingImages := false
	for _, img := range prepareResp.Images {
		if img.Digest == "" || img.Digest == "N/A" {
			hasMissingImages = true
			fmt.Printf("❌ Missing image for service: %s\n", img.Service)
		}
	}

	if hasMissingImages {
		return fmt.Errorf("cannot create stack: some services have missing images")
	}

	// Create stack with request_id
	fmt.Println("Creating stack...")
	identifier, err := apiClient.CreateStack(blueprintName, envName, prepareResp.RequestID)
	if err != nil {
		return fmt.Errorf("failed to create stack: %w", err)
	}

	fmt.Printf("✅ Stack created successfully\n")
	fmt.Printf("ID: %s\n", identifier)

	return nil
}
