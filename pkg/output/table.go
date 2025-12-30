package output

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// PrintTable prints a table with proper alignment and styling
func PrintTable(w io.Writer, headers []string, rows [][]string) {
	// Create table with no borders and left-aligned headers
	table := tablewriter.NewTable(w,
		tablewriter.WithSymbols(tw.NewSymbols(tw.StyleNone)),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
	)

	// Set headers
	headerVals := make([]any, len(headers))
	for i, h := range headers {
		headerVals[i] = h
	}
	table.Header(headerVals...)

	// Append rows
	for _, row := range rows {
		rowVals := make([]any, len(row))
		for i, val := range row {
			rowVals[i] = val
		}
		_ = table.Append(rowVals...)
	}

	// Render
	_ = table.Render()
}
