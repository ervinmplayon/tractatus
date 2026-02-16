package basic

import (
	"context"
	"fmt"

	"github.com/ervinmplayon/tractatus/internal/inventory"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Will reuse `github.client`,
// Detector is unnecessary
type Client struct {
	client *github.Client
	org    string // because reusability
}

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

type BasicRepository struct {
	Name    string
	Owner   string
	RepoUrl string
}

type DataSource struct {
	client          *Client
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
		excludeArchived: excludeArchived,
	}, nil
}

func (ds *DataSource) Name() string {
	return "GitHub"
}

func (c *Client) ListRepositories(ctx context.Context, excludeArchived bool) ([]*BasicRepository, error) {
	var allRepos []*BasicRepository
	options := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, c.org, options)
		if err != nil {
			return nil, fmt.Errorf("error [github.basic.ListRepositories]: failed to list repositories: %w", err)
		}
		for _, repo := range repos {
			if excludeArchived {
				continue
			}
			allRepos = append(allRepos, &BasicRepository{
				Name:    repo.GetName(),
				Owner:   *repo.GetOwner().Name,
				RepoUrl: repo.GetHTMLURL(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return allRepos, nil
}

func (ds *DataSource) Collect(ctx context.Context) ([]*inventory.ResourceInfo, error) {
	repos, err := ds.client.ListRepositories(ctx, ds.excludeArchived)
	if err != nil {
		return nil, fmt.Errorf("error [github.basic.Collect] failed to list repositories: %w", err)
	}
	var resources []*inventory.ResourceInfo
	for _, repo := range repos {
		resource := &inventory.ResourceInfo{
			AppName: repo.Name,
			Owner:   repo.Owner,
			RepoURL: repo.RepoUrl,
		}
		resources = append(resources, resource)
	}
	return resources, nil
}
