// ABOUTME: This file implements the SQS-backed WebhookPublisher for async event processing.
// ABOUTME: It serializes webhook events to JSON and sends them to an SQS queue.
package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jduncan/josh-bot/internal/domain"
)

// SQSClient is the subset of the SQS API used by the publisher.
type SQSClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// Publisher sends webhook events to an SQS queue for async processing.
type Publisher struct {
	client   SQSClient
	queueURL string
}

// NewPublisher creates a new SQS publisher.
func NewPublisher(client SQSClient, queueURL string) *Publisher {
	return &Publisher{client: client, queueURL: queueURL}
}

// Publish serializes the webhook event to JSON and sends it to the SQS queue.
func (p *Publisher) Publish(ctx context.Context, event domain.WebhookEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal webhook event: %w", err)
	}

	msgBody := string(body)
	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &p.queueURL,
		MessageBody: &msgBody,
	})
	return err
}
