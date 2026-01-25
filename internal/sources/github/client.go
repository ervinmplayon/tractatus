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
