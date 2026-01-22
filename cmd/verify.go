package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	apicompose "github.com/lissto-dev/api/pkg/compose"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/lissto-dev/cli/pkg/output"
)

var verifyCmd = &cobra.Command{
	Use:   "verify [compose-file]",
	Short: "Verify a Docker Compose file",
	Long: `Validate a Docker Compose file and show detailed information.

This command checks:
- YAML syntax
- Docker Compose schema validity
- Service definitions
- Network and volume configurations
- Environment variable references

Environment variables:
  LISSTO_COMPOSE_FILE  Override compose file path (used when no argument provided)

Examples:
  # Verify a compose file
  lissto verify docker-compose.yaml
  
  # Verify with full output
  lissto verify compose.yaml --verbose
  
  # Verify quietly (only errors)
  lissto verify compose.yaml --quiet
  
  # Verify with raw parser output (for debugging)
  lissto verify compose.yaml --raw
  
  # Verify using environment variable
  LISSTO_COMPOSE_FILE=docker-compose.yaml lissto verify`,
	Args: cobra.MaximumNArgs(1),
	RunE: runVerify,
}

func init() {
	verifyCmd.Flags().BoolP("verbose", "v", false, "Show verbose output including warnings")
	verifyCmd.Flags().BoolP("quiet", "q", false, "Only show errors, suppress warnings")
	verifyCmd.Flags().Bool("raw", false, "Show raw parser output (for debugging)")
}

func runVerify(cmd *cobra.Command, args []string) error {
	// Load environment variable overrides
	overrides := cmdutil.LoadEnvOverrides()

	// Determine compose file: argument > env var
	var composePath string
	if len(args) > 0 {
		composePath = args[0]
	} else if overrides.HasComposeFile() {
		composePath = overrides.ComposeFile
		fmt.Printf("📄 Using compose file from %s: %s\n", cmdutil.EnvOverrideComposeFile, composePath)
	} else {
		return fmt.Errorf("compose file required: provide as argument or set %s", cmdutil.EnvOverrideComposeFile)
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	raw, _ := cmd.Flags().GetBool("raw")

	// Silence all logs by default (we capture warnings internally)
	logrus.SetLevel(logrus.PanicLevel)

	// Read file
	data, err := os.ReadFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var validationResult *apicompose.ValidationResult

	if raw {
		// Raw mode: let all output flow naturally from the parser
		logrus.SetLevel(logrus.WarnLevel)
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: false,
			FullTimestamp:    true,
		})
		fmt.Println("🔍 Running validation with raw parser output...")
		fmt.Println()

		// Call validation without capturing warnings (warnings will be printed as they occur)
		validationResult, err = apicompose.ValidateComposeRaw(string(data))
		if err != nil {
			return err
		}

		// Show simple result
		fmt.Println()
		if validationResult.Valid {
			fmt.Println("✅ Result: Valid")
		} else {
			fmt.Println("❌ Result: Invalid")
			if len(validationResult.Errors) > 0 {
				fmt.Println("\nErrors:")
				for _, errMsg := range validationResult.Errors {
					fmt.Printf("  - %s\n", errMsg)
				}
			}
		}
	} else {
		// Normal mode: validate using shared logic (captures warnings internally)
		validationResult, err = apicompose.ValidateCompose(string(data))
		if err != nil {
			return err
		}

		// Prepare template data
		templateData := &output.VerifyTemplateData{
			Valid:        validationResult.Valid,
			Verbose:      verbose,
			Metadata:     validationResult.Metadata,
			Errors:       validationResult.Errors,
			Warnings:     validationResult.Warnings,
			WarningCount: len(validationResult.Warnings),
		}

		// Display results using template
		if err := output.PrintVerificationResultToStdout(templateData); err != nil {
			return fmt.Errorf("failed to display results: %w", err)
		}
	}

	// Exit with error code if invalid
	if !validationResult.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}
