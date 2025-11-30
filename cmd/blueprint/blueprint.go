package blueprint

import (
	"github.com/spf13/cobra"
)

// BlueprintCmd represents the blueprint command
var BlueprintCmd = &cobra.Command{
	Use:   "blueprint",
	Short: "Manage blueprints",
	Long:  `Manage Lissto blueprints. Blueprints are user-specific or global.`,
}

func init() {
	BlueprintCmd.AddCommand(listCmd)
	BlueprintCmd.AddCommand(getCmd)
	BlueprintCmd.AddCommand(createCmd)
	BlueprintCmd.AddCommand(deleteCmd)
}
