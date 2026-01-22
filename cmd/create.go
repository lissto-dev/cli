package cmd

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/interactive"
	"github.com/lissto-dev/cli/pkg/output"
	controllerconfig "github.com/lissto-dev/controller/pkg/config"
	"github.com/spf13/cobra"
)

var (
	createBlueprint      string
	createBranch         string
	createTag            string
	createCommit         string
	createEnv            string
	createNonInteractive bool
)

// StackCreateResult represents the JSON output for stack create command
type StackCreateResult struct {
	StackID     string         `json:"stack_id"`
	BlueprintID string         `json:"blueprint_id"`
	Environment string         `json:"environment"`
	Exposed     []ExposedEntry `json:"exposed,omitempty"`
}

// ExposedEntry represents an exposed service URL
type ExposedEntry struct {
	Service string `json:"service"`
	URL     string `json:"url"`
}

// createCmd represents the unified create command (parent)
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create blueprints or stacks (intelligent wizard)",
	Long: `Create blueprints or stacks with an intelligent wizard.

When run without subcommands, automatically detects your situation:
- If no blueprints exist, starts blueprint creation wizard
- If blueprints exist, shows unified selector to deploy or create new

Subcommands:
  stack      - Explicitly create a stack
  blueprint  - Explicitly create a blueprint

Examples:
  # Intelligent wizard mode
  lissto create

  # Explicit stack creation
  lissto create stack --blueprint my-blueprint

  # Explicit blueprint creation
  lissto create blueprint`,
	RunE: runCreateRouter,
}

// createStackCmd represents the explicit stack creation subcommand
var createStackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Create a new stack (interactive)",
	Long: `Create a new stack with an interactive workflow.

The create command guides you through:
1. Selecting a blueprint (or use --blueprint flag)
2. Preparing and previewing images
3. Optionally customizing branch/tag/commit
4. Deploying the stack

Examples:
  # Interactive mode - select blueprint and confirm deployment
  lissto create stack

  # Specify blueprint, interactive for the rest
  lissto create stack --blueprint my-blueprint

  # Non-interactive with all parameters
  lissto create stack --blueprint my-blueprint --env dev --branch main

  # Specify branch/tag/commit
  lissto create stack --blueprint my-blueprint --branch develop
  lissto create stack --blueprint my-blueprint --tag v1.2.3
  lissto create stack --blueprint my-blueprint --commit abc123

  # Output in different formats
  lissto create stack --blueprint my-blueprint --output json`,
	RunE: runCreateStack,
}

// createBlueprintCmd represents the explicit blueprint creation subcommand
var createBlueprintCmd = &cobra.Command{
	Use:   "blueprint",
	Short: "Create a new blueprint (wizard)",
	Long: `Create a new blueprint with an interactive wizard.

The wizard will:
1. Auto-detect compose files in current directory
2. Detect git repository
3. Check for existing blueprints from same repository
4. Create or override blueprint safely

Examples:
  # Auto-detect compose file
  lissto create blueprint

  # The power-user command 'lissto blueprint create <file>' is still available`,
	RunE: runCreateBlueprintWizard,
}

func init() {
	// Add subcommands
	createCmd.AddCommand(createStackCmd)
	createCmd.AddCommand(createBlueprintCmd)

	// Move flags to stack subcommand
	createStackCmd.Flags().StringVar(&createBlueprint, "blueprint", "", "Blueprint to deploy")
	createStackCmd.Flags().StringVar(&createBranch, "branch", "", "Git branch to use for image resolution")
	createStackCmd.Flags().StringVar(&createTag, "tag", "", "Git tag to use for image resolution")
	createStackCmd.Flags().StringVar(&createCommit, "commit", "", "Git commit hash to use for image resolution")
	createStackCmd.Flags().StringVar(&createEnv, "env", "", "Environment to deploy to")
	createStackCmd.Flags().BoolVar(&createNonInteractive, "non-interactive", false, "Run in non-interactive mode (fail if required info is missing)")
}

// runCreateRouter is the smart router for bare 'lissto create' command
func runCreateRouter(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current context
	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("no active context. Run 'lissto login' first: %w", err)
	}

	// Create API client
	fmt.Println("🔌 Connecting to Lissto API...")
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// List blueprints to determine routing
	fmt.Println("🔍 Checking for existing blueprints...")
	blueprints, err := apiClient.ListBlueprints(true)
	if err != nil {
		return fmt.Errorf("failed to list blueprints: %w", err)
	}

	// If no blueprints, start blueprint creation wizard
	if len(blueprints) == 0 {
		fmt.Println("✨ No blueprints found. Let's create your first blueprint!")
		return runCreateBlueprintWizard(cmd, args)
	}

	// Blueprints exist - show unified selector
	action, selectedBlueprint, err := interactive.SelectBlueprintOrCreate(blueprints)
	if err != nil {
		return fmt.Errorf("selection cancelled: %w", err)
	}

	if action == interactive.ActionCreateAdditional {
		// User wants to create new blueprint
		return runCreateBlueprintWizard(cmd, args)
	}

	// User selected a blueprint to deploy - set it and run stack creation
	createBlueprint = selectedBlueprint.ID
	return runCreateStack(cmd, args)
}

// runCreateBlueprintWizard handles the blueprint creation wizard flow
func runCreateBlueprintWizard(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current context
	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("no active context. Run 'lissto login' first: %w", err)
	}

	// Create API client
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Run the blueprint creation wizard
	createdBlueprint, err := blueprintWizardFlow(cmd, apiClient)
	if err != nil {
		return err
	}

	// Step 11: Prompt "What would you like to do next?"
	action, err := interactive.ConfirmNextAction()
	if err != nil {
		// User cancelled, that's okay - blueprint was created
		return nil
	}

	if action == interactive.ActionDeployThisBlueprint && createdBlueprint != nil {
		// Step 12: Set the blueprint and run stack deployment
		createBlueprint = createdBlueprint.ID
		fmt.Println()
		return runCreateStack(cmd, args)
	}

	return nil
}

func runCreateStack(cmd *cobra.Command, args []string) error {
	// Create output context for handling quiet mode (JSON/YAML output)
	out := cmdutil.NewOutputContext(cmd)

	// Load all environment overrides once
	overrides := cmdutil.LoadEnvOverrides()

	// Get API client (handles env var auth and config-based auth)
	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	// Track if blueprint was selected interactively (to show/hide Back button)
	blueprintWasInteractive := createBlueprint == ""

	// Step 1: Determine environment (once, outside blueprint loop)
	envToUse := createEnv
	if envToUse == "" {
		// Check if global --env flag is set
		envToUse = envName
	}

	if envToUse == "" {
		// In CI/CD mode with env vars, environment must be provided
		if overrides.IsCICDMode() {
			return fmt.Errorf("--env flag is required when using environment variable authentication")
		}

		// Try to get existing envs
		envs, err := apiClient.ListEnvs()
		if err != nil {
			return fmt.Errorf("failed to list environments: %w", err)
		}

		if len(envs) > 0 {
			if createNonInteractive {
				// Use first env in non-interactive mode
				envToUse = envs[0].Name
				out.Printf("Using environment: %s\n", envToUse)
			} else {
				// Interactive env selection
				selectedEnv, err := interactive.SelectEnv(envs)
				if err != nil {
					return fmt.Errorf("environment selection cancelled: %w", err)
				}
				envToUse = selectedEnv.Name
			}
		} else {
			// No envs exist, create default
			user, err := apiClient.GetCurrentUser()
			if err != nil {
				return fmt.Errorf("failed to get current user: %w", err)
			}

			envToUse = user.Name
			out.Printf("Creating default environment: %s\n", envToUse)
			_, err = apiClient.CreateEnv(envToUse)
			if err != nil {
				return fmt.Errorf("failed to create environment: %w", err)
			}
		}
	}

	// Step 2: Blueprint selection loop (allows going back from preview)
	var selectedBlueprint *client.BlueprintResponse
blueprintLoop:
	for {
		if createBlueprint != "" {
			// Blueprint provided via flag, skip selection
			out.Printf("Using blueprint: %s\n", createBlueprint)
			bp, err := apiClient.GetBlueprint(createBlueprint)
			if err != nil {
				return fmt.Errorf("failed to get blueprint: %w", err)
			}
			selectedBlueprint = bp
		} else {
			// Interactive blueprint selection
			if createNonInteractive {
				return fmt.Errorf("--blueprint is required in non-interactive mode")
			}

			fmt.Println("\nFetching blueprints...")
			blueprints, err := apiClient.ListBlueprints(true) // Include global
			if err != nil {
				return fmt.Errorf("failed to list blueprints: %w", err)
			}

			if len(blueprints) == 0 {
				return fmt.Errorf("no blueprints available")
			}

			selectedBlueprint, err = interactive.SelectBlueprint(blueprints)
			if err != nil {
				return fmt.Errorf("blueprint selection cancelled: %w", err)
			}
		}

		// Validation: Check for duplicate stacks
		// Step 1: Check for exact blueprint match in current environment
		existingStacks, err := apiClient.ListStacks(envToUse)
		if err != nil {
			return fmt.Errorf("failed to list existing stacks: %w", err)
		}

		// Check for exact blueprint ID match
		for _, stack := range existingStacks {
			if stack.Spec.BlueprintReference == selectedBlueprint.ID {
				fmt.Printf("\n❌ Error: Stack with this blueprint already exists: %s\n", stack.Name)
				fmt.Printf("💡 Tip: Use 'lissto update' to update the stack with new images\n\n")
				return fmt.Errorf("stack '%s' already deployed with blueprint '%s'", stack.Name, selectedBlueprint.ID)
			}
		}

		// Step 2: Check for same repository blueprint (if repository info available)
		// Get detailed info to access repository annotation
		selectedBlueprintDetailed, err := apiClient.GetBlueprintDetailed(selectedBlueprint.ID)
		if err == nil && selectedBlueprintDetailed.Metadata.Annotations["lissto.dev/repository"] != "" && !createNonInteractive {
			selectedRepo := selectedBlueprintDetailed.Metadata.Annotations["lissto.dev/repository"]
			// Repository is already normalized in the annotation
			normalizedSelectedRepo := selectedRepo

			fmt.Printf("🔍 Checking for existing stacks from repository: %s\n", normalizedSelectedRepo)

			// Check existing stacks for same repository
			var matchingStacks []string
			for _, stack := range existingStacks {
				// Get blueprint details for each stack
				stackBlueprintID := stack.Spec.BlueprintReference
				if stackBlueprintID == selectedBlueprint.ID {
					continue // Already checked for exact match above
				}

				// Get the blueprint detailed to check its repository
				stackBlueprint, err := apiClient.GetBlueprintDetailed(stackBlueprintID)
				if err != nil {
					// Skip if we can't get blueprint details
					fmt.Printf("  ⚠️  Warning: Could not fetch blueprint %s: %v\n", stackBlueprintID, err)
					continue
				}

				// Extract repository from annotations
				if repo, ok := stackBlueprint.Metadata.Annotations["lissto.dev/repository"]; ok && repo != "" {
					normalizedStackRepo := controllerconfig.NormalizeRepositoryURL(repo)
					fmt.Printf("  📦 Stack %s uses repository: %s\n", stack.Name, normalizedStackRepo)
					if normalizedStackRepo == normalizedSelectedRepo {
						matchingStacks = append(matchingStacks, stack.Name)
					}
				} else {
					fmt.Printf("  ℹ️  Stack %s has no repository annotation (might be an old blueprint)\n", stack.Name)
				}
			}

			// If we found matching repositories, warn the user
			if len(matchingStacks) > 0 {
				fmt.Printf("\n⚠️  Warning: Found existing stack(s) from the same repository:\n")
				for _, stackName := range matchingStacks {
					fmt.Printf("  - %s\n", stackName)
				}
				fmt.Println()

				action, err := interactive.ConfirmDuplicateRepoAction()
				if err != nil {
					return fmt.Errorf("cancelled: %w", err)
				}

				switch action {
				case interactive.ActionUpdateExisting:
					// Suggest using lissto update command
					fmt.Println("\n💡 Please run 'lissto update' to update the existing stack")
					return fmt.Errorf("use 'lissto update' to update existing stacks")
				case interactive.ActionDeployAnyway:
					fmt.Println("\n⚠️  Proceeding with deployment (risky)...")
					// Continue with create flow
				case interactive.ActionCancel:
					return fmt.Errorf("deployment cancelled by user")
				}
			}
		}

		// Step 3: Prepare and preview loop
		var prepareResp *client.PrepareStackResponse
		for {
			// Prepare stack
			fmt.Println("\nPreparing stack...")
			var err error
			prepareResp, err = apiClient.PrepareStack(
				selectedBlueprint.ID,
				envToUse,
				createCommit,
				createBranch,
				createTag,
				true, // detailed
			)
			if err != nil {
				fmt.Printf("❌ Failed to prepare stack: %v\n", err)

				if createNonInteractive {
					return fmt.Errorf("failed to prepare stack: %w", err)
				}

				// Ask what user wants to do
				var action string
				var retryErr error
				if blueprintWasInteractive {
					action, retryErr = interactive.ConfirmRetryWithBack()
				} else {
					action, retryErr = interactive.ConfirmRetry()
				}
				if retryErr != nil {
					return fmt.Errorf("failed to prepare stack: %w", err)
				}

				switch action {
				case interactive.ActionTryAnotherBranchTag:
					// Get new branch/tag/commit
					branch, tag, commit, promptErr := interactive.PromptBranchTag()
					if promptErr != nil {
						return fmt.Errorf("cancelled: %w", promptErr)
					}

					// Update for next iteration
					createBranch = branch
					createTag = tag
					createCommit = commit
					continue
				case interactive.ActionBackToBlueprint:
					// Reset branch/tag/commit for fresh start
					createBranch = ""
					createTag = ""
					createCommit = ""
					continue blueprintLoop
				case interactive.ActionCancel:
					return fmt.Errorf("failed to prepare stack: %w", err)
				}
			}

			// Display preview
			format := outputFormat
			if format == "" {
				format = outputFormatTable
			}

			if format == outputFormatTable {
				output.PrintImagePreview(os.Stdout, prepareResp.Images, prepareResp.Exposed)
			} else {
				err = output.PrintImagePreviewWithFormat(format, prepareResp)
				if err != nil {
					return fmt.Errorf("failed to print preview: %w", err)
				}
			}

			// Check for missing images
			if output.HasMissingImages(prepareResp.Images) {
				fmt.Println("❌ Cannot deploy: Some services have missing images.")

				if createNonInteractive {
					return fmt.Errorf("deployment blocked: missing images")
				}

				// Ask what user wants to do
				var action string
				if blueprintWasInteractive {
					action, err = interactive.ConfirmRetryWithBack()
				} else {
					action, err = interactive.ConfirmRetry()
				}
				if err != nil {
					return fmt.Errorf("deployment cancelled: missing images")
				}

				switch action {
				case interactive.ActionTryAnotherBranchTag:
					// Get new branch/tag/commit
					branch, tag, commit, err := interactive.PromptBranchTag()
					if err != nil {
						return fmt.Errorf("cancelled: %w", err)
					}

					// Update for next iteration
					createBranch = branch
					createTag = tag
					createCommit = commit
					continue
				case interactive.ActionBackToBlueprint:
					// Reset branch/tag/commit for fresh start
					createBranch = ""
					createTag = ""
					createCommit = ""
					continue blueprintLoop
				case interactive.ActionCancel:
					return fmt.Errorf("deployment cancelled: missing images")
				}
			}

			// Step 4: Confirm deployment or modify
			if createNonInteractive {
				// Non-interactive mode, proceed directly
				break
			}

			// Use appropriate confirmation based on whether blueprint was interactive
			var action string
			if blueprintWasInteractive {
				action, err = interactive.ConfirmDeploymentWithBack()
			} else {
				action, err = interactive.ConfirmDeployment()
			}
			if err != nil {
				return fmt.Errorf("cancelled: %w", err)
			}

			switch action {
			case interactive.ActionDeploy:
				// Proceed to deployment - exit the loop
			case interactive.ActionTryAnotherBranchTag:
				// Get new branch/tag/commit
				branch, tag, commit, err := interactive.PromptBranchTag()
				if err != nil {
					return fmt.Errorf("cancelled: %w", err)
				}

				// Update for next iteration
				createBranch = branch
				createTag = tag
				createCommit = commit
				continue
			case interactive.ActionBackToBlueprint:
				// Reset branch/tag/commit for fresh start
				createBranch = ""
				createTag = ""
				createCommit = ""
				continue blueprintLoop
			case interactive.ActionCancel:
				return fmt.Errorf("deployment cancelled by user")
			}

			// Break out of loop after successful confirmation
			break
		}

		// Step 5: Create stack
		out.Println("\nCreating stack...")
		stackID, err := apiClient.CreateStack(selectedBlueprint.ID, envToUse, prepareResp.RequestID)
		if err != nil {
			return fmt.Errorf("failed to create stack: %w", err)
		}

		// Prepare result for structured output
		result := StackCreateResult{
			StackID:     stackID,
			BlueprintID: selectedBlueprint.ID,
			Environment: envToUse,
		}

		// Add exposed URLs
		for _, exp := range prepareResp.Exposed {
			result.Exposed = append(result.Exposed, ExposedEntry{
				Service: exp.Service,
				URL:     "https://" + exp.URL,
			})
		}

		// Use unified output pattern: JSON/YAML for structured, custom for human-readable
		if err := out.PrintResult(result, func() {
			fmt.Printf("✅ Stack created successfully!\n")
			fmt.Printf("Stack ID: %s\n", stackID)

			// Show exposed URLs if any
			if len(prepareResp.Exposed) > 0 {
				fmt.Println("\n🔗 Exposed services:")
				for _, exp := range prepareResp.Exposed {
					fmt.Printf("  - %s: https://%s\n", exp.Service, exp.URL)
				}
			}
		}); err != nil {
			return err
		}

		// Successfully created stack, break out of blueprint loop
		break blueprintLoop
	}

	return nil
}
