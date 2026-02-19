// ABOUTME: This file tests webhook domain types and HMAC signature validation.
// ABOUTME: It verifies ID generation format and cryptographic signature correctness.
package domain

import (
	"strings"
	"testing"
)

func TestWebhookEventID_HasPrefix(t *testing.T) {
	id := WebhookEventID()
	if !strings.HasPrefix(id, "webhook#") {
		t.Errorf("expected prefix 'webhook#', got %q", id)
	}
}

func TestWebhookEventID_HasCorrectLength(t *testing.T) {
	id := WebhookEventID()
	// "webhook#" (8 chars) + 16 hex chars (8 bytes) = 24
	if len(id) != 24 {
		t.Errorf("expected length 24, got %d for %q", len(id), id)
	}
}

func TestWebhookEventID_IsUnique(t *testing.T) {
	id1 := WebhookEventID()
	id2 := WebhookEventID()
	if id1 == id2 {
		t.Errorf("expected unique IDs, got %q twice", id1)
	}
}

func TestValidateWebhookSignature_ValidSignature(t *testing.T) {
	body := `{"type":"message","source":"k8-one","payload":{"text":"hello"}}`
	secret := "test-secret"
	// Pre-computed: echo -n '<body>' | openssl dgst -sha256 -hmac "test-secret" -hex
	// Using the function itself to compute for now, then validate round-trip
	sig := ComputeWebhookSignature(body, secret)
	if !ValidateWebhookSignature(body, "sha256="+sig, secret) {
		t.Error("expected valid signature to pass validation")
	}
}

func TestValidateWebhookSignature_InvalidSignature(t *testing.T) {
	body := `{"type":"message"}`
	secret := "test-secret"
	if ValidateWebhookSignature(body, "sha256=deadbeef", secret) {
		t.Error("expected invalid signature to fail validation")
	}
}

func TestValidateWebhookSignature_MissingPrefix(t *testing.T) {
	body := `{"type":"message"}`
	secret := "test-secret"
	sig := ComputeWebhookSignature(body, secret)
	// Missing "sha256=" prefix should fail
	if ValidateWebhookSignature(body, sig, secret) {
		t.Error("expected signature without sha256= prefix to fail validation")
	}
}

func TestValidateWebhookSignature_EmptySignature(t *testing.T) {
	body := `{"type":"message"}`
	secret := "test-secret"
	if ValidateWebhookSignature(body, "", secret) {
		t.Error("expected empty signature to fail validation")
	}
}

func TestValidateWebhookSignature_DifferentBody(t *testing.T) {
	secret := "test-secret"
	sig := ComputeWebhookSignature(`{"type":"message"}`, secret)
	if ValidateWebhookSignature(`{"type":"alert"}`, "sha256="+sig, secret) {
		t.Error("expected signature for different body to fail validation")
	}
}

func TestComputeWebhookSignature_Deterministic(t *testing.T) {
	body := `{"type":"test"}`
	secret := "secret"
	sig1 := ComputeWebhookSignature(body, secret)
	sig2 := ComputeWebhookSignature(body, secret)
	if sig1 != sig2 {
		t.Errorf("expected deterministic signatures, got %q and %q", sig1, sig2)
	}
}
