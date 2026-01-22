package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// OutputContext provides context-aware output handling for commands.
// It suppresses progress/status messages when structured output (JSON/YAML) is requested,
// ensuring clean machine-readable output for CI/CD pipelines.
type OutputContext struct {
	cmd    *cobra.Command
	format string
}

// NewOutputContext creates a new output context from a command
func NewOutputContext(cmd *cobra.Command) *OutputContext {
	return &OutputContext{
		cmd:    cmd,
		format: GetOutputFormat(cmd),
	}
}

// IsQuiet returns true if output format requires quiet mode (json/yaml).
// In quiet mode, progress messages and emojis should be suppressed.
func (o *OutputContext) IsQuiet() bool {
	return o.format == "json" || o.format == "yaml"
}

// Printf prints formatted text only if not in quiet mode
func (o *OutputContext) Printf(format string, a ...interface{}) {
	if !o.IsQuiet() {
		fmt.Printf(format, a...)
	}
}

// Println prints a line only if not in quiet mode
func (o *OutputContext) Println(a ...interface{}) {
	if !o.IsQuiet() {
		fmt.Println(a...)
	}
}

// PrintResult outputs the result using the unified PrintOutput pattern.
// For JSON/YAML formats, it serializes the data struct.
// For default format, it calls the customFormatter function.
func (o *OutputContext) PrintResult(data interface{}, customFormatter func()) error {
	return PrintOutput(o.cmd, data, customFormatter)
}
