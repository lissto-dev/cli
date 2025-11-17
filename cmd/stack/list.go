package stack

import (
	"fmt"
	"os"
	"time"

	"github.com/lissto-dev/cli/pkg/k8s"
	"github.com/lissto-dev/cli/pkg/output"
	"github.com/lissto-dev/cli/pkg/types"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stacks",
	Long: `List all stacks.

Examples:
  # List stacks with default format
  lissto stack list

  # List stacks with wide format (shows blueprint ID)
  lissto stack list -o wide

  # List stacks in a specific environment
  lissto stack list --env dev`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	apiClient, envName, err := getAPIClientAndEnv(cmd)
	if err != nil {
		return err
	}

	stacks, err := apiClient.ListStacks(envName)
	if err != nil {
		return fmt.Errorf("failed to list stacks: %w", err)
	}

	format := getOutputFormat(cmd)
	if format == "json" {
		return output.PrintJSON(os.Stdout, stacks)
	} else if format == "yaml" {
		return output.PrintYAML(os.Stdout, stacks)
	}

	// Check if no stacks exist
	if len(stacks) == 0 {
		fmt.Println("No stacks found. Use 'lissto create' to create a new stack.")
		return nil
	}

	// Table format - check if wide format is requested
	isWide := format == "wide"
	var headers []string
	if isWide {
		headers = []string{"NAME", "ENV", "BLUEPRINT", "BLUEPRINT ID", "AGE"}
	} else {
		headers = []string{"NAME", "ENV", "BLUEPRINT", "AGE"}
	}

	var rows [][]string
	for _, stack := range stacks {
		// Calculate age using time.Since
		duration := time.Since(stack.CreationTimestamp.Time)
		age := k8s.FormatAge(duration)

		// Get blueprint title from annotations, fallback to blueprint reference
		blueprintTitle := types.GetBlueprintTitle(&stack)
		if blueprintTitle == "" {
			blueprintTitle = stack.Spec.BlueprintReference
		}

		// Get environment from spec
		env := stack.Spec.Env

		// Build row based on format
		var row []string
		if isWide {
			row = []string{stack.Name, env, blueprintTitle, stack.Spec.BlueprintReference, age}
		} else {
			row = []string{stack.Name, env, blueprintTitle, age}
		}
		rows = append(rows, row)
	}
	output.PrintTable(os.Stdout, headers, rows)

	return nil
}
