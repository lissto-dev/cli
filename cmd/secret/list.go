package secret

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
	Short: "List all secrets (keys only)",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	secrets, err := apiClient.ListSecrets()
	if err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}

	if len(secrets) == 0 {
		fmt.Println("No secrets found")
		return nil
	}

	return cmdutil.PrintOutput(cmd, secrets, func() {
		headers := []string{"NAME", "SCOPE", "ENV", "REPOSITORY", "KEY", "UPDATED"}
		var rows [][]string
		for _, s := range secrets {
			if len(s.Keys) == 0 {
				// If no keys, show one row with empty key
				rows = append(rows, []string{s.Name, s.Scope, s.Env, s.Repository, "", ""})
			} else {
				// Create a row for each key
				for _, key := range s.Keys {
					// Get update time for this specific key
					updated := ""
					if s.KeyUpdatedAt != nil {
						if timestamp, ok := s.KeyUpdatedAt[key]; ok {
							updatedTime := time.Unix(timestamp, 0)
							updated = k8s.FormatAge(time.Since(updatedTime))
						}
					}

					rows = append(rows, []string{s.Name, s.Scope, s.Env, s.Repository, key, updated})
				}
			}
		}
		output.PrintTable(os.Stdout, headers, rows)
	})
}
