package k8s

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// LogOptions contains options for streaming logs
type LogOptions struct {
	Follow     bool
	Timestamps bool
	TailLines  *int64
	Since      *time.Duration
	Container  string
}

// StreamLogs streams logs from a pod/container
func (c *Client) StreamLogs(ctx context.Context, namespace, podName string, opts LogOptions) (io.ReadCloser, error) {
	podLogOpts := &corev1.PodLogOptions{
		Follow:     opts.Follow,
		Timestamps: opts.Timestamps,
	}

	if opts.TailLines != nil {
		podLogOpts.TailLines = opts.TailLines
	}

	if opts.Since != nil {
		seconds := int64(opts.Since.Seconds())
		podLogOpts.SinceSeconds = &seconds
	}

	if opts.Container != "" {
		podLogOpts.Container = opts.Container
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	return req.Stream(ctx)
}

// LogLine represents a single log line with metadata
type LogLine struct {
	PodName   string
	Container string
	Message   string
	Timestamp time.Time
}

// StreamLogsMulti streams logs from multiple pods and multiplexes them
func (c *Client) StreamLogsMulti(ctx context.Context, namespace string, pods []corev1.Pod, opts LogOptions, output chan<- LogLine) error {
	errCh := make(chan error, len(pods))

	for _, pod := range pods {
		pod := pod // Capture for goroutine
		go func() {
			// Determine which containers to stream from
			containers := []string{}
			if opts.Container != "" {
				containers = []string{opts.Container}
			} else {
				// Stream from all containers
				for _, c := range pod.Spec.Containers {
					containers = append(containers, c.Name)
				}
			}

			for _, container := range containers {
				containerOpts := opts
				containerOpts.Container = container

				stream, err := c.StreamLogs(ctx, namespace, pod.Name, containerOpts)
				if err != nil {
					errCh <- fmt.Errorf("failed to stream logs from pod %s container %s: %w", pod.Name, container, err)
					continue
				}

				// Read and send log lines
				scanner := bufio.NewScanner(stream)
				for scanner.Scan() {
					select {
					case <-ctx.Done():
						stream.Close()
						return
					case output <- LogLine{
						PodName:   pod.Name,
						Container: container,
						Message:   scanner.Text(),
						Timestamp: time.Now(),
					}:
					}
				}

				stream.Close()

				if err := scanner.Err(); err != nil && err != io.EOF {
					errCh <- fmt.Errorf("error reading logs from pod %s: %w", pod.Name, err)
				}
			}

			errCh <- nil
		}()
	}

	// Wait for all goroutines
	var lastErr error
	for i := 0; i < len(pods); i++ {
		if err := <-errCh; err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GetPodContainers returns the list of containers in a pod
func (c *Client) GetPodContainers(ctx context.Context, namespace, podName string) ([]string, error) {
	pod, err := c.GetPod(ctx, namespace, podName)
	if err != nil {
		return nil, err
	}

	var containers []string
	for _, c := range pod.Spec.Containers {
		containers = append(containers, c.Name)
	}
	return containers, nil
}
