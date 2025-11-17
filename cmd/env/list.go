package env

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/pkg/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	apiClient, err := getAPIClient()
	if err != nil {
		return err
	}

	envs, err := apiClient.ListEnvs()
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	format := getOutputFormat(cmd)
	if format == "json" {
		return output.PrintJSON(os.Stdout, envs)
	} else if format == "yaml" {
		return output.PrintYAML(os.Stdout, envs)
	}

	// Table format
	headers := []string{"NAME", "ID"}
	var rows [][]string
	for _, env := range envs {
		rows = append(rows, []string{env.Name, env.ID})
	}
	output.PrintTable(os.Stdout, headers, rows)

	return nil
}

