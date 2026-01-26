package aws

import (
	"context"
	"fmt"

	"github.com/ervinmplayon/tractatus/internal/config"
	"github.com/ervinmplayon/tractatus/internal/inventory"
)

type DataSource struct {
	accountName string
	account     *config.Account
	useProfile  bool
}

func NewDataSource(accountName string, account *config.Account, useProfile bool) *DataSource {
	return &DataSource{
		accountName: accountName,
		account:     account,
		useProfile:  useProfile,
	}
}

// Returns the name of this data source
func (ds *DataSource) Name() string {
	return "AWS"
}

// Fetches resources from AWS
func (ds *DataSource) Collect(ctx context.Context) ([]*inventory.ResourceInfo, error) {
	// Create AWS client
	client, err := NewClient(ctx, ds.accountName, ds.useProfile, ds.account)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Get resources
	resources, err := client.GetResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	// Transform to ResourceInfo
	var resourceInfos []*inventory.ResourceInfo
	for _, res := range resources {
		info := enrichResource(res)
		resourceInfos = append(resourceInfos, &info)
	}

	return resourceInfos, nil
}

// Extracts and enriches resource information from tags
func enrichResource(res Resource) inventory.ResourceInfo {
	info := inventory.ResourceInfo{
		Platform:     res.Platform,
		Account:      res.Account,
		ARN:          res.ARN,
		ResourceTags: res.Tags,
	}

	// Extract App Name
	if name, exists := res.Tags["Name"]; exists {
		info.AppName = name
	} else if logicalID, exists := res.Tags["aws:cloudformation:logical-id"]; exists {
		info.AppName = logicalID
	} else {
		info.AppName = "Unknown"
	}

	// Extract Owner
	if owner, exists := res.Tags["owned-by"]; exists {
		info.Owner = owner
	} else if team, exists := res.Tags["team"]; exists {
		info.Owner = team
	} else {
		info.Owner = "Unknown"
	}

	// Extract Team
	if team, exists := res.Tags["team"]; exists {
		info.Team = team
	} else if owner, exists := res.Tags["owned-by"]; exists {
		info.Team = owner
	} else {
		info.Team = "Unknown"
	}

	// Extract Stack Name
	if stackName, exists := res.Tags["aws:cloudformation:stack-name"]; exists {
		info.StackName = stackName
		info.HasCICD = true
		info.CICDPlatform = "CloudFormation"
	} else {
		info.StackName = "None"
		info.HasCICD = false
	}

	return info
}
