// ABOUTME: This file contains tests for the Lambda adapter.
// ABOUTME: It verifies API key validation and request routing logic.
package lambda

import (
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jduncan/josh-bot/internal/adapters/mock"
)

func TestRouter_ValidAPIKey(t *testing.T) {
	t.Setenv("API_KEY", "test-secret-key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		Path:    "/v1/status",
		Headers: map[string]string{"x-api-key": "test-secret-key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRouter_InvalidAPIKey(t *testing.T) {
	t.Setenv("API_KEY", "test-secret-key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		Path:    "/v1/status",
		Headers: map[string]string{"x-api-key": "wrong-key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRouter_MissingAPIKey(t *testing.T) {
	t.Setenv("API_KEY", "test-secret-key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		Path:    "/v1/status",
		Headers: map[string]string{},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRouter_NoAPIKeyConfigured(t *testing.T) {
	t.Setenv("API_KEY", "")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		Path:    "/v1/status",
		Headers: map[string]string{},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 when no API key configured, got %d", resp.StatusCode)
	}
}

func TestRouter_StatusEndpoint(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		Path:    "/v1/status",
		Headers: map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Body, "Refining Go backends") {
		t.Errorf("unexpected body: %v", resp.Body)
	}
	if !strings.Contains(resp.Body, `"status":"ok"`) {
		t.Errorf("missing status field in body: %v", resp.Body)
	}
}

func TestRouter_ProjectsEndpoint(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		Path:    "/v1/projects",
		Headers: map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Body, "Modular AWS Backend") {
		t.Errorf("unexpected body: %v", resp.Body)
	}
}

func TestRouter_NotFound(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		Path:    "/v1/unknown",
		Headers: map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
