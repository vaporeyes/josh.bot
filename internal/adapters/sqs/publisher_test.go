// ABOUTME: This file contains tests for the SQS webhook publisher.
// ABOUTME: It verifies message serialization and error handling using a mock SQS client.
package sqs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jduncan/josh-bot/internal/domain"
)

// mockSQSClient is a test double for the SQS SendMessage API.
type mockSQSClient struct {
	input *sqs.SendMessageInput
	err   error
}

func (m *mockSQSClient) SendMessage(_ context.Context, input *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	m.input = input
	return &sqs.SendMessageOutput{}, m.err
}

func TestSQSPublisher_Publish_SendsMessage(t *testing.T) {
	mock := &mockSQSClient{}
	pub := NewPublisher(mock, "https://sqs.us-east-1.amazonaws.com/123456789/test-queue")

	event := domain.WebhookEvent{
		ID:      "webhook#abc123",
		Type:    "message",
		Source:  "k8-one",
		Payload: map[string]any{"text": "hello"},
	}

	err := pub.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.input == nil {
		t.Fatal("expected SendMessage to be called")
	}
	if *mock.input.QueueUrl != "https://sqs.us-east-1.amazonaws.com/123456789/test-queue" {
		t.Errorf("unexpected queue URL: %s", *mock.input.QueueUrl)
	}

	// Verify the message body is valid JSON matching the event
	var decoded domain.WebhookEvent
	if err := json.Unmarshal([]byte(*mock.input.MessageBody), &decoded); err != nil {
		t.Fatalf("failed to unmarshal message body: %v", err)
	}
	if decoded.Type != "message" {
		t.Errorf("expected type 'message', got '%s'", decoded.Type)
	}
	if decoded.Source != "k8-one" {
		t.Errorf("expected source 'k8-one', got '%s'", decoded.Source)
	}
}

func TestSQSPublisher_Publish_SQSError_ReturnsError(t *testing.T) {
	mock := &mockSQSClient{err: errors.New("SQS unavailable")}
	pub := NewPublisher(mock, "https://sqs.us-east-1.amazonaws.com/123456789/test-queue")

	event := domain.WebhookEvent{
		Type:    "message",
		Source:  "test",
		Payload: map[string]any{"text": "hello"},
	}

	err := pub.Publish(context.Background(), event)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, mock.err) {
		t.Errorf("expected SQS error, got: %v", err)
	}
}
