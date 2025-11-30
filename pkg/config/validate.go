package config

import (
	"fmt"

	"github.com/lissto-dev/cli/pkg/k8s"
)

// ValidateContext checks if the saved context matches the current k8s context
// Returns a warning message if they differ, empty string if they match
func ValidateContext(ctx *Context) (string, error) {
	// Get current k8s context
	currentKubeContext, err := k8s.GetCurrentKubeContext()
	if err != nil {
		return "", fmt.Errorf("failed to get current k8s context: %w", err)
	}

	// Compare with saved context
	if currentKubeContext != ctx.KubeContext {
		warning := fmt.Sprintf(
			"⚠️  Warning: Current k8s context is '%s' but saved context is '%s'\n"+
				"   This may lead to unexpected behavior. Switch contexts with 'kubectl config use-context %s'",
			currentKubeContext, ctx.KubeContext, ctx.KubeContext,
		)
		return warning, nil
	}

	return "", nil
}

// ValidateAndWarn validates the context and prints a warning if contexts don't match
// Deprecated: Use ValidateAndFail for safety
func ValidateAndWarn(ctx *Context) error {
	warning, err := ValidateContext(ctx)
	if err != nil {
		// Don't fail on validation errors, just return
		return nil
	}

	if warning != "" {
		fmt.Println(warning)
		fmt.Println()
	}

	return nil
}

// ValidateAndFail validates the context and fails if contexts don't match
// This ensures operations are executed against the correct Kubernetes cluster
func ValidateAndFail(ctx *Context) error {
	// Get current k8s context
	currentKubeContext, err := k8s.GetCurrentKubeContext()
	if err != nil {
		return fmt.Errorf("failed to get current k8s context: %w", err)
	}

	// Compare with saved context
	if currentKubeContext != ctx.KubeContext {
		return fmt.Errorf(
			"context mismatch: current k8s context is '%s' but Lissto context expects '%s'\n"+
				"Switch contexts with: kubectl config use-context %s",
			currentKubeContext, ctx.KubeContext, ctx.KubeContext,
		)
	}

	return nil
}
