package stack

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/pkg/output"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <stack-name>",
	Short: "Get stack details",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	stackName := args[0]

	apiClient, envName, err := getAPIClientAndEnv(cmd)
	if err != nil {
		return err
	}

	identifier, err := apiClient.GetStack(stackName, envName)
	if err != nil {
		return fmt.Errorf("failed to get stack: %w", err)
	}

	format := getOutputFormat(cmd)
	if format == "json" {
		return output.PrintJSON(os.Stdout, map[string]string{"id": identifier})
	} else if format == "yaml" {
		return output.PrintYAML(os.Stdout, map[string]string{"id": identifier})
	}

	// Human-readable format
	fmt.Printf("Stack ID: %s\n", identifier)

	return nil
}

