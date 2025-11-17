package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lissto-dev/cli/pkg/client"
	"github.com/lissto-dev/cli/pkg/config"
	"github.com/lissto-dev/cli/pkg/k8s"
	"github.com/spf13/cobra"
)

var (
	loginContextName      string
	loginServiceName      string
	loginServiceNamespace string
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login [api-key]",
	Short: "Login to Lissto API",
	Long: `Login to the Lissto API and create a new context.
This command discovers the API endpoint from your current Kubernetes cluster,
authenticates with the provided API key, and sets up a new context for future commands.

The API key can be provided as an argument or will be prompted interactively.
The context name is automatically set based on your k8s context. Use --name to override.

Examples:
  lissto login                          # Interactive mode, prompts for API key
  lissto login abc123                   # Provide API key as argument
  lissto login --service my-api --namespace my-ns  # Custom service location`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginContextName, "name", "", "Name for the context (defaults to k8s context)")
	loginCmd.Flags().StringVar(&loginServiceName, "service", "lissto-api", "Name of the Lissto API service")
	loginCmd.Flags().StringVar(&loginServiceNamespace, "namespace", "lissto-system", "Namespace of the Lissto API service")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Step 1: Get current k8s context
	kubeContext, err := k8s.GetCurrentKubeContext()
	if err != nil {
		return fmt.Errorf("failed to get current k8s context: %w\nMake sure you have a valid kubeconfig", err)
	}

	fmt.Printf("Using Kubernetes context: %s\n", kubeContext)

	// Step 2: Get API key (from arg or prompt)
	var apiKey string
	if len(args) > 0 {
		apiKey = args[0]
	} else {
		// Interactive prompt for API key
		prompt := &survey.Password{
			Message: "Enter your API key:",
		}
		if err := survey.AskOne(prompt, &apiKey); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
	}

	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Step 3: Create k8s client for current context
	fmt.Println("Connecting to Kubernetes cluster...")
	k8sClient, err := k8s.NewClientWithContext(kubeContext)
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes: %w", err)
	}

	// Step 4: Discover API endpoint with fast discovery (opens port-forward once, gets all info)
	fmt.Printf("Discovering Lissto API service (%s/%s)...\n", loginServiceNamespace, loginServiceName)
	discoveryInfo, err := k8sClient.DiscoverAPIEndpointFast(
		context.Background(),
		loginServiceName,
		loginServiceNamespace,
	)
	if err != nil {
		return fmt.Errorf("failed to discover API endpoint: %w\nMake sure the service exists in the cluster", err)
	}

	// Use public URL if available, otherwise use the port-forward URL we already established
	apiURL := discoveryInfo.PublicURL
	if apiURL == "" {
		apiURL = discoveryInfo.PortForwardURL
	}

	// Step 5: Test authentication
	fmt.Println("Authenticating...")
	apiClient := client.NewClient(apiURL, apiKey)

	user, err := apiClient.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("✓ Logged in as: %s (role: %s)\n", user.Name, user.Role)

	// Step 6: Determine context name
	ctxName := loginContextName
	if ctxName == "" {
		// Use k8s context name as default
		ctxName = kubeContext
	}

	// Step 7: Load or create config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if context already exists
	if _, err := cfg.GetContext(ctxName); err == nil {
		return fmt.Errorf("context '%s' already exists. Use a different name with --name flag or delete the existing context first with 'lissto context delete %s'", ctxName, ctxName)
	}

	// Step 8: Create and save new context with discovered API info
	ctx := config.Context{
		Name:             ctxName,
		KubeContext:      kubeContext,
		ServiceName:      loginServiceName,
		ServiceNamespace: loginServiceNamespace,
		APIKey:           apiKey,
		APIUrl:           discoveryInfo.PublicURL, // Cache public URL (empty if not available)
		APIID:            discoveryInfo.APIID,     // Cache API instance ID
	}
	cfg.AddOrUpdateContext(ctx)
	cfg.CurrentContext = ctxName

	// Step 9: Fetch and cache environments
	envList, err := apiClient.ListEnvs()
	if err != nil {
		fmt.Printf("Warning: failed to fetch environments: %v\n", err)
	} else {
		envCache := &config.EnvCache{
			TTL: 300, // 5 minutes
		}

		var envs []config.EnvInfo
		for _, env := range envList {
			// Parse namespace from ID (format: "namespace/envname")
			namespace := ""
			if idx := strings.Index(env.ID, "/"); idx != -1 {
				namespace = env.ID[:idx]
			}
			envs = append(envs, config.EnvInfo{
				Name:      env.Name,
				Namespace: namespace,
			})
		}
		envCache.UpdateEnvs(envs)

		if err := config.SaveEnvCache(envCache); err != nil {
			fmt.Printf("Warning: failed to save environment cache: %v\n", err)
		} else {
			fmt.Printf("✓ Discovered %d environment(s):\n", len(envs))
			for _, env := range envs {
				fmt.Printf("  - %s\n", env.Name)
			}
		}

		// Set default environment
		if len(envs) > 0 {
			// Prefer user's own environment, otherwise use first one
			defaultEnv := envs[0].Name
			for _, env := range envs {
				if strings.Contains(env.Name, user.Name) {
					defaultEnv = env.Name
					break
				}
			}
			cfg.CurrentEnv = defaultEnv
			fmt.Printf("✓ Set current environment to: %s\n", defaultEnv)
		}
	}

	// Step 10: Save config
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Context '%s' created and set as current\n", ctxName)
	fmt.Println("\nReady to use Lissto CLI!")
	fmt.Println("Try: lissto status")

	return nil
}
