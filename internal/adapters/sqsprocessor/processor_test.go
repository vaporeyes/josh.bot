// ABOUTME: This file contains tests for the SQS webhook processor.
// ABOUTME: It verifies event deserialization, DynamoDB writes, and partial batch failure handling.
package sqsprocessor

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jduncan/josh-bot/internal/adapters/mock"
	"github.com/jduncan/josh-bot/internal/domain"
)

// errorWebhookService wraps mock.WebhookService and injects errors for specific message IDs.
type errorWebhookService struct {
	mock.WebhookService
	failForIDs map[string]error
	created    []domain.WebhookEvent
}

func (s *errorWebhookService) CreateWebhookEvent(_ context.Context, event domain.WebhookEvent) error {
	if err, ok := s.failForIDs[event.ID]; ok {
		return err
	}
	s.created = append(s.created, event)
	return nil
}

func TestProcessor_Handle_SingleRecord_Success(t *testing.T) {
	ws := &errorWebhookService{failForIDs: map[string]error{}}
	proc := NewProcessor(ws)

	event := domain.WebhookEvent{
		ID:      "webhook#abc123",
		Type:    "message",
		Source:  "k8-one",
		Payload: map[string]any{"text": "hello"},
	}
	body, _ := json.Marshal(event)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "msg-1", Body: string(body)},
		},
	}

	resp, err := proc.Handle(context.Background(), sqsEvent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("expected no failures, got %d", len(resp.BatchItemFailures))
	}
	if len(ws.created) != 1 {
		t.Fatalf("expected 1 created event, got %d", len(ws.created))
	}
	if ws.created[0].Type != "message" {
		t.Errorf("expected type 'message', got '%s'", ws.created[0].Type)
	}
}

func TestProcessor_Handle_MultipleRecords_AllSuccess(t *testing.T) {
	ws := &errorWebhookService{failForIDs: map[string]error{}}
	proc := NewProcessor(ws)

	events1 := []domain.WebhookEvent{
		{ID: "webhook#aaa", Type: "message", Source: "bot-a", Payload: map[string]any{"n": 1}},
		{ID: "webhook#bbb", Type: "alert", Source: "bot-b", Payload: map[string]any{"n": 2}},
	}

	var records []events.SQSMessage
	for i, e := range events1 {
		body, _ := json.Marshal(e)
		records = append(records, events.SQSMessage{
			MessageId: "msg-" + string(rune('1'+i)),
			Body:      string(body),
		})
	}

	resp, err := proc.Handle(context.Background(), events.SQSEvent{Records: records})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BatchItemFailures) != 0 {
		t.Errorf("expected no failures, got %d", len(resp.BatchItemFailures))
	}
	if len(ws.created) != 2 {
		t.Errorf("expected 2 created events, got %d", len(ws.created))
	}
}

func TestProcessor_Handle_InvalidJSON_RecordFails(t *testing.T) {
	ws := &errorWebhookService{failForIDs: map[string]error{}}
	proc := NewProcessor(ws)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "msg-bad", Body: `{not valid json`},
		},
	}

	resp, err := proc.Handle(context.Background(), sqsEvent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(resp.BatchItemFailures))
	}
	if resp.BatchItemFailures[0].ItemIdentifier != "msg-bad" {
		t.Errorf("expected failed message ID 'msg-bad', got '%s'", resp.BatchItemFailures[0].ItemIdentifier)
	}
}

func TestProcessor_Handle_WebhookServiceError_RecordFails(t *testing.T) {
	ws := &errorWebhookService{
		failForIDs: map[string]error{
			"webhook#fail": errors.New("DynamoDB write failed"),
		},
	}
	proc := NewProcessor(ws)

	event := domain.WebhookEvent{
		ID:      "webhook#fail",
		Type:    "alert",
		Source:  "test",
		Payload: map[string]any{"msg": "kaboom"},
	}
	body, _ := json.Marshal(event)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "msg-fail", Body: string(body)},
		},
	}

	resp, err := proc.Handle(context.Background(), sqsEvent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(resp.BatchItemFailures))
	}
	if resp.BatchItemFailures[0].ItemIdentifier != "msg-fail" {
		t.Errorf("expected failed message ID 'msg-fail', got '%s'", resp.BatchItemFailures[0].ItemIdentifier)
	}
}

func TestProcessor_Handle_PartialFailure_OnlyFailedRecordInResponse(t *testing.T) {
	ws := &errorWebhookService{
		failForIDs: map[string]error{
			"webhook#bad": errors.New("write failed"),
		},
	}
	proc := NewProcessor(ws)

	goodEvent := domain.WebhookEvent{ID: "webhook#good", Type: "message", Source: "bot", Payload: map[string]any{}}
	badEvent := domain.WebhookEvent{ID: "webhook#bad", Type: "alert", Source: "bot", Payload: map[string]any{}}
	goodBody, _ := json.Marshal(goodEvent)
	badBody, _ := json.Marshal(badEvent)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "msg-good", Body: string(goodBody)},
			{MessageId: "msg-bad", Body: string(badBody)},
		},
	}

	resp, err := proc.Handle(context.Background(), sqsEvent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(resp.BatchItemFailures))
	}
	if resp.BatchItemFailures[0].ItemIdentifier != "msg-bad" {
		t.Errorf("expected failed message ID 'msg-bad', got '%s'", resp.BatchItemFailures[0].ItemIdentifier)
	}
	if len(ws.created) != 1 {
		t.Errorf("expected 1 successful creation, got %d", len(ws.created))
	}
}
