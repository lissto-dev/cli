package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// PrintJSON prints data in JSON format
func PrintJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}
