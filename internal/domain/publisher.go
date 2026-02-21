// ABOUTME: This file defines the WebhookPublisher interface for async webhook event publishing.
// ABOUTME: It decouples the ingestor from the storage backend, enabling SQS or other queues.
package domain

import "context"

// WebhookPublisher publishes validated webhook events to an async processing queue.
// AIDEV-NOTE: Ingestor validates HMAC then publishes; processor reads from queue and writes to DynamoDB.
type WebhookPublisher interface {
	Publish(ctx context.Context, event WebhookEvent) error
}
