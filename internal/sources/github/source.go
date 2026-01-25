package github

import (
	"context"
	"fmt"

	"github.com/ervinmplayon/tractatus/internal/inventory"
)

type DataSource struct {
	client          *Client
	detector        *Detector
	excludeArchived bool
}

func NewDataSource(token, org string, excludeArchived bool) (*DataSource, error) {
	ctx := context.Background()
	client, err := NewClient(ctx, token, org)
	if err != nil {
		return nil, fmt.Errorf("newDataSource error: %w", err)
	}
	return &DataSource{
		client:          client,
		detector:        NewDetector(),
		excludeArchived: excludeArchived,
	}, nil
}

func (ds *DataSource) Name() string {
	return "GitHub"
}

// Fetches all repositories and analyzes them
func (ds *DataSource) Collect(ctx context.Context) ([]*inventory.ResourceInfo, error) {
	repos, err := ds.client.ListRepositories(ctx, ds.excludeArchived)
	if err != nil {
		return nil, fmt.Errorf("collect failed to list repositories: %w", err)
	}

	var resources []*inventory.ResourceInfo

	// Analyze each repository
	for _, repo := range repos {
		// Skip EKS repositories
		if ds.detector.IsEKS(repo.Files) {
			continue
		}

		info := ds.analyzeRepository(ctx, repo)
		resources = append(resources, info)
	}

	return resources, nil
}

// Analyze a single repository
func (ds *DataSource) analyzeRepository(ctx context.Context, repo *Repository) *inventory.ResourceInfo {
	info := &inventory.ResourceInfo{
		AppName:        repo.Name,
		GitHubRepo:     repo.Name,
		RepoURL:        repo.HTMLURL,
		IsArchived:     repo.IsArchived,
		LastCommitter:  repo.LastCommitter,
		LastCommitDate: repo.LastCommitDate,
	}

	// Detect CI/CD
	hasCICD, cicdPlatform := ds.detector.DetectCICD(repo.Files)
	info.HasCICD = hasCICD
	info.CICDPlatform = cicdPlatform

	// Detect tests
	hasTests, testFramework := ds.detector.DetectTests(repo.Files)
	info.HasTests = hasTests
	info.TestFramework = testFramework

	// Detect platform
	info.Platform = ds.detector.DetectPlatform(repo.Files)

	// Detect CODEOWNERS
	info.HasCodeOwners = ds.detector.DetectCodeOwners(repo.Files)

	// If CODEOWNERS exists, fetch and parse it
	if info.HasCodeOwners {
		codeownersContent, err := ds.getCodeOwnersContent(ctx, repo.Name, repo.Files)
		if err == nil {
			info.CodeOwners = ds.detector.ParseCodeOwners(codeownersContent)

			// Set Owner and Team from CODEOWNERS
			if len(info.CodeOwners) > 0 {
				info.Owner = info.CodeOwners[0]
				info.Team = info.CodeOwners[0]
			}
		}
	}

	// If no owner found, set to Unknown
	if info.Owner == "" {
		info.Owner = "Unknown"
	}
	if info.Team == "" {
		info.Team = "Unknown"
	}

	return info
}

// Fetches the CODEOWNERS file content
func (ds *DataSource) getCodeOwnersContent(ctx context.Context, repoName string, files []string) (string, error) {
	// Try common CODEOWNERS locations
	codeownersLocations := []string{
		"CODEOWNERS",
		".github/CODEOWNERS",
		"docs/CODEOWNERS",
	}

	for _, location := range codeownersLocations {
		// Check if this location exists in files
		found := false
		for _, file := range files {
			if file == location {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		// Try to fetch the content
		content, err := ds.client.GetFileContent(ctx, repoName, location)
		if err == nil {
			return content, nil
		}
	}

	return "", fmt.Errorf("getCodeOwnersContent CODEOWNERS file not found")
}
