package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// APIDiscoveryInfo contains the discovered API information
type APIDiscoveryInfo struct {
	PublicURL       string // Public URL if configured (empty if not available)
	PortForwardURL  string // Port-forward URL (available if no public URL)
	APIID           string // API instance ID
	StopPortForward func() // Function to stop the port-forward (nil if public URL exists)
}

// DiscoverAPIEndpointFast discovers the API endpoint with public URL preference
// It establishes a port-forward connection ONCE, then queries /health?info=true to get
// public URL and API ID. If public URL exists, closes the port-forward immediately.
// If no public URL, keeps the port-forward open and returns it for continued use.
func (c *Client) DiscoverAPIEndpointFast(ctx context.Context, serviceName, namespace string) (*APIDiscoveryInfo, error) {
	// Establish port-forward to get initial connection (only once!)
	localPort := 8080
	portForwardURL, stopFunc, err := c.SetupPortForward(ctx, serviceName, namespace, localPort)
	if err != nil {
		return nil, fmt.Errorf("failed to setup initial connection: %w", err)
	}

	// Create HTTP client and call /health?info=true through the port-forward
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(portForwardURL + "/health?info=true")
	if err != nil {
		stopFunc() // Clean up on error
		return nil, fmt.Errorf("failed to get API info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		stopFunc() // Clean up on error
		return nil, fmt.Errorf("API info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var apiInfo struct {
		PublicURL string `json:"public_url"`
		APIID     string `json:"api_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiInfo); err != nil {
		stopFunc() // Clean up on error
		return nil, fmt.Errorf("failed to parse API info: %w", err)
	}

	// Decision time: close port-forward if we have a public URL
	if apiInfo.PublicURL != "" {
		// We have a public URL - close the port-forward, we don't need it
		stopFunc()
		return &APIDiscoveryInfo{
			PublicURL:       apiInfo.PublicURL,
			PortForwardURL:  "", // Not needed
			APIID:           apiInfo.APIID,
			StopPortForward: nil, // Already closed
		}, nil
	}

	// No public URL - keep the port-forward open and return it
	return &APIDiscoveryInfo{
		PublicURL:       "",             // Not available
		PortForwardURL:  portForwardURL, // Keep using this
		APIID:           apiInfo.APIID,
		StopPortForward: stopFunc, // Caller can close it later if needed
	}, nil
}

// DiscoverAPIEndpoint discovers the Lissto API endpoint from the cluster
// Returns just the port-forward URL (simpler version without API info)
func (c *Client) DiscoverAPIEndpoint(ctx context.Context, serviceName, namespace string) (string, error) {
	// use port-forward for all service types
	localPort := 8080 // Default local port
	url, _, err := c.SetupPortForward(ctx, serviceName, namespace, localPort)
	if err != nil {
		return "", fmt.Errorf("failed to setup port-forward: %w", err)
	}

	return url, nil
}

// SetupPortForward sets up port-forwarding to the API service
// Returns the local endpoint and a cleanup function to stop the port-forward
func (c *Client) SetupPortForward(ctx context.Context, serviceName, namespace string, localPort int) (string, func(), error) {
	// Get the service to find the target port
	service, err := c.GetService(ctx, namespace, serviceName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get service: %w", err)
	}

	if len(service.Spec.Ports) == 0 {
		return "", nil, fmt.Errorf("service has no ports defined")
	}

	// Note: targetPort would be used in actual port-forward implementation
	// targetPort := int(service.Spec.Ports[0].Port)

	// Find a pod backing this service
	selector := service.Spec.Selector
	pods, err := c.ListPods(ctx, namespace, selector)
	if err != nil {
		return "", nil, fmt.Errorf("failed to list pods for service: %w", err)
	}

	if len(pods) == 0 {
		return "", nil, fmt.Errorf("no pods found for service %s", serviceName)
	}

	// Use the first running pod
	var targetPod *corev1.Pod
	for i := range pods {
		if pods[i].Status.Phase == corev1.PodRunning {
			targetPod = &pods[i]
			break
		}
	}

	if targetPod == nil {
		return "", nil, fmt.Errorf("no running pods found for service %s", serviceName)
	}

	// Check if the port is available
	if !isPortAvailable(localPort) {
		// Port is in use - try to find an available port
		availablePort := findAvailablePort(localPort)
		if availablePort == 0 {
			return "", nil, fmt.Errorf(
				"port %d is already in use and no alternative ports available",
				localPort)
		}
		localPort = availablePort
	}

	// Get target port from service (the port the container is listening on)
	// TargetPort can be a name or a number
	targetPortValue := service.Spec.Ports[0].TargetPort
	var targetPort int

	if targetPortValue.IntVal != 0 {
		// TargetPort is a number
		targetPort = int(targetPortValue.IntVal)
	} else if targetPortValue.StrVal != "" {
		// TargetPort is a named port - need to resolve it from the pod's container ports
		for _, container := range targetPod.Spec.Containers {
			for _, port := range container.Ports {
				if port.Name == targetPortValue.StrVal {
					targetPort = int(port.ContainerPort)
					break
				}
			}
			if targetPort != 0 {
				break
			}
		}
	}

	// If targetPort is still 0, use the service port as fallback
	if targetPort == 0 {
		targetPort = int(service.Spec.Ports[0].Port)
	}

	// Set up actual port-forwarding (silently)
	stopFunc, err := c.startPortForward(ctx, namespace, targetPod.Name, localPort, targetPort)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start port-forward: %w", err)
	}

	url := fmt.Sprintf("http://localhost:%d", localPort)

	return url, stopFunc, nil
}

// startPortForward starts a port-forward to a pod and returns a cleanup function
func (c *Client) startPortForward(ctx context.Context, namespace, podName string, localPort, remotePort int) (func(), error) {
	// Build the port-forward URL
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	// Create SPDY transport
	transport, upgrader, err := spdy.RoundTripperFor(c.restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SPDY transport: %w", err)
	}

	// Create port forwarder
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)

	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}

	// Discard output from port-forward
	out := io.Discard
	errOut := io.Discard

	forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		return nil, fmt.Errorf("failed to create port forwarder: %w", err)
	}

	// Start port-forwarding in background
	go func() {
		if err := forwarder.ForwardPorts(); err != nil {
			// Silently ignore errors when stopped intentionally
		}
	}()

	// Wait for port-forward to be ready
	select {
	case <-readyChan:
		// Return cleanup function that closes the stopChan
		stopFunc := func() {
			close(stopChan)
		}
		return stopFunc, nil
	case <-time.After(10 * time.Second):
		close(stopChan)
		return nil, fmt.Errorf("timeout waiting for port-forward to be ready")
	case <-ctx.Done():
		close(stopChan)
		return nil, fmt.Errorf("context cancelled while waiting for port-forward")
	}
}

// isPortAvailable checks if a port is available on localhost
func isPortAvailable(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// findAvailablePort tries to find an available port starting from the given port
func findAvailablePort(startPort int) int {
	// Try the next 100 ports
	for port := startPort + 1; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return 0
}

// verifyEndpoint checks if an endpoint is reachable
func (c *Client) verifyEndpoint(url string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Try to reach the health endpoint or root
	resp, err := client.Get(url + "/health")
	if err != nil {
		// Try root if health fails
		resp, err = client.Get(url)
		if err != nil {
			return fmt.Errorf("endpoint not reachable: %w", err)
		}
	}
	defer resp.Body.Close()

	// Any 2xx, 3xx, or even 401/403 means the endpoint exists
	if resp.StatusCode < 500 {
		return nil
	}

	return fmt.Errorf("endpoint returned status %d", resp.StatusCode)
}

// GetService gets a service by namespace and name
func (c *Client) GetService(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	service, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	return service, nil
}
