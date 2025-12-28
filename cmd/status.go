package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/cmdutil"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/k8s"
	"github.com/lissto-dev/cli/pkg/output"
	"github.com/lissto-dev/cli/pkg/status"
	"github.com/lissto-dev/cli/pkg/types"
	envv1alpha1 "github.com/lissto-dev/controller/api/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

// Output format constants
const (
	outputFormatJSON  = "json"
	outputFormatYAML  = "yaml"
	outputFormatTable = "table"
)

// Pod status constants
const (
	podStatusError   = "Error"
	podStatusPending = "Pending"
)

var (
	statusEnvFilter string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of active environments and stacks",
	Long: `Display the status of all active environments and their stacks.
	
Shows deployment status, services, and pod-level details.

Output formats:
  (default)    Detailed view with emojis and pod status
  -o table     Compact table view
  -o json      Raw JSON output
  -o yaml      Raw YAML output`,
	RunE:          runStatus,
	SilenceUsage:  true,
	SilenceErrors: false,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringVar(&statusEnvFilter, "env", "", "Filter by environment name")
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current context
	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return fmt.Errorf("no active context. Run 'lissto login' first: %w", err)
	}

	// Create API client with k8s discovery and validation
	apiClient, err := client.NewClientFromConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// List all stacks (pass empty string to get all)
	stacks, err := apiClient.ListStacks("")
	if err != nil {
		return fmt.Errorf("failed to list stacks: %w", err)
	}

	if len(stacks) == 0 {
		fmt.Println("No stacks found.")
		fmt.Println("Use 'lissto create' to create a new stack.")
		return nil
	}

	// Group stacks by environment
	envGroups := groupStacksByEnv(stacks, statusEnvFilter)

	if len(envGroups) == 0 {
		if statusEnvFilter != "" {
			return fmt.Errorf("no stacks found in environment '%s'", statusEnvFilter)
		}
		return fmt.Errorf("no stacks found")
	}

	// Get output format
	format := cmdutil.GetOutputFormat(cmd)

	// Handle different output formats
	switch format {
	case outputFormatJSON:
		return output.PrintJSON(os.Stdout, stacks)
	case outputFormatYAML:
		return output.PrintYAML(os.Stdout, stacks)
	case outputFormatTable:
		return printTableStatus(envGroups)
	default:
		return printPrettyStatus(envGroups, apiClient)
	}
}

// groupStacksByEnv groups stacks by environment name
func groupStacksByEnv(stacks []envv1alpha1.Stack, envFilter string) map[string][]envv1alpha1.Stack {
	groups := make(map[string][]envv1alpha1.Stack)

	for _, stack := range stacks {
		env := stack.Spec.Env
		if env == "" {
			env = "unknown"
		}

		// Apply filter if specified
		if envFilter != "" && env != envFilter {
			continue
		}

		groups[env] = append(groups[env], stack)
	}

	return groups
}

// printTableStatus prints compact table format
func printTableStatus(envGroups map[string][]envv1alpha1.Stack) error {
	headers := []string{"ENV", "STACK", "STATUS", "SERVICES", "AGE"}
	var rows [][]string

	// Try to create k8s client for pod status checking
	k8sClient, _ := k8s.NewClient()
	hasErrors := false
	hasUnknown := false

	// Sort environments for consistent output
	envs := make([]string, 0, len(envGroups))
	for env := range envGroups {
		envs = append(envs, env)
	}
	sort.Strings(envs)

	for _, env := range envs {
		stacks := envGroups[env]

		// Sort stacks by creation time (newest first)
		sort.Slice(stacks, func(i, j int) bool {
			return stacks[i].CreationTimestamp.After(stacks[j].CreationTimestamp.Time)
		})

		for _, stack := range stacks {
			// Parse stack status
			stackStatus := status.ParseStackStatus(stack.Status.Conditions)

			// Check pod status if k8s client is available
			if k8sClient != nil {
				podStatus := checkStackPodsStatus(k8sClient, &stack)
				switch podStatus {
				case status.StateUnknown:
					stackStatus.State = status.StateUnknown
					hasUnknown = true
				case podStatusError:
					stackStatus.State = podStatusError
					hasErrors = true
				case podStatusPending:
					stackStatus.State = status.StateDeploying
				}
			}

			// Get stack display name (blueprint title if available, otherwise stack name)
			stackDisplay := types.GetStackDisplayName(&stack)

			// Parse service statuses
			services := status.ParseServiceStatuses(&stack)
			ready, total := status.CountReadyServices(services)
			servicesStr := fmt.Sprintf("%d/%d", ready, total)

			// Calculate age
			age := time.Since(stack.CreationTimestamp.Time)
			ageStr := k8s.FormatAge(age)

			rows = append(rows, []string{
				env,
				stackDisplay,
				stackStatus.State,
				servicesStr,
				ageStr,
			})
		}
	}

	output.PrintTable(os.Stdout, headers, rows)

	// Show hint if there are errors
	if hasErrors {
		_, _ = fmt.Fprintln(os.Stdout, "\nℹ️  Some pods are in error state. Use 'lissto status -o pretty' for details.")
	}

	// Show hint if there are unknown statuses
	if hasUnknown {
		_, _ = fmt.Fprintln(os.Stdout, "\n⚠️  Could not find pods for some stacks. Check your cluster context with 'kubectl config current-context'")
	}

	return nil
}

// printPrettyStatus prints detailed format with emojis and pod status
func printPrettyStatus(envGroups map[string][]envv1alpha1.Stack, apiClient *client.Client) error {
	printer := output.NewPrettyPrinter(os.Stdout)

	// Try to create k8s client (may fail if no kubeconfig)
	k8sClient, err := k8s.NewClient()
	k8sAvailable := err == nil

	if !k8sAvailable {
		fmt.Fprintf(os.Stderr, "⚠️  Kubernetes access unavailable - pod details not shown\n")
		fmt.Fprintf(os.Stderr, "   Error: %v\n\n", err)
	}

	// Sort environments for consistent output
	envs := make([]string, 0, len(envGroups))
	for env := range envGroups {
		envs = append(envs, env)
	}
	sort.Strings(envs)

	for envIdx, env := range envs {
		if envIdx > 0 {
			printer.PrintNewline()
		}

		printer.PrintHeader(fmt.Sprintf("Environment: %s", env))

		stacks := envGroups[env]

		// Sort stacks by creation time (newest first)
		sort.Slice(stacks, func(i, j int) bool {
			return stacks[i].CreationTimestamp.After(stacks[j].CreationTimestamp.Time)
		})

		for stackIdx, stack := range stacks {
			if stackIdx > 0 {
				printer.PrintDivider()
			}

			// Stack header with blueprint title if available
			printer.PrintNewline()
			stackDisplay := types.GetStackDisplayName(&stack)
			_, _ = fmt.Fprintf(os.Stdout, "Stack: %s\n", stackDisplay)

			// Stack status - check actual pod status if k8s available
			stackStatus := status.ParseStackStatus(stack.Status.Conditions)
			if k8sAvailable {
				podStatus := checkStackPodsStatus(k8sClient, &stack)
				switch podStatus {
				case status.StateUnknown:
					stackStatus.State = status.StateUnknown
					stackStatus.Symbol = status.SymbolUnknown
					stackStatus.Reason = "Can't find pods - check cluster context"
				case podStatusError:
					stackStatus.State = podStatusError
					stackStatus.Symbol = status.SymbolFailed
					stackStatus.Reason = "Pod issues detected"
				case podStatusPending:
					stackStatus.State = status.StateDeploying
					stackStatus.Symbol = status.SymbolDeploying
					stackStatus.Reason = "Pods starting"
				}
			}

			_, _ = fmt.Fprintf(os.Stdout, "Status: %s %s", stackStatus.Symbol, stackStatus.State)
			if stackStatus.Reason != "" {
				_, _ = fmt.Fprintf(os.Stdout, " (%s)", stackStatus.Reason)
			}
			_, _ = fmt.Fprintf(os.Stdout, "\n")

			// Creation time
			formatted, timeAgo := output.FormatTimestamp(stack.CreationTimestamp.Time)
			_, _ = fmt.Fprintf(os.Stdout, "Created: %s (%s)\n", formatted, timeAgo)

			// Parse services
			services := status.ParseServiceStatuses(&stack)
			if len(services) == 0 {
				printer.PrintNewline()
				printer.PrintIndentedLine(1, "No services configured")
				continue
			}

			// Fetch blueprint for categorization
			blueprintContent := fetchBlueprintMetadata(apiClient, stack.Spec.BlueprintReference)

			// 1. Display URLs table
			displayURLsTable(&stack, services, k8sClient, k8sAvailable)

			// 2. Categorize services
			regularServices, jobs, infra := categorizeServices(services, k8sClient, &stack, k8sAvailable, blueprintContent)

			// 3. Display categorized pods tables with category-specific headers
			_, _ = fmt.Fprintf(os.Stdout, "\n")
			displayCategorizedPodsTable(regularServices, jobs, infra, k8sClient, &stack, k8sAvailable)
		}
	}

	// Show helpful hints
	printer.PrintNewline()
	_, _ = fmt.Fprintln(os.Stdout, "💡 Tip: Use 'lissto logs' to view logs, 'lissto update' to update images")

	return nil
}

// fetchBlueprintMetadata fetches blueprint service metadata for categorization
func fetchBlueprintMetadata(apiClient *client.Client, blueprintRef string) *client.ServiceMetadata {
	if apiClient == nil || blueprintRef == "" {
		return nil
	}

	// API now accepts scoped IDs directly
	blueprint, err := apiClient.GetBlueprint(blueprintRef)
	if err != nil {
		return nil
	}

	return &blueprint.Content
}

// displayURLsTable displays services with exposed URLs
func displayURLsTable(stack *envv1alpha1.Stack, services []status.ServiceStatus, k8sClient *k8s.Client, k8sAvailable bool) {
	// Filter services with URLs
	type urlRow struct {
		Service string
		URL     string
		Ready   string
		Age     string
	}

	urlServices := make([]urlRow, 0, len(services))
	for _, svc := range services {
		if svc.URL == "" {
			continue
		}

		// Calculate service age from stack creation timestamp
		serviceAge := time.Since(stack.CreationTimestamp.Time)
		ageStr := k8s.FormatAge(serviceAge)

		// Default ready status
		readyStatus := "⚪ (unknown)"

		if k8sAvailable {
			// Fetch pods for this service
			pods, err := fetchServicePods(k8sClient, stack, svc.Name)
			if err == nil {
				// Check service readiness (Service, Endpoints, Ingress, Pods)
				ctx := context.Background()
				readiness := k8sClient.CheckServiceReadiness(ctx, stack.Namespace, svc.Name, pods, serviceAge)
				readyStatus = k8s.FormatReadinessStatus(readiness, serviceAge)
			} else if serviceAge < time.Minute {
				readyStatus = "⚪ (starting up..)"
			} else {
				readyStatus = "⚪ (unknown)"
			}
		} else if serviceAge < time.Minute {
			readyStatus = "⚪ (starting up..)"
		}

		urlServices = append(urlServices, urlRow{
			Service: svc.Name,
			URL:     fmt.Sprintf("https://%s", svc.URL),
			Ready:   readyStatus,
			Age:     ageStr,
		})
	}

	if len(urlServices) == 0 {
		return
	}

	// Sort by service name
	sort.Slice(urlServices, func(i, j int) bool {
		return urlServices[i].Service < urlServices[j].Service
	})

	headers := []string{"NAME", "URL", "READY", "AGE"}
	rows := make([][]string, 0, len(urlServices))
	for _, u := range urlServices {
		rows = append(rows, []string{u.Service, u.URL, u.Ready, u.Age})
	}
	output.PrintTable(os.Stdout, headers, rows)
}

// displayCategorizedPodsTable displays all pods in a single table with category headers
func displayCategorizedPodsTable(services, jobs, infra []status.ServiceStatus, k8sClient *k8s.Client, stack *envv1alpha1.Stack, k8sAvailable bool) {
	if !k8sAvailable {
		return
	}

	// Display regular services
	if len(services) > 0 {
		headers := []string{"SERVICE", "POD NAME", "STATUS", "RESTARTS", "AGE"}
		rows := buildPodRows(services, k8sClient, stack, false)
		if len(rows) > 0 {
			output.PrintTable(os.Stdout, headers, rows)
		}
	}

	// Display infrastructure
	if len(infra) > 0 {
		if len(services) > 0 {
			_, _ = fmt.Fprintf(os.Stdout, "\n")
		}
		headers := []string{"INFRA", "POD NAME", "STATUS", "RESTARTS", "AGE"}
		rows := buildPodRows(infra, k8sClient, stack, false)
		if len(rows) > 0 {
			output.PrintTable(os.Stdout, headers, rows)
		}
	}

	// Display jobs
	if len(jobs) > 0 {
		if len(services) > 0 || len(infra) > 0 {
			_, _ = fmt.Fprintf(os.Stdout, "\n")
		}
		headers := []string{"JOBS", "POD NAME", "STATUS", "RESTARTS", "AGE"}
		rows := buildPodRows(jobs, k8sClient, stack, true)
		if len(rows) > 0 {
			output.PrintTable(os.Stdout, headers, rows)
		}
	}
}

// buildPodRows builds table rows for a list of services
func buildPodRows(services []status.ServiceStatus, k8sClient *k8s.Client, stack *envv1alpha1.Stack, isJobGroup bool) [][]string {
	var rows [][]string

	for _, svc := range services {
		pods, err := fetchServicePods(k8sClient, stack, svc.Name)
		if err != nil || len(pods) == 0 {
			// Show service with no pods
			rows = append(rows, []string{
				svc.Name,
				"❓ No pods found",
				"-",
				"-",
				"-",
			})
			continue
		}

		for _, pod := range pods {
			podStatus := k8s.ParsePodStatus(&pod)

			// Check if completed (for graying)
			isCompleted := isJobGroup && pod.Status.Phase == corev1.PodSucceeded

			// Format fields
			serviceName := svc.Name
			podName := podStatus.Name
			phase := podStatus.Phase
			restarts := formatRestartCountWithHelpers(podStatus.Restarts, isCompleted)
			age := k8s.FormatAge(podStatus.Age)

			if isCompleted {
				// Gray out completed jobs
				serviceName = output.Gray(serviceName)
				podName = output.Gray(podName)
				phase = output.Gray(phase)
				age = output.Gray(age)
			}

			rows = append(rows, []string{
				serviceName,
				podName,
				phase,
				restarts,
				age,
			})
		}
	}

	// For jobs, sort so failed/active jobs appear first
	if isJobGroup && len(rows) > 0 {
		sort.SliceStable(rows, func(i, j int) bool {
			// Check if rows contain gray ANSI codes (completed jobs)
			iIsCompleted := strings.Contains(rows[i][2], "\033[90m") // Check STATUS column
			jIsCompleted := strings.Contains(rows[j][2], "\033[90m")

			// Non-completed (active/failed) jobs should come first
			if iIsCompleted != jIsCompleted {
				return !iIsCompleted // i comes first if it's NOT completed
			}
			// Otherwise maintain original order
			return false
		})
	}

	return rows
}

// formatRestartCountWithHelpers formats restart count with yellow highlighting if > 0
func formatRestartCountWithHelpers(restarts int32, isCompleted bool) string {
	countStr := fmt.Sprintf("%d", restarts)
	if isCompleted {
		return output.Gray(countStr)
	}
	if restarts > 0 {
		// Yellow color for restarted pods
		return output.Yellow(countStr)
	}
	return countStr
}

// checkStackPodsStatus checks the overall pod status for a stack
// Returns: status.StateReady, podStatusPending, podStatusError, or status.StateUnknown
func checkStackPodsStatus(k8sClient *k8s.Client, stack *envv1alpha1.Stack) string {
	ctx := context.Background()

	// Query all pods for this stack
	labels := map[string]string{
		"lissto.dev/stack": stack.Name,
	}

	pods, err := k8sClient.ListPods(ctx, stack.Namespace, labels)
	if err != nil {
		// Error accessing pods (e.g., wrong cluster context, no permissions)
		return status.StateUnknown
	}

	if len(pods) == 0 {
		// No pods found - likely wrong cluster or stack failed to deploy
		return status.StateUnknown
	}

	hasError := false
	hasPending := false
	allRunning := true

	// Check all pods
	for _, pod := range pods {
		phase := pod.Status.Phase

		// Check for explicit failure
		if phase == corev1.PodFailed {
			hasError = true
			continue
		}

		// Check for pending state
		if phase == corev1.PodPending {
			hasPending = true
			allRunning = false
			continue
		}

		// Check if running
		if phase != corev1.PodRunning {
			allRunning = false
		}

		// Check container statuses for any issues
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				// Check for error states
				if reason == "CrashLoopBackOff" ||
					reason == "ImagePullBackOff" ||
					reason == "ErrImagePull" ||
					reason == "CreateContainerError" ||
					reason == "InvalidImageName" {
					hasError = true
				} else {
					// Other waiting reasons mean still starting
					hasPending = true
					allRunning = false
				}
			}
			// Check if container has terminated
			if cs.State.Terminated != nil && cs.State.Terminated.ExitCode != 0 {
				hasError = true
			}
			// Check if container is not ready
			if !cs.Ready {
				allRunning = false
			}
		}
	}

	if hasError {
		return podStatusError
	}

	if hasPending || !allRunning {
		return podStatusPending
	}

	return status.StateReady
}

// fetchServicePods queries k8s for pods belonging to a service
func fetchServicePods(k8sClient *k8s.Client, stack *envv1alpha1.Stack, serviceName string) ([]corev1.Pod, error) {
	ctx := context.Background()

	// Query pods with stack label only
	labels := map[string]string{
		"lissto.dev/stack": stack.Name,
	}

	pods, err := k8sClient.ListPods(ctx, stack.Namespace, labels)
	if err != nil {
		return nil, err
	}

	// Filter pods by service name using multiple matching strategies
	var servicePods []corev1.Pod
	for _, pod := range pods {
		matched := false

		// Strategy 1: Check lissto.dev/service label
		if pod.Labels != nil && pod.Labels["lissto.dev/service"] == serviceName {
			matched = true
		}

		// Strategy 2: Check io.kompose.service label (from kompose conversion)
		if !matched && pod.Labels != nil && pod.Labels["io.kompose.service"] == serviceName {
			matched = true
		}

		// Strategy 3: Pod name prefix matching (e.g., "bo-67db85fc78-lhs9t" matches "bo")
		if !matched && strings.HasPrefix(pod.Name, serviceName+"-") {
			matched = true
		}

		if matched {
			servicePods = append(servicePods, pod)
		}
	}

	return servicePods, nil
}

// categorizeServices categorizes services into regular services, jobs, and infra
func categorizeServices(services []status.ServiceStatus, k8sClient *k8s.Client, stack *envv1alpha1.Stack, k8sAvailable bool, blueprintContent *client.ServiceMetadata) (regularServices, jobs, infra []status.ServiceStatus) {
	// Create lookup map for infrastructure services from blueprint
	infraMap := make(map[string]bool)
	if blueprintContent != nil {
		for _, name := range blueprintContent.Infra {
			infraMap[name] = true
		}
	}

	for _, svc := range services {
		// Determine service category based on pod characteristics
		if k8sAvailable {
			pods, err := fetchServicePods(k8sClient, stack, svc.Name)
			if err == nil && len(pods) > 0 {
				pod := pods[0] // Check first pod to determine type

				// Check restart policy to identify jobs
				if pod.Spec.RestartPolicy == corev1.RestartPolicyNever ||
					pod.Spec.RestartPolicy == corev1.RestartPolicyOnFailure {
					jobs = append(jobs, svc)
					continue
				}
			}
		}

		// Check if it's an infra component (from blueprint)
		if infraMap[svc.Name] {
			infra = append(infra, svc)
		} else {
			regularServices = append(regularServices, svc)
		}
	}

	// Sort each category by name
	sort.Slice(regularServices, func(i, j int) bool {
		return regularServices[i].Name < regularServices[j].Name
	})
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Name < jobs[j].Name
	})
	sort.Slice(infra, func(i, j int) bool {
		return infra[i].Name < infra[j].Name
	})

	return regularServices, jobs, infra
}
