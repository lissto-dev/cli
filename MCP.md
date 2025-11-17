# MCP Integration for AI Assistants

Lissto includes a Model Context Protocol (MCP) server that lets AI assistants like Claude Desktop and Cursor manage your infrastructure through natural language.

## What is MCP?

MCP is an open standard that enables AI assistants to interact with external tools. With Lissto's MCP server, your AI assistant can manage environments, blueprints, stacks, view logs, and more.

## Prerequisites

Before using MCP, you need:

1. **Lissto CLI installed** (see [README.md](./README.md))
2. **Active lissto context** - authenticate first:

```bash
lissto login
```

The MCP server uses your current context for all operations.

## Configuration

### For Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS):

```json
{
  "mcpServers": {
    "lissto": {
      "command": "lissto",
      "args": ["mcp"]
    }
  }
}
```

### For Cursor

Add to your Cursor MCP settings (Settings → MCP or `~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "lissto": {
      "command": "lissto",
      "args": ["mcp"]
    }
  }
}
```

**Note:** If `lissto` isn't in your PATH, use the full path (find it with `which lissto`).

## Usage Examples

Once configured, restart your AI assistant and use natural language:

**Managing Stacks:**
- "List all my stacks"
- "Create a stack from the nginx blueprint"
- "Delete the old-api stack"
- "Show me the status of my dev stacks"

**Working with Blueprints:**
- "List all blueprints"
- "Create a blueprint from this docker-compose file: [paste yaml]"
- "Show me details of the backend blueprint"


**Monitoring:**
- "Show me logs from the frontend service"
- "Get the last 50 lines of logs from my-stack"
- "What's the status of all pods in dev environment?"

Your AI assistant automatically calls the appropriate Lissto operations to fulfill these requests.

## Troubleshooting

### "No active context" error

Login first:
```bash
lissto login
```

### Server not responding

1. Verify lissto is installed:
   ```bash
   which lissto
   lissto --version
   ```

2. Test the MCP server manually:
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | lissto mcp
   ```

3. Enable debug logging:
   ```json
   {
     "mcpServers": {
       "lissto": {
         "command": "lissto",
         "args": ["mcp", "--log-file", "/tmp/lissto-mcp.log"]
       }
     }
   }
   ```
   
   Then view logs: `tail -f /tmp/lissto-mcp.log`

4. Restart your AI assistant

## Security Note

The MCP server runs with your lissto credentials and can perform destructive operations (create, delete). Ensure your AI assistant is configured to confirm destructive actions.

## Resources

- **[Model Context Protocol](https://modelcontextprotocol.io/)** - MCP specification
- **[Lissto CLI README](./README.md)** - Main CLI documentation
