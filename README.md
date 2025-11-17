# Lissto CLI

Deploy and manage your stacks on Kubernetes with a simple command-line interface. Lissto lets you convert docker-compose files into Kubernetes resources and manage them across multiple environments.

## Installation

### Homebrew (Recommended)

```bash
brew install lissto-dev/tap/lissto
```

### Download Binary

Download pre-built binaries from the [releases page](https://github.com/lissto-dev/cli/releases).

## Quick Start

### 1. Login

Connect to your Lissto API server:

```bash
# Login (interactive)
lissto login
```

### 2. Common Commands

```bash
# Create a blueprint from docker-compose
lissto blueprint create docker-compose.yaml

# List blueprints
lissto blueprint list

# Create a stack from a blueprint (interactive)
lissto create

# View status across all environments (interactive)
lissto status

# View logs (interactive)
lissto logs

# View help for any command
lissto --help
```

### 3. MCP Integration for AI Assistants

Lissto includes a Model Context Protocol (MCP) server that lets AI assistants like Claude and Cursor manage your infrastructure.

**Quick Setup:**

1. Add to your MCP settings (e.g., `~/Library/Application Support/Claude/claude_desktop_config.json`):

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

2. Restart your AI assistant

Your AI can now manage environments, blueprints, stacks, view logs, and more through natural language.

**Learn more:** See [MCP.md](./MCP.md) for detailed setup and capabilities.

## Global Flags

- `--output, -o`: Output format (json, yaml, wide)
- `--context`: Override current context
- `--env`: Override current environment

## Available Commands

- `login` - Authenticate with Lissto API
- `context` - Manage API contexts
- `env` - Manage environments
- `blueprint` - Manage blueprints
- `stack` - Manage stacks
- `status` - View stack status
- `logs` - View pod logs
- `update` - Update stacks
- `admin` - Admin operations (requires admin role)
- `mcp` - Start MCP server for AI assistants

Run `lissto <command> --help` for detailed information on each command.

## Documentation

- **[MCP Integration](./MCP.md)** - Model Context Protocol setup for AI assistants
- **[Development Guide](./DEVELOPMENT.md)** - Building and contributing to Lissto CLI

## Support

- Documentation: [https://docs.lissto.dev](https://docs.lissto.dev)
- Issues: [GitHub Issues](https://github.com/lissto-dev/cli/issues)



