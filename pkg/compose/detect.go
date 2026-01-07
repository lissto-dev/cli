package compose

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	apicompose "github.com/lissto-dev/api/pkg/compose"
	"github.com/lissto-dev/controller/pkg/config"
	"github.com/sirupsen/logrus"
)

var composeFilePatterns = []string{
	"docker-compose.yaml",
	"docker-compose.yml",
	"compose.yaml",
	"compose.yml",
}

// silenceLoggers silences both logrus and standard log, returns cleanup function
func silenceLoggers() func() {
	oldLogrusLevel := logrus.GetLevel()
	oldLogrusOutput := logrus.StandardLogger().Out
	logrus.SetLevel(logrus.PanicLevel)
	logrus.SetOutput(io.Discard)

	oldLogOutput := log.Writer()
	log.SetOutput(io.Discard)

	return func() {
		logrus.SetLevel(oldLogrusLevel)
		logrus.SetOutput(oldLogrusOutput)
		log.SetOutput(oldLogOutput)
	}
}

// validateComposeContent checks if compose content is valid
func validateComposeContent(data []byte) bool {
	_, err := apicompose.ParseBlueprintMetadata(string(data), config.RepoConfig{})
	return err == nil
}

// DetectComposeFiles searches for valid compose files in the given directory (deprecated, use DetectComposeFilesQuiet)
func DetectComposeFiles(dir string) ([]string, error) {
	return DetectComposeFilesQuiet(dir)
}

// DetectComposeFilesQuiet searches for valid compose files with ALL warnings silenced
// This is used during auto-detection to avoid cluttering output
func DetectComposeFilesQuiet(dir string) ([]string, error) {
	cleanup := silenceLoggers()
	defer cleanup()

	var validFiles []string
	for _, pattern := range composeFilePatterns {
		path := filepath.Join(dir, pattern)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if validateComposeContent(data) {
			validFiles = append(validFiles, path)
		}
	}

	return validFiles, nil
}

// ValidateComposeFile checks if a file is a valid Docker Compose file
// Uses the API's compose parser which leverages compose-spec's official parser
// Warnings are silenced during validation to avoid noise during file detection
func ValidateComposeFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	cleanup := silenceLoggers()
	defer cleanup()

	return validateComposeContent(data), nil
}

// SelectComposeFile prompts the user to select from multiple compose files
// If only one file, returns it automatically
// If no files, returns error
func SelectComposeFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no valid compose files found")
	}

	if len(files) == 1 {
		fmt.Printf("🔍 Detected compose file: %s\n", filepath.Base(files[0]))
		return files[0], nil
	}

	// Multiple files - prompt user to select
	options := make([]string, len(files))
	for i, f := range files {
		options[i] = filepath.Base(f)
	}

	var selectedIndex int
	prompt := &survey.Select{
		Message: "Multiple compose files found. Which one would you like to use?",
		Options: options,
	}

	err := survey.AskOne(prompt, &selectedIndex)
	if err != nil {
		return "", err
	}

	fmt.Printf("🔍 Using compose file: %s\n", filepath.Base(files[selectedIndex]))
	return files[selectedIndex], nil
}
