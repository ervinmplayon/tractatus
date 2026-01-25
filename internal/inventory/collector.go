package inventory

import (
	"context"
	"fmt"
)

// Manages resource collection form multiple AWS accounts
type Collector struct{}

func NewCollector() *Collector {
	return &Collector{}
}

// Represents the complete inventory of resources
type Inventory struct {
	Resources []*ResourceInfo
}

// Represents enriched resource information
type ResourceInfo struct {
	// Common fields
	AppName  string
	Owner    string
	Team     string
	Platform string

	// AWS-specific fields
	StackName    string
	HasCICD      bool
	Account      string
	ARN          string
	ResourceTags map[string]string // Keep all tags for reference

	// GitHub-specific fields
	GitHubRepo     string
	LastCommitter  string
	LastCommitDate string
	HasCodeOwners  bool
	CodeOwners     []string
	HasTests       bool
	TestFramework  string // "pytest", "jest", "go test", etc.
	CICDPlatform   string // "CircleCI", "GitHub Actions", "CloudFormation", etc.
	RepoURL        string
	IsArchived     bool
}

type DataSource interface {
	Collect(ctx context.Context) ([]*ResourceInfo, error)
	Name() string
}

// Collects inventory from a single data source
func (c *Collector) CollectFromSource(ctx context.Context, source DataSource) (*Inventory, error) {
	resources, err := source.Collect(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", source.Name(), err)
	}

	return &Inventory{
		Resources: resources,
	}, nil
}

// Combines multiple inventories into one
func MergeInventories(inventories []*Inventory) *Inventory {
	merged := &Inventory{
		Resources: make([]*ResourceInfo, 0),
	}

	for _, inv := range inventories {
		merged.Resources = append(merged.Resources, inv.Resources...)
	}

	return merged
}
