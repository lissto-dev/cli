package mcp

import (
	"context"
	"fmt"
	"io"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/k8s"
	"github.com/lissto-dev/cli/pkg/status"
	corev1 "k8s.io/api/core/v1"
)

// Logger interface for handlers
type Logger interface {
	log(format string, args ...interface{})
}

// ExecuteTool executes a tool with the given arguments
func ExecuteTool(name string, args map[string]interface{}, logger Logger) (interface{}, error) {
	switch name {
	// Environment tools
	case "lissto_env_list":
		return handleEnvList(args, logger)
	case "lissto_env_get":
		return handleEnvGet(args, logger)
	case "lissto_env_create":
		return handleEnvCreate(args, logger)
	case "lissto_env_current":
		return handleEnvCurrent(args, logger)

	// Blueprint tools
	case "lissto_blueprint_list":
		return handleBlueprintList(args, logger)
	case "lissto_blueprint_get":
		return handleBlueprintGet(args, logger)
	case "lissto_blueprint_create":
		return handleBlueprintCreate(args, logger)
	case "lissto_blueprint_delete":
		return handleBlueprintDelete(args, logger)

	// Stack tools
	case "lissto_stack_list":
		return handleStackList(args, logger)
	case "lissto_stack_get":
		return handleStackGet(args, logger)
	case "lissto_stack_create":
		return handleStackCreate(args, logger)
	case "lissto_stack_delete":
		return handleStackDelete(args, logger)

	// Admin tools
	case "lissto_admin_apikey_create":
		return handleAdminAPIKeyCreate(args, logger)

	// Status and logs tools
	case "lissto_status":
		return handleStatus(args, logger)
	case "lissto_logs":
		return handleLogs(args, logger)

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// Helper to get API client from current context
func getAPIClient() (*client.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("no active context. Run 'lissto login' first: %w", err)
	}

	// Create API client with k8s discovery and validation
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize API client: %w", err)
	}
	return apiClient, nil
}

// Helper to get string from args
func getString(args map[string]interface{}, key string, defaultVal string) string {
	if val, ok := args[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

// Helper to get int from args
func getInt(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key]; ok {
		// Handle float64 from JSON unmarshaling
		if f, ok := val.(float64); ok {
			return int(f)
		}
		if i, ok := val.(int); ok {
			return i
		}
	}
	return defaultVal
}

// Environment handlers
func handleEnvList(args map[string]interface{}, logger Logger) (interface{}, error) {
	logger.log("→ handleEnvList: Getting API client")
	apiClient, err := getAPIClient()
	if err != nil {
		logger.log("→ handleEnvList: Failed to get API client: %v", err)
		return nil, err
	}

	logger.log("→ handleEnvList: Calling apiClient.ListEnvs()")
	envs, err := apiClient.ListEnvs()
	if err != nil {
		logger.log("→ handleEnvList: API call failed: %v", err)
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	logger.log("→ handleEnvList: Successfully retrieved %d environments", len(envs))
	result := map[string]interface{}{
		"environments": envs,
		"count":        len(envs),
	}
	return result, nil
}

func handleEnvGet(args map[string]interface{}, logger Logger) (interface{}, error) {
	logger.log("→ handleEnvGet: args=%+v", args)
	name := getString(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	env, err := apiClient.GetEnv(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}

	return env, nil
}

func handleEnvCreate(args map[string]interface{}, logger Logger) (interface{}, error) {
	logger.log("→ handleEnvCreate: args=%+v", args)
	name := getString(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	identifier, err := apiClient.CreateEnv(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	return map[string]interface{}{
		"identifier": identifier,
		"message":    fmt.Sprintf("Environment '%s' created successfully", name),
	}, nil
}

func handleEnvCurrent(args map[string]interface{}, logger Logger) (interface{}, error) {
	logger.log("→ handleEnvCurrent: args=%+v", args)
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return map[string]interface{}{
		"current_env": cfg.CurrentEnv,
		"context":     cfg.CurrentContext,
	}, nil
}

// Blueprint handlers
func handleBlueprintList(args map[string]interface{}, logger Logger) (interface{}, error) {
	// Always include global blueprints (scope determined by the api, not flag)
	logger.log("→ handleBlueprintList: Listing all blueprints (user + global)")

	apiClient, err := getAPIClient()
	if err != nil {
		logger.log("→ handleBlueprintList: Failed to get API client: %v", err)
		return nil, err
	}

	logger.log("→ handleBlueprintList: Calling apiClient.ListBlueprints()")
	blueprints, err := apiClient.ListBlueprints(true)
	if err != nil {
		logger.log("→ handleBlueprintList: API call failed: %v", err)
		return nil, fmt.Errorf("failed to list blueprints: %w", err)
	}

	logger.log("→ handleBlueprintList: Successfully retrieved %d blueprints", len(blueprints))
	return map[string]interface{}{
		"blueprints": blueprints,
		"count":      len(blueprints),
	}, nil
}

func handleBlueprintGet(args map[string]interface{}, logger Logger) (interface{}, error) {
	name := getString(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	blueprint, err := apiClient.GetBlueprint(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get blueprint: %w", err)
	}

	return blueprint, nil
}

func handleBlueprintCreate(args map[string]interface{}, logger Logger) (interface{}, error) {
	compose := getString(args, "compose", "")
	if compose == "" {
		return nil, fmt.Errorf("compose is required")
	}

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	req := client.CreateBlueprintRequest{
		Compose:    compose,
		Branch:     getString(args, "branch", ""),
		Author:     getString(args, "author", ""),
		Repository: getString(args, "repository", ""),
	}

	identifier, err := apiClient.CreateBlueprint(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create blueprint: %w", err)
	}

	return map[string]interface{}{
		"identifier": identifier,
		"message":    "Blueprint created successfully",
	}, nil
}

func handleBlueprintDelete(args map[string]interface{}, logger Logger) (interface{}, error) {
	name := getString(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	if err := apiClient.DeleteBlueprint(name); err != nil {
		return nil, fmt.Errorf("failed to delete blueprint: %w", err)
	}

	return map[string]interface{}{
		"message": fmt.Sprintf("Blueprint '%s' deleted successfully", name),
	}, nil
}

// Stack handlers
func handleStackList(args map[string]interface{}, logger Logger) (interface{}, error) {
	env := getString(args, "env", "")
	logger.log("→ handleStackList: env=%v", env)

	apiClient, err := getAPIClient()
	if err != nil {
		logger.log("→ handleStackList: Failed to get API client: %v", err)
		return nil, err
	}

	logger.log("→ handleStackList: Calling apiClient.ListStacks()")
	stacks, err := apiClient.ListStacks(env)
	if err != nil {
		logger.log("→ handleStackList: API call failed: %v", err)
		return nil, fmt.Errorf("failed to list stacks: %w", err)
	}

	logger.log("→ handleStackList: Successfully retrieved %d stacks", len(stacks))
	return map[string]interface{}{
		"stacks": stacks,
		"count":  len(stacks),
	}, nil
}

func handleStackGet(args map[string]interface{}, logger Logger) (interface{}, error) {
	name := getString(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	env := getString(args, "env", "")

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	identifier, err := apiClient.GetStack(name, env)
	if err != nil {
		return nil, fmt.Errorf("failed to get stack: %w", err)
	}

	return map[string]interface{}{
		"identifier": identifier,
	}, nil
}

func handleStackCreate(args map[string]interface{}, logger Logger) (interface{}, error) {
	blueprintName := getString(args, "blueprint_name", "")
	if blueprintName == "" {
		return nil, fmt.Errorf("blueprint_name is required")
	}

	env := getString(args, "env", "")

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	// First prepare the stack to get request_id
	prepareResp, err := apiClient.PrepareStack(blueprintName, env, "", "", "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare stack: %w", err)
	}

	// Check for missing images
	for _, img := range prepareResp.Images {
		if img.Digest == "" || img.Digest == "N/A" {
			return nil, fmt.Errorf("cannot create stack: service '%s' has missing image", img.Service)
		}
	}

	// Create stack with request_id
	identifier, err := apiClient.CreateStack(blueprintName, env, prepareResp.RequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to create stack: %w", err)
	}

	return map[string]interface{}{
		"identifier": identifier,
		"message":    fmt.Sprintf("Stack created from blueprint '%s'", blueprintName),
	}, nil
}

func handleStackDelete(args map[string]interface{}, logger Logger) (interface{}, error) {
	name := getString(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	env := getString(args, "env", "")

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	if err := apiClient.DeleteStack(name, env); err != nil {
		return nil, fmt.Errorf("failed to delete stack: %w", err)
	}

	return map[string]interface{}{
		"message": fmt.Sprintf("Stack '%s' deleted successfully", name),
	}, nil
}

// Admin handlers
func handleAdminAPIKeyCreate(args map[string]interface{}, logger Logger) (interface{}, error) {
	name := getString(args, "name", "")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	role := getString(args, "role", "user")

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	req := client.CreateAPIKeyRequest{
		Name: name,
		Role: role,
	}

	result, err := apiClient.CreateAPIKey(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return map[string]interface{}{
		"api_key": result.APIKey,
		"name":    result.Name,
		"role":    result.Role,
		"message": "API key created successfully. IMPORTANT: Save this key securely, it cannot be retrieved later.",
	}, nil
}

// Status handler
func handleStatus(args map[string]interface{}, logger Logger) (interface{}, error) {
	envFilter := getString(args, "env", "")

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	// List all stacks
	stacks, err := apiClient.ListStacks("")
	if err != nil {
		return nil, fmt.Errorf("failed to list stacks: %w", err)
	}

	if len(stacks) == 0 {
		return map[string]interface{}{
			"stacks":  []interface{}{},
			"message": "No stacks found",
		}, nil
	}

	// Initialize K8s client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	// Collect status for each stack
	var stackStatuses []map[string]interface{}

	for _, stack := range stacks {
		// Filter by environment if specified
		if envFilter != "" && stack.Namespace != envFilter {
			continue
		}

		// Parse stack status from conditions
		stackStatusParsed := status.ParseStackStatus(stack.Status.Conditions)

		stackStatus := map[string]interface{}{
			"name":      stack.Name,
			"namespace": stack.Namespace,
			"blueprint": stack.Spec.BlueprintReference,
			"state":     stackStatusParsed.State,
			"reason":    stackStatusParsed.Reason,
		}

		// Get pods for this stack using label selector
		labels := map[string]string{
			"lissto.dev/stack": stack.Name,
		}
		pods, err := k8sClient.ListPods(context.Background(), stack.Namespace, labels)
		if err == nil {
			podStatuses := []map[string]interface{}{}
			for _, pod := range pods {
				podStatus := map[string]interface{}{
					"name":   pod.Name,
					"phase":  string(pod.Status.Phase),
					"ready":  isPodReady(&pod),
					"reason": getPodReason(&pod),
				}
				podStatuses = append(podStatuses, podStatus)
			}
			stackStatus["pods"] = podStatuses
			stackStatus["pod_count"] = len(pods)
		}

		stackStatuses = append(stackStatuses, stackStatus)
	}

	return map[string]interface{}{
		"stacks": stackStatuses,
		"count":  len(stackStatuses),
	}, nil
}

// Logs handler
func handleLogs(args map[string]interface{}, logger Logger) (interface{}, error) {
	stackFilter := getString(args, "stack", "")
	envFilter := getString(args, "env", "")
	serviceFilter := getString(args, "service", "")
	podFilter := getString(args, "pod", "")
	tail := int64(getInt(args, "tail", 100))
	maxPods := getInt(args, "max_pods", 5)

	apiClient, err := getAPIClient()
	if err != nil {
		return nil, err
	}

	// List stacks to filter
	stacks, err := apiClient.ListStacks(envFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list stacks: %w", err)
	}

	// Initialize K8s client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	var logEntries []map[string]interface{}
	podsProcessed := 0

	for _, stack := range stacks {
		// Filter by stack name if specified
		if stackFilter != "" && stack.Name != stackFilter {
			continue
		}

		// Get pods for this stack using label selector
		labels := map[string]string{
			"lissto.dev/stack": stack.Name,
		}
		pods, err := k8sClient.ListPods(context.Background(), stack.Namespace, labels)
		if err != nil {
			continue
		}

		for _, pod := range pods {
			if podsProcessed >= maxPods {
				break
			}

			// Filter by pod name if specified
			if podFilter != "" && pod.Name != podFilter {
				continue
			}

			// Filter by service label if specified
			if serviceFilter != "" {
				if serviceName, ok := pod.Labels["app"]; !ok || serviceName != serviceFilter {
					continue
				}
			}

			// Get logs for each container in the pod
			for _, container := range pod.Spec.Containers {
				// Stream logs using k8s client
				opts := k8s.LogOptions{
					Follow:     false,
					Timestamps: false,
					TailLines:  &tail,
					Container:  container.Name,
				}

				stream, err := k8sClient.StreamLogs(context.Background(), pod.Namespace, pod.Name, opts)
				if err != nil {
					continue
				}

				// Read all logs from stream
				logBytes, err := io.ReadAll(stream)
				stream.Close()
				if err != nil {
					continue
				}

				logEntry := map[string]interface{}{
					"stack":     stack.Name,
					"namespace": pod.Namespace,
					"pod":       pod.Name,
					"container": container.Name,
					"logs":      string(logBytes),
				}

				if serviceName, ok := pod.Labels["app"]; ok {
					logEntry["service"] = serviceName
				}

				logEntries = append(logEntries, logEntry)
			}

			podsProcessed++
		}
	}

	return map[string]interface{}{
		"log_entries":    logEntries,
		"count":          len(logEntries),
		"pods_processed": podsProcessed,
	}, nil
}

// Helper functions for pod status
func isPodReady(pod *corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

func getPodReason(pod *corev1.Pod) string {
	// Check container statuses for reasons
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			return cs.State.Waiting.Reason
		}
		if cs.State.Terminated != nil {
			return cs.State.Terminated.Reason
		}
	}

	// Check pod conditions
	for _, cond := range pod.Status.Conditions {
		if cond.Status != corev1.ConditionTrue && cond.Reason != "" {
			return cond.Reason
		}
	}

	return string(pod.Status.Phase)
}
