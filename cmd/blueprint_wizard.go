package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	apicompose "github.com/lissto-dev/api/pkg/compose"
	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/lissto-dev/cli/pkg/compose"
	"github.com/lissto-dev/cli/pkg/interactive"
	controllerconfig "github.com/lissto-dev/controller/pkg/config"
	"github.com/spf13/cobra"
)

// findGitRepo searches upward from the given directory to find a .git directory
func findGitRepo(startDir string) (string, error) {
	absPath, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	currentDir := absPath
	for {
		gitDir := filepath.Join(currentDir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return currentDir, nil
		}

		// Move up one directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// We've reached the root
			return "", fmt.Errorf("no git repository found")
		}
		currentDir = parent
	}
}

// getGitRemote gets the remote URL from the git repository
func getGitRemote(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	remote := strings.TrimSpace(string(output))
	if remote == "" {
		return "", fmt.Errorf("no git remote 'origin' configured")
	}

	return remote, nil
}

// inferRepositoryFromFile attempts to infer the repository from the docker-compose file's location
func inferRepositoryFromFile(composeFile string) (string, error) {
	// Get the directory containing the compose file
	dir := filepath.Dir(composeFile)

	// Find the git repository
	repoPath, err := findGitRepo(dir)
	if err != nil {
		return "", fmt.Errorf("no git repository found in or above %s", dir)
	}

	// Get the remote URL
	remote, err := getGitRemote(repoPath)
	if err != nil {
		return "", fmt.Errorf("found git repository at %s but %w", repoPath, err)
	}

	return remote, nil
}

// blueprintWizardFlow orchestrates the complete blueprint creation wizard
func blueprintWizardFlow(_ *cobra.Command, apiClient *client.Client) (*client.BlueprintResponse, error) {
	var selectedFile string
	var repository string

	// Load environment variable overrides
	overrides := cmdutil.LoadEnvOverrides()

	// Check for compose file override
	if overrides.HasComposeFile() {
		// Validate the file exists
		if _, err := os.Stat(overrides.ComposeFile); err != nil {
			return nil, fmt.Errorf("compose file from %s not found: %s", cmdutil.EnvOverrideComposeFile, overrides.ComposeFile)
		}
		selectedFile = overrides.ComposeFile
		fmt.Printf("📄 Using compose file from %s: %s\n", cmdutil.EnvOverrideComposeFile, selectedFile)
	} else {
		// Step 1: Detect compose files in current directory (with warnings silenced)
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		composeFiles, err := compose.DetectComposeFilesQuiet(currentDir)
		if err != nil {
			return nil, fmt.Errorf("failed to detect compose files: %w", err)
		}

		if len(composeFiles) == 0 {
			return nil, fmt.Errorf("no valid compose files found in current directory.\nSuggestion: Use 'lissto blueprint create <file>' or set %s", cmdutil.EnvOverrideComposeFile)
		}

		// Step 2: Select compose file (auto or prompt)
		selectedFile, err = compose.SelectComposeFile(composeFiles)
		if err != nil {
			return nil, fmt.Errorf("compose file selection cancelled: %w", err)
		}
	}

	// Check for repository override
	if overrides.HasRepository() {
		repository = overrides.Repository
		fmt.Printf("📦 Using repository from %s: %s\n", cmdutil.EnvOverrideRepository, repository)
	} else {
		// Step 3: Detect git repository
		var err error
		repository, err = inferRepositoryFromFile(selectedFile)
		if err != nil {
			return nil, fmt.Errorf("failed to detect git repository: %w\nSuggestion: Set %s to specify the repository", err, cmdutil.EnvOverrideRepository)
		}
	}

	// Step 4: Normalize repository URL
	normalizedRepo := controllerconfig.NormalizeRepositoryURL(repository)
	if !overrides.HasRepository() {
		fmt.Printf("📦 Detected repository: %s\n", normalizedRepo)
	}

	// Step 5: Validate compose file with warning detection
	fmt.Println("\n📋 Validating compose file...")
	composeContent, err := os.ReadFile(selectedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	validationResult, err := apicompose.ValidateCompose(string(composeContent))
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if !validationResult.Valid {
		fmt.Println("❌ Compose file is invalid")
		for _, errMsg := range validationResult.Errors {
			fmt.Printf("  - %s\n", errMsg)
		}
		return nil, fmt.Errorf("compose file validation failed")
	}

	// Show concise validation result
	if len(validationResult.Warnings) > 0 {
		fmt.Printf("⚠️  Compose file is valid but %d warning(s) found\n", len(validationResult.Warnings))
		fmt.Printf("💡 Run 'lissto verify %s --verbose' to see details\n", filepath.Base(selectedFile))
	} else {
		fmt.Println("✅ Compose file is valid")
	}

	// Step 6: Check for existing blueprints for this repository
	existingBlueprints, err := apiClient.FindBlueprintsByRepository(normalizedRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing blueprints: %w", err)
	}

	var blueprintIDToDelete string
	var shouldOverride bool

	if len(existingBlueprints) > 0 {
		// Step 7: Handle existing blueprint
		latestBP := existingBlueprints[0] // Already sorted by newest first

		action, err := interactive.ConfirmBlueprintAction(latestBP)
		if err != nil {
			return nil, fmt.Errorf("cancelled: %w", err)
		}

		switch action {
		case interactive.ActionOverrideBlueprint:
			shouldOverride = true
			blueprintIDToDelete = latestBP.ID

			// Step 8: Check for active stacks using this blueprint
			env, err := cmdutil.GetOrCreateDefaultEnv(apiClient, createEnv, false)
			if err != nil {
				return nil, fmt.Errorf("failed to determine environment: %w", err)
			}

			stacks, err := apiClient.FindStacksByBlueprint(latestBP.ID, env)
			if err != nil {
				return nil, fmt.Errorf("failed to check for active stacks: %w", err)
			}

			if len(stacks) > 0 {
				// Stacks are using this blueprint
				stackNames := make([]string, len(stacks))
				for i, s := range stacks {
					stackNames[i] = s.Name
				}

				stackAction, err := interactive.ConfirmStackDeletion(stackNames)
				if err != nil {
					return nil, fmt.Errorf("cancelled: %w", err)
				}

				switch stackAction {
				case interactive.ActionDeleteStacksContinue:
					// Delete all stacks using this blueprint
					fmt.Println("\nDeleting stacks...")
					for _, stack := range stacks {
						fmt.Printf("  Deleting stack: %s\n", stack.Name)
						if err := apiClient.DeleteStack(stack.Name, env); err != nil {
							return nil, fmt.Errorf("failed to delete stack %s: %w", stack.Name, err)
						}
					}
					fmt.Println("✅ Stacks deleted successfully")

				case interactive.ActionCreateVersionInstead:
					shouldOverride = false
					blueprintIDToDelete = ""

				case interactive.ActionCancel:
					return nil, fmt.Errorf("cancelled by user")
				}
			}

		case interactive.ActionCreateNewVersion:
			shouldOverride = false

		case interactive.ActionCancel:
			return nil, fmt.Errorf("cancelled by user")
		}
	}

	// Step 9: Delete old blueprint if overriding
	if shouldOverride && blueprintIDToDelete != "" {
		fmt.Printf("Deleting old blueprint: %s\n", blueprintIDToDelete)
		if err := apiClient.DeleteBlueprint(blueprintIDToDelete); err != nil {
			return nil, fmt.Errorf("failed to delete old blueprint: %w", err)
		}
	}

	// Step 10: Create new blueprint
	fmt.Println("\nCreating blueprint...")
	req := client.CreateBlueprintRequest{
		Compose:    string(composeContent),
		Repository: normalizedRepo,
	}

	identifier, err := apiClient.CreateBlueprint(req)
	if err != nil {
		// Check if it's a repository configuration error
		if strings.Contains(err.Error(), "is not configured") || strings.Contains(err.Error(), "not allowed") {
			fmt.Printf("\n❌ Repository configuration error\n")
			fmt.Printf("   Detected: %s\n\n", normalizedRepo)
			fmt.Printf("💡 This repository is not in your Lissto configuration.\n")
			fmt.Printf("   If using SSH host aliases (e.g., github.com-lissto), try:\n")
			fmt.Printf("   • Set %s=git@github.com:org/repo.git\n", cmdutil.EnvOverrideRepository)
			fmt.Printf("   • Or use: lissto blueprint create <file> --repository git@github.com:org/repo.git\n")
		}
		return nil, fmt.Errorf("failed to create blueprint: %w", err)
	}

	fmt.Printf("✅ Blueprint created successfully!\n")
	fmt.Printf("Blueprint ID: %s\n\n", identifier)

	// Fetch the created blueprint to return
	createdBP, err := apiClient.GetBlueprint(identifier)
	if err != nil {
		// Don't fail the whole operation, just return nil
		fmt.Printf("⚠️  Warning: Could not fetch created blueprint details: %v\n", err)
		return nil, nil
	}

	return createdBP, nil
}
