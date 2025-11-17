package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lissto-dev/cli/pkg/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpLogFile string
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server",
	Long: `Start an MCP server that exposes lissto operations as tools.

The MCP server communicates via JSON-RPC 2.0 over stdin/stdout and can be
used by AI assistants like Claude Desktop and Cursor.

The server uses your current lissto context (configured via 'lissto login')
for authentication and API access.

Example usage:
  lissto mcp

Configuration for Cursor:
  Add to your MCP settings:
  {
    "mcpServers": {
      "lissto": {
        "command": "/path/to/lissto",
        "args": ["mcp"]
      }
    }
  }

Available tools:
  - Environment management (list, get, create, current)
  - Blueprint management (list, get, create, delete)
  - Stack management (list, get, create, delete)
  - Admin operations (API key creation, force delete)
  - Status and logs (get stack status, retrieve logs)

Prerequisites:
  - Run 'lissto login' to configure your context
  - Ensure you have a valid API key and active context`,
	RunE:          runMCP,
	SilenceUsage:  true,
	SilenceErrors: false,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVar(&mcpLogFile, "log-file", "/tmp/lissto-mcp.log", "Path to log file for debugging MCP server")
}

func runMCP(cmd *cobra.Command, args []string) error {
	// Create MCP server with optional logging
	server, err := mcp.NewServer(os.Stdin, os.Stdout, mcpLogFile)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}
	defer server.Close()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Run()
	}()

	// Wait for either server error or shutdown signal
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("MCP server error: %w", err)
		}
		return nil
	case sig := <-sigChan:
		fmt.Fprintf(os.Stderr, "\nReceived signal %v, shutting down...\n", sig)
		return nil
	}
}
