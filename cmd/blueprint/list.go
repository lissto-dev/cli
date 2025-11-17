package blueprint

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/pkg/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all blueprints",
	Long:  `List all blueprints (both user and global).`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	apiClient, err := getAPIClient()
	if err != nil {
		return err
	}

	// Always include global blueprints (API returns both by default)
	blueprints, err := apiClient.ListBlueprints(true)
	if err != nil {
		return fmt.Errorf("failed to list blueprints: %w", err)
	}

	format := getOutputFormat(cmd)
	if format == "json" {
		return output.PrintJSON(os.Stdout, blueprints)
	} else if format == "yaml" {
		return output.PrintYAML(os.Stdout, blueprints)
	}

	// Table format
	headers := []string{"ID", "TITLE", "AGE"}
	var rows [][]string
	for _, bp := range blueprints {
		age := output.ExtractBlueprintAge(bp.ID)
		rows = append(rows, []string{bp.ID, bp.Title, age})
	}
	output.PrintTable(os.Stdout, headers, rows)

	return nil
}
