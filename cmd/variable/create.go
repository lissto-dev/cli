package variable

import (
	"fmt"
	"strings"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Scope constants
const scopeEnv = "env"

var (
	createScope      string
	createEnv        string
	createRepository string
)

var createCmd = &cobra.Command{
	Use:   "create KEY=value [KEY=value...]",
	Short: "Create or update variable config",
	Long: `Create a new variable config or merge keys into an existing one.

If a config already exists for the same scope/env, new keys are merged in.
Rejects only if keys conflict (same key with different value).

Examples:
  # Create env-scoped variables (uses current env)
  lissto variable create KEY1=value1 KEY2=value2

  # Add more keys to the same env (merges)
  lissto variable create KEY3=value3

  # Create with explicit env
  lissto variable create KEY1=value1 --env production

  # Create repo-scoped variables
  lissto variable create KEY=value --scope repo --repository github.com/org/app

  # Create global variables (admin only)
  lissto variable create KEY=value --scope global
`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&createScope, "scope", "s", "", "Scope: env, repo, or global (default: env)")
	createCmd.Flags().StringVarP(&createEnv, "env", "e", "", "Environment name (default: current env)")
	createCmd.Flags().StringVarP(&createRepository, "repository", "r", "", "Repository (required for scope=repo)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Default scope to "env"
	scope := createScope
	if scope == "" {
		scope = scopeEnv
	}

	// Default env to current env from config
	env := createEnv
	if scope == scopeEnv && env == "" {
		env = cmdutil.GetCurrentEnv()
		if env == "" {
			return fmt.Errorf("env is required for scope=env. Set with --env or run 'lissto env use <env>'")
		}
	}

	// Parse KEY=value arguments
	data, err := cmdutil.ParseKeyValueArgs(args)
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
	req := &client.CreateVariableRequest{
		Name:       name,
		Scope:      scope,
		Env:        env,
		Repository: createRepository,
		Data:       data,
	}

	variable, err := apiClient.CreateVariable(req)
	if err != nil {
		// Check if it's a conflict error (409 or "already exists")
		if strings.Contains(err.Error(), "409") || strings.Contains(strings.ToLower(err.Error()), "already exists") {
			// Variable exists - try to merge keys
			fmt.Printf("Variable '%s' already exists, merging keys...\n", name)

			// Get existing variable (pass scope for correct namespace resolution)
			existing, err := apiClient.GetVariable(name, scope, env, createRepository)
			if err != nil {
				return fmt.Errorf("failed to get existing variable: %w", err)
			}

			// Check for conflicting keys (same key, different value)
			conflicts := []string{}
			for key, newValue := range data {
				if existingValue, exists := existing.Data[key]; exists && existingValue != newValue {
					conflicts = append(conflicts, fmt.Sprintf("%s (existing: %s, new: %s)", key, existingValue, newValue))
				}
			}

			if len(conflicts) > 0 {
				return fmt.Errorf("key conflicts detected:\n  %s\n\nUse 'lissto variable update %s' to overwrite",
					strings.Join(conflicts, "\n  "), name)
			}

			// Merge: combine existing + new data
			mergedData := make(map[string]string)
			for k, v := range existing.Data {
				mergedData[k] = v
			}
			for k, v := range data {
				mergedData[k] = v
			}

			// Update with merged data
			updateReq := &client.UpdateVariableRequest{
				Data: mergedData,
			}
			variable, err = apiClient.UpdateVariable(name, scope, env, createRepository, updateReq)
			if err != nil {
				return fmt.Errorf("failed to merge variable: %w", err)
			}

			fmt.Printf("✅ Variable '%s' updated with new keys\n", variable.Name)
			fmt.Printf("ID: %s\n", variable.ID)
			fmt.Printf("Scope: %s\n", variable.Scope)
			if variable.Env != "" {
				fmt.Printf("Env: %s\n", variable.Env)
			}
			fmt.Printf("Keys: %d (added %d)\n", len(variable.Data), len(data))
			return nil
		}
		return fmt.Errorf("failed to create variable: %w", err)
	}

	// Success - created new
	fmt.Printf("✅ Variable '%s' created successfully\n", variable.Name)
	fmt.Printf("ID: %s\n", variable.ID)
	fmt.Printf("Scope: %s\n", variable.Scope)
	if variable.Env != "" {
		fmt.Printf("Env: %s\n", variable.Env)
	}
	fmt.Printf("Keys: %d\n", len(variable.Data))
	return nil
}
