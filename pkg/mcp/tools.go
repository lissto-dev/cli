package mcp

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// GetAllTools returns all available MCP tools
func GetAllTools() []Tool {
	return []Tool{
		// Environment tools
		{
			Name:        "lissto_env_list",
			Description: "List all available environments",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "lissto_env_get",
			Description: "Get details of a specific environment",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Environment name",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "lissto_env_create",
			Description: "Create a new environment",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Environment name",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "lissto_env_current",
			Description: "Get the current active environment from config",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},

		// Blueprint tools
		{
			Name:        "lissto_blueprint_list",
			Description: "List all blueprints (user and optionally global)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"include_global": map[string]interface{}{
						"type":        "boolean",
						"description": "Include global blueprints in the list",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "lissto_blueprint_get",
			Description: "Get details of a specific blueprint",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Blueprint name/ID",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "lissto_blueprint_create",
			Description: "Create a new blueprint from docker-compose YAML content",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"compose": map[string]interface{}{
						"type":        "string",
						"description": "Docker compose YAML content",
					},
					"branch": map[string]interface{}{
						"type":        "string",
						"description": "Git branch name (optional)",
					},
					"author": map[string]interface{}{
						"type":        "string",
						"description": "Author name (optional)",
					},
					"repository": map[string]interface{}{
						"type":        "string",
						"description": "Repository name (optional)",
					},
					"global": map[string]interface{}{
						"type":        "boolean",
						"description": "Create as global blueprint (requires admin)",
						"default":     false,
					},
				},
				"required": []string{"compose"},
			},
		},
		{
			Name:        "lissto_blueprint_delete",
			Description: "Delete a blueprint",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Blueprint name/ID to delete",
					},
				},
				"required": []string{"name"},
			},
		},

		// Stack tools
		{
			Name:        "lissto_stack_list",
			Description: "List all stacks in current or specified environment",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Environment name (optional, defaults to current)",
					},
				},
			},
		},
		{
			Name:        "lissto_stack_get",
			Description: "Get details of a specific stack",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Stack name",
					},
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Environment name (optional, defaults to current)",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "lissto_stack_create",
			Description: "Create a new stack from a blueprint",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"blueprint_name": map[string]interface{}{
						"type":        "string",
						"description": "Blueprint name/ID to use",
					},
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Environment name (optional, defaults to current)",
					},
				},
				"required": []string{"blueprint_name"},
			},
		},
		{
			Name:        "lissto_stack_delete",
			Description: "Delete a stack",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Stack name to delete",
					},
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Environment name (optional, defaults to current)",
					},
				},
				"required": []string{"name"},
			},
		},

		// Admin tools
		{
			Name:        "lissto_admin_apikey_create",
			Description: "Create a new API key (admin only)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "User name for the API key",
					},
					"role": map[string]interface{}{
						"type":        "string",
						"description": "Role for the API key (admin, user, deploy)",
						"default":     "user",
						"enum":        []string{"admin", "user", "deploy"},
					},
					"slack_user_id": map[string]interface{}{
						"type":        "string",
						"description": "Slack user ID (optional)",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "lissto_admin_blueprint_delete",
			Description: "Force delete a blueprint including global blueprints (admin only)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Blueprint name/ID to delete",
					},
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Environment name (optional)",
					},
				},
				"required": []string{"name"},
			},
		},

		// Variable tools
		{
			Name:        "lissto_variable_list",
			Description: "List all variables (env and global)",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "lissto_variable_get",
			Description: "Get details of a specific variable config",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Variable config name or ID",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "lissto_variable_create",
			Description: "Create or update a variable config. Name is auto-generated based on env/scope (e.g., 'staging', 'repo-myapp'). If a config already exists, new keys are merged in. Rejects only if keys conflict (same key, different value). Provide environment variables as key-value pairs in the data field.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": map[string]interface{}{
						"type":        "string",
						"description": "Scope: env, repo, or global (default: env)",
						"default":     "env",
						"enum":        []string{"env", "repo", "global"},
					},
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Environment name (default: current env from context)",
					},
					"repository": map[string]interface{}{
						"type":        "string",
						"description": "Repository URL (required for scope=repo)",
					},
					"data": map[string]interface{}{
						"type":        "object",
						"description": "Environment variable key-value pairs (e.g., {\"DB_HOST\": \"localhost\", \"PORT\": \"8080\"})",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"data"},
			},
		},
		{
			Name:        "lissto_variable_update",
			Description: "Update a variable config's data",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Variable config name or ID",
					},
					"data": map[string]interface{}{
						"type":        "object",
						"description": "New key-value pairs (replaces existing)",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"name", "data"},
			},
		},
		{
			Name:        "lissto_variable_delete",
			Description: "Delete a variable config",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Variable config name or ID",
					},
				},
				"required": []string{"name"},
			},
		},

		// Secret tools
		{
			Name:        "lissto_secret_list",
			Description: "List all secrets (keys only, no values)",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "lissto_secret_get",
			Description: "Get details of a specific secret config (keys only, no values)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Secret config name or ID",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "lissto_secret_create",
			Description: "Create a new secret config. Name is auto-generated based on env/scope (e.g., 'staging', 'repo-myapp'). If a config already exists, this operation will FAIL. To add/update keys in an existing secret, you must explicitly use lissto_secret_set instead (irreversible operation requires explicit intent).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"scope": map[string]interface{}{
						"type":        "string",
						"description": "Scope: env, repo, or global (default: env)",
						"default":     "env",
						"enum":        []string{"env", "repo", "global"},
					},
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Environment name (default: current env from context)",
					},
					"repository": map[string]interface{}{
						"type":        "string",
						"description": "Repository URL (required for scope=repo)",
					},
					"secrets": map[string]interface{}{
						"type":        "object",
						"description": "Secret key-value pairs (e.g., {\"API_KEY\": \"secret123\", \"DB_PASSWORD\": \"pass456\"})",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"secrets"},
			},
		},
		{
			Name:        "lissto_secret_set",
			Description: "Set/update secret values (merges with existing)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Secret config name or ID",
					},
					"secrets": map[string]interface{}{
						"type":        "object",
						"description": "Key-value pairs to set",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"name", "secrets"},
			},
		},
		{
			Name:        "lissto_secret_delete",
			Description: "Delete a secret config",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Secret config name or ID",
					},
				},
				"required": []string{"name"},
			},
		},

		// Status and logs tools
		{
			Name:        "lissto_status",
			Description: "Get detailed status of stacks and their pods",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Filter by environment name (optional)",
					},
				},
			},
		},
		{
			Name:        "lissto_logs",
			Description: "Get recent logs from stack pods (not streaming, returns last N lines)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"stack": map[string]interface{}{
						"type":        "string",
						"description": "Filter by stack name (optional)",
					},
					"env": map[string]interface{}{
						"type":        "string",
						"description": "Filter by environment (optional)",
					},
					"service": map[string]interface{}{
						"type":        "string",
						"description": "Filter by service name (optional)",
					},
					"pod": map[string]interface{}{
						"type":        "string",
						"description": "Filter by specific pod name (optional)",
					},
					"tail": map[string]interface{}{
						"type":        "integer",
						"description": "Number of lines to show from end of logs",
						"default":     100,
					},
					"max_pods": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of pods to get logs from",
						"default":     5,
					},
				},
			},
		},
	}
}
