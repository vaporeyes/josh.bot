// ABOUTME: This file contains tests for the DynamoDB-backed WebhookService.
// ABOUTME: It uses the shared mockDynamoDBClient to test without hitting real DynamoDB.
package dynamodb

import (
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
	err := svc.CreateWebhookEvent(event)
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
}

func TestWebhookService_GetWebhookEvents_NoFilters(t *testing.T) {
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{
			Items: []map[string]types.AttributeValue{
				webhookItemFixture("webhook#aaa", "message", "k8-one"),
				webhookItemFixture("webhook#bbb", "alert", "cookbot"),
			},
		},
	}
	svc := NewWebhookService(mock, "test-table")

	events, err := svc.GetWebhookEvents("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	// Verify scan filter uses begins_with
	if mock.scanInput == nil {
		t.Fatal("expected Scan to be called")
	}
	if !strings.Contains(*mock.scanInput.FilterExpression, "begins_with(id, :prefix)") {
		t.Errorf("expected begins_with filter, got %q", *mock.scanInput.FilterExpression)
	}
}

func TestWebhookService_GetWebhookEvents_FilterByType(t *testing.T) {
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{
			Items: []map[string]types.AttributeValue{
				webhookItemFixture("webhook#aaa", "alert", "cookbot"),
			},
		},
	}
	svc := NewWebhookService(mock, "test-table")

	events, err := svc.GetWebhookEvents("alert", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	// Verify type filter was included
	filter := *mock.scanInput.FilterExpression
	if !strings.Contains(filter, "#t = :eventType") {
		t.Errorf("expected type filter in expression, got %q", filter)
	}
}

func TestWebhookService_GetWebhookEvents_FilterBySource(t *testing.T) {
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{
			Items: []map[string]types.AttributeValue{
				webhookItemFixture("webhook#aaa", "message", "k8-one"),
			},
		},
	}
	svc := NewWebhookService(mock, "test-table")

	events, err := svc.GetWebhookEvents("", "k8-one")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}

	filter := *mock.scanInput.FilterExpression
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

	event, err := svc.GetWebhookEvent("abc123")
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

	_, err := svc.GetWebhookEvent("nonexistent")
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
	event, err := svc.GetWebhookEvent("webhook#abc123")
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
