package interactive

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/output"
)

// FormatAlignedColumns formats multiple columns of data with proper alignment
// Each column is a slice of strings. Returns a slice of aligned strings.
func FormatAlignedColumns(columns ...[]string) []string {
	if len(columns) == 0 {
		return nil
	}

	numRows := len(columns[0])
	if numRows == 0 {
		return nil
	}

	// Find max width for each column
	maxWidths := make([]int, len(columns))
	for colIdx, column := range columns {
		for _, value := range column {
			if len(value) > maxWidths[colIdx] {
				maxWidths[colIdx] = len(value)
			}
		}
	}

	// Build aligned rows
	result := make([]string, numRows)
	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		parts := make([]string, len(columns))
		for colIdx, column := range columns {
			// Last column doesn't need padding
			if colIdx == len(columns)-1 {
				parts[colIdx] = column[rowIdx]
			} else {
				parts[colIdx] = fmt.Sprintf("%-*s", maxWidths[colIdx], column[rowIdx])
			}
		}
		result[rowIdx] = strings.Join(parts, "   ")
	}

	return result
}

// SelectBlueprint prompts the user to select a blueprint interactively
func SelectBlueprint(blueprints []client.BlueprintResponse) (*client.BlueprintResponse, error) {
	if len(blueprints) == 0 {
		return nil, fmt.Errorf("no blueprints available")
	}

	// Collect data for columns
	titles := make([]string, len(blueprints))
	ages := make([]string, len(blueprints))
	services := make([]string, len(blueprints))

	for i, bp := range blueprints {
		title := bp.Title
		if title == "" {
			title = bp.ID
		}

		// Add scope indicator (global vs user)
		scope := "🌐" // Global icon
		if !strings.HasPrefix(bp.ID, "global/") {
			scope = "👤" // User icon
		}
		titles[i] = scope + " " + title

		ages[i] = output.ExtractBlueprintAge(bp.ID)

		// Build services and infra display
		var parts []string
		if len(bp.Content.Services) > 0 {
			parts = append(parts, "Services: "+strings.Join(bp.Content.Services, ", "))
		}
		if len(bp.Content.Infra) > 0 {
			parts = append(parts, "Infra: "+strings.Join(bp.Content.Infra, ", "))
		}

		if len(parts) > 0 {
			services[i] = strings.Join(parts, "    ")
		}
	}

	// Format aligned options
	options := FormatAlignedColumns(titles, ages, services)

	var selectedIndex int
	prompt := &survey.Select{
		Message:  "Choose a blueprint:",
		Options:  options,
		PageSize: 10,
	}

	err := survey.AskOne(prompt, &selectedIndex)
	if err != nil {
		return nil, err
	}

	return &blueprints[selectedIndex], nil
}

// ConfirmDeployment asks the user what they want to do after seeing the preview
func ConfirmDeployment() (string, error) {
	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Deploy",
			"Try another branch/tag",
			"Cancel",
		},
		Default: "Deploy",
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		return "", err
	}

	return action, nil
}

// ConfirmUpdate asks the user what they want to do after seeing the update preview
func ConfirmUpdate() (string, error) {
	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Apply Update",
			"Try another branch/tag",
			"Cancel",
		},
		Default: "Apply Update",
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		return "", err
	}

	return action, nil
}

// ConfirmDeploymentWithBack asks the user what they want to do, including option to go back
func ConfirmDeploymentWithBack() (string, error) {
	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Deploy",
			"Try another branch/tag",
			"Back to blueprint selection",
			"Cancel",
		},
		Default: "Deploy",
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		return "", err
	}

	return action, nil
}

// ConfirmRetry asks the user what they want to do after a failure
func ConfirmRetry() (string, error) {
	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Try another branch/tag",
			"Cancel",
		},
		Default: "Try another branch/tag",
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		return "", err
	}

	return action, nil
}

// ConfirmRetryWithBack asks the user what they want to do after a failure, including option to go back
func ConfirmRetryWithBack() (string, error) {
	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{
			"Try another branch/tag",
			"Back to blueprint selection",
			"Cancel",
		},
		Default: "Try another branch/tag",
	}

	err := survey.AskOne(prompt, &action)
	if err != nil {
		return "", err
	}

	return action, nil
}

// PromptBranchTag prompts for branch, tag, or commit (single input for simplicity)
func PromptBranchTag() (branch, tag, commit string, err error) {
	var value string
	inputPrompt := &survey.Input{
		Message: "Enter branch/tag/commit:",
		Help:    "This will be used to resolve images. Can be a branch name, tag, or commit hash.",
		Default: "main",
	}

	err = survey.AskOne(inputPrompt, &value)
	if err != nil {
		return "", "", "", err
	}

	if value == "" {
		return "", "", "", fmt.Errorf("no value provided")
	}

	// Use as branch by default - the API will try multiple resolution methods
	branch = value
	return branch, "", "", nil
}

// ConfirmAction asks for a yes/no confirmation
func ConfirmAction(message string, defaultValue bool) (bool, error) {
	var confirmed bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}

	err := survey.AskOne(prompt, &confirmed)
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

// SelectEnv prompts the user to select an environment
func SelectEnv(envs []client.EnvResponse) (*client.EnvResponse, error) {
	if len(envs) == 0 {
		return nil, fmt.Errorf("no environments available")
	}

	if len(envs) == 1 {
		// If only one env, use it automatically
		return &envs[0], nil
	}

	// Create options
	options := make([]string, len(envs))
	for i, env := range envs {
		options[i] = env.Name
	}

	var selectedIndex int
	prompt := &survey.Select{
		Message:  "Choose an environment:",
		Options:  options,
		PageSize: 10,
	}

	err := survey.AskOne(prompt, &selectedIndex)
	if err != nil {
		return nil, err
	}

	return &envs[selectedIndex], nil
}

// SelectStack prompts the user to select a stack
func SelectStack(stacks interface{}) (interface{}, error) {
	// Handle both []types.Stack and []interface{}
	var stackList []interface{}
	var count int

	switch v := stacks.(type) {
	case []interface{}:
		stackList = v
		count = len(v)
	default:
		// Try to convert from typed slice using reflection-like approach
		// For now, return error if not []interface{}
		return nil, fmt.Errorf("unsupported stack list type")
	}

	if count == 0 {
		return nil, fmt.Errorf("no stacks available")
	}

	// Create options with blueprint title and env
	options := make([]string, count)
	for i, s := range stackList {
		// Type assert to access stack fields
		stack, ok := s.(map[string]interface{})
		if !ok {
			// Try as Stack type from types package
			return nil, fmt.Errorf("invalid stack type")
		}

		// Get blueprint title from annotations
		title := ""
		if metadata, ok := stack["metadata"].(map[string]interface{}); ok {
			if annotations, ok := metadata["annotations"].(map[string]interface{}); ok {
				if blueprintTitle, ok := annotations["lissto.dev/blueprint-title"].(string); ok {
					title = blueprintTitle
				}
			}
			// Fallback to stack name
			if title == "" {
				if name, ok := metadata["name"].(string); ok {
					title = name
				}
			}
		}

		// Get env name
		envName := "unknown"
		if spec, ok := stack["spec"].(map[string]interface{}); ok {
			if env, ok := spec["env"].(string); ok {
				envName = env
			}
		}

		options[i] = fmt.Sprintf("%s (env: %s)", title, envName)
	}

	var selectedIndex int
	prompt := &survey.Select{
		Message:  "Choose a stack to update:",
		Options:  options,
		PageSize: 10,
	}

	err := survey.AskOne(prompt, &selectedIndex)
	if err != nil {
		return nil, err
	}

	return stackList[selectedIndex], nil
}
