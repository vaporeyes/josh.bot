// ABOUTME: This file contains tests for the DynamoDB batch writer.
// ABOUTME: Tests BatchWriteItem chunking, retry logic, and edge cases.
package domain

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockBatchWriter implements BatchWriteClient for testing.
type mockBatchWriter struct {
	calls             int
	totalItems        int
	returnUnprocessed bool
	unprocessedOnce   bool
	err               error
}

func (m *mockBatchWriter) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	m.calls++
	for _, reqs := range params.RequestItems {
		m.totalItems += len(reqs)
	}

	output := &dynamodb.BatchWriteItemOutput{}

	// Simulate unprocessed items on first call only
	if m.returnUnprocessed && !m.unprocessedOnce {
		m.unprocessedOnce = true
		// Return one unprocessed item
		for tableName, reqs := range params.RequestItems {
			if len(reqs) > 0 {
				output.UnprocessedItems = map[string][]types.WriteRequest{
					tableName: {reqs[0]},
				}
			}
			break
		}
	}

	return output, nil
}

func makeTestLifts(n int) []Lift {
	lifts := make([]Lift, n)
	for i := range lifts {
		lifts[i] = Lift{
			ID:           fmt.Sprintf("lift#test#exercise#%d", i+1),
			Date:         "2022-05-11 04:20:50",
			WorkoutName:  "Test Workout",
			Duration:     "60m",
			ExerciseName: "Squat (Barbell)",
			SetOrder:     fmt.Sprintf("%d", i+1),
			Weight:       225.0,
			Reps:         5.0,
		}
	}
	return lifts
}

func TestBatchWriteLifts_Success(t *testing.T) {
	mock := &mockBatchWriter{}
	lifts := makeTestLifts(60) // 60 items = 3 batches of 25, 25, 10

	err := BatchWriteLifts(context.Background(), mock, "test-table", lifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.calls != 3 {
		t.Errorf("expected 3 batch calls, got %d", mock.calls)
	}
	if mock.totalItems != 60 {
		t.Errorf("expected 60 total items written, got %d", mock.totalItems)
	}
}

func TestBatchWriteLifts_ExactBatchSize(t *testing.T) {
	mock := &mockBatchWriter{}
	lifts := makeTestLifts(25) // Exactly one batch

	err := BatchWriteLifts(context.Background(), mock, "test-table", lifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.calls != 1 {
		t.Errorf("expected 1 batch call, got %d", mock.calls)
	}
}

func TestBatchWriteLifts_Empty(t *testing.T) {
	mock := &mockBatchWriter{}

	err := BatchWriteLifts(context.Background(), mock, "test-table", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.calls != 0 {
		t.Errorf("expected 0 batch calls for empty input, got %d", mock.calls)
	}
}

func TestBatchWriteLifts_UnprocessedRetry(t *testing.T) {
	mock := &mockBatchWriter{returnUnprocessed: true}
	lifts := makeTestLifts(5)

	err := BatchWriteLifts(context.Background(), mock, "test-table", lifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have initial call + retry call
	if mock.calls < 2 {
		t.Errorf("expected at least 2 calls (initial + retry), got %d", mock.calls)
	}
}

func TestBatchWriteLifts_Error(t *testing.T) {
	mock := &mockBatchWriter{err: fmt.Errorf("throttled")}
	lifts := makeTestLifts(5)

	err := BatchWriteLifts(context.Background(), mock, "test-table", lifts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBatchWriteLifts_MarshalCheck(t *testing.T) {
	// Verify that lifts are properly marshaled into DynamoDB items
	mock := &mockBatchWriter{}
	lifts := makeTestLifts(1)

	err := BatchWriteLifts(context.Background(), mock, "test-table", lifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we can roundtrip a lift through marshal/unmarshal
	item, err := attributevalue.MarshalMap(lifts[0])
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var roundtrip Lift
	if err := attributevalue.UnmarshalMap(item, &roundtrip); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if roundtrip.ID != lifts[0].ID {
		t.Errorf("roundtrip ID mismatch: got %q, want %q", roundtrip.ID, lifts[0].ID)
	}
}
