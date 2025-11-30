package variable

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var updateData []string

var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update a variable config",
	Long: `Update an existing variable config.

Examples:
  # Update variable data (replaces all data)
  lissto variable update my-vars --data KEY1=newvalue1 --data KEY2=newvalue2
`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().StringArrayVarP(&updateData, "data", "d", []string{}, "Data in KEY=value format (can be repeated)")
	updateCmd.MarkFlagRequired("data")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Parse data
	data, err := cmdutil.ParseKeyValueArgs(updateData)
	if err != nil {
		return err
	}

	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	req := &client.UpdateVariableRequest{
		Data: data,
	}

	// Use default scope (env) - TODO: add scope flags
	variable, err := apiClient.UpdateVariable(name, "", "", "", req)
	if err != nil {
		return fmt.Errorf("failed to update variable: %w", err)
	}

	fmt.Printf("Variable '%s' updated successfully\n", variable.Name)
	fmt.Printf("Keys: %d\n", len(variable.Data))

	return nil
}
