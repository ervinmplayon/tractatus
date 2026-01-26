package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Wrap the Github API client
type Client struct {
	client *github.Client
	org    string // because reusability
}

// Create a new GHub API client
func NewClient(ctx context.Context, token, org string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("newClient: github token is required")
	}
	if org == "" {
		return nil, fmt.Errorf("newClient: github organization is required")
	}

	// Create OAuth2 token source and create Github client
	toke_src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	toke_client := oauth2.NewClient(ctx, toke_src)
	client := github.NewClient(toke_client)

	return &Client{
		client: client,
		org:    org,
	}, nil
}

// Represents a Github repo with its file tree
type Repository struct {
	Name           string
	IsArchived     bool
	DefaultBranch  string
	HTMLURL        string
	Files          []string // List of file/directory paths at root
	LastCommitter  string
	LastCommitDate string
}

// Fetch all the repon in an org
func (c *Client) ListRepositories(ctx context.Context, excludeArchived bool) ([]*Repository, error) {
	var allRepos []*Repository

	options := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, c.org, options)
		if err != nil {
			return nil, fmt.Errorf("listRepositories: failed to list repositories: %w", err)
		}

		for _, repo := range repos {
			// Skip archived repos if requested
			if excludeArchived && repo.GetArchived() {
				continue
			}

			// Get file tree for the repository
			files, err := c.getFileTree(ctx, repo.GetName(), repo.GetDefaultBranch())
			if err != nil {
				// Log warning but continue
				fmt.Printf("Warning: failed to get file tree for %s: %v\n", repo.GetName(), err)
				files = []string{}
			}

			// Get last commit info
			lastCommitter, lastCommitDate, err := c.getLastCommit(ctx, repo.GetName(), repo.GetDefaultBranch())
			if err != nil {
				// Log warning but continue
				fmt.Printf("Warning: failed to get last commit for %s: %v\n", repo.GetName(), err)
			}

			allRepos = append(allRepos, &Repository{
				Name:           repo.GetName(),
				IsArchived:     repo.GetArchived(),
				DefaultBranch:  repo.GetDefaultBranch(),
				HTMLURL:        repo.GetHTMLURL(),
				Files:          files,
				LastCommitter:  lastCommitter,
				LastCommitDate: lastCommitDate,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return allRepos, nil
}

// Gets the list of the files and directories at the root of a repository
func (c *Client) getFileTree(ctx context.Context, repoName, branch string) ([]string, error) {
	if branch == "" {
		branch = "main" // some repos have a non-main default branch but this is a good fallback for now
	}

	tree, _, err := c.client.Git.GetTree(ctx, c.org, repoName, branch, false)
	if err != nil {
		// Try master as fallback
		tree, _, err = c.client.Git.GetTree(ctx, c.org, repoName, "master", false)
		if err != nil {
			return nil, fmt.Errorf("getFileTree error: %w", err)
		}
	}

	// This part of the code does not run IF the "main", "master" branches above return due to errs.
	var files []string
	for _, entry := range tree.Entries {
		files = append(files, entry.GetPath())
	}

	return files, nil
}

// Returns the last commiter and the commit date
func (c *Client) getLastCommit(ctx context.Context, repoName, branch string) (string, string, error) {
	if branch == "" {
		branch = "main"
	}

	commits, _, err := c.client.Repositories.ListCommits(ctx, c.org, repoName, &github.CommitsListOptions{
		SHA: branch,
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})

	// Rethink returning empty string, more helpful returned msg?
	if err != nil {
		return "", "", err
	}

	if len(commits) == 0 {
		return "", "", nil
	}

	commit := commits[0]
	committer := "Unknown"
	if commit.Commit != nil && commit.Commit.Committer != nil && commit.Commit.Committer.Name != nil {
		committer = commit.Commit.Committer.GetName()
	}

	date := ""
	if commit.Commit != nil && commit.Commit.Committer != nil && commit.Commit.Committer.Date != nil {
		date = commit.Commit.Committer.Date.Format("2006-01-02")
	}

	return committer, date, nil
}

// Fecth the content of a specific file
func (c *Client) GetFileContent(ctx context.Context, repoName, filePath string) (string, error) {
	fileContent, _, _, err := c.client.Repositories.GetContents(ctx, c.org, repoName, filePath, nil)
	if err != nil {
		return "", fmt.Errorf("[getFileContent] error: %w", err)
	}

	// Safety check: if fileContent is nil, it's not a file (could be a directory)
	if fileContent == nil {
		return "", fmt.Errorf("[getFileContent] path %s is not a file", filePath)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("[getFileContent] failed to decode content for %s: %w", filePath, err)
	}

	return content, nil
}
