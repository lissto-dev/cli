package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client
type Client struct {
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
}

// NewClient creates a new Kubernetes client using the current context
func NewClient() (*Client, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset:  clientset,
		restConfig: config,
	}, nil
}

// NewClientWithContext creates a new Kubernetes client for a specific kubeconfig context
func NewClientWithContext(kubeContext string) (*Client, error) {
	config, err := getKubeConfigWithContext(kubeContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig for context %s: %w", kubeContext, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset:  clientset,
		restConfig: config,
	}, nil
}

// GetCurrentKubeContext returns the current context name from kubeconfig
func GetCurrentKubeContext() (string, error) {
	// Try KUBECONFIG env var first
	var kubeconfigPath string
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		kubeconfigPath = kubeconfig
	} else {
		// Try default location
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Load the kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	if config.CurrentContext == "" {
		return "", fmt.Errorf("no current context set in kubeconfig")
	}

	return config.CurrentContext, nil
}

// getKubeConfig loads kubeconfig from standard locations
func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// Try KUBECONFIG env var
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	// Try default location
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	kubeconfig := filepath.Join(home, ".kube", "config")
	if _, err := os.Stat(kubeconfig); err != nil {
		return nil, fmt.Errorf("kubeconfig not found: %w", err)
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// getKubeConfigWithContext loads kubeconfig for a specific context
func getKubeConfigWithContext(contextName string) (*rest.Config, error) {
	// Determine kubeconfig path
	var kubeconfigPath string
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		kubeconfigPath = kubeconfig
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Load the kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Check if the context exists
	if _, exists := config.Contexts[contextName]; !exists {
		return nil, fmt.Errorf("context %s not found in kubeconfig", contextName)
	}

	// Build config for the specific context
	return clientcmd.NewNonInteractiveClientConfig(*config, contextName, &clientcmd.ConfigOverrides{}, nil).ClientConfig()
}

// ListPods queries pods by namespace and label selector
func (c *Client) ListPods(ctx context.Context, namespace string, labels map[string]string) ([]corev1.Pod, error) {
	// Build label selector
	labelSelector := ""
	for k, v := range labels {
		if labelSelector != "" {
			labelSelector += ","
		}
		labelSelector += fmt.Sprintf("%s=%s", k, v)
	}

	opts := metav1.ListOptions{}
	if labelSelector != "" {
		opts.LabelSelector = labelSelector
	}

	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return podList.Items, nil
}

// GetPod gets a specific pod by namespace and name
func (c *Client) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}
	return pod, nil
}
