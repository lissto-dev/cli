package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/k8s"
	envv1alpha1 "github.com/lissto-dev/controller/api/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var (
	logsFollow     bool
	logsTimestamps bool
	logsTail       int64
	logsSince      string
	logsStack      string
	logsService    string
	logsPod        string
	logsContainer  string
	logsEnv        string
	logsMaxPods    int
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream logs from stack pods",
	Long: `Stream logs from pods. By default streams from all stacks.

Use filters to narrow down what logs to stream:
  --stack      Filter by stack name
  --env        Filter by environment
  --service    Filter by service name
  --pod        Filter by specific pod name
  --container  Filter by container name
  --max-pods   Maximum number of pods to stream (default 10)

Examples:
  # Stream logs from all stacks (default)
  lissto logs

  # Stream logs from all stacks in an environment
  lissto logs --env dev

  # Stream logs from a specific stack
  lissto logs --stack my-stack

  # Follow logs from a specific service across all stacks
  lissto logs --service frontend -f

  # Combine filters
  lissto logs --env dev --service api --tail 100

  # Show last 100 lines from specific pod
  lissto logs --pod frontend-abc123 --tail 100

  # Show logs from last 5 minutes
  lissto logs --since 5m

  # Allow more pods to stream
  lissto logs --max-pods 50`,
	Args:          cobra.NoArgs,
	RunE:          runLogs,
	SilenceUsage:  true,
	SilenceErrors: false,
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().BoolVar(&logsTimestamps, "timestamps", false, "Include timestamps in output")
	logsCmd.Flags().Int64Var(&logsTail, "tail", 10, "Number of lines to show from end of logs (use -1 for all)")
	logsCmd.Flags().StringVar(&logsSince, "since", "", "Show logs since duration (e.g., 5s, 2m, 3h)")
	logsCmd.Flags().StringVar(&logsStack, "stack", "", "Filter by stack name")
	logsCmd.Flags().StringVar(&logsService, "service", "", "Filter by service name")
	logsCmd.Flags().StringVar(&logsPod, "pod", "", "Filter by specific pod name")
	logsCmd.Flags().StringVar(&logsContainer, "container", "", "Filter by container name")
	logsCmd.Flags().StringVar(&logsEnv, "env", "", "Filter by environment")
	logsCmd.Flags().IntVar(&logsMaxPods, "max-pods", 10, "Maximum number of pods to stream logs from")
}

func runLogs(cmd *cobra.Command, args []string) error {
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

	// Get stacks
	allStacks, err := apiClient.ListStacks("")
	if err != nil {
		return fmt.Errorf("failed to list stacks: %w", err)
	}

	if len(allStacks) == 0 {
		return fmt.Errorf("no stacks found")
	}

	// Filter stacks by provided filters
	var targetStacks []interface{}
	for _, stack := range allStacks {
		// Filter by stack name
		if logsStack != "" && stack.Name != logsStack {
			continue
		}

		// Filter by environment
		if logsEnv != "" && stack.Spec.Env != logsEnv {
			continue
		}

		targetStacks = append(targetStacks, stack)
	}

	if len(targetStacks) == 0 {
		return fmt.Errorf("no stacks match the filters")
	}

	// Create k8s client
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Collect all pods from target stacks
	podCtx := context.Background()
	var allPods []corev1.Pod

	for _, s := range targetStacks {
		stack := s.(envv1alpha1.Stack)

		labels := map[string]string{
			"lissto.dev/stack": stack.Name,
		}

		pods, err := k8sClient.ListPods(podCtx, stack.Namespace, labels)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to list pods for stack %s: %v\n", stack.Name, err)
			continue
		}

		allPods = append(allPods, pods...)
	}

	if len(allPods) == 0 {
		return fmt.Errorf("no pods found")
	}

	// Filter pods by service/pod name
	filteredPods := filterPods(allPods, logsService, logsPod)
	if len(filteredPods) == 0 {
		return fmt.Errorf("no pods match the filters")
	}

	// Check max-pods limit
	if len(filteredPods) > logsMaxPods {
		return fmt.Errorf("found %d pods but max-pods is set to %d. Use --max-pods to increase the limit or add filters (--service, --pod, --env)",
			len(filteredPods), logsMaxPods)
	}

	// Parse log options
	logOpts := k8s.LogOptions{
		Follow:     logsFollow,
		Timestamps: logsTimestamps,
		Container:  logsContainer,
	}

	if logsTail >= 0 {
		logOpts.TailLines = &logsTail
	}

	if logsSince != "" {
		duration, err := parseDuration(logsSince)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
		logOpts.Since = &duration
	}

	// Setup signal handling for graceful shutdown
	logCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nShutting down...")
		cancel()
	}()

	// Stream logs
	logChan := make(chan k8s.LogLine, 100)
	errChan := make(chan error, 1)

	// Show info about what we're streaming
	if logsStack != "" {
		fmt.Fprintf(os.Stderr, "📡 Streaming logs from %d pod(s) in stack '%s'...\n", len(filteredPods), logsStack)
	} else {
		fmt.Fprintf(os.Stderr, "📡 Streaming logs from %d pod(s) across %d stack(s)...\n", len(filteredPods), len(targetStacks))
	}
	if logsFollow {
		fmt.Fprintln(os.Stderr, "Press Ctrl+C to stop.")
	}
	fmt.Fprintln(os.Stderr)

	go func() {
		// Group pods by namespace for streaming
		podsByNamespace := make(map[string][]corev1.Pod)
		for _, pod := range filteredPods {
			podsByNamespace[pod.Namespace] = append(podsByNamespace[pod.Namespace], pod)
		}

		// Stream from each namespace
		var streamErr error
		for namespace, pods := range podsByNamespace {
			err := k8sClient.StreamLogsMulti(logCtx, namespace, pods, logOpts, logChan)
			if err != nil {
				streamErr = err
			}
		}

		errChan <- streamErr
		close(logChan)
	}()

	// Print logs
	colors := []string{
		"\033[36m", // Cyan
		"\033[33m", // Yellow
		"\033[35m", // Magenta
		"\033[32m", // Green
		"\033[34m", // Blue
		"\033[31m", // Red
	}
	reset := "\033[0m"

	podColors := make(map[string]string)
	colorIdx := 0

	for logLine := range logChan {
		// Assign color to pod if not already assigned
		if _, exists := podColors[logLine.PodName]; !exists {
			podColors[logLine.PodName] = colors[colorIdx%len(colors)]
			colorIdx++
		}

		color := podColors[logLine.PodName]
		prefix := fmt.Sprintf("%s[%s]%s", color, logLine.PodName, reset)

		if logsContainer == "" && logLine.Container != "" {
			prefix = fmt.Sprintf("%s[%s/%s]%s", color, logLine.PodName, logLine.Container, reset)
		}

		fmt.Fprintf(os.Stdout, "%s %s\n", prefix, logLine.Message)
	}

	// Check for errors
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

// filterPods filters pods by service name or pod name
func filterPods(pods []corev1.Pod, serviceName, podName string) []corev1.Pod {
	if podName != "" {
		// Filter by specific pod name
		for _, pod := range pods {
			if pod.Name == podName {
				return []corev1.Pod{pod}
			}
		}
		return nil
	}

	if serviceName != "" {
		// Filter by service name using labels or name prefix
		var filtered []corev1.Pod
		for _, pod := range pods {
			if pod.Labels != nil && pod.Labels["lissto.dev/service"] == serviceName {
				filtered = append(filtered, pod)
				continue
			}
			if pod.Labels != nil && pod.Labels["io.kompose.service"] == serviceName {
				filtered = append(filtered, pod)
				continue
			}
			if strings.HasPrefix(pod.Name, serviceName+"-") {
				filtered = append(filtered, pod)
			}
		}
		return filtered
	}

	return pods
}

// parseDuration parses duration strings like "5s", "2m", "3h"
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
