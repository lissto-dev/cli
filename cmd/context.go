package cmd

import (
	"fmt"
	"os"

	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/output"
	"github.com/spf13/cobra"
)

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage contexts",
	Long:  `Manage Lissto API contexts. Contexts store API connection information.`,
}

// contextListCmd lists all contexts
var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contexts",
	RunE:  runContextList,
}

// contextCurrentCmd shows the current context
var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current context",
	RunE:  runContextCurrent,
}

// contextUseCmd switches to a different context
var contextUseCmd = &cobra.Command{
	Use:   "use <context-name>",
	Short: "Switch to a different context",
	Args:  cobra.ExactArgs(1),
	RunE:  runContextUse,
}

// contextDeleteCmd deletes a context
var contextDeleteCmd = &cobra.Command{
	Use:   "delete <context-name>",
	Short: "Delete a context",
	Args:  cobra.ExactArgs(1),
	RunE:  runContextDelete,
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextCurrentCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextDeleteCmd)
}

func runContextList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Contexts) == 0 {
		fmt.Println("No contexts found. Use 'lissto login' to create one.")
		return nil
	}

	if outputFormat == "json" {
		return output.PrintJSON(os.Stdout, cfg.Contexts)
	} else if outputFormat == "yaml" {
		return output.PrintYAML(os.Stdout, cfg.Contexts)
	}

	// Table format
	headers := []string{"NAME", "K8S CONTEXT", "SERVICE", "NAMESPACE", "CURRENT"}
	var rows [][]string
	for _, ctx := range cfg.Contexts {
		current := ""
		if ctx.Name == cfg.CurrentContext {
			current = "*"
		}
		rows = append(rows, []string{ctx.Name, ctx.KubeContext, ctx.ServiceName, ctx.ServiceNamespace, current})
	}
	output.PrintTable(os.Stdout, headers, rows)

	return nil
}

func runContextCurrent(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentContext == "" {
		return fmt.Errorf("no context selected")
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return err
	}

	if outputFormat == "json" {
		return output.PrintJSON(os.Stdout, ctx)
	} else if outputFormat == "yaml" {
		return output.PrintYAML(os.Stdout, ctx)
	}

	fmt.Printf("Current context: %s\n", ctx.Name)
	fmt.Printf("Kubernetes context: %s\n", ctx.KubeContext)
	fmt.Printf("Service: %s/%s\n", ctx.ServiceNamespace, ctx.ServiceName)
	if ctx.APIUrl != "" {
		fmt.Printf("API URL: %s\n", ctx.APIUrl)
	}

	return nil
}

func runContextUse(cmd *cobra.Command, args []string) error {
	contextName := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.SetCurrentContext(contextName); err != nil {
		return err
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Switched to context: %s\n", contextName)

	return nil
}

func runContextDelete(cmd *cobra.Command, args []string) error {
	contextName := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.DeleteContext(contextName); err != nil {
		return err
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Deleted context: %s\n", contextName)
	if cfg.CurrentContext == "" && len(cfg.Contexts) > 0 {
		fmt.Printf("Hint: Set a new current context with 'lissto context use <name>'\n")
	}

	return nil
}
