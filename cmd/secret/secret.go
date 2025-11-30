package secret

import (
	"github.com/spf13/cobra"
)

// SecretCmd represents the secret command
var SecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets",
	Long:  `Manage Lissto secrets. Secrets can be scoped to env, repo, or global. Values are write-only.`,
}

func init() {
	SecretCmd.AddCommand(listCmd)
	SecretCmd.AddCommand(getCmd)
	SecretCmd.AddCommand(createCmd)
	SecretCmd.AddCommand(setCmd)
	SecretCmd.AddCommand(deleteCmd)
}
