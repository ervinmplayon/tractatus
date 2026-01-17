package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ervinmplayon/tractatus/internal/inventory"
)

// Stdout ---------------------------------------------------------------------------------
// Writes table format to sdtout
type StdoutTableWriter struct{}

func NewStdoutTableWriter() *StdoutTableWriter {
	return &StdoutTableWriter{}
}

// Outputs the inventory as a table to stdout
func (w *StdoutTableWriter) Write(inv *inventory.Inventory) error {
	return writeTable(os.Stdout, inv)
}

// Stdout ---------------------------------------------------------------------------------

// File ------------------------------------------------------------------------------------
// Writes table format to a file
type FiletTableWriter struct {
	filepath string
}

func NewFileTableWriter(filepath string) *FiletTableWriter {
	return &FiletTableWriter{filepath: filepath}
}

// Outputs the inventory as a table to a file
func (w *FiletTableWriter) Write(inv *inventory.Inventory) error {
	file, err := os.Create(w.filepath)
	if err != nil {
		return fmt.Errorf("fileTableWriter: failed to create file: %w", err)
	}
	defer file.Close()

	return writeTable(file, inv)
}

// File ------------------------------------------------------------------------------------

// Writes the inventory as a formatted table
func writeTable(writer io.Writer, inv *inventory.Inventory) error {
	if len(inv.Resources) == 0 {
		fmt.Fprintln(writer, "writeTable: No resources found.")
		return nil
	}

	// calculate the column widths and print header
	widths := calculateColumnWidths(inv)
	printTableRow(writer, widths,
		"App Name",
		"Owner",
		"Team",
		"Platform",
		"Stack Name",
		"CI/CD",
		"Account",
	)

	// print separator and print rows
	printTableSeparator(writer, widths)
	for _, res := range inv.Resources {
		cicd := "No"
		if res.HasCICD {
			cicd = "Yes"
		}

		printTableRow(writer, widths,
			res.AppName,
			res.Owner,
			res.Team,
			res.Platform,
			res.StackName,
			cicd,
			res.Account,
		)
	}

	return nil
}

// Determines the width needed for each column
func calculateColumnWidths(inv *inventory.Inventory) []int {
	headers := []string{"App Name", "Owner", "Team", "Platform", "Stack Name", "CI/CD", "Account"}
	widths := make([]int, len(headers))

	// starting with header widths
	for i, header := range headers {
		widths[i] = len(header)
	}

	// check resource data
	for _, res := range inv.Resources {
		values := []string{
			res.AppName,
			res.Owner,
			res.Team,
			res.Platform,
			res.StackName,
			formatBool(res.HasCICD),
			res.Account,
		}

		for i, val := range values {
			if len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}

	return widths
}

// Prints a single row with proper padding
func printTableRow(writer io.Writer, widths []int, values ...string) {
	fmt.Fprint(writer, "|")
	for i, val := range values {
		fmt.Fprintf(writer, "%-*s | ", widths[i], val)
	}
	fmt.Fprintln(writer)
}

// Prints a separator line
func printTableSeparator(writer io.Writer, widths []int) {
	fmt.Fprint(writer, "|")
	for _, width := range widths {
		fmt.Fprintf(writer, "-%s-|", strings.Repeat("-", width))
	}
	fmt.Fprintln(writer)
}

// Converts boolean to Yes/No string
func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
