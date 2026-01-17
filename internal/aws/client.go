package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/ervinmplayon/tractatus/internal/config"
)

// Client wraps AWS SDK clients
type Client struct {
	taggingClient *resourcegroupstaggingapi.Client
	accountName   string
}

// Creates a new AWS client for the given account
func NewClient(ctx context.Context, accountName string, account config.Account, useProfile bool) (*Client, error) {
	var cfg aws.Config
	var err error
	if useProfile {
		// Use AWS credential profile (profile name = account name)
		cfg, err = awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(account.Region),
			awsconfig.WithSharedConfigProfile(accountName), // Uses account name as profile name
		)
	} else {
		cfg, err = awsconfig.LoadDefaultConfig(
			ctx,
			awsconfig.WithRegion(account.Region),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				account.AccessKeyID,
				account.SecretAccessKey,
				account.SessionToken, // This can be an empty string
			)),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("newClient: failed to load AWS config: %w", err)
	}

	return &Client{
		taggingClient: resourcegroupstaggingapi.NewFromConfig(cfg),
		accountName:   accountName,
	}, nil
}

// ResourceTypes that we want to query (non-EKS compute resources)
var ResourceTypes = []string{
	"ec2:instance",
	"lambda:function",
	"ecs:service",
	"ecs:cluster",
	"elasticbeanstalk:application",
	"elasticbeanstalk:environment",
	"lightsail:instance",
	"apprunner:service",
}

// Represents a single AWS resource with its metadata
type Resource struct {
	ARN      string
	Tags     map[string]string
	Platform string
	Account  string
}

// Fetch all non-EKS resources
func (c *Client) GetResources(ctx context.Context) ([]Resource, error) {
	var allResources []Resource
	var paginationToken *string

	for {
		input := &resourcegroupstaggingapi.GetResourcesInput{
			ResourceTypeFilters: ResourceTypes,
			ResourcesPerPage:    aws.Int32(100),
		}

		if paginationToken != nil {
			input.PaginationToken = paginationToken
		}

		// GRAB ALL OF THEM!
		result, err := c.taggingClient.GetResources(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("getResources: failed to get resources: %w", err)
		}

		for _, mapping := range result.ResourceTagMappingList {
			resource := c.processResource(mapping)

			// filter out EKS resources
			if !isEKSResource(resource.Tags) {
				allResources = append(allResources, resource)
			}
		}

		// check for more pages
		if result.PaginationToken == nil || *result.PaginationToken == "" {
			break
		}
		paginationToken = result.PaginationToken
	}

	return allResources, nil
}

// Converts AWS resource tag mapping to our Resource struct
func (c *Client) processResource(mapping types.ResourceTagMapping) Resource {
	// convert the tags to map
	tags := make(map[string]string)
	for _, tag := range mapping.Tags {
		if tag.Key != nil && tag.Value != nil {
			tags[*tag.Key] = *tag.Value
		}
	}

	// extract the platform from ARN
	// ARN format: arn:aws:service:region:account:resource
	platform := extractPlatformFromARN(*mapping.ResourceARN)

	return Resource{
		ARN:      *mapping.ResourceARN,
		Tags:     tags,
		Platform: platform,
		Account:  c.accountName,
	}
}

// Check if a resource belongs to EKS
func isEKSResource(tags map[string]string) bool {
	// Check for EKS-specific tags
	if _, exists := tags["aws:eks:cluster-name"]; exists {
		return true
	}
	if _, exists := tags["eks:nodegroup-name"]; exists {
		return true
	}
	if _, exists := tags["eks:cluster-name"]; exists {
		return true
	}
	return false
}

// Parse the ARN to get the service name
func extractPlatformFromARN(arn string) string {
	// ARN format: arn:aws:service:region:account:resource
	// Example: arn:aws:ec2:us-east-1:123456789:instance/i-123456

	parts := parseARN(arn)
	if len(parts) < 3 {
		return "unknown"
	}

	service := parts[2]

	// Map service names to friendly names
	platformMap := map[string]string{
		"ec2":              "EC2",
		"lambda":           "Lambda",
		"ecs":              "ECS",
		"elasticbeanstalk": "Elastic Beanstalk",
		"lightsail":        "Lightsail",
		"apprunner":        "App Runner",
	}

	if friendly, exists := platformMap[service]; exists {
		return friendly
	}
	return service
}

// Split the ARN into its components
func parseARN(arn string) []string {
	// simple split by colon
	parts := []string{}
	current := ""
	for _, char := range arn {
		if char == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
