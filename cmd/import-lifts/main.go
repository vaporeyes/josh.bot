// ABOUTME: This file is the CLI entrypoint for importing workout CSV data into DynamoDB.
// ABOUTME: Usage: go run cmd/import-lifts/main.go [--dry-run] [--table TABLE] <csv-file>
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/jduncan/josh-bot/internal/domain"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Parse CSV and report stats without writing to DynamoDB")
	tableName := flag.String("table", "", "DynamoDB table name (defaults to TABLE_NAME env var)")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: import-lifts [--dry-run] [--table TABLE] <csv-file>\n")
		os.Exit(1)
	}
	csvPath := flag.Arg(0)

	// Resolve table name
	table := *tableName
	if table == "" {
		table = os.Getenv("TABLE_NAME")
	}
	if table == "" && !*dryRun {
		log.Fatal("TABLE_NAME environment variable or --table flag required")
	}

	// Parse CSV
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("open %s: %v", csvPath, err)
	}
	defer f.Close()

	lifts, err := domain.ParseLiftsCSV(f)
	if err != nil {
		log.Fatalf("parse CSV: %v", err)
	}

	// Collect stats
	exercises := make(map[string]bool)
	workouts := make(map[string]bool)
	for _, l := range lifts {
		exercises[l.ExerciseName] = true
		workouts[l.Date] = true
	}

	fmt.Printf("Parsed %d sets across %d workouts and %d exercises\n",
		len(lifts), len(workouts), len(exercises))

	if len(lifts) > 0 {
		fmt.Printf("Date range: %s to %s\n", lifts[0].Date, lifts[len(lifts)-1].Date)
		fmt.Printf("Sample ID: %s\n", lifts[0].ID)
	}

	if *dryRun {
		fmt.Println("Dry run complete. No data written.")
		return
	}

	// Write to DynamoDB
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	fmt.Printf("Writing %d items to %s in batches of 25...\n", len(lifts), table)
	if err := domain.BatchWriteLifts(ctx, client, table, lifts); err != nil {
		log.Fatalf("batch write: %v", err)
	}

	fmt.Printf("Done. Imported %d lift sets into %s.\n", len(lifts), table)
}
