// ABOUTME: This file contains tests for the DynamoDB-backed WebhookService.
// ABOUTME: It uses the shared mockDynamoDBClient to test without hitting real DynamoDB.
package dynamodb

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

func TestWebhookService_CreateWebhookEvent(t *testing.T) {
	mock := &mockDynamoDBClient{
		putOutput: &dynamodb.PutItemOutput{},
	}
	svc := NewWebhookService(mock, "test-table")

	event := webhookEventFixture()
	err := svc.CreateWebhookEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.putInput == nil {
		t.Fatal("expected PutItem to be called")
	}

	// Verify the ID was set with webhook# prefix
	idAttr := mock.putInput.Item["id"]
	idVal, ok := idAttr.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatal("expected id to be a string attribute")
	}
	if !strings.HasPrefix(idVal.Value, "webhook#") {
		t.Errorf("expected id to start with 'webhook#', got %q", idVal.Value)
	}

	// Verify created_at was set
	if _, ok := mock.putInput.Item["created_at"]; !ok {
		t.Error("expected created_at to be set")
	}

	// Verify item_type is set for GSI
	typeAttr, ok := mock.putInput.Item["item_type"]
	if !ok {
		t.Fatal("expected 'item_type' in put item")
	}
	if typeAttr.(*types.AttributeValueMemberS).Value != "webhook" {
		t.Errorf("expected item_type 'webhook', got %q", typeAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestWebhookService_GetWebhookEvents_NoFilters(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				webhookItemFixture("webhook#aaa", "message", "k8-one"),
				webhookItemFixture("webhook#bbb", "alert", "cookbot"),
			},
		},
	}
	svc := NewWebhookService(mock, "test-table")

	events, err := svc.GetWebhookEvents(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	// Verify Query was used on GSI
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	if *mock.queryInput.IndexName != "item-type-index" {
		t.Errorf("expected index 'item-type-index', got %q", *mock.queryInput.IndexName)
	}
}

func TestWebhookService_GetWebhookEvents_FilterByType(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				webhookItemFixture("webhook#aaa", "alert", "cookbot"),
			},
		},
	}
	svc := NewWebhookService(mock, "test-table")

	events, err := svc.GetWebhookEvents(context.Background(), "alert", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	// Verify type filter was included
	filter := *mock.queryInput.FilterExpression
	if !strings.Contains(filter, "#t = :eventType") {
		t.Errorf("expected type filter in expression, got %q", filter)
	}
}

func TestWebhookService_GetWebhookEvents_FilterBySource(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				webhookItemFixture("webhook#aaa", "message", "k8-one"),
			},
		},
	}
	svc := NewWebhookService(mock, "test-table")

	events, err := svc.GetWebhookEvents(context.Background(), "", "k8-one")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	filter := *mock.queryInput.FilterExpression
	if !strings.Contains(filter, "#s = :source") {
		t.Errorf("expected source filter in expression, got %q", filter)
	}
}

func TestWebhookService_GetWebhookEvent_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: webhookItemFixture("webhook#abc123", "message", "k8-one"),
		},
	}
	svc := NewWebhookService(mock, "test-table")

	event, err := svc.GetWebhookEvent(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Source != "k8-one" {
		t.Errorf("expected source 'k8-one', got %q", event.Source)
	}
}

func TestWebhookService_GetWebhookEvent_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}
	svc := NewWebhookService(mock, "test-table")

	_, err := svc.GetWebhookEvent(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing event")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %q", err.Error())
	}
}

func TestWebhookService_GetWebhookEvent_FullID(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: webhookItemFixture("webhook#abc123", "message", "k8-one"),
		},
	}
	svc := NewWebhookService(mock, "test-table")

	// Pass full ID with prefix
	event, err := svc.GetWebhookEvent(context.Background(), "webhook#abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Type != "message" {
		t.Errorf("expected type 'message', got %q", event.Type)
	}
}

// --- Test fixtures ---

func webhookEventFixture() webhookEventInput {
	return webhookEventInput{
		Type:   "message",
		Source: "k8-one",
	}
}

// webhookEventInput is a minimal struct for CreateWebhookEvent test input.
type webhookEventInput = domain.WebhookEvent

func webhookItemFixture(id, eventType, source string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":         &types.AttributeValueMemberS{Value: id},
		"type":       &types.AttributeValueMemberS{Value: eventType},
		"source":     &types.AttributeValueMemberS{Value: source},
		"created_at": &types.AttributeValueMemberS{Value: "2026-02-19T10:00:00Z"},
	}
}
