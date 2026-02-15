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

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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

func TestRouter_InvalidAPIKey_ProtectedRoute(t *testing.T) {
	t.Setenv("API_KEY", "test-secret-key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/v1/status",
		Headers:    map[string]string{"x-api-key": "wrong-key"},
		Body:       `{"status":"busy"}`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRouter_MissingAPIKey_ProtectedRoute(t *testing.T) {
	t.Setenv("API_KEY", "test-secret-key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/v1/status",
		Headers:    map[string]string{},
		Body:       `{"status":"busy"}`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRouter_GetStatus_NoAPIKey_PublicRoute(t *testing.T) {
	t.Setenv("API_KEY", "test-secret-key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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
		t.Errorf("expected 200 for public GET /v1/status, got %d", resp.StatusCode)
	}
}

func TestRouter_NoAPIKeyConfigured(t *testing.T) {
	t.Setenv("API_KEY", "")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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

func TestRouter_PostProject_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/v1/projects",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{"slug":"new-proj","name":"New Project","stack":"Go","description":"A thing","url":"https://github.com/vaporeyes/new","status":"active"}`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestRouter_PostProject_InvalidJSON(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/v1/projects",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{bad json`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRouter_GetProject_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/v1/projects/modular-aws-backend",
		Headers:    map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}

	var project domain.Project
	if err := json.Unmarshal([]byte(resp.Body), &project); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if project.Name != "Modular AWS Backend" {
		t.Errorf("expected name 'Modular AWS Backend', got '%s'", project.Name)
	}
}

func TestRouter_PutProject_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/v1/projects/modular-aws-backend",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{"status":"archived"}`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestRouter_DeleteProject_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/v1/projects/modular-aws-backend",
		Headers:    map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestRouter_GetLinks_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/v1/links",
		Headers:    map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Body, "The Go Blog") {
		t.Errorf("unexpected body: %v", resp.Body)
	}
}

func TestRouter_GetLinks_FilterByTag(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod:            "GET",
		Path:                  "/v1/links",
		Headers:               map[string]string{"x-api-key": "key"},
		QueryStringParameters: map[string]string{"tag": "aws"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Body, "DynamoDB") {
		t.Errorf("expected DynamoDB link in filtered results: %v", resp.Body)
	}
	if strings.Contains(resp.Body, "Go Blog") {
		t.Errorf("expected Go Blog to be filtered out: %v", resp.Body)
	}
}

func TestRouter_PostLink_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/v1/links",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{"url":"https://example.com","title":"Example","tags":["test"]}`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestRouter_PostLink_InvalidJSON(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/v1/links",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{bad json`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRouter_GetLink_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/v1/links/a1b2c3d4e5f6",
		Headers:    map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestRouter_PutLink_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/v1/links/a1b2c3d4e5f6",
		Headers:    map[string]string{"x-api-key": "key"},
		Body:       `{"title":"Updated Title"}`,
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestRouter_DeleteLink_Success(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
	req := events.APIGatewayProxyRequest{
		HTTPMethod: "DELETE",
		Path:       "/v1/links/a1b2c3d4e5f6",
		Headers:    map[string]string{"x-api-key": "key"},
	}

	resp, err := adapter.Router(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}
}

func TestRouter_NotFound(t *testing.T) {
	t.Setenv("API_KEY", "key")

	adapter := NewAdapter(mock.NewBotService(), mock.NewMetricsService())
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
