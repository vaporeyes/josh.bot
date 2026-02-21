// ABOUTME: This file provides a mock implementation of WebhookPublisher for testing.
// ABOUTME: It records published events and supports error injection.
package mock

import (
	"context"

	"github.com/jduncan/josh-bot/internal/domain"
)

// WebhookPublisher is a mock implementation of domain.WebhookPublisher.
type WebhookPublisher struct {
	Published []domain.WebhookEvent
	Err       error
}

// NewWebhookPublisher creates a new mock WebhookPublisher.
func NewWebhookPublisher() *WebhookPublisher {
	return &WebhookPublisher{}
}

// Publish records the event and returns the configured error.
func (p *WebhookPublisher) Publish(_ context.Context, event domain.WebhookEvent) error {
	if p.Err != nil {
		return p.Err
	}
	p.Published = append(p.Published, event)
	return nil
}
