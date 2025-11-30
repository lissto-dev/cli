package admin

import (
	"github.com/spf13/cobra"
)

// AdminCmd represents the admin command
var AdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin commands",
	Long:  `Admin-only commands for managing Lissto resources and API keys.`,
}

func init() {
	AdminCmd.AddCommand(apikeyCmd)
}
