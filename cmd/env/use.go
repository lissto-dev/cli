package env

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/config"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <env-name>",
	Short: "Set the active environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runUse,
}

func runUse(cmd *cobra.Command, args []string) error {
	envName := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetCurrentEnv(envName); err != nil {
		return err
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Switched to environment: %s\n", envName)

	return nil
}
