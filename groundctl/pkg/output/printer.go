package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// Format represents the output format.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// Printer handles output formatting.
type Printer struct {
	Format Format
	Writer io.Writer
}

// NewPrinter creates a printer with the given format.
func NewPrinter(format string) *Printer {
	f := FormatTable
	if strings.EqualFold(format, "json") {
		f = FormatJSON
	}
	return &Printer{
		Format: f,
		Writer: os.Stdout,
	}
}

// PrintTable renders data as a formatted table.
func (p *Printer) PrintTable(headers []string, rows [][]string) {
	if p.Format == FormatJSON {
		var result []map[string]string
		for _, row := range rows {
			item := make(map[string]string)
			for i, h := range headers {
				if i < len(row) {
					item[h] = row[i]
				}
			}
			result = append(result, item)
		}
		p.PrintJSON(result)
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Fprintf(p.Writer, "%-*s  ", widths[i], h)
	}
	fmt.Fprintln(p.Writer)

	// Print rows
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) {
				fmt.Fprintf(p.Writer, "%-*s  ", widths[i], col)
			}
		}
		fmt.Fprintln(p.Writer)
	}
}

// PrintJSON renders data as formatted JSON.
func (p *Printer) PrintJSON(data interface{}) {
	enc := json.NewEncoder(p.Writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
}

// PrintSuccess prints a success message.
func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf("ok: "+format+"\n", args...)
}

// PrintError prints an error message.
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
}
