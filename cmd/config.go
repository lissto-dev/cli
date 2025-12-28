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
  disable-update-check  Whether automatic update checks are disabled`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configSetCmd sets a configuration value
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  disable-update-check  Set to 'true' to disable automatic update checks, 'false' to enable

Examples:
  lissto config set disable-update-check true
  lissto config set disable-update-check false`,
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
	case "disable-update-check":
		fmt.Printf("%v\n", cfg.DisableUpdateCheck)
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
	case "disable-update-check":
		switch value {
		case "true", "1", "yes":
			cfg.DisableUpdateCheck = true
		case "false", "0", "no":
			cfg.DisableUpdateCheck = false
		default:
			return fmt.Errorf("invalid value for disable-update-check: %s (use 'true' or 'false')", value)
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

	configValues := map[string]interface{}{
		"disable-update-check": cfg.DisableUpdateCheck,
	}

	if outputFormat == "json" {
		return output.PrintJSON(os.Stdout, configValues)
	} else if outputFormat == "yaml" {
		return output.PrintYAML(os.Stdout, configValues)
	}

	// Table format
	headers := []string{"KEY", "VALUE"}
	rows := [][]string{
		{"disable-update-check", fmt.Sprintf("%v", cfg.DisableUpdateCheck)},
	}
	output.PrintTable(os.Stdout, headers, rows)

	return nil
}
