package variable

import (
	"fmt"
	"os"
	"time"

	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/lissto-dev/cli/pkg/k8s"
	"github.com/lissto-dev/cli/pkg/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all variables",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	variables, err := apiClient.ListVariables()
	if err != nil {
		return fmt.Errorf("failed to list variables: %w", err)
	}

	if len(variables) == 0 {
		fmt.Println("No variables found")
		return nil
	}

	return cmdutil.PrintOutput(cmd, variables, func() {
		headers := []string{"NAME", "SCOPE", "ENV", "REPOSITORY", "KEY", "VALUE", "UPDATED"}
		var rows [][]string
		for _, v := range variables {
			keys := cmdutil.GetKeysFromMap(v.Data)

			if len(keys) == 0 {
				// If no keys, show one row with empty values
				rows = append(rows, []string{v.Name, v.Scope, v.Env, v.Repository, "", "", ""})
			} else {
				// Create a row for each key-value pair
				for _, key := range keys {
					// Get update time for this specific key
					updated := ""
					if v.KeyUpdatedAt != nil {
						if timestamp, ok := v.KeyUpdatedAt[key]; ok {
							updatedTime := time.Unix(timestamp, 0)
							updated = k8s.FormatAge(time.Since(updatedTime))
						}
					}

					rows = append(rows, []string{v.Name, v.Scope, v.Env, v.Repository, key, v.Data[key], updated})
				}
			}
		}
		output.PrintTable(os.Stdout, headers, rows)
	})
}
