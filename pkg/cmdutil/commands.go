package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeleteConfig configures a generic delete command
type DeleteConfig struct {
	Use          string
	ResourceType string
	DeleteFunc   func(name string) error
}

// NewDeleteCommand creates a standardized delete command
func NewDeleteCommand(cfg DeleteConfig) *cobra.Command {
	return &cobra.Command{
		Use:   cfg.Use,
		Short: fmt.Sprintf("Delete a %s", cfg.ResourceType),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := cfg.DeleteFunc(name); err != nil {
				return fmt.Errorf("failed to delete %s: %w", cfg.ResourceType, err)
			}
			fmt.Printf("%s '%s' deleted successfully\n", cfg.ResourceType, name)
			return nil
		},
	}
}

// ListConfig configures a generic list command
type ListConfig struct {
	Use          string
	ResourceType string
	ListFunc     func() (interface{}, error)
	EmptyMessage string
	Formatter    func(cmd *cobra.Command, items interface{}) error
}

// NewListCommand creates a standardized list command
func NewListCommand(cfg ListConfig) *cobra.Command {
	return &cobra.Command{
		Use:   cfg.Use,
		Short: fmt.Sprintf("List all %ss", cfg.ResourceType),
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := cfg.ListFunc()
			if err != nil {
				return fmt.Errorf("failed to list %ss: %w", cfg.ResourceType, err)
			}

			// Use custom formatter if provided
			if cfg.Formatter != nil {
				return cfg.Formatter(cmd, items)
			}

			return nil
		},
	}
}



