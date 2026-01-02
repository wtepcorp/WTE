package ui

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

// Table wraps tablewriter for consistent table output
type Table struct {
	table *tablewriter.Table
}

// NewTable creates a new table
func NewTable(headers []string) *Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	return &Table{table: table}
}

// NewBorderedTable creates a table with borders
func NewBorderedTable(headers []string) *Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(true)
	table.SetRowLine(false)

	return &Table{table: table}
}

// Append adds a row to the table
func (t *Table) Append(row []string) {
	t.table.Append(row)
}

// AppendBulk adds multiple rows to the table
func (t *Table) AppendBulk(rows [][]string) {
	t.table.AppendBulk(rows)
}

// Render outputs the table
func (t *Table) Render() {
	if Quiet {
		return
	}
	t.table.Render()
}

// SetColWidth sets the minimum width for a column
func (t *Table) SetColWidth(width int) {
	t.table.SetColWidth(width)
}

// SetColMinWidth sets the minimum width for a specific column
func (t *Table) SetColMinWidth(col int, width int) {
	t.table.SetColMinWidth(col, width)
}

// SetAlignment sets the alignment for all columns
func (t *Table) SetAlignment(align int) {
	t.table.SetAlignment(align)
}

// SetHeaderColor sets colors for headers
func (t *Table) SetHeaderColor(colors ...tablewriter.Colors) {
	t.table.SetHeaderColor(colors...)
}

// SetColumnColor sets colors for columns
func (t *Table) SetColumnColor(colors ...tablewriter.Colors) {
	t.table.SetColumnColor(colors...)
}

// Rich appends a row with colors
func (t *Table) Rich(row []string, colors []tablewriter.Colors) {
	t.table.Rich(row, colors)
}

// StatusTable creates a table for showing status information
func StatusTable(data map[string]string) {
	if Quiet {
		return
	}

	table := NewTable([]string{"Property", "Value"})
	for key, value := range data {
		table.Append([]string{key, value})
	}
	table.Render()
}

// KeyValueTable creates a simple key-value table
func KeyValueTable(title string, data [][2]string) {
	if Quiet {
		return
	}

	Header(title)
	for _, item := range data {
		Printf("  %-20s %s\n", item[0]+":", item[1])
	}
	Println()
}
