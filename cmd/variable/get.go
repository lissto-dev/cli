package variable

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var (
	getScope      string
	getEnv        string
	getRepository string
)

var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a specific variable",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func init() {
	getCmd.Flags().StringVar(&getScope, "scope", "env", "Scope: env, repo, or global")
	getCmd.Flags().StringVar(&getEnv, "env", "", "Environment name (defaults to current env for scope=env)")
	getCmd.Flags().StringVar(&getRepository, "repository", "", "Repository for scope=repo")
}

func runGet(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Default env for scope=env
	env := getEnv
	if getScope == "env" && env == "" {
		env = cmdutil.GetCurrentEnv()
	}

	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	variable, err := apiClient.GetVariable(name, getScope, env, getRepository)
	if err != nil {
		return fmt.Errorf("failed to get variable: %w", err)
	}

	return cmdutil.PrintOutput(cmd, variable, func() {
		fmt.Printf("Name:       %s\n", variable.Name)
		fmt.Printf("Scope:      %s\n", variable.Scope)
		if variable.Env != "" {
			fmt.Printf("Env:        %s\n", variable.Env)
		}
		if variable.Repository != "" {
			fmt.Printf("Repository: %s\n", variable.Repository)
		}
		fmt.Println("Data:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for k, v := range variable.Data {
			_, _ = fmt.Fprintf(w, "  %s\t= %s\n", k, v)
		}
		_ = w.Flush()
	})
}
