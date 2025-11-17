package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// PrintTable prints data in a table format
func PrintTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)

	// Print headers
	fmt.Fprintln(tw, strings.Join(headers, "\t"))

	// Print rows
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	tw.Flush()
}
