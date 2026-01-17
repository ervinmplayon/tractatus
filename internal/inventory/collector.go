package inventory

import (
	"context"
	"fmt"
	"sync"

	awsclient "github.com/ervinmplayon/tractatus/internal/aws"
	"github.com/ervinmplayon/tractatus/internal/config"
)

// Manages resource collection form multiple AWS accounts
type Collector struct{}

func NewCollector() *Collector {
	return &Collector{}
}

// Represents the complete inventory of resources
type Inventory struct {
	Resources []ResourceInfo
}

// Represents enriched resource information
type ResourceInfo struct {
	AppName      string
	Owner        string
	Team         string
	Platform     string
	StackName    string
	HasCICD      bool
	Account      string
	ARN          string
	ResourceTags map[string]string // keeping these for reference
}

// Represent the result form querying a single account
type accountResult struct {
	inventory *Inventory
	err       error
}

// Collect the inventory from multiple AWS accounts concurrently
func (c *Collector) CollectFromAccounts(cfg *config.Config, accountNames []string) ([]*Inventory, []error) {
	ctx := context.Background()

	// this channel collects results
	results := make(chan accountResult, len(accountNames))
	var wg sync.WaitGroup

	// launch goroutine for each account
	for _, accountName := range accountNames {
		wg.Add(1)
		go func(accName string) {
			defer wg.Done()

			account := cfg.Accounts[accName]
			inventory, err := c.collectFromAccount(ctx, accName, account)
			results <- accountResult{
				inventory: inventory,
				err:       err,
			}
		}(accountName)
	}

	// Wait for all goroutines to complete and close channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// collect results
	var inventories []*Inventory
	var errors []error

	for result := range results {
		if result.err != nil {
			errors = append(errors, result.err)
		} else if result.inventory != nil {
			inventories = append(inventories, result.inventory)
		}
	}

	return inventories, errors
}

// Collects inventory from a single AWS account
func (c *Collector) collectFromAccount(ctx context.Context, accountName string, account config.Account) (*Inventory, error) {
	// Create AWS client
	client, err := awsclient.NewClient(ctx, accountName, account)
	if err != nil {
		return nil, fmt.Errorf("collectFromAccount: account '%s': failed to create AWS client: %w", accountName, err)
	}

	// get resources
	resources, err := client.GetResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("collectFromAccount: account '%s': failed to get resources: %w", accountName, err)
	}

	inventory := &Inventory{
		// create a slice of ResourceInfos with length 0 and capacity len(resources)
		Resources: make([]ResourceInfo, 0, len(resources)),
	}

	for _, res := range resources {
		info := c.enrichResources(res)
		inventory.Resources = append(inventory.Resources, info)
	}

	return inventory, nil
}

// Extracts and enriches information from tags
func (c *Collector) enrichResources(res awsclient.Resource) ResourceInfo {
	info := ResourceInfo{
		Platform:     res.Platform,
		Account:      res.Account,
		ARN:          res.ARN,
		ResourceTags: res.Tags,
	}

	// extract App Name, Priority: Name tag, Fallback: CloudFormation logical ID
	if name, exists := res.Tags["Name"]; exists {
		info.AppName = name
	} else if logicalID, exists := res.Tags["aws:cloudformation:logical-id"]; exists {
		info.AppName = logicalID
	} else {
		info.AppName = "Unknown"
	}

	// extract Owner, Priority: owned-by tag, Fallback: team tag
	if owner, exists := res.Tags["owned-by"]; exists {
		info.Owner = owner
	} else if team, exists := res.Tags["team"]; exists {
		info.Owner = team
	} else {
		info.Owner = "Unknown"
	}

	// extract Team
	if team, exists := res.Tags["team"]; exists {
		info.Team = team
	} else if owner, exists := res.Tags["owned-by"]; exists {
		info.Team = owner
	} else {
		info.Team = "Unknown"
	}

	// extract Stack Name
	if stackName, exists := res.Tags["aws:cloudformation:stack-name"]; exists {
		info.StackName = stackName
		info.HasCICD = true // If managed by CloudFormation, likely has CI/CD
	} else {
		info.StackName = "None"
		info.HasCICD = false
	}

	return info
}

// Combines multiple inventories into one
func MergeInventories(inventories []*Inventory) *Inventory {
	merged := &Inventory{
		Resources: make([]ResourceInfo, 0),
	}

	for _, inv := range inventories {
		merged.Resources = append(merged.Resources, inv.Resources...)
	}

	return merged
}
