package secret

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

var (
	createScope      string
	createEnv        string
	createRepository string
	createYes        bool
)

var createCmd = &cobra.Command{
	Use:   "create KEY=value [KEY=value...]",
	Short: "Create or update secret config",
	Long: `Create a new secret config or add keys to an existing one.

If a config already exists for the same scope/env, you'll be prompted to confirm
before merging keys (since secrets are write-only and we can't detect conflicts).
Use --yes to skip confirmation.

Examples:
  # Create env-scoped secrets (uses current env)
  lissto secret create KEY1=value1 KEY2=value2

  # Add more keys to the same env (prompts for confirmation)
  lissto secret create KEY3=value3

  # Skip confirmation prompt
  lissto secret create KEY3=value3 --yes

  # Update an existing key (overwrites with confirmation)
  lissto secret create KEY1=newvalue

  # Create with explicit env
  lissto secret create KEY1=value1 --env production

  # Create repo-scoped secrets
  lissto secret create KEY=value --scope repo --repository github.com/org/app

  # Create global secrets (admin only)
  lissto secret create KEY=value --scope global
`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&createScope, "scope", "s", "", "Scope: env, repo, or global (default: env)")
	createCmd.Flags().StringVarP(&createEnv, "env", "e", "", "Environment name (default: current env)")
	createCmd.Flags().StringVarP(&createRepository, "repository", "r", "", "Repository (required for scope=repo)")
	createCmd.Flags().BoolVarP(&createYes, "yes", "y", false, "Skip confirmation prompt for overwriting existing secrets")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Default scope to "env"
	scope := createScope
	if scope == "" {
		scope = "env"
	}

	// Default env to current env from config
	env := createEnv
	if scope == "env" && env == "" {
		env = cmdutil.GetCurrentEnv()
		if env == "" {
			return fmt.Errorf("env is required for scope=env. Set with --env or run 'lissto env use <env>'")
		}
	}

	// Parse KEY=value arguments
	secrets, err := cmdutil.ParseKeyValueArgs(args)
	if err != nil {
		return err
	}

	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	// Generate name based on scope
	name := cmdutil.GenerateResourceName(scope, env, createRepository)

	// Try to create
	req := &client.CreateSecretRequest{
		Name:       name,
		Scope:      scope,
		Env:        env,
		Repository: createRepository,
		Secrets:    secrets,
	}

	secret, err := apiClient.CreateSecret(req)
	if err != nil {
		// Check if it's a conflict error (409 or "already exists")
		if strings.Contains(err.Error(), "409") || strings.Contains(strings.ToLower(err.Error()), "already exists") {
			// Secret exists - get it to check for key overlaps
			existing, err := apiClient.GetSecret(name, scope, env, createRepository)
			if err != nil {
				return fmt.Errorf("failed to get existing secret: %w", err)
			}

			// Check if any keys overlap (already exist)
			overlapping := []string{}
			for newKey := range secrets {
				for _, existingKey := range existing.Keys {
					if newKey == existingKey {
						overlapping = append(overlapping, newKey)
						break
					}
				}
			}

			// Only prompt if there are overlapping keys (potential overwrites)
			if len(overlapping) > 0 {
				fmt.Printf("⚠️  Secret '%s' already exists.\n", name)
				fmt.Printf("Existing keys: %v\n", existing.Keys)
				fmt.Printf("New keys to add: %v\n", cmdutil.GetKeysFromMap(secrets))
				fmt.Printf("Keys that will be OVERWRITTEN: %v\n", overlapping)
				fmt.Println("\n⚠️  WARNING: Since secrets are write-only, we cannot detect if values differ.")

				// Ask for confirmation unless --yes flag is set
				if !createYes {
					fmt.Print("\nDo you want to proceed? (yes/no): ")
					reader := bufio.NewReader(os.Stdin)
					response, err := reader.ReadString('\n')
					if err != nil {
						return fmt.Errorf("failed to read confirmation: %w", err)
					}
					response = strings.TrimSpace(strings.ToLower(response))

					if response != "yes" && response != "y" {
						return fmt.Errorf("operation cancelled by user")
					}
				}
			} else {
				// No overlapping keys - safe to add without prompting
				fmt.Printf("Secret '%s' already exists, adding %d new keys...\n", name, len(secrets))
			}

			// Use set/update endpoint to add new keys (API merges automatically)
			setReq := &client.SetSecretRequest{
				Secrets: secrets,
			}
			secret, err = apiClient.UpdateSecret(name, scope, env, createRepository, setReq)
			if err != nil {
				return fmt.Errorf("failed to add keys to secret: %w", err)
			}

			fmt.Printf("\n✅ Secret '%s' updated with new keys\n", secret.Name)
			fmt.Printf("ID: %s\n", secret.ID)
			fmt.Printf("Scope: %s\n", secret.Scope)
			if secret.Env != "" {
				fmt.Printf("Env: %s\n", secret.Env)
			}
			fmt.Printf("Total keys: %d\n", len(secret.Keys))
			return nil
		}
		return fmt.Errorf("failed to create secret: %w", err)
	}

	// Success - created new
	fmt.Printf("✅ Secret '%s' created successfully\n", secret.Name)
	fmt.Printf("ID: %s\n", secret.ID)
	fmt.Printf("Scope: %s\n", secret.Scope)
	if secret.Env != "" {
		fmt.Printf("Env: %s\n", secret.Env)
	}
	fmt.Printf("Keys: %d\n", len(secret.Keys))
	return nil
}
