// ABOUTME: This file implements the SQS webhook event processor Lambda handler.
// ABOUTME: It reads webhook events from SQS, writes them to DynamoDB, and reports partial failures.
package sqsprocessor

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jduncan/josh-bot/internal/domain"
)

// Processor reads webhook events from SQS and persists them via WebhookService.
// AIDEV-NOTE: Uses ReportBatchItemFailures so only failed records retry.
type Processor struct {
	webhookService domain.WebhookService
}

// NewProcessor creates a new SQS webhook event processor.
func NewProcessor(ws domain.WebhookService) *Processor {
	return &Processor{webhookService: ws}
}

// Handle processes an SQS batch of webhook events, returning partial failures.
func (p *Processor) Handle(ctx context.Context, sqsEvent events.SQSEvent) (events.SQSEventResponse, error) {
	var failures []events.SQSBatchItemFailure

	for _, record := range sqsEvent.Records {
		var event domain.WebhookEvent
		if err := json.Unmarshal([]byte(record.Body), &event); err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal webhook event",
				"message_id", record.MessageId, "error", err)
			failures = append(failures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
			continue
		}

		if err := p.webhookService.CreateWebhookEvent(ctx, event); err != nil {
			slog.ErrorContext(ctx, "failed to write webhook event",
				"message_id", record.MessageId, "event_id", event.ID, "error", err)
			failures = append(failures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
			continue
		}

		slog.InfoContext(ctx, "processed webhook event",
			"message_id", record.MessageId, "event_id", event.ID, "type", event.Type)
	}

	return events.SQSEventResponse{BatchItemFailures: failures}, nil
}
