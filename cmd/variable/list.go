package variable

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/lissto-dev/cli/pkg/cmdutil"
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
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSCOPE\tENV\tREPOSITORY\tKEYS")
		for _, v := range variables {
			keyCount := len(v.Data)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", v.Name, v.Scope, v.Env, v.Repository, keyCount)
		}
		w.Flush()
	})
}
