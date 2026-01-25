package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ervinmplayon/tractatus/internal/config"
	"github.com/ervinmplayon/tractatus/internal/inventory"
	"github.com/ervinmplayon/tractatus/internal/output"
	awssource "github.com/ervinmplayon/tractatus/internal/sources/aws"
	githubsource "github.com/ervinmplayon/tractatus/internal/sources/github"
)

func main() {
	// Define CLI flags
	source := flag.String("source", "github", "Data source: github, aws")

	// GitHub flags
	githubOrg := flag.String("github-org", "", "GitHub organization name")
	githubToken := flag.String("github-token", "", "GitHub personal access token (or use GITHUB_TOKEN env var)")
	excludeArchived := flag.Bool("exclude-archived", true, "Exclude archived repositories")

	// AWS flags
	accountsFlag := flag.String("account", "", "AWS account name(s) from config (comma-separated for multiple)")
	useProfile := flag.Bool("use-profile", true, "Use AWS credential profiles instead of config.json")
	configPath := flag.String("config", "config.json", "Path to config file")

	// Output flags
	formatFlag := flag.String("format", "table", "Output format: table, markdown")
	outputFlag := flag.String("output", "stdout", "Output destination: stdout or file path")

	flag.Parse()

	var dataSource inventory.DataSource
	var err error

	// Determine the DataSource here: github vs aws.
	switch *source {
	case "github":
		// Get token from 1. flag or 2. environment variable (backup)
		token := *githubToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}
		if token == "" {
			log.Fatal("Error: GitHub token required. Use --github-token flag or set GITHUB_TOKEN environment variable")
		}
		fmt.Fprintf(os.Stderr, "Collecting inventory from Github org: %s\n", *githubOrg)
		dataSource, err = githubsource.NewDataSource(token, *githubOrg, *excludeArchived)
		if err != nil {
			log.Fatalf("Failed to create Github data source: %v", err)
		}

	case "aws":
		if *accountsFlag == "" {
			log.Fatal("Error: --account flag is required for AWS source")
		}

		// Load configuration if not using profiles
		var cfg *config.Config
		if !*useProfile {
			cfg, err = config.LoadConfig(*configPath)
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}
		}

		// Support is limited to single account (extend to multiple later)
		accountName := *accountsFlag
		var account *config.Account
		if cfg != nil {
			if acc, exists := cfg.Accounts[accountName]; !exists {
				log.Fatalf("Error: Account '%s' not found in config", accountName)
			} else {
				account = &acc
			}
		}
		fmt.Fprintf(os.Stderr, "Collecting inventory from AWS account: %s\n", accountName)
		if *useProfile {
			fmt.Fprintf(os.Stderr, "Using AWS credential profiles from ~/.aws/\n")
		}
		dataSource = awssource.NewDataSource(accountName, account, *useProfile)

	default:
		log.Fatalf("Error: Unknown source '%s'. Use 'github' or 'aws'", *source)
	}

	// Collect inventory
	collector := inventory.NewCollector()
	ctx := context.Background()
	result, err := collector.CollectFromSource(ctx, dataSource)
	if err != nil {
		log.Fatalf("Failed to collect inventory: %v", err)
	}
	if len(result.Resources) == 0 {
		log.Fatal("Error: No resources found")
	}

	// Create appropriate output writer
	var writer output.OutputWriter
	switch *formatFlag {
	case "table":
		if *outputFlag == "stdout" {
			writer = output.NewStdoutTableWriter()
		} else {
			writer = output.NewFileTableWriter(*outputFlag)
		}
	case "markdown":
		if *outputFlag == "stdout" {
			writer = output.NewStdoutMarkdownWriter()
		} else {
			writer = output.NewFileMarkdownWriter(*outputFlag)
		}
	default:
		log.Fatalf("Error: Unknown format '%s'. Use 'table' or 'markdown'", *formatFlag)
	}

	// Write output
	if err := writer.Write(result); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	fmt.Fprintf(os.Stderr, "\nSuccessfully processed %d resources from %s\n",
		len(result.Resources), *source)
}
