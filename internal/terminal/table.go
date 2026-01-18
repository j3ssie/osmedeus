package terminal

import (
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
)

// NewTable creates a new table with consistent styling (no borders)
func NewTable(w io.Writer, header []string) *tablewriter.Table {
	if w == nil {
		w = os.Stdout
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader(header)

	// Clean styling without borders
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	if colorEnabled {
		colors := make([]tablewriter.Colors, len(header))
		for i := range colors {
			colors[i] = tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor}
		}
		table.SetHeaderColor(colors...)
	}

	return table
}

// NewBorderedTable creates a table with borders
func NewBorderedTable(w io.Writer, header []string) *tablewriter.Table {
	if w == nil {
		w = os.Stdout
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader(header)
	table.SetBorder(true)
	table.SetRowLine(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	if colorEnabled {
		colors := make([]tablewriter.Colors, len(header))
		for i := range colors {
			colors[i] = tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor}
		}
		table.SetHeaderColor(colors...)
	}

	return table
}

// NewSimpleTable creates a minimal table without header styling
func NewSimpleTable(w io.Writer, header []string) *tablewriter.Table {
	if w == nil {
		w = os.Stdout
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader(header)
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("  ")
	table.SetRowSeparator("")
	table.SetHeaderLine(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	return table
}

// TableWithColors creates a table and applies color functions to columns
func TableWithColors(w io.Writer, header []string, columnColors []func(string) string) *tablewriter.Table {
	table := NewTable(w, header)
	// Note: tablewriter doesn't support per-cell coloring via functions
	// Colors need to be applied to the data before appending
	return table
}
