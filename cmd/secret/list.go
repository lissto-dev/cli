package secret

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/lissto-dev/cli/pkg/cmdutil"
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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSCOPE\tENV\tREPOSITORY\tKEYS")
		for _, s := range secrets {
			keys := strings.Join(s.Keys, ", ")
			if len(keys) > 40 {
				keys = keys[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.Name, s.Scope, s.Env, s.Repository, keys)
		}
		w.Flush()
	})
}
