package cmd

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/cmd/admin"
	"github.com/lissto-dev/cli/cmd/blueprint"
	"github.com/lissto-dev/cli/cmd/env"
	"github.com/lissto-dev/cli/cmd/secret"
	"github.com/lissto-dev/cli/cmd/stack"
	"github.com/lissto-dev/cli/cmd/variable"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	contextName  string
	envName      string
	showVersion  bool
)

// Version information (set via ldflags during build)
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "lissto",
	Short: "Lissto CLI - Manage your Lissto resources",
	Long: `Lissto CLI is a command-line tool for managing Lissto resources
including blueprints, stacks, and environments.`,
	SilenceUsage: true, // Don't show usage on errors
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("lissto version %s\n", Version)
			fmt.Printf("  commit: %s\n", Commit)
			fmt.Printf("  built at: %s\n", Date)
			return
		}
		_ = cmd.Help()
	},
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
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")

	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(env.EnvCmd)
	rootCmd.AddCommand(blueprint.BlueprintCmd)
	rootCmd.AddCommand(stack.StackCmd)
	rootCmd.AddCommand(variable.VariableCmd)
	rootCmd.AddCommand(secret.SecretCmd)
	rootCmd.AddCommand(admin.AdminCmd)
}
