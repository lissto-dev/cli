package cmd

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/output"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `Manage Lissto CLI configuration settings.`,
}

// configGetCmd gets a configuration value
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value.

Available keys:
  settings.update-check  Whether automatic update checks are enabled (true/false)`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  settings.update-check  Set to 'true' to enable automatic update checks, 'false' to disable

Examples:
  lissto config set settings.update-check true
  lissto config set settings.update-check false`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

// configListCmd lists all configuration values
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE:  runConfigList,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "settings.update-check":
		fmt.Printf("%t\n", cfg.Settings.UpdateCheck)
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "settings.update-check":
		switch value {
		case "true":
			cfg.Settings.UpdateCheck = true
		case "false":
			cfg.Settings.UpdateCheck = false
		default:
			return fmt.Errorf("invalid value for settings.update-check: %s (use 'true' or 'false')", value)
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s to %s\n", key, value)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch outputFormat {
	case outputFormatJSON:
		return output.PrintJSON(os.Stdout, cfg.Settings)
	case outputFormatYAML:
		return output.PrintYAML(os.Stdout, cfg.Settings)
	}

	// Table format
	headers := []string{"KEY", "VALUE"}
	rows := [][]string{
		{"settings.update-check", fmt.Sprintf("%t", cfg.Settings.UpdateCheck)},
	}
	output.PrintTable(os.Stdout, headers, rows)

	return nil
}
