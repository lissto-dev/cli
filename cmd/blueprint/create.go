package blueprint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/lissto-dev/cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	createBranch     string
	createAuthor     string
	createRepository string
	createOutput     string
)

// BlueprintCreateResult represents the JSON output for blueprint create command
type BlueprintCreateResult struct {
	ID         string `json:"id"`
	Repository string `json:"repository"`
	Branch     string `json:"branch,omitempty"`
	Author     string `json:"author,omitempty"`
}

var createCmd = &cobra.Command{
	Use:   "create <docker-compose-file>",
	Short: "Create a new blueprint",
	Long: `Create a new blueprint from a docker-compose file.

Optional flags:
  --branch          Branch name (for CI/CD workflows)
  --author          Author name (for CI/CD workflows)
  --repository      Repository name/URL (overrides auto-detection)
  --output          Output format (json, yaml)

Environment variables:
  LISSTO_REPOSITORY    Override repository auto-detection
  LISSTO_COMPOSE_FILE  Override compose file path (used when no argument provided)
  LISSTO_API_KEY       API key for CI/CD authentication
  LISSTO_API_URL       API URL for CI/CD authentication`,
	Args:          cobra.MaximumNArgs(1),
	RunE:          runCreate,
	SilenceUsage:  true, // Don't show usage on errors
	SilenceErrors: false,
}

func init() {
	createCmd.Flags().StringVar(&createBranch, "branch", "", "Branch name (for CI/CD workflows)")
	createCmd.Flags().StringVar(&createAuthor, "author", "", "Author name (for CI/CD workflows)")
	createCmd.Flags().StringVar(&createRepository, "repository", "", "Repository name/URL (used for blueprint title)")
	createCmd.Flags().StringVarP(&createOutput, "output", "o", "", "Output format (json, yaml)")
}

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

func runCreate(_ *cobra.Command, args []string) error {
	// Load environment variable overrides
	overrides := cmdutil.LoadOverrides()
	isJSONOutput := createOutput == "json" || createOutput == "yaml"

	// Determine compose file: argument > env var
	var composeFile string
	if len(args) > 0 {
		composeFile = args[0]
	} else if overrides.HasComposeFile() {
		composeFile = overrides.ComposeFile
		if !isJSONOutput {
			fmt.Printf("📄 Using compose file from %s: %s\n", cmdutil.EnvOverrideComposeFile, composeFile)
		}
	} else {
		return fmt.Errorf("compose file required: provide as argument or set %s", cmdutil.EnvOverrideComposeFile)
	}

	apiClient, err := cmdutil.GetAPIClient()
	if err != nil {
		return err
	}

	// Read docker-compose file
	composeContent, err := os.ReadFile(composeFile)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose file: %w", err)
	}

	// Determine repository: flag > env var > auto-detect
	repository := createRepository
	if repository == "" {
		if overrides.HasRepository() {
			repository = overrides.Repository
			if !isJSONOutput {
				fmt.Printf("📦 Using repository from %s: %s\n", cmdutil.EnvOverrideRepository, repository)
			}
		} else {
			inferredRepo, err := inferRepositoryFromFile(composeFile)
			if err != nil {
				return fmt.Errorf("failed to infer repository: %w\nPlease specify --repository or set %s", err, cmdutil.EnvOverrideRepository)
			}
			repository = inferredRepo
		}
	}

	// Build request (scope determined by API based on repository)
	req := client.CreateBlueprintRequest{
		Compose:    string(composeContent),
		Branch:     createBranch,
		Author:     createAuthor,
		Repository: repository,
	}

	identifier, err := apiClient.CreateBlueprint(req)
	if err != nil {
		return fmt.Errorf("failed to create blueprint: %w", err)
	}

	// Prepare result
	result := BlueprintCreateResult{
		ID:         identifier,
		Repository: repository,
		Branch:     createBranch,
		Author:     createAuthor,
	}

	// Output based on format
	switch createOutput {
	case "json":
		return output.PrintJSON(os.Stdout, result)
	case "yaml":
		return output.PrintYAML(os.Stdout, result)
	default:
		fmt.Printf("Blueprint created successfully\n")
		fmt.Printf("ID: %s\n", identifier)
		return nil
	}
}
