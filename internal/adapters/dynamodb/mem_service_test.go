// ABOUTME: This file tests the DynamoDB-backed MemService for reading claude-mem data.
// ABOUTME: It uses a mock DynamoDBClient to test Query/Scan/GetItem operations on the mem table.
package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// --- Observation Tests ---

func TestGetObservations_WithType(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":               &types.AttributeValueMemberS{Value: "obs#42"},
					"type":             &types.AttributeValueMemberS{Value: "decision"},
					"source_id":        &types.AttributeValueMemberN{Value: "42"},
					"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
					"session_id":       &types.AttributeValueMemberS{Value: "sess-abc"},
					"title":            &types.AttributeValueMemberS{Value: "Chose DynamoDB"},
					"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
					"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	obs, err := svc.GetObservations(context.Background(), "decision", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(obs))
	}
	if obs[0].ID != "obs#42" {
		t.Errorf("expected id 'obs#42', got '%s'", obs[0].ID)
	}
	if obs[0].Title != "Chose DynamoDB" {
		t.Errorf("expected title 'Chose DynamoDB', got '%s'", obs[0].Title)
	}
	// Verify Query was used (not Scan) when type is specified
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called when type is specified")
	}
}

func TestGetObservations_FilterByProject(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":               &types.AttributeValueMemberS{Value: "obs#42"},
					"type":             &types.AttributeValueMemberS{Value: "decision"},
					"source_id":        &types.AttributeValueMemberN{Value: "42"},
					"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
					"session_id":       &types.AttributeValueMemberS{Value: "sess-abc"},
					"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
					"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	obs, err := svc.GetObservations(context.Background(), "decision", "josh.bot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(obs))
	}
	// Verify filter expression includes project
	if mock.queryInput.FilterExpression == nil {
		t.Error("expected filter expression for project")
	}
}

func TestGetObservations_NoTypeFilter(t *testing.T) {
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":               &types.AttributeValueMemberS{Value: "obs#42"},
					"type":             &types.AttributeValueMemberS{Value: "decision"},
					"source_id":        &types.AttributeValueMemberN{Value: "42"},
					"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
					"session_id":       &types.AttributeValueMemberS{Value: "sess-abc"},
					"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
					"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	obs, err := svc.GetObservations(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(obs))
	}
	// Verify Scan was used (not Query) when no type filter
	if mock.scanInput == nil {
		t.Fatal("expected Scan to be called when no type filter")
	}
}

func TestGetObservation_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":               &types.AttributeValueMemberS{Value: "obs#42"},
				"type":             &types.AttributeValueMemberS{Value: "decision"},
				"source_id":        &types.AttributeValueMemberN{Value: "42"},
				"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
				"session_id":       &types.AttributeValueMemberS{Value: "sess-abc"},
				"title":            &types.AttributeValueMemberS{Value: "Chose DynamoDB"},
				"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
				"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	obs, err := svc.GetObservation(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ID != "obs#42" {
		t.Errorf("expected id 'obs#42', got '%s'", obs.ID)
	}
	if obs.Title != "Chose DynamoDB" {
		t.Errorf("expected title 'Chose DynamoDB', got '%s'", obs.Title)
	}
}

func TestGetObservation_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	_, err := svc.GetObservation(context.Background(), "999")
	if err == nil {
		t.Error("expected error for missing observation, got nil")
	}
}

// --- Summary Tests ---

func TestGetSummaries_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":               &types.AttributeValueMemberS{Value: "summary#10"},
					"type":             &types.AttributeValueMemberS{Value: "summary"},
					"source_id":        &types.AttributeValueMemberN{Value: "10"},
					"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
					"session_id":       &types.AttributeValueMemberS{Value: "sess-xyz"},
					"request":          &types.AttributeValueMemberS{Value: "Add mem endpoints"},
					"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
					"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	summaries, err := svc.GetSummaries(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Request != "Add mem endpoints" {
		t.Errorf("expected request 'Add mem endpoints', got '%s'", summaries[0].Request)
	}
}

func TestGetSummaries_FilterByProject(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":               &types.AttributeValueMemberS{Value: "summary#10"},
					"type":             &types.AttributeValueMemberS{Value: "summary"},
					"source_id":        &types.AttributeValueMemberN{Value: "10"},
					"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
					"session_id":       &types.AttributeValueMemberS{Value: "sess-xyz"},
					"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
					"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	summaries, err := svc.GetSummaries(context.Background(), "josh.bot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	if mock.queryInput.FilterExpression == nil {
		t.Error("expected filter expression for project")
	}
}

func TestGetSummary_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":               &types.AttributeValueMemberS{Value: "summary#10"},
				"type":             &types.AttributeValueMemberS{Value: "summary"},
				"source_id":        &types.AttributeValueMemberN{Value: "10"},
				"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
				"session_id":       &types.AttributeValueMemberS{Value: "sess-xyz"},
				"request":          &types.AttributeValueMemberS{Value: "Add mem endpoints"},
				"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
				"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	summary, err := svc.GetSummary(context.Background(), "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Request != "Add mem endpoints" {
		t.Errorf("expected request 'Add mem endpoints', got '%s'", summary.Request)
	}
}

func TestGetSummary_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	_, err := svc.GetSummary(context.Background(), "999")
	if err == nil {
		t.Error("expected error for missing summary, got nil")
	}
}

// --- Prompt Tests ---

func TestGetPrompts_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":               &types.AttributeValueMemberS{Value: "prompt#5"},
					"type":             &types.AttributeValueMemberS{Value: "prompt"},
					"source_id":        &types.AttributeValueMemberN{Value: "5"},
					"session_id":       &types.AttributeValueMemberS{Value: "sess-abc"},
					"prompt_number":    &types.AttributeValueMemberN{Value: "3"},
					"prompt_text":      &types.AttributeValueMemberS{Value: "Show me the plan"},
					"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
					"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	prompts, err := svc.GetPrompts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0].PromptText != "Show me the plan" {
		t.Errorf("expected prompt_text 'Show me the plan', got '%s'", prompts[0].PromptText)
	}
}

func TestGetPrompt_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":               &types.AttributeValueMemberS{Value: "prompt#5"},
				"type":             &types.AttributeValueMemberS{Value: "prompt"},
				"source_id":        &types.AttributeValueMemberN{Value: "5"},
				"session_id":       &types.AttributeValueMemberS{Value: "sess-abc"},
				"prompt_number":    &types.AttributeValueMemberN{Value: "3"},
				"prompt_text":      &types.AttributeValueMemberS{Value: "Show me the plan"},
				"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
				"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	prompt, err := svc.GetPrompt(context.Background(), "5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt.PromptText != "Show me the plan" {
		t.Errorf("expected prompt_text 'Show me the plan', got '%s'", prompt.PromptText)
	}
}

func TestGetPrompt_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	_, err := svc.GetPrompt(context.Background(), "999")
	if err == nil {
		t.Error("expected error for missing prompt, got nil")
	}
}

// --- Stats Tests ---

func TestGetStats_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{
			Items: []map[string]types.AttributeValue{
				{"id": &types.AttributeValueMemberS{Value: "obs#1"}, "type": &types.AttributeValueMemberS{Value: "decision"}, "project": &types.AttributeValueMemberS{Value: "josh.bot"}},
				{"id": &types.AttributeValueMemberS{Value: "obs#2"}, "type": &types.AttributeValueMemberS{Value: "feature"}, "project": &types.AttributeValueMemberS{Value: "josh.bot"}},
				{"id": &types.AttributeValueMemberS{Value: "obs#3"}, "type": &types.AttributeValueMemberS{Value: "decision"}, "project": &types.AttributeValueMemberS{Value: "other"}},
				{"id": &types.AttributeValueMemberS{Value: "summary#1"}, "type": &types.AttributeValueMemberS{Value: "summary"}, "project": &types.AttributeValueMemberS{Value: "josh.bot"}},
				{"id": &types.AttributeValueMemberS{Value: "prompt#1"}, "type": &types.AttributeValueMemberS{Value: "prompt"}},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	stats, err := svc.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalObservations != 3 {
		t.Errorf("expected 3 observations, got %d", stats.TotalObservations)
	}
	if stats.TotalSummaries != 1 {
		t.Errorf("expected 1 summary, got %d", stats.TotalSummaries)
	}
	if stats.TotalPrompts != 1 {
		t.Errorf("expected 1 prompt, got %d", stats.TotalPrompts)
	}
	if stats.ByType["decision"] != 2 {
		t.Errorf("expected 2 decisions, got %d", stats.ByType["decision"])
	}
	if stats.ByProject["josh.bot"] != 3 {
		t.Errorf("expected 3 items for josh.bot, got %d", stats.ByProject["josh.bot"])
	}
}

// --- Pagination Tests ---

func TestGetStats_Paginated(t *testing.T) {
	// Simulate DynamoDB returning results across two pages
	mock := &mockDynamoDBClient{
		scanOutputs: []*dynamodb.ScanOutput{
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "obs#1"}, "type": &types.AttributeValueMemberS{Value: "decision"}, "project": &types.AttributeValueMemberS{Value: "josh.bot"}},
					{"id": &types.AttributeValueMemberS{Value: "obs#2"}, "type": &types.AttributeValueMemberS{Value: "feature"}, "project": &types.AttributeValueMemberS{Value: "josh.bot"}},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "obs#2"},
				},
			},
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "obs#3"}, "type": &types.AttributeValueMemberS{Value: "decision"}, "project": &types.AttributeValueMemberS{Value: "other"}},
					{"id": &types.AttributeValueMemberS{Value: "summary#1"}, "type": &types.AttributeValueMemberS{Value: "summary"}, "project": &types.AttributeValueMemberS{Value: "josh.bot"}},
					{"id": &types.AttributeValueMemberS{Value: "prompt#1"}, "type": &types.AttributeValueMemberS{Value: "prompt"}},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	stats, err := svc.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalObservations != 3 {
		t.Errorf("expected 3 observations, got %d", stats.TotalObservations)
	}
	if stats.TotalSummaries != 1 {
		t.Errorf("expected 1 summary, got %d", stats.TotalSummaries)
	}
	if stats.TotalPrompts != 1 {
		t.Errorf("expected 1 prompt, got %d", stats.TotalPrompts)
	}
	if mock.scanCallNum != 2 {
		t.Errorf("expected 2 scan calls, got %d", mock.scanCallNum)
	}
}

func TestGetObservations_NoType_Paginated(t *testing.T) {
	// scanByPrefix should paginate across multiple pages
	mock := &mockDynamoDBClient{
		scanOutputs: []*dynamodb.ScanOutput{
			{
				Items: []map[string]types.AttributeValue{
					{
						"id":               &types.AttributeValueMemberS{Value: "obs#1"},
						"type":             &types.AttributeValueMemberS{Value: "decision"},
						"source_id":        &types.AttributeValueMemberN{Value: "1"},
						"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
						"session_id":       &types.AttributeValueMemberS{Value: "sess-abc"},
						"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
						"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
					},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "obs#1"},
				},
			},
			{
				Items: []map[string]types.AttributeValue{
					{
						"id":               &types.AttributeValueMemberS{Value: "obs#2"},
						"type":             &types.AttributeValueMemberS{Value: "feature"},
						"source_id":        &types.AttributeValueMemberN{Value: "2"},
						"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
						"session_id":       &types.AttributeValueMemberS{Value: "sess-def"},
						"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T13:00:00Z"},
						"created_at_epoch": &types.AttributeValueMemberN{Value: "1739624400"},
					},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	obs, err := svc.GetObservations(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(obs) != 2 {
		t.Errorf("expected 2 observations across pages, got %d", len(obs))
	}
	if mock.scanCallNum != 2 {
		t.Errorf("expected 2 scan calls, got %d", mock.scanCallNum)
	}
}

func TestGetSummaries_Paginated(t *testing.T) {
	// queryByType should paginate across multiple pages
	mock := &mockDynamoDBClient{
		queryOutputs: []*dynamodb.QueryOutput{
			{
				Items: []map[string]types.AttributeValue{
					{
						"id":               &types.AttributeValueMemberS{Value: "summary#1"},
						"type":             &types.AttributeValueMemberS{Value: "summary"},
						"source_id":        &types.AttributeValueMemberN{Value: "1"},
						"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
						"session_id":       &types.AttributeValueMemberS{Value: "sess-abc"},
						"request":          &types.AttributeValueMemberS{Value: "First request"},
						"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T12:00:00Z"},
						"created_at_epoch": &types.AttributeValueMemberN{Value: "1739620800"},
					},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "summary#1"},
				},
			},
			{
				Items: []map[string]types.AttributeValue{
					{
						"id":               &types.AttributeValueMemberS{Value: "summary#2"},
						"type":             &types.AttributeValueMemberS{Value: "summary"},
						"source_id":        &types.AttributeValueMemberN{Value: "2"},
						"project":          &types.AttributeValueMemberS{Value: "josh.bot"},
						"session_id":       &types.AttributeValueMemberS{Value: "sess-def"},
						"request":          &types.AttributeValueMemberS{Value: "Second request"},
						"created_at":       &types.AttributeValueMemberS{Value: "2026-02-15T13:00:00Z"},
						"created_at_epoch": &types.AttributeValueMemberN{Value: "1739624400"},
					},
				},
			},
		},
	}

	svc := NewMemService(mock, "josh-bot-mem")
	summaries, err := svc.GetSummaries(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("expected 2 summaries across pages, got %d", len(summaries))
	}
	if mock.queryCallNum != 2 {
		t.Errorf("expected 2 query calls, got %d", mock.queryCallNum)
	}
}
