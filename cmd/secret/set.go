package secret

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var setSecrets []string

var setCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set/update secret values",
	Long: `Set or update secret values for an existing secret config.

This merges new values with existing ones (doesn't remove existing keys).

Examples:
  # Set new secret values
  lissto secret set my-secrets --secret KEY1=newvalue1 --secret KEY2=newvalue2
`,
	Args: cobra.ExactArgs(1),
	RunE: runSet,
}

func init() {
	setCmd.Flags().StringArrayVarP(&setSecrets, "secret", "k", []string{}, "Secret in KEY=value format (can be repeated)")
	setCmd.MarkFlagRequired("secret")
}

func runSet(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Parse secrets
	secrets, err := cmdutil.ParseKeyValueArgs(setSecrets)
	if err != nil {
		return err
	}

	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	req := &client.SetSecretRequest{
		Secrets: secrets,
	}

	// Use default scope (env) - TODO: add scope flags
	secret, err := apiClient.UpdateSecret(name, "", "", "", req)
	if err != nil {
		return fmt.Errorf("failed to set secrets: %w", err)
	}

	fmt.Printf("Secret '%s' updated successfully\n", secret.Name)
	fmt.Printf("Keys: %d\n", len(secret.Keys))

	return nil
}
