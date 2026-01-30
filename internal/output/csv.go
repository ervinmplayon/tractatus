package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/ervinmplayon/tractatus/internal/inventory"
)

// write to std out-------------------------------------------------------------------
type StdoutCSVWriter struct{}

func NewStdoutCSVWriter() *StdoutCSVWriter {
	return &StdoutCSVWriter{}
}

func (w *StdoutCSVWriter) Write(inv *inventory.Inventory) error {
	return writeCSV(os.Stdout, inv)
}

// write to std out-------------------------------------------------------------------

// write to file-----------------------------------------------------------------------
type FileCSVWriter struct {
	filepath string
}

func NewFileWriter(filepath string) *FileCSVWriter {
	return &FileCSVWriter{filepath: filepath}
}

func (w *FileCSVWriter) Write(inv *inventory.Inventory) error {
	file, err := os.Create(w.filepath)
	if err != nil {
		return fmt.Errorf("error [FileCSVWriter.Write()] failed to create file: %w", err)
	}
	defer file.Close()
	return writeCSV(file, inv)
}

// write to file-----------------------------------------------------------------------

func writeCSV(file *os.File, inv *inventory.Inventory) error {
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if len(inv.Resources) == 0 {
		return nil
	}
	isGitHub := len(inv.Resources) > 0 && inv.Resources[0].GitHubRepo != ""
	if isGitHub {
		return writeGitHubCSV(writer, inv)
	}
	return writeAWSCSV(writer, inv)
}

func writeGitHubCSV(writer *csv.Writer, inv *inventory.Inventory) error {
	header := []string{
		"Repo Name",
		"Owner(s)",
		"Last Committer",
		"Last Commit Date",
		"Platform",
		"CI/CD Platform",
		"Has Tests",
		"Test Framework",
		"Repo URL",
		"Is Archived",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, res := range inv.Resources {
		// Format owners
		owners := "Unknown"
		if res.HasCodeOwners && len(res.CodeOwners) > 0 {
			owners = strings.Join(res.CodeOwners, "; ")
		}
		cicd := res.CICDPlatform
		if cicd == "" {
			cicd = "No"
		}
		hasTests := "No"
		if res.HasTests {
			hasTests = "Yes"
		}
		testFramework := res.TestFramework
		if testFramework == "" {
			testFramework = "N/A"
		}
		archived := "No"
		if res.IsArchived {
			archived = "Yes"
		}
		row := []string{
			res.AppName,
			owners,
			res.LastCommitter,
			res.LastCommitDate,
			res.Platform,
			cicd,
			hasTests,
			testFramework,
			res.RepoURL,
			archived,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func writeAWSCSV(writer *csv.Writer, inv *inventory.Inventory) error {
	header := []string{
		"App Name",
		"Owner",
		"Team",
		"Platform",
		"Stack Name",
		"Has CI/CD",
		"CI/CD Platform",
		"Account",
		"ARN",
	}

	if err := writer.Write(header); err != nil {
		return err
	}

	for _, res := range inv.Resources {
		hasCICD := "No"
		if res.HasCICD {
			hasCICD = "Yes"
		}
		cicdPlatform := res.CICDPlatform
		if cicdPlatform == "" {
			cicdPlatform = "N/A"
		}
		row := []string{
			res.AppName,
			res.Owner,
			res.Team,
			res.Platform,
			res.StackName,
			hasCICD,
			cicdPlatform,
			res.Account,
			res.ARN,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
