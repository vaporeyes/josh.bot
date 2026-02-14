// ABOUTME: This file contains tests for the Lambda adapter.
// ABOUTME: It verifies API key validation and request routing logic.
package lambda

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jduncan/josh-bot/internal/adapters/mock"
	"github.com/jduncan/josh-bot/internal/domain"
)

func TestRouter_ValidAPIKey(t *testing.T) {
	t.Setenv("API_KEY", "test-secret-key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/v1/status",
		Headers:    map[string]string{"x-api-key": "test-secret-key"},
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
		HTTPMethod: "GET",
		Path:       "/v1/status",
		Headers:    map[string]string{"x-api-key": "wrong-key"},
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
		HTTPMethod: "GET",
		Path:       "/v1/status",
		Headers:    map[string]string{},
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
		HTTPMethod: "GET",
		Path:       "/v1/status",
		Headers:    map[string]string{},
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
		HTTPMethod: "GET",
		Path:       "/v1/status",
		Headers:    map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var status domain.Status
	if err := json.Unmarshal([]byte(resp.Body), &status); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if status.Name != "Josh Duncan" {
		t.Errorf("expected name 'Josh Duncan', got '%s'", status.Name)
	}
	if status.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", status.Status)
	}
	if len(status.Links) == 0 {
		t.Error("expected links to be non-empty")
	}
	if len(status.Interests) == 0 {
		t.Error("expected interests to be non-empty")
	}
}

func TestRouter_ProjectsEndpoint(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/v1/projects",
		Headers:    map[string]string{"x-api-key": "key"},
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

func TestRouter_PutStatus_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/v1/status",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{"current_activity": "Deploying josh.bot", "availability": "Heads down"}`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestRouter_PutStatus_InvalidJSON(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/v1/status",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{not valid json`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRouter_PostStatus_MethodNotAllowed(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/v1/status",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{"status": "busy"}`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 405 {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestRouter_PutProjects_MethodNotAllowed(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/v1/projects",
		Headers:    map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 405 {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestRouter_NotFound(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/v1/unknown",
		Headers:    map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
