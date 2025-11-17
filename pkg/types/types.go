package types

import (
	"fmt"

	envv1alpha1 "github.com/lissto-dev/controller/api/v1alpha1"
)

// Re-export types from controller
type (
	Blueprint     = envv1alpha1.Blueprint
	BlueprintList = envv1alpha1.BlueprintList
	BlueprintSpec = envv1alpha1.BlueprintSpec

	Stack     = envv1alpha1.Stack
	StackList = envv1alpha1.StackList
	StackSpec = envv1alpha1.StackSpec

	Env     = envv1alpha1.Env
	EnvList = envv1alpha1.EnvList
	EnvSpec = envv1alpha1.EnvSpec
)

// GetBlueprintTitle extracts the blueprint title from stack annotations
func GetBlueprintTitle(stack *Stack) string {
	if stack.Annotations != nil {
		if title, ok := stack.Annotations["lissto.dev/blueprint-title"]; ok && title != "" {
			return title
		}
	}
	return ""
}

// GetStackDisplayName returns a user-friendly display name for a stack.
// If a blueprint title exists, returns "blueprint-title (stack-name)", otherwise just "stack-name"
func GetStackDisplayName(stack *Stack) string {
	blueprintTitle := GetBlueprintTitle(stack)
	if blueprintTitle != "" {
		return fmt.Sprintf("%s (%s)", blueprintTitle, stack.Name)
	}
	return stack.Name
}
