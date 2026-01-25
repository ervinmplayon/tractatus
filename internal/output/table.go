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
		fmt.Fprintln(writer, "No resources found.")
		return nil
	}

	// Detect if this is GitHub or AWS data
	isGitHub := len(inv.Resources) > 0 && inv.Resources[0].GitHubRepo != ""

	if isGitHub {
		return writeGitHubTable(writer, inv)
	}
	return writeAWSTable(writer, inv)
}

// Writes GitHub inventory as a table
func writeGitHubTable(writer io.Writer, inv *inventory.Inventory) error {
	// Calculate column widths
	widths := calculateGitHubColumnWidths(inv)

	// Print header
	printTableRow(writer, widths,
		"Repo Name",
		"Owner",
		"Last Committer",
		"CODEOWNERS",
		"Platform",
		"CI/CD",
		"Tests",
	)

	// Print separator
	printTableSeparator(writer, widths)

	// Print rows
	for _, res := range inv.Resources {
		cicd := res.CICDPlatform
		if cicd == "" {
			cicd = "No"
		}

		codeowners := "No"
		if res.HasCodeOwners {
			codeowners = fmt.Sprintf("Yes (%d)", len(res.CodeOwners))
		}

		tests := "No"
		if res.HasTests {
			tests = "Yes"
			if res.TestFramework != "" {
				tests = fmt.Sprintf("Yes (%s)", res.TestFramework)
			}
		}

		printTableRow(writer, widths,
			res.AppName,
			res.Owner,
			res.LastCommitter,
			codeowners,
			res.Platform,
			cicd,
			tests,
		)
	}

	return nil
}

// Writes AWS inventory as a table
func writeAWSTable(writer io.Writer, inv *inventory.Inventory) error {
	// Calculate column widths
	widths := calculateAWSColumnWidths(inv)

	// Print header
	printTableRow(writer, widths,
		"App Name",
		"Owner",
		"Team",
		"Platform",
		"Stack Name",
		"CI/CD",
		"Account",
	)

	// Print separator
	printTableSeparator(writer, widths)

	// Print rows
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

// Determines the width needed for each GitHub column
func calculateGitHubColumnWidths(inv *inventory.Inventory) []int {
	headers := []string{"Repo Name", "Owner", "Last Committer", "CODEOWNERS", "Platform", "CI/CD", "Tests"}
	widths := make([]int, len(headers))

	// Start with header widths
	for i, header := range headers {
		widths[i] = len(header)
	}

	// Check resource data
	for _, res := range inv.Resources {
		codeowners := "No"
		if res.HasCodeOwners {
			codeowners = fmt.Sprintf("Yes (%d)", len(res.CodeOwners))
		}

		cicd := res.CICDPlatform
		if cicd == "" {
			cicd = "No"
		}

		tests := "No"
		if res.HasTests {
			tests = "Yes"
			if res.TestFramework != "" {
				tests = fmt.Sprintf("Yes (%s)", res.TestFramework)
			}
		}

		values := []string{
			res.AppName,
			res.Owner,
			res.LastCommitter,
			codeowners,
			res.Platform,
			cicd,
			tests,
		}

		for i, val := range values {
			if len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}

	return widths
}

// Determines the width needed for each AWS column
func calculateAWSColumnWidths(inv *inventory.Inventory) []int {
	headers := []string{"App Name", "Owner", "Team", "Platform", "Stack Name", "CI/CD", "Account"}
	widths := make([]int, len(headers))

	// Start with header widths
	for i, header := range headers {
		widths[i] = len(header)
	}

	// Check resource data
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
	fmt.Fprint(writer, "| ")
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
