package env

import (
	"github.com/spf13/cobra"
)

// EnvCmd represents the env command
var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environments",
	Long:  `Manage Lissto environments. Environments provide isolated namespaces for your resources.`,
}

func init() {
	EnvCmd.AddCommand(listCmd)
	EnvCmd.AddCommand(getCmd)
	EnvCmd.AddCommand(createCmd)
	EnvCmd.AddCommand(useCmd)
	EnvCmd.AddCommand(currentCmd)
}
