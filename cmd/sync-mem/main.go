// ABOUTME: CLI tool for syncing claude-mem SQLite data into DynamoDB.
// ABOUTME: Usage: go run cmd/sync-mem/main.go [--db PATH] [--table TABLE] [--since EPOCH] [--project PROJECT] [--dry-run]
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := flag.String("db", defaultDBPath(), "Path to claude-mem SQLite database")
	tableName := flag.String("table", "", "DynamoDB table name (defaults to TABLE_NAME env var or josh-bot-mem)")
	sinceEpoch := flag.Int64("since", 0, "Only sync rows created after this epoch (for incremental sync)")
	project := flag.String("project", "", "Only sync data for this project")
	dryRun := flag.Bool("dry-run", false, "Read SQLite and report stats without writing to DynamoDB")
	flag.Parse()

	// Resolve table name
	table := *tableName
	if table == "" {
		table = os.Getenv("TABLE_NAME")
	}
	if table == "" {
		table = "josh-bot-mem"
	}

	// Open SQLite (read-only)
	db, err := sql.Open("sqlite", *dbPath+"?mode=ro")
	if err != nil {
		log.Fatalf("open sqlite %s: %v", *dbPath, err)
	}
	defer db.Close()

	// Query all three tables
	observations, err := queryObservations(db, *sinceEpoch, *project)
	if err != nil {
		log.Fatalf("query observations: %v", err)
	}

	summaries, err := querySummaries(db, *sinceEpoch, *project)
	if err != nil {
		log.Fatalf("query summaries: %v", err)
	}

	prompts, err := queryPrompts(db, *sinceEpoch)
	if err != nil {
		log.Fatalf("query prompts: %v", err)
	}

	// Report stats
	total := len(observations) + len(summaries) + len(prompts)
	fmt.Printf("Found %d items to sync:\n", total)
	fmt.Printf("  observations: %d\n", len(observations))
	fmt.Printf("  summaries:    %d\n", len(summaries))
	fmt.Printf("  prompts:      %d\n", len(prompts))

	if total == 0 {
		fmt.Println("Nothing to sync.")
		return
	}

	// Find max epoch for incremental sync tracking
	maxEpoch := maxEpochFrom(observations, summaries, prompts)
	fmt.Fprintf(os.Stderr, "max_epoch=%d\n", maxEpoch)

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

	// Combine all items into one slice of write requests
	allItems := make([]map[string]types.AttributeValue, 0, total)
	allItems = append(allItems, observations...)
	allItems = append(allItems, summaries...)
	allItems = append(allItems, prompts...)

	fmt.Printf("Writing %d items to %s in batches of 25...\n", total, table)
	if err := batchWriteItems(ctx, client, table, allItems); err != nil {
		log.Fatalf("batch write: %v", err)
	}

	fmt.Printf("Done. Synced %d items into %s.\n", total, table)
}

func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".claude-mem/claude-mem.db"
	}
	return home + "/.claude-mem/claude-mem.db"
}

// queryObservations reads observations from SQLite and returns DynamoDB items.
func queryObservations(db *sql.DB, sinceEpoch int64, project string) ([]map[string]types.AttributeValue, error) {
	query := "SELECT id, memory_session_id, project, text, type, title, subtitle, facts, narrative, concepts, files_read, files_modified, prompt_number, discovery_tokens, created_at, created_at_epoch FROM observations WHERE created_at_epoch > ?"
	args := []any{sinceEpoch}

	if project != "" {
		query += " AND project = ?"
		args = append(args, project)
	}
	query += " ORDER BY id"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var items []map[string]types.AttributeValue
	for rows.Next() {
		var (
			id              int64
			sessionID       string
			proj            string
			text            sql.NullString
			obsType         string
			title           sql.NullString
			subtitle        sql.NullString
			facts           sql.NullString
			narrative       sql.NullString
			concepts        sql.NullString
			filesRead       sql.NullString
			filesModified   sql.NullString
			promptNumber    sql.NullInt64
			discoveryTokens sql.NullInt64
			createdAt       string
			createdAtEpoch  int64
		)
		if err := rows.Scan(&id, &sessionID, &proj, &text, &obsType, &title, &subtitle, &facts, &narrative, &concepts, &filesRead, &filesModified, &promptNumber, &discoveryTokens, &createdAt, &createdAtEpoch); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		item := map[string]types.AttributeValue{
			"id":               strAttr("obs#" + strconv.FormatInt(id, 10)),
			"type":             strAttr(obsType),
			"source_id":        numAttr(id),
			"session_id":       strAttr(sessionID),
			"project":          strAttr(proj),
			"created_at":       strAttr(createdAt),
			"created_at_epoch": numAttr(createdAtEpoch),
		}

		setIfValid(item, "title", title)
		setIfValid(item, "subtitle", subtitle)
		setIfValid(item, "narrative", narrative)
		setIfValid(item, "text", text)
		setIfValid(item, "facts", facts)
		setIfValid(item, "concepts", concepts)
		setIfValid(item, "files_read", filesRead)
		setIfValid(item, "files_modified", filesModified)
		if promptNumber.Valid {
			item["prompt_number"] = numAttr(promptNumber.Int64)
		}
		if discoveryTokens.Valid && discoveryTokens.Int64 > 0 {
			item["discovery_tokens"] = numAttr(discoveryTokens.Int64)
		}

		items = append(items, item)
	}
	return items, rows.Err()
}

// querySummaries reads session summaries from SQLite and returns DynamoDB items.
func querySummaries(db *sql.DB, sinceEpoch int64, project string) ([]map[string]types.AttributeValue, error) {
	query := "SELECT id, memory_session_id, project, request, investigated, learned, completed, next_steps, files_read, files_edited, notes, prompt_number, discovery_tokens, created_at, created_at_epoch FROM session_summaries WHERE created_at_epoch > ?"
	args := []any{sinceEpoch}

	if project != "" {
		query += " AND project = ?"
		args = append(args, project)
	}
	query += " ORDER BY id"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var items []map[string]types.AttributeValue
	for rows.Next() {
		var (
			id              int64
			sessionID       string
			proj            string
			request         sql.NullString
			investigated    sql.NullString
			learned         sql.NullString
			completed       sql.NullString
			nextSteps       sql.NullString
			filesRead       sql.NullString
			filesEdited     sql.NullString
			notes           sql.NullString
			promptNumber    sql.NullInt64
			discoveryTokens sql.NullInt64
			createdAt       string
			createdAtEpoch  int64
		)
		if err := rows.Scan(&id, &sessionID, &proj, &request, &investigated, &learned, &completed, &nextSteps, &filesRead, &filesEdited, &notes, &promptNumber, &discoveryTokens, &createdAt, &createdAtEpoch); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		item := map[string]types.AttributeValue{
			"id":               strAttr("summary#" + strconv.FormatInt(id, 10)),
			"type":             strAttr("summary"),
			"source_id":        numAttr(id),
			"session_id":       strAttr(sessionID),
			"project":          strAttr(proj),
			"created_at":       strAttr(createdAt),
			"created_at_epoch": numAttr(createdAtEpoch),
		}

		setIfValid(item, "request", request)
		setIfValid(item, "investigated", investigated)
		setIfValid(item, "learned", learned)
		setIfValid(item, "completed", completed)
		setIfValid(item, "next_steps", nextSteps)
		setIfValid(item, "files_read", filesRead)
		setIfValid(item, "files_edited", filesEdited)
		setIfValid(item, "notes", notes)
		if promptNumber.Valid {
			item["prompt_number"] = numAttr(promptNumber.Int64)
		}
		if discoveryTokens.Valid && discoveryTokens.Int64 > 0 {
			item["discovery_tokens"] = numAttr(discoveryTokens.Int64)
		}

		items = append(items, item)
	}
	return items, rows.Err()
}

// queryPrompts reads user prompts from SQLite and returns DynamoDB items.
func queryPrompts(db *sql.DB, sinceEpoch int64) ([]map[string]types.AttributeValue, error) {
	// AIDEV-NOTE: user_prompts has no project column, so --project filter does not apply here.
	query := "SELECT id, content_session_id, prompt_number, prompt_text, created_at, created_at_epoch FROM user_prompts WHERE created_at_epoch > ? ORDER BY id"

	rows, err := db.Query(query, sinceEpoch)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var items []map[string]types.AttributeValue
	for rows.Next() {
		var (
			id             int64
			sessionID      string
			promptNumber   int64
			promptText     string
			createdAt      string
			createdAtEpoch int64
		)
		if err := rows.Scan(&id, &sessionID, &promptNumber, &promptText, &createdAt, &createdAtEpoch); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		item := map[string]types.AttributeValue{
			"id":               strAttr("prompt#" + strconv.FormatInt(id, 10)),
			"type":             strAttr("prompt"),
			"source_id":        numAttr(id),
			"session_id":       strAttr(sessionID),
			"prompt_number":    numAttr(promptNumber),
			"prompt_text":      strAttr(promptText),
			"created_at":       strAttr(createdAt),
			"created_at_epoch": numAttr(createdAtEpoch),
		}

		items = append(items, item)
	}
	return items, rows.Err()
}

// batchWriteItems writes pre-built DynamoDB items using the same chunking/retry pattern as BatchWriteLifts.
func batchWriteItems(ctx context.Context, client domain.BatchWriteClient, tableName string, items []map[string]types.AttributeValue) error {
	if len(items) == 0 {
		return nil
	}

	requests := make([]types.WriteRequest, len(items))
	for i, item := range items {
		requests[i] = types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		}
	}

	// Process in chunks of 25 (DynamoDB BatchWriteItem limit)
	const batchSize = 25
	for i := 0; i < len(requests); i += batchSize {
		end := min(i+batchSize, len(requests))
		chunk := requests[i:end]

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				tableName: chunk,
			},
		}

		if err := batchWriteWithRetry(ctx, client, input); err != nil {
			return fmt.Errorf("batch write at offset %d: %w", i, err)
		}
	}

	return nil
}

// batchWriteWithRetry retries unprocessed items with exponential backoff.
func batchWriteWithRetry(ctx context.Context, client domain.BatchWriteClient, input *dynamodb.BatchWriteItemInput) error {
	// AIDEV-NOTE: mirrors domain.batchWriteWithRetry but called from main to avoid exporting internals.
	maxRetries := 5
	for attempt := 0; attempt <= maxRetries; attempt++ {
		output, err := client.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("dynamodb BatchWriteItem: %w", err)
		}

		if len(output.UnprocessedItems) == 0 {
			return nil
		}

		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
			input.RequestItems = output.UnprocessedItems
		}
	}
	return fmt.Errorf("still have unprocessed items after %d retries", maxRetries)
}

// strAttr builds a DynamoDB string attribute.
func strAttr(s string) types.AttributeValue {
	return &types.AttributeValueMemberS{Value: s}
}

// numAttr builds a DynamoDB number attribute.
func numAttr(n int64) types.AttributeValue {
	return &types.AttributeValueMemberN{Value: strconv.FormatInt(n, 10)}
}

// setIfValid adds a string attribute to the item if the NullString is valid and non-empty.
func setIfValid(item map[string]types.AttributeValue, key string, val sql.NullString) {
	if val.Valid && val.String != "" {
		item[key] = strAttr(val.String)
	}
}

// maxEpochFrom finds the highest created_at_epoch across all item sets.
func maxEpochFrom(observations, summaries, prompts []map[string]types.AttributeValue) int64 {
	var maxVal int64
	for _, items := range [][]map[string]types.AttributeValue{observations, summaries, prompts} {
		for _, item := range items {
			if attr, ok := item["created_at_epoch"]; ok {
				if n, ok := attr.(*types.AttributeValueMemberN); ok {
					if v, err := strconv.ParseInt(n.Value, 10, 64); err == nil && v > maxVal {
						maxVal = v
					}
				}
			}
		}
	}
	return maxVal
}
