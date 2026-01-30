package output

import (
	"encoding/csv"
	"strings"

	"github.com/ervinmplayon/tractatus/internal/inventory"
)

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
