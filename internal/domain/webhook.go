// ABOUTME: This file defines domain types and helpers for inbound webhook events.
// ABOUTME: It provides HMAC-SHA256 signature validation for bot-to-bot communication.
package domain

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// WebhookEvent represents an inbound webhook event from another bot.
// AIDEV-NOTE: Events are immutable once received (append-only log).
type WebhookEvent struct {
	ID        string         `json:"id" dynamodbav:"id"`
	Type      string         `json:"type" dynamodbav:"type"`
	Source    string         `json:"source" dynamodbav:"source"`
	Payload   map[string]any `json:"payload" dynamodbav:"payload"`
	CreatedAt string         `json:"created_at" dynamodbav:"created_at"`
}

// WebhookEventID generates a random ID with a "webhook#" prefix.
// AIDEV-NOTE: Random IDs since webhook events have no natural dedup key.
func WebhookEventID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "webhook#" + hex.EncodeToString(b)
}

// WebhookService defines operations for webhook event storage and retrieval.
type WebhookService interface {
	CreateWebhookEvent(event WebhookEvent) error
	GetWebhookEvents(eventType, source string) ([]WebhookEvent, error)
	GetWebhookEvent(id string) (WebhookEvent, error)
}

// ComputeWebhookSignature computes an HMAC-SHA256 hex digest of the body using the secret.
func ComputeWebhookSignature(body, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateWebhookSignature checks that the provided signature header matches the
// expected HMAC-SHA256 of the body. The signature must be in "sha256=<hex>" format.
// Uses constant-time comparison to prevent timing attacks.
func ValidateWebhookSignature(body, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	provided := strings.TrimPrefix(signature, "sha256=")
	expected := ComputeWebhookSignature(body, secret)
	return hmac.Equal([]byte(provided), []byte(expected))
}
