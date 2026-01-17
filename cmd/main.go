package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ervinmplayon/tractatus/internal/config"
	"github.com/ervinmplayon/tractatus/internal/inventory"
	"github.com/ervinmplayon/tractatus/internal/output"
)

func main() {
	// Define CLI flags
	accountsFlag := flag.String("account", "", "AWS account name(s) from config (comma-separated for multiple)")
	formatFlag := flag.String("format", "table", "Output format: table, markdown")
	outputFlag := flag.String("output", "stdout", "Output destination: stdout or file path")
	// TODO: change this frfr
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	// Validate required flags
	if *accountsFlag == "" {
		log.Fatal("Error: --account flag is required")
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Parse account names
	accountNames := strings.Split(*accountsFlag, ",")
	for i := range accountNames {
		accountNames[i] = strings.TrimSpace(accountNames[i])
	}

	// Validate accounts exist in config
	for _, name := range accountNames {
		if _, exists := cfg.Accounts[name]; !exists {
			log.Fatalf("Error: Account '%s' not found in config", name)
		}
	}

	// Concurrently collect inventory from all accounts specified
	collector := inventory.NewCollector()
	results, errors := collector.CollectFromAccounts(cfg, accountNames)

	// Log any errors but continue
	for _, err := range errors {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	if len(results) == 0 {
		log.Fatal("Error: No inventory data collected")
	}

	// Merge results from all accounts
	mergedInventory := inventory.MergeInventories(results)

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
	if err := writer.Write(mergedInventory); err != nil {
		log.Fatalf("Failed to write output: %v", err)
	}

	fmt.Fprintf(os.Stderr, "\n✓ Successfully processed %d resources from %d account(s)\n",
		len(mergedInventory.Resources), len(accountNames))
}
