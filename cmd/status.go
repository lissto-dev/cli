package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/k8s"
	"github.com/lissto-dev/cli/pkg/output"
	"github.com/lissto-dev/cli/pkg/status"
	"github.com/lissto-dev/cli/pkg/types"
	envv1alpha1 "github.com/lissto-dev/controller/api/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
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
	format := getOutputFormat(cmd)

	// Handle different output formats
	switch format {
	case "json":
		return output.PrintJSON(os.Stdout, stacks)
	case "yaml":
		return output.PrintYAML(os.Stdout, stacks)
	case "table":
		return printTableStatus(envGroups)
	default:
		return printPrettyStatus(envGroups)
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
	var envs []string
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
				if podStatus == "Unknown" {
					stackStatus.State = "Unknown"
					hasUnknown = true
				} else if podStatus == "Error" {
					stackStatus.State = "Error"
					hasErrors = true
				} else if podStatus == "Pending" {
					stackStatus.State = "Deploying"
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
		fmt.Fprintln(os.Stdout, "\nℹ️  Some pods are in error state. Use 'lissto status -o pretty' for details.")
	}

	// Show hint if there are unknown statuses
	if hasUnknown {
		fmt.Fprintln(os.Stdout, "\n⚠️  Could not find pods for some stacks. Check your cluster context with 'kubectl config current-context'")
	}

	return nil
}

// printPrettyStatus prints detailed format with emojis and pod status
func printPrettyStatus(envGroups map[string][]envv1alpha1.Stack) error {
	printer := output.NewPrettyPrinter(os.Stdout)

	// Try to create k8s client (may fail if no kubeconfig)
	k8sClient, err := k8s.NewClient()
	k8sAvailable := err == nil

	if !k8sAvailable {
		fmt.Fprintf(os.Stderr, "⚠️  Kubernetes access unavailable - pod details not shown\n")
		fmt.Fprintf(os.Stderr, "   Error: %v\n\n", err)
	}

	// Sort environments for consistent output
	var envs []string
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
			fmt.Fprintf(os.Stdout, "Stack: %s\n", stackDisplay)

			// Stack status - check actual pod status if k8s available
			stackStatus := status.ParseStackStatus(stack.Status.Conditions)
			if k8sAvailable {
				podStatus := checkStackPodsStatus(k8sClient, &stack)
				if podStatus == "Unknown" {
					stackStatus.State = "Unknown"
					stackStatus.Symbol = "❓"
					stackStatus.Reason = "Can't find pods - check cluster context"
				} else if podStatus == "Error" {
					stackStatus.State = "Error"
					stackStatus.Symbol = "❌"
					stackStatus.Reason = "Pod issues detected"
				} else if podStatus == "Pending" {
					stackStatus.State = "Deploying"
					stackStatus.Symbol = "⏳"
					stackStatus.Reason = "Pods starting"
				}
			}

			fmt.Fprintf(os.Stdout, "Status: %s %s", stackStatus.Symbol, stackStatus.State)
			if stackStatus.Reason != "" {
				fmt.Fprintf(os.Stdout, " (%s)", stackStatus.Reason)
			}
			fmt.Fprintf(os.Stdout, "\n")

			// Creation time
			formatted, timeAgo := output.FormatTimestamp(stack.CreationTimestamp.Time)
			fmt.Fprintf(os.Stdout, "Created: %s (%s)\n", formatted, timeAgo)

			// Services
			services := status.ParseServiceStatuses(&stack)
			if len(services) > 0 {
				printer.PrintNewline()
				fmt.Fprintf(os.Stdout, "Services:\n")

				// Sort services by name
				sort.Slice(services, func(i, j int) bool {
					return services[i].Name < services[j].Name
				})

				for _, svc := range services {
					// Check actual pod status if k8s is available
					serviceSymbol := svc.Symbol
					if k8sAvailable {
						pods, err := fetchServicePods(k8sClient, &stack, svc.Name)
						if err == nil && len(pods) > 0 {
							// Update service symbol based on actual pod status
							serviceSymbol = getServiceSymbolFromPods(pods)
						}
					}

					printer.PrintSubSection(serviceSymbol, svc.Name)

					// Image
					if svc.Image != "" {
						printer.PrintIndentedLine(2, fmt.Sprintf("🐳 %s", svc.Image))
					}

					// URL
					if svc.URL != "" {
						printer.PrintIndentedLine(2, fmt.Sprintf("🔗 https://%s", svc.URL))
					}

					// Pod status (if k8s available)
					if k8sAvailable {
						pods, err := fetchServicePods(k8sClient, &stack, svc.Name)
						if err == nil && len(pods) > 0 {
							printer.PrintIndentedLine(2, "Pods:")
							for _, pod := range pods {
								podStatus := k8s.ParsePodStatus(&pod)
								podLine := fmt.Sprintf("%s | %s | %s | %s",
									podStatus.Name,
									podStatus.Phase,
									k8s.FormatRestarts(podStatus.Restarts),
									k8s.FormatAge(podStatus.Age))
								printer.PrintBullet(3, podLine)
							}
						}
					}

					printer.PrintNewline()
				}
			} else {
				printer.PrintNewline()
				printer.PrintIndentedLine(1, "No services configured")
			}
		}
	}

	// Show helpful hints
	printer.PrintNewline()
	fmt.Fprintln(os.Stdout, "💡 Tip: Use 'lissto logs' to view logs, 'lissto update' to update images")

	return nil
}

// getServiceSymbolFromPods determines the service status symbol based on pod states
func getServiceSymbolFromPods(pods []corev1.Pod) string {
	if len(pods) == 0 {
		return "❓"
	}

	allRunning := true
	hasError := false

	for _, pod := range pods {
		phase := pod.Status.Phase

		// Check for failures
		if phase == corev1.PodFailed {
			return "❌"
		}

		// Check if pod is not running
		if phase != corev1.PodRunning {
			allRunning = false
		}

		// Check container statuses
		for _, cs := range pod.Status.ContainerStatuses {
			// Check for error states
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				if reason == "CrashLoopBackOff" ||
					reason == "ImagePullBackOff" ||
					reason == "ErrImagePull" ||
					reason == "CreateContainerError" ||
					reason == "InvalidImageName" {
					hasError = true
				}
			}

			// Check if container terminated with error
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
		return "❌"
	}

	if allRunning {
		return "✅"
	}

	return "⏳"
}

// checkStackPodsStatus checks the overall pod status for a stack
// Returns: "Ready", "Pending", "Error", or "Unknown"
func checkStackPodsStatus(k8sClient *k8s.Client, stack *envv1alpha1.Stack) string {
	ctx := context.Background()

	// Query all pods for this stack
	labels := map[string]string{
		"lissto.dev/stack": stack.Name,
	}

	pods, err := k8sClient.ListPods(ctx, stack.Namespace, labels)
	if err != nil {
		// Error accessing pods (e.g., wrong cluster context, no permissions)
		return "Unknown"
	}

	if len(pods) == 0 {
		// No pods found - likely wrong cluster or stack failed to deploy
		return "Unknown"
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
		return "Error"
	}

	if hasPending || !allRunning {
		return "Pending"
	}

	return "Ready"
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

func getOutputFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("output")
	if format == "" {
		format = "pretty"
	}
	return format
}
