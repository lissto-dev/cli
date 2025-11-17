package cmd

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/cmd/admin"
	"github.com/lissto-dev/cli/cmd/blueprint"
	"github.com/lissto-dev/cli/cmd/env"
	"github.com/lissto-dev/cli/cmd/stack"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	contextName  string
	envName      string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "lissto",
	Short: "Lissto CLI - Manage your Lissto resources",
	Long: `Lissto CLI is a command-line tool for managing Lissto resources
including blueprints, stacks, and environments.`,
	SilenceUsage: true, // Don't show usage on errors
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (json, yaml, wide)")
	rootCmd.PersistentFlags().StringVar(&contextName, "context", "", "Override current context")
	rootCmd.PersistentFlags().StringVar(&envName, "env", "", "Override current environment")

	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(env.EnvCmd)
	rootCmd.AddCommand(blueprint.BlueprintCmd)
	rootCmd.AddCommand(stack.StackCmd)
	rootCmd.AddCommand(admin.AdminCmd)
}
