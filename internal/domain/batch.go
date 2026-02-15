// ABOUTME: This file implements DynamoDB BatchWriteItem for bulk lift imports.
// ABOUTME: It chunks items into 25-item batches and retries unprocessed items.
package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// AIDEV-NOTE: DynamoDB BatchWriteItem max is 25 items per call.
const batchWriteMaxItems = 25

// BatchWriteClient is the interface for DynamoDB batch write operations.
type BatchWriteClient interface {
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

// BatchWriteLifts writes lifts to DynamoDB in batches of 25.
// It retries unprocessed items with exponential backoff.
func BatchWriteLifts(ctx context.Context, client BatchWriteClient, tableName string, lifts []Lift) error {
	if len(lifts) == 0 {
		return nil
	}

	// Build all write requests upfront
	requests := make([]types.WriteRequest, 0, len(lifts))
	for _, lift := range lifts {
		item, err := attributevalue.MarshalMap(lift)
		if err != nil {
			return fmt.Errorf("marshal lift %s: %w", lift.ID, err)
		}
		requests = append(requests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
	}

	// Process in chunks of 25
	for i := 0; i < len(requests); i += batchWriteMaxItems {
		end := min(i+batchWriteMaxItems, len(requests))
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

// batchWriteWithRetry executes a BatchWriteItem and retries unprocessed items.
func batchWriteWithRetry(ctx context.Context, client BatchWriteClient, input *dynamodb.BatchWriteItemInput) error {
	maxRetries := 5
	for attempt := 0; attempt <= maxRetries; attempt++ {
		output, err := client.BatchWriteItem(ctx, input)
		if err != nil {
			return fmt.Errorf("dynamodb BatchWriteItem: %w", err)
		}

		if len(output.UnprocessedItems) == 0 {
			return nil
		}

		// Retry unprocessed items with backoff
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
