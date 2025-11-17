package output

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// PrintYAML prints data in YAML format
func PrintYAML(w io.Writer, data interface{}) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}
	return nil
}
