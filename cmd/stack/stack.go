package stack

import (
	"github.com/spf13/cobra"
)

// StackCmd represents the stack command
var StackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Manage stacks",
	Long:  `Manage Lissto stacks. Stacks are deployed instances of blueprints in environments.`,
}

func init() {
	StackCmd.AddCommand(listCmd)
	StackCmd.AddCommand(getCmd)
	StackCmd.AddCommand(createCmd)
	StackCmd.AddCommand(deleteCmd)
}
