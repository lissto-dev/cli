package blueprint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/spf13/cobra"
)

var (
	createBranch     string
	createAuthor     string
	createRepository string
)

var createCmd = &cobra.Command{
	Use:   "create <docker-compose-file>",
	Short: "Create a new blueprint",
	Long: `Create a new blueprint from a docker-compose file.

The repository will be automatically inferred from the git repository where the 
docker-compose file is located (searching upward in the directory tree). If no 
git repository is found or no remote is configured, the command will fail.

Optional flags:
  --branch          Branch name (for CI/CD workflows)
  --author          Author name (for CI/CD workflows)
  --repository      Repository name/URL (overrides auto-detection)`,
	Args:          cobra.ExactArgs(1),
	RunE:          runCreate,
	SilenceUsage:  true, // Don't show usage on errors
	SilenceErrors: false,
}

func init() {
	createCmd.Flags().StringVar(&createBranch, "branch", "", "Branch name (for CI/CD workflows)")
	createCmd.Flags().StringVar(&createAuthor, "author", "", "Author name (for CI/CD workflows)")
	createCmd.Flags().StringVar(&createRepository, "repository", "", "Repository name/URL (used for blueprint title)")
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

func runCreate(cmd *cobra.Command, args []string) error {
	composeFile := args[0]

	apiClient, err := getAPIClient()
	if err != nil {
		return err
	}

	// Read docker-compose file
	composeContent, err := os.ReadFile(composeFile)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose file: %w", err)
	}

	// Infer repository if not provided
	repository := createRepository
	if repository == "" {
		inferredRepo, err := inferRepositoryFromFile(composeFile)
		if err != nil {
			return fmt.Errorf("failed to infer repository: %w. Please specify --repository explicitly", err)
		}
		repository = inferredRepo
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

	fmt.Printf("Blueprint created successfully\n")
	fmt.Printf("ID: %s\n", identifier)

	return nil
}
