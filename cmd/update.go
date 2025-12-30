package cmd

import (
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/interactive"
	"github.com/lissto-dev/cli/pkg/k8s"
	"github.com/lissto-dev/cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	updateStack          string
	updateBranch         string
	updateCommit         string
	updateTag            string
	updateYes            bool
	updateNonInteractive bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a stack with new images",
	Long: `Interactively update a stack with new images from a branch, tag, or commit.

This command allows you to update an existing stack with new container images.
By default, it will guide you through an interactive process to:
  1. Select a stack (if not specified with --stack)
  2. Choose a branch/tag/commit (if not specified with flags)
  3. Preview the changes
  4. Confirm the update

Examples:
  # Interactive update (most common)
  lissto update

  # Update a specific stack
  lissto update --stack my-stack

  # Update with a specific branch
  lissto update --stack my-stack --branch develop

  # Update with auto-confirmation
  lissto update --stack my-stack --branch main --yes`,
	RunE:          runUpdate,
	SilenceUsage:  true,
	SilenceErrors: false,
}

func init() {
	updateCmd.Flags().StringVar(&updateStack, "stack", "", "Stack name to update")
	updateCmd.Flags().StringVar(&updateBranch, "branch", "", "Git branch for image resolution")
	updateCmd.Flags().StringVar(&updateCommit, "commit", "", "Git commit for image resolution")
	updateCmd.Flags().StringVar(&updateTag, "tag", "", "Git tag for image resolution")
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "Skip confirmation prompt")
	updateCmd.Flags().BoolVar(&updateNonInteractive, "non-interactive", false, "Disable interactive prompts")
}

func runUpdate(cmd *cobra.Command, args []string) error {
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

	// Get environment from flag or config
	envToUse := envName
	if envToUse == "" {
		envToUse = cfg.CurrentEnv
	}

	if envToUse == "" {
		return fmt.Errorf("no environment selected. Use --env flag or 'lissto env use <name>'")
	}

	// Create API client
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Step 1: List stacks in current environment
	stacks, err := apiClient.ListStacks(envToUse)
	if err != nil {
		return fmt.Errorf("failed to list stacks: %w", err)
	}

	if len(stacks) == 0 {
		return fmt.Errorf("no stacks found in environment '%s'", envToUse)
	}

	// Step 2: Select stack
	var selectedStack *types.Stack
	if updateStack != "" {
		// Find stack by name
		found := false
		for i := range stacks {
			if stacks[i].Name == updateStack {
				selectedStack = &stacks[i]
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("stack '%s' not found in environment '%s'", updateStack, envToUse)
		}
	} else if len(stacks) == 1 {
		// Only one stack, use it automatically
		selectedStack = &stacks[0]
	} else {
		// Interactive stack selection
		if updateNonInteractive {
			return fmt.Errorf("--stack is required in non-interactive mode")
		}

		// Collect data for columns
		titles := make([]string, len(stacks))
		envs := make([]string, len(stacks))
		ages := make([]string, len(stacks))

		for i, stack := range stacks {
			// Get blueprint title from annotations, fallback to stack name
			title := types.GetBlueprintTitle(&stack)
			if title == "" {
				title = stack.Name
			}

			// Calculate age
			duration := time.Since(stack.CreationTimestamp.Time)
			ages[i] = k8s.FormatAge(duration)

			titles[i] = title
			envs[i] = "env: " + stack.Spec.Env
		}

		// Format aligned options using helper
		options := interactive.FormatAlignedColumns(titles, ages, envs)

		// Show selection prompt
		var selectedIndex int
		prompt := &survey.Select{
			Message:  "Choose a stack to update:",
			Options:  options,
			PageSize: 10,
		}

		if err := survey.AskOne(prompt, &selectedIndex); err != nil {
			return fmt.Errorf("stack selection cancelled: %w", err)
		}

		selectedStack = &stacks[selectedIndex]
	}

	// Extract stack details
	stackName := selectedStack.Name
	blueprintRef := selectedStack.Spec.BlueprintReference
	stackEnv := selectedStack.Spec.Env
	currentImages := selectedStack.Spec.Images

	if stackName == "" || blueprintRef == "" || stackEnv == "" {
		return fmt.Errorf("failed to extract stack details")
	}

	// Show stack display with blueprint title if available
	stackDisplay := types.GetStackDisplayName(selectedStack)

	fmt.Printf("\n📦 Updating: %s (env: %s)\n", stackDisplay, stackEnv)

	// Step 3: Branch/Tag/Commit selection loop
	branch := updateBranch
	tag := updateTag
	commit := updateCommit
	skipBranchPrompt := branch != "" || tag != "" || commit != ""

	var prepareResp *client.PrepareStackResponse
	for {
		// Prompt for branch/tag/commit if not provided via flags
		if !skipBranchPrompt && !updateNonInteractive {
			fmt.Println("Enter branch/tag/commit for image resolution:")
			b, t, c, err := interactive.PromptBranchTag()
			if err != nil {
				return fmt.Errorf("cancelled: %w", err)
			}
			branch = b
			tag = t
			commit = c
		}

		// Step 4: Prepare stack to get new images
		fmt.Println("\nPreparing update...")
		prepareResp, err = apiClient.PrepareStack(
			blueprintRef,
			stackEnv,
			commit,
			branch,
			tag,
			true, // detailed
		)
		if err != nil {
			fmt.Printf("❌ Failed to prepare update: %v\n", err)

			if updateNonInteractive || updateYes {
				return fmt.Errorf("failed to prepare update: %w", err)
			}

			// Ask what user wants to do
			action, retryErr := interactive.ConfirmRetry()
			if retryErr != nil {
				return fmt.Errorf("failed to prepare update: %w", err)
			}

			switch action {
			case interactive.ActionTryAnotherBranchTag:
				// Reset and allow reprompting
				skipBranchPrompt = false
				branch = ""
				tag = ""
				commit = ""
				continue
			case interactive.ActionCancel:
				return fmt.Errorf("update cancelled")
			}
		}

		// Validate we have images
		if prepareResp == nil || len(prepareResp.Images) == 0 {
			return fmt.Errorf("no images returned from prepare")
		}

		// Check for missing images
		hasMissing := false
		for _, img := range prepareResp.Images {
			if img.Digest == "" || img.Digest == "N/A" {
				hasMissing = true
				break
			}
		}

		if hasMissing {
			fmt.Println("\n❌ Some services have missing images:")
			for _, img := range prepareResp.Images {
				if img.Digest == "" || img.Digest == "N/A" {
					fmt.Printf("  - %s\n", img.Service)
				}
			}

			if updateNonInteractive || updateYes {
				return fmt.Errorf("cannot update: some services have missing images")
			}

			action, retryErr := interactive.ConfirmRetry()
			if retryErr != nil {
				return fmt.Errorf("update cancelled")
			}

			switch action {
			case interactive.ActionTryAnotherBranchTag:
				skipBranchPrompt = false
				branch = ""
				tag = ""
				commit = ""
				continue
			case interactive.ActionCancel:
				return fmt.Errorf("update cancelled")
			}
		}

		// Successfully prepared
		break
	}

	// Step 5: Display comparison - only show changes in diff style
	hasChanges := false
	var changedServices []string

	for _, img := range prepareResp.Images {
		currentImageInfo := ""
		if currentImages != nil {
			if imgInfo, ok := currentImages[img.Service]; ok {
				currentImageInfo = imgInfo.Image
			}
		}

		newImage := img.Image
		if newImage == "" {
			newImage = img.Digest
		}

		// Only track changes
		if img.Digest != "" && currentImageInfo != newImage {
			hasChanges = true
			changedServices = append(changedServices, img.Service)
		}
	}

	// Show preview based on whether there are changes
	if !hasChanges {
		fmt.Println("\nℹ️  No new images found")

		if updateYes || updateNonInteractive {
			// Non-interactive mode with no changes - just exit
			return nil
		}

		// Interactive mode - only offer to try different branch or cancel
		for {
			action, err := interactive.ConfirmRetry()
			if err != nil {
				return nil
			}

			switch action {
			case interactive.ActionTryAnotherBranchTag:
				// Get new branch/tag/commit
				branch, tag, commit, err := interactive.PromptBranchTag()
				if err != nil {
					return fmt.Errorf("cancelled: %w", err)
				}

				// Restart prepare loop
				updateBranch = branch
				updateTag = tag
				updateCommit = commit
				return runUpdate(cmd, args)
			case interactive.ActionCancel:
				return nil
			}
		}
	} else {
		// Show git-style diff for changed services only
		fmt.Println("\n📋 Image Updates:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for _, img := range prepareResp.Images {
			currentImageInfo := ""
			if currentImages != nil {
				if imgInfo, ok := currentImages[img.Service]; ok {
					currentImageInfo = imgInfo.Image
				}
			}

			newImage := img.Image
			if newImage == "" {
				newImage = img.Digest
			}

			// Only show changed services
			if img.Digest != "" && currentImageInfo != newImage {
				fmt.Printf("\n%s:\n", img.Service)
				if currentImageInfo != "" {
					fmt.Printf("  \033[31m- %s (old)\033[0m\n", currentImageInfo)
				}
				fmt.Printf("  \033[32m+ %s (new)\033[0m\n", newImage)
			}
		}
		fmt.Println()
	}

	// Step 6: Confirm update (only if there are changes)
	if !updateYes && !updateNonInteractive && hasChanges {
		for {
			action, err := interactive.ConfirmUpdate()
			if err != nil {
				return fmt.Errorf("update cancelled: %w", err)
			}

			switch action {
			case interactive.ActionApplyUpdate:
				// Continue with update
				goto applyUpdate
			case interactive.ActionTryAnotherBranchTag:
				// Get new branch/tag/commit
				branch, tag, commit, err := interactive.PromptBranchTag()
				if err != nil {
					return fmt.Errorf("cancelled: %w", err)
				}

				// Restart prepare loop
				updateBranch = branch
				updateTag = tag
				updateCommit = commit
				return runUpdate(cmd, args)
			case interactive.ActionCancel:
				return fmt.Errorf("update cancelled")
			}
		}
	}

applyUpdate:
	// Step 7: Build images map and update stack
	fmt.Println("Applying update...")
	imagesMap := make(map[string]interface{})
	for _, img := range prepareResp.Images {
		imagesMap[img.Service] = map[string]interface{}{
			"digest": img.Digest,
			"image":  img.Image,
		}
	}

	if err := apiClient.UpdateStack(stackName, imagesMap); err != nil {
		return fmt.Errorf("failed to update stack: %w", err)
	}

	// Step 8: Success message
	fmt.Printf("\n✅ Stack '%s' updated successfully\n", stackName)

	// Show updated services count
	if len(changedServices) == 1 {
		fmt.Printf("Updated 1 service: %s\n", changedServices[0])
	} else {
		fmt.Printf("Updated %d services\n", len(changedServices))
	}

	return nil
}
