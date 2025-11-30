package variable

import (
	"github.com/spf13/cobra"
)

// VariableCmd represents the variable command
var VariableCmd = &cobra.Command{
	Use:   "variable",
	Short: "Manage environment variables",
	Long:  `Manage Lissto environment variables. Variables can be scoped to env, repo, or global.`,
}

func init() {
	VariableCmd.AddCommand(listCmd)
	VariableCmd.AddCommand(getCmd)
	VariableCmd.AddCommand(createCmd)
	VariableCmd.AddCommand(updateCmd)
	VariableCmd.AddCommand(deleteCmd)
}
