package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ervinmplayon/tractatus/internal/inventory"
)

// Stdout ---------------------------------------------------------------------------------
// Writes markdown format to stdout
type StdoutMarkdownWriter struct{}

func NewStdoutMarkdownWriter() *StdoutMarkdownWriter {
	return &StdoutMarkdownWriter{}
}

// Ouputs the inventory as markdown to stdout
func (w *StdoutMarkdownWriter) Write(inv *inventory.Inventory) error {
	return writeMarkdown(os.Stdout, inv)
}

// Stdout ---------------------------------------------------------------------------------

// File ------------------------------------------------------------------------------------
// Writes markdown format to a file
type FileMarkdownWriter struct {
	filepath string
}

func NewFileMarkdownWriter(filepath string) *FileMarkdownWriter {
	return &FileMarkdownWriter{filepath: filepath}
}

// Outputs the inventory as markdown to a file
func (w *FileMarkdownWriter) Write(inv *inventory.Inventory) error {
	file, err := os.Create(w.filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return writeMarkdown(file, inv)
}

// File ------------------------------------------------------------------------------------

// Writes the inventory in Confluencee-compatible markdown format
func writeMarkdown(writer io.Writer, inv *inventory.Inventory) error {
	// Header
	fmt.Fprintln(writer, "# Resource Inventory")
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "**Generated**: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintln(writer)

	if len(inv.Resources) == 0 {
		fmt.Fprintln(writer, "No resources found.")
		return nil
	}

	// Detect if this is GitHub or AWS data
	isGitHub := len(inv.Resources) > 0 && inv.Resources[0].GitHubRepo != ""

	if isGitHub {
		return writeGitHubMarkdown(writer, inv)
	}
	return writeAWSMarkdown(writer, inv)
}

// Writes GitHub inventory as markdown
func writeGitHubMarkdown(writer io.Writer, inv *inventory.Inventory) error {
	// Summary statistics
	summary := generateGitHubSummary(inv)
	fmt.Fprintln(writer, "## Summary")
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "- **Total Repositories**: %d\n", summary.TotalResources)

	for platform, count := range summary.ByPlatform {
		fmt.Fprintf(writer, "- **%s**: %d\n", platform, count)
	}
	fmt.Fprintln(writer)

	fmt.Fprintf(writer, "- **With CI/CD**: %d\n", summary.WithCICD)
	fmt.Fprintf(writer, "- **With Tests**: %d\n", summary.WithTests)
	fmt.Fprintf(writer, "- **With CODEOWNERS**: %d\n", summary.WithCodeOwners)
	fmt.Fprintln(writer)

	// Resources table
	fmt.Fprintln(writer, "## Repositories")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "| Repo Name | Owner | Last Committer | CODEOWNERS | Platform | CI/CD | Tests |")
	fmt.Fprintln(writer, "|-----------|-------|----------------|------------|----------|-------|-------|")

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

		fmt.Fprintf(writer, "| %s | %s | %s | %s | %s | %s | %s |\n",
			escapeMarkdown(res.AppName),
			escapeMarkdown(res.Owner),
			escapeMarkdown(res.LastCommitter),
			escapeMarkdown(codeowners),
			escapeMarkdown(res.Platform),
			escapeMarkdown(cicd),
			escapeMarkdown(tests),
		)
	}

	return nil
}

// Writes AWS inventory as markdown
func writeAWSMarkdown(writer io.Writer, inv *inventory.Inventory) error {
	// Summary statistics
	summary := generateAWSSummary(inv)
	fmt.Fprintln(writer, "## Summary")
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "- **Total Resources**: %d\n", summary.TotalResources)

	for platform, count := range summary.ByPlatform {
		fmt.Fprintf(writer, "- **%s**: %d\n", platform, count)
	}
	fmt.Fprintln(writer)

	fmt.Fprintf(writer, "- **Resources with CI/CD**: %d\n", summary.WithCICD)
	fmt.Fprintf(writer, "- **Resources without CI/CD**: %d\n", summary.WithoutCICD)
	fmt.Fprintln(writer)

	// Resources table
	fmt.Fprintln(writer, "## Resources")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "| App Name | Owner | Team | Platform | Stack Name | CI/CD | Account |")
	fmt.Fprintln(writer, "|----------|-------|------|----------|------------|-------|---------|")

	for _, res := range inv.Resources {
		cicd := "No"
		if res.HasCICD {
			cicd = "Yes"
		}

		fmt.Fprintf(writer, "| %s | %s | %s | %s | %s | %s | %s |\n",
			escapeMarkdown(res.AppName),
			escapeMarkdown(res.Owner),
			escapeMarkdown(res.Team),
			escapeMarkdown(res.Platform),
			escapeMarkdown(res.StackName),
			cicd,
			escapeMarkdown(res.Account),
		)
	}

	return nil
}

// Creates summary statistics for GitHub inventory
func generateGitHubSummary(inv *inventory.Inventory) Summary {
	summary := Summary{
		TotalResources: len(inv.Resources),
		ByPlatform:     make(map[string]int),
		ByAccount:      make(map[string]int),
	}

	for _, res := range inv.Resources {
		summary.ByPlatform[res.Platform]++

		if res.HasCICD {
			summary.WithCICD++
		}

		if res.HasTests {
			summary.WithTests++
		}

		if res.HasCodeOwners {
			summary.WithCodeOwners++
		}
	}

	return summary
}

// Creates summary statistics for AWS inventory
func generateAWSSummary(inv *inventory.Inventory) Summary {
	summary := Summary{
		TotalResources: len(inv.Resources),
		ByPlatform:     make(map[string]int),
		ByAccount:      make(map[string]int),
	}

	for _, res := range inv.Resources {
		summary.ByPlatform[res.Platform]++
		summary.ByAccount[res.Account]++

		if res.HasCICD {
			summary.WithCICD++
		} else {
			summary.WithoutCICD++
		}
	}

	return summary
}

// Contains statistics about the inventory
type Summary struct {
	TotalResources int
	ByPlatform     map[string]int
	ByAccount      map[string]int
	WithCICD       int
	WithoutCICD    int
	WithTests      int
	WithCodeOwners int
}

// Escapes special markdown characters
func escapeMarkdown(s string) string {
	// Basic escaping for pipe characters which can break tables
	replacer := strings.NewReplacer(
		"|", "\\|",
	)
	return replacer.Replace(s)
}
