package k8s

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// TrafficReadiness contains the readiness status of a service
type TrafficReadiness struct {
	IsReady        bool
	ServiceExists  bool
	EndpointsReady bool
	IngressReady   bool
	PodsReady      bool
	FailureReason  string // "no endpoints", "no lb yet", "pod not ready", "starting up..", etc.
}

// CheckServiceReadiness checks if a service is ready to receive traffic
// This checks: Service exists, Endpoints are ready, Ingress has LB, and Pods are ready
func (c *Client) CheckServiceReadiness(ctx context.Context, namespace, serviceName string, pods []corev1.Pod, serviceAge time.Duration) TrafficReadiness {
	readiness := TrafficReadiness{
		IsReady:        false,
		ServiceExists:  false,
		EndpointsReady: false,
		IngressReady:   false,
		PodsReady:      false,
	}

	// 1. Check if Service exists
	_, err := c.GetService(ctx, namespace, serviceName)
	if err != nil {
		readiness.FailureReason = "no service"
		return readiness
	}
	readiness.ServiceExists = true

	// 2. Check if Ingress is ready (has load balancer address)
	ingress, err := c.GetIngressForService(ctx, namespace, serviceName)
	if err != nil || !hasIngressLoadBalancer(ingress) {
		readiness.FailureReason = "no lb yet"
		return readiness
	}
	readiness.IngressReady = true

	// 3. Check if all pods are ready
	if !arePodsReady(pods) {
		readiness.FailureReason = "pod not ready"
		return readiness
	}
	readiness.PodsReady = true

	// 4. Check if EndpointSlices have ready addresses
	endpointSlices, err := c.ListEndpointSlices(ctx, namespace, serviceName)
	if err != nil || !hasReadyEndpointSlices(endpointSlices) {
		readiness.FailureReason = "no endpoints"
		return readiness
	}
	readiness.EndpointsReady = true

	// All checks passed
	readiness.IsReady = true
	readiness.FailureReason = ""
	return readiness
}

// hasReadyEndpointSlices checks if endpoint slices have at least one ready endpoint
func hasReadyEndpointSlices(endpointSlices []discoveryv1.EndpointSlice) bool {
	if len(endpointSlices) == 0 {
		return false
	}

	for _, slice := range endpointSlices {
		for _, endpoint := range slice.Endpoints {
			// Check if endpoint is ready
			// The Ready field is a pointer to bool, nil means true (for backward compatibility)
			if endpoint.Conditions.Ready == nil || *endpoint.Conditions.Ready {
				return true
			}
		}
	}

	return false
}

// hasIngressLoadBalancer checks if an ingress has a load balancer address
func hasIngressLoadBalancer(ingress *networkingv1.Ingress) bool {
	if ingress == nil {
		return false
	}

	// Check if the ingress has load balancer ingress status
	// This indicates that the ingress controller has provisioned the load balancer
	return len(ingress.Status.LoadBalancer.Ingress) > 0
}

// arePodsReady checks if all pods in the list are ready
func arePodsReady(pods []corev1.Pod) bool {
	if len(pods) == 0 {
		return false
	}

	for _, pod := range pods {
		// Skip completed/succeeded pods (e.g., jobs)
		if pod.Status.Phase == corev1.PodSucceeded {
			continue
		}

		// Check if pod is running
		if pod.Status.Phase != corev1.PodRunning {
			return false
		}

		// Check if all containers are ready
		allContainersReady := true
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				allContainersReady = false
				break
			}
		}

		if !allContainersReady {
			return false
		}
	}

	return true
}

// FormatReadinessStatus formats the readiness status for display
// Shows "🟢" if ready, or "⚪ (reason)" if not ready
// If service is < 1 minute old and not ready, shows "⚪ (starting up..)"
func FormatReadinessStatus(readiness TrafficReadiness, serviceAge time.Duration) string {
	if readiness.IsReady {
		return "🟢"
	}

	// If service is less than 1 minute old, show generic "starting up.." message
	if serviceAge < time.Minute {
		return "⚪ (starting up..)"
	}

	// Service is older than 1 minute and not ready - show specific reason
	if readiness.FailureReason != "" {
		return "⚪ (" + readiness.FailureReason + ")"
	}

	return "⚪ (unknown)"
}
