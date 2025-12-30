package env

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/config"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current environment",
	RunE:  runCurrent,
}

func runCurrent(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	currentEnv, err := cfg.GetCurrentEnv()
	if err != nil {
		return fmt.Errorf("no environment selected. Use 'lissto env use <name>' to select one")
	}

	fmt.Printf("Current environment: %s\n", currentEnv)

	return nil
}
