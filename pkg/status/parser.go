package status

import (
	"strings"

	envv1alpha1 "github.com/lissto-dev/controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StackStatus represents the overall status of a stack
type StackStatus struct {
	State   string // Ready, Deploying, Failed, Unknown
	Symbol  string // ✅ ⏳ ❌ ❓
	Reason  string
	Message string
}

// ServiceStatus represents the status of a service within a stack
type ServiceStatus struct {
	Name   string
	State  string // Ready, Deploying, Failed, Unknown
	Symbol string // ✅ ⏳ ❌ ❓
	Image  string
	URL    string
}

// ParseStackStatus extracts the overall stack status from conditions
func ParseStackStatus(conditions []metav1.Condition) StackStatus {
	status := StackStatus{
		State:  "Unknown",
		Symbol: "❓",
	}

	// Look for Ready condition
	for _, cond := range conditions {
		if cond.Type == "Ready" {
			if cond.Status == metav1.ConditionTrue {
				status.State = "Ready"
				status.Symbol = "✅"
			} else {
				status.Reason = cond.Reason
				status.Message = cond.Message

				// Determine if it's deploying or failed based on reason
				if strings.Contains(strings.ToLower(cond.Reason), "fail") ||
					strings.Contains(strings.ToLower(cond.Reason), "error") {
					status.State = "Failed"
					status.Symbol = "❌"
				} else {
					status.State = "Deploying"
					status.Symbol = "⏳"
				}
			}
			return status
		}
	}

	// If no Ready condition, check if we're still deploying
	if len(conditions) > 0 {
		status.State = "Deploying"
		status.Symbol = "⏳"
	}

	return status
}

// ParseServiceStatuses extracts per-service status from conditions
func ParseServiceStatuses(stack *envv1alpha1.Stack) []ServiceStatus {
	var services []ServiceStatus
	serviceMap := make(map[string]*ServiceStatus)

	// First, extract service info from spec.images
	for serviceName, imageInfo := range stack.Spec.Images {
		status := &ServiceStatus{
			Name:   serviceName,
			State:  "Unknown",
			Symbol: "❓",
			Image:  imageInfo.Image,
			URL:    imageInfo.URL,
		}

		serviceMap[serviceName] = status
	}

	// Then, update status from conditions
	for _, cond := range stack.Status.Conditions {
		// Look for Resource-deployment-serviceName conditions
		if strings.HasPrefix(cond.Type, "Resource-deployment-") {
			// Extract service name: "Resource-deployment-bo" -> "bo"
			serviceName := strings.TrimPrefix(cond.Type, "Resource-deployment-")

			if svc, exists := serviceMap[serviceName]; exists {
				if cond.Status == metav1.ConditionTrue {
					svc.State = "Ready"
					svc.Symbol = "✅"
				} else {
					// Check if it's a failure or just deploying
					if strings.Contains(strings.ToLower(cond.Reason), "fail") ||
						strings.Contains(strings.ToLower(cond.Reason), "error") {
						svc.State = "Failed"
						svc.Symbol = "❌"
					} else {
						svc.State = "Deploying"
						svc.Symbol = "⏳"
					}
				}
			}
		}
	}

	// Convert map to slice
	for _, svc := range serviceMap {
		services = append(services, *svc)
	}

	return services
}

// CountReadyServices counts how many services are ready
func CountReadyServices(services []ServiceStatus) (ready, total int) {
	total = len(services)
	for _, svc := range services {
		if svc.State == "Ready" {
			ready++
		}
	}
	return ready, total
}
