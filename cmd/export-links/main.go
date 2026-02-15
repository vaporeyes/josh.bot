// ABOUTME: CLI tool for exporting links from the josh-bot-data DynamoDB table.
// ABOUTME: Usage: go run cmd/export-links/main.go [--tag TAG] [--since DATE] [--before DATE] [--format json|urls] [--table TABLE]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

func main() {
	tag := flag.String("tag", "", "Filter links by tag (e.g. 'go', 'aws')")
	since := flag.String("since", "", "Only links created on or after this date (YYYY-MM-DD)")
	before := flag.String("before", "", "Only links created before this date (YYYY-MM-DD)")
	format := flag.String("format", "json", "Output format: json (full objects) or urls (one URL per line, for piping to archivebox)")
	tableName := flag.String("table", "", "DynamoDB table name (defaults to TABLE_NAME env var)")
	flag.Parse()

	// Resolve table name
	table := *tableName
	if table == "" {
		table = os.Getenv("TABLE_NAME")
	}
	if table == "" {
		log.Fatal("TABLE_NAME environment variable or --table flag required")
	}

	// Validate format
	*format = strings.ToLower(*format)
	if *format != "json" && *format != "urls" {
		log.Fatalf("invalid format %q: must be 'json' or 'urls'", *format)
	}

	// Build DynamoDB scan filter
	filterParts := []string{"begins_with(id, :prefix)"}
	exprValues := map[string]types.AttributeValue{
		":prefix": &types.AttributeValueMemberS{Value: "link#"},
	}

	if *tag != "" {
		filterParts = append(filterParts, "contains(tags, :tag)")
		exprValues[":tag"] = &types.AttributeValueMemberS{Value: *tag}
	}
	if *since != "" {
		filterParts = append(filterParts, "created_at >= :since")
		exprValues[":since"] = &types.AttributeValueMemberS{Value: *since}
	}
	if *before != "" {
		filterParts = append(filterParts, "created_at < :before")
		exprValues[":before"] = &types.AttributeValueMemberS{Value: *before}
	}

	filterExpr := strings.Join(filterParts, " AND ")

	// Connect to DynamoDB
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("load AWS config: %v", err)
	}
	client := dynamodb.NewFromConfig(cfg)

	// Scan with pagination
	var links []domain.Link
	var lastKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:                 &table,
			FilterExpression:          &filterExpr,
			ExpressionAttributeValues: exprValues,
		}
		if lastKey != nil {
			input.ExclusiveStartKey = lastKey
		}

		output, err := client.Scan(ctx, input)
		if err != nil {
			log.Fatalf("dynamodb Scan: %v", err)
		}

		for _, item := range output.Items {
			var l domain.Link
			if err := attributevalue.UnmarshalMap(item, &l); err != nil {
				log.Fatalf("unmarshal link: %v", err)
			}
			links = append(links, l)
		}

		if output.LastEvaluatedKey == nil {
			break
		}
		lastKey = output.LastEvaluatedKey
	}

	// Output
	switch *format {
	case "urls":
		for _, l := range links {
			fmt.Println(l.URL)
		}
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(links); err != nil {
			log.Fatalf("encode JSON: %v", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Exported %d links\n", len(links))
}
