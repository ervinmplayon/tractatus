package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/ervinmplayon/tractatus/internal/inventory"
	"github.com/google/go-github/v57/github"
)

// A DataSource needs the client to hook into platform, the detector for file detection
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
		codeownersContent, err := ds.getCodeOwnersContent(ctx, repo.Name)
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
func (ds *DataSource) getCodeOwnersContent(ctx context.Context, repoName string) (string, error) {
	// Try common CODEOWNERS locations
	codeownersLocations := []string{
		"CODEOWNERS",
		".github/CODEOWNERS",
		"docs/CODEOWNERS",
		"workflows/CODEOWNERS",
	}

	for _, location := range codeownersLocations {
		content, err := ds.client.GetFileContent(ctx, repoName, location)
		if err != nil {
			var gerr *github.ErrorResponse

			// Safe error type check, it will "reach inside" the wrapper error, find the original Github
			// error rather than just converting it to a flat string.
			if errors.As(err, &gerr) && gerr.Response.StatusCode == 404 {
				// It's a 404, just move to the next location
				continue
			}

			// For any other error, wrap the ORIGINAL err safely with %w
			return "", fmt.Errorf("[getCodeOwnersContent] API error at %s: %w", location, err)
		}

		if content != "" {
			return content, nil
		}
	}

	return "", fmt.Errorf("[getCodeOwnersContent] CODEOWNERS file not found")
}
