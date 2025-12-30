package cmd

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/interactive"
	"github.com/lissto-dev/cli/pkg/output"
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

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new stack (interactive)",
	Long: `Create a new stack with an interactive workflow.

The create command guides you through:
1. Selecting a blueprint (or use --blueprint flag)
2. Preparing and previewing images
3. Optionally customizing branch/tag/commit
4. Deploying the stack

Examples:
  # Interactive mode - select blueprint and confirm deployment
  lissto create

  # Specify blueprint, interactive for the rest
  lissto create --blueprint my-blueprint

  # Non-interactive with all parameters
  lissto create --blueprint my-blueprint --env dev --branch main

  # Specify branch/tag/commit
  lissto create --blueprint my-blueprint --branch develop
  lissto create --blueprint my-blueprint --tag v1.2.3
  lissto create --blueprint my-blueprint --commit abc123

  # Output in different formats
  lissto create --blueprint my-blueprint --output json`,
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVar(&createBlueprint, "blueprint", "", "Blueprint to deploy")
	createCmd.Flags().StringVar(&createBranch, "branch", "", "Git branch to use for image resolution")
	createCmd.Flags().StringVar(&createTag, "tag", "", "Git tag to use for image resolution")
	createCmd.Flags().StringVar(&createCommit, "commit", "", "Git commit hash to use for image resolution")
	createCmd.Flags().StringVar(&createEnv, "env", "", "Environment to deploy to")
	createCmd.Flags().BoolVar(&createNonInteractive, "non-interactive", false, "Run in non-interactive mode (fail if required info is missing)")
}

func runCreate(cmd *cobra.Command, args []string) error {
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

	// Create API client with k8s discovery and validation
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
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
		// Try to get existing envs
		envs, err := apiClient.ListEnvs()
		if err != nil {
			return fmt.Errorf("failed to list environments: %w", err)
		}

		if len(envs) > 0 {
			if createNonInteractive {
				// Use first env in non-interactive mode
				envToUse = envs[0].Name
				fmt.Printf("Using environment: %s\n", envToUse)
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
			fmt.Printf("Creating default environment: %s\n", envToUse)
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
			fmt.Printf("Using blueprint: %s\n", createBlueprint)
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
				case "Try another branch/tag":
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
		fmt.Println("\nCreating stack...")
		stackID, err := apiClient.CreateStack(selectedBlueprint.ID, envToUse, prepareResp.RequestID)
		if err != nil {
			return fmt.Errorf("failed to create stack: %w", err)
		}

		fmt.Printf("✅ Stack created successfully!\n")
		fmt.Printf("Stack ID: %s\n", stackID)

		// Show exposed URLs if any
		if len(prepareResp.Exposed) > 0 {
			fmt.Println("\n🔗 Exposed services:")
			for _, exp := range prepareResp.Exposed {
				fmt.Printf("  - %s: https://%s\n", exp.Service, exp.URL)
			}
		}

		// Successfully created stack, break out of blueprint loop
		break blueprintLoop
	}

	return nil
}
