package k8s

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// PodStatus represents a simplified pod status for display
type PodStatus struct {
	Name         string
	Phase        string
	Restarts     int32
	Age          time.Duration
	Ready        bool
	StatusSymbol string
}

// ParsePodStatus extracts status information from a pod
func ParsePodStatus(pod *corev1.Pod) PodStatus {
	status := PodStatus{
		Name:  pod.Name,
		Phase: string(pod.Status.Phase),
		Age:   time.Since(pod.CreationTimestamp.Time),
	}

	// Calculate total restarts and ready status
	for _, cs := range pod.Status.ContainerStatuses {
		status.Restarts += cs.RestartCount
		if cs.Ready {
			status.Ready = true
		}
	}

	// If no containers are ready, pod is not ready
	if len(pod.Status.ContainerStatuses) > 0 {
		allReady := true
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				allReady = false
				break
			}
		}
		status.Ready = allReady
	}

	// Determine status symbol based on phase and ready state
	switch pod.Status.Phase {
	case corev1.PodRunning:
		if status.Ready {
			status.StatusSymbol = "✅"
		} else {
			status.StatusSymbol = "⏳"
		}
	case corev1.PodPending:
		status.StatusSymbol = "⏳"
	case corev1.PodSucceeded:
		status.StatusSymbol = "✅"
	case corev1.PodFailed:
		status.StatusSymbol = "❌"
	default:
		status.StatusSymbol = "❓"
	}

	// Check for crash loop or error states
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			reason := cs.State.Waiting.Reason
			if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" || reason == "ErrImagePull" {
				status.StatusSymbol = "❌"
				status.Phase = reason
				break
			}
		}
	}

	return status
}

// FormatAge formats a duration into a human-readable age string
func FormatAge(d time.Duration) string {
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%dd", days)
}

// FormatRestarts formats restart count
func FormatRestarts(count int32) string {
	if count == 0 {
		return "0 restarts"
	} else if count == 1 {
		return "1 restart"
	}
	return fmt.Sprintf("%d restarts", count)
}
