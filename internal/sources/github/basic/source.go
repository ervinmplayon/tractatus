package basic

import (
	"context"
	"fmt"

	"github.com/ervinmplayon/tractatus/internal/sources/github"
)

// Will reuse `github.client`,
// Detector is unnecessary

type BasicRepository struct {
	Name    string
	Owner   string
	RepoUrl string
}

type DataSource struct {
	client          *github.Client
	excludeArchived bool
}

func NewDataSource(token, org string, excludeArchived bool) (*DataSource, error) {
	ctx := context.Background()
	client, err := github.NewClient(ctx, token, org)
	if err != nil {
		return nil, fmt.Errorf("newDataSource error: %w", err)
	}
	return &DataSource{
		client:          client,
		excludeArchived: excludeArchived,
	}, nil
}
