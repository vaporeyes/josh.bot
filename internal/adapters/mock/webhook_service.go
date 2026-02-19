// ABOUTME: This file provides a mock implementation of WebhookService for testing.
// ABOUTME: It returns hardcoded webhook events and supports type/source filtering.
package mock

import (
	"fmt"

	"github.com/jduncan/josh-bot/internal/domain"
)

// WebhookService is a mock implementation of domain.WebhookService.
type WebhookService struct{}

// NewWebhookService creates a new mock WebhookService.
func NewWebhookService() *WebhookService {
	return &WebhookService{}
}

// CreateWebhookEvent is a no-op in the mock adapter.
func (s *WebhookService) CreateWebhookEvent(event domain.WebhookEvent) error {
	return nil
}

// GetWebhookEvents returns hardcoded events, optionally filtered by type and source.
func (s *WebhookService) GetWebhookEvents(eventType, source string) ([]domain.WebhookEvent, error) {
	events := []domain.WebhookEvent{
		{
			ID:        "webhook#abc123def456",
			Type:      "message",
			Source:    "k8-one",
			Payload:   map[string]any{"text": "hello from k8-one"},
			CreatedAt: "2026-02-19T10:00:00Z",
		},
		{
			ID:        "webhook#fed987cba654",
			Type:      "alert",
			Source:    "cookbot",
			Payload:   map[string]any{"level": "warning", "message": "oven too hot"},
			CreatedAt: "2026-02-19T11:00:00Z",
		},
	}

	if eventType == "" && source == "" {
		return events, nil
	}

	var filtered []domain.WebhookEvent
	for _, e := range events {
		if eventType != "" && e.Type != eventType {
			continue
		}
		if source != "" && e.Source != source {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered, nil
}

// GetWebhookEvent returns a hardcoded event by ID.
func (s *WebhookService) GetWebhookEvent(id string) (domain.WebhookEvent, error) {
	events, _ := s.GetWebhookEvents("", "")
	for _, e := range events {
		if e.ID == id || e.ID == "webhook#"+id {
			return e, nil
		}
	}
	return domain.WebhookEvent{}, fmt.Errorf("webhook event %q not found", id)
}
