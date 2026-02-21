// ABOUTME: This file implements the AWS Lambda adapter for the josh.bot API.
// ABOUTME: It handles API Gateway events, validates API keys, and routes requests to domain services.
package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jduncan/josh-bot/internal/domain"
)

// errorResponse maps domain errors to the appropriate HTTP status code and body.
func errorResponse(err error) (events.APIGatewayProxyResponse, error) {
	var notFound *domain.NotFoundError
	if errors.As(err, &notFound) {
		body := `{"error":"` + notFound.Resource + ` not found"}`
		return jsonResponse(404, body), nil
	}
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		body := `{"error":"` + validationErr.Error() + `"}`
		return jsonResponse(400, body), nil
	}
	return jsonResponse(500, `{"error":"internal server error"}`), err
}

// Adapter wraps domain services and handles Lambda API Gateway events.
type Adapter struct {
	service          domain.BotService
	metricsService   domain.MetricsService
	memService       domain.MemService
	diaryService     domain.DiaryService
	webhookService   domain.WebhookService
	webhookPublisher domain.WebhookPublisher
	webhookSecret    string
}

// NewAdapter creates a new Lambda adapter for the given services.
func NewAdapter(service domain.BotService, metricsService domain.MetricsService, memService domain.MemService) *Adapter {
	return &Adapter{service: service, metricsService: metricsService, memService: memService}
}

// SetDiaryService sets the diary service for the adapter.
// AIDEV-NOTE: Separate setter avoids changing NewAdapter signature for all callers.
func (a *Adapter) SetDiaryService(ds domain.DiaryService) {
	a.diaryService = ds
}

// SetWebhookService sets the webhook service and shared secret for the adapter.
// AIDEV-NOTE: Separate setter avoids changing NewAdapter signature for all callers.
func (a *Adapter) SetWebhookService(ws domain.WebhookService, secret string) {
	a.webhookService = ws
	a.webhookSecret = secret
}

// SetWebhookPublisher sets the publisher for async webhook event processing.
// AIDEV-NOTE: When set, POST /v1/webhooks publishes to queue instead of writing to DynamoDB directly.
func (a *Adapter) SetWebhookPublisher(p domain.WebhookPublisher) {
	a.webhookPublisher = p
}

// isPublicRoute returns true for routes that don't require API key auth.
func isPublicRoute(method, path string) bool {
	if method != "GET" {
		return false
	}
	return path == "/v1/status" || path == "/v1/metrics"
}

// isWebhookPost returns true for POST /v1/webhooks which uses HMAC auth instead of API key.
func isWebhookPost(method, path string) bool {
	return method == "POST" && path == "/v1/webhooks"
}

// Router handles API Gateway proxy requests with API key validation.
func (a *Adapter) Router(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// AIDEV-NOTE: client IP from CF-Connecting-IP (Cloudflare) with X-Forwarded-For fallback
	clientIP := req.Headers["cf-connecting-ip"]
	if clientIP == "" {
		clientIP = req.Headers["x-forwarded-for"]
	}
	if clientIP == "" {
		clientIP = req.RequestContext.Identity.SourceIP
	}
	slog.InfoContext(ctx, "request", "method", req.HTTPMethod, "path", req.Path, "client_ip", clientIP)

	// Handle CORS preflight
	if req.HTTPMethod == "OPTIONS" {
		return jsonResponse(204, ""), nil
	}

	// Validate API key from x-api-key header (skip for public routes and webhook POST)
	if !isPublicRoute(req.HTTPMethod, req.Path) && !isWebhookPost(req.HTTPMethod, req.Path) {
		expectedKey := os.Getenv("API_KEY")
		if expectedKey != "" && req.Headers["x-api-key"] != expectedKey {
			return jsonResponse(401, `{"error":"unauthorized"}`), nil
		}
	}

	// AIDEV-NOTE: Idempotency check on POST requests with X-Idempotency-Key header.
	// Returns cached response for duplicate keys within TTL (24h).
	idempotencyKey := req.Headers["x-idempotency-key"]
	if req.HTTPMethod == "POST" && idempotencyKey != "" {
		fullKey := domain.IdempotencyKey(req.Path, idempotencyKey)
		record, err := a.service.GetIdempotencyRecord(ctx, fullKey)
		if err != nil {
			slog.WarnContext(ctx, "idempotency lookup failed", "key", fullKey, "error", err)
		}
		if record != nil {
			slog.InfoContext(ctx, "idempotency hit", "key", fullKey)
			resp := jsonResponse(record.StatusCode, record.Body)
			slog.InfoContext(ctx, "response", "method", req.HTTPMethod, "path", req.Path, "status", resp.StatusCode, "client_ip", clientIP)
			return resp, nil
		}
	}

	var resp events.APIGatewayProxyResponse
	var routeErr error

	switch {
	case req.Path == "/v1/status":
		resp, routeErr = a.handleStatus(ctx, req)
	case req.Path == "/v1/projects":
		resp, routeErr = a.handleProjects(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/projects/"):
		slug := strings.TrimPrefix(req.Path, "/v1/projects/")
		resp, routeErr = a.handleProject(ctx, req, slug)
	case req.Path == "/v1/metrics":
		resp, routeErr = a.handleMetrics(ctx, req)
	case req.Path == "/v1/links":
		resp, routeErr = a.handleLinks(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/links/"):
		id := strings.TrimPrefix(req.Path, "/v1/links/")
		resp, routeErr = a.handleLink(ctx, req, id)
	case req.Path == "/v1/notes":
		resp, routeErr = a.handleNotes(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/notes/"):
		id := strings.TrimPrefix(req.Path, "/v1/notes/")
		resp, routeErr = a.handleNote(ctx, req, id)
	case req.Path == "/v1/til":
		resp, routeErr = a.handleTILs(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/til/"):
		id := strings.TrimPrefix(req.Path, "/v1/til/")
		resp, routeErr = a.handleTIL(ctx, req, id)
	case req.Path == "/v1/log":
		resp, routeErr = a.handleLogEntries(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/log/"):
		id := strings.TrimPrefix(req.Path, "/v1/log/")
		resp, routeErr = a.handleLogEntry(ctx, req, id)
	case req.Path == "/v1/mem/observations":
		resp, routeErr = a.handleMemObservations(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/mem/observations/"):
		id := strings.TrimPrefix(req.Path, "/v1/mem/observations/")
		resp, routeErr = a.handleMemObservation(ctx, req, id)
	case req.Path == "/v1/mem/summaries":
		resp, routeErr = a.handleMemSummaries(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/mem/summaries/"):
		id := strings.TrimPrefix(req.Path, "/v1/mem/summaries/")
		resp, routeErr = a.handleMemSummary(ctx, req, id)
	case req.Path == "/v1/mem/prompts":
		resp, routeErr = a.handleMemPrompts(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/mem/prompts/"):
		id := strings.TrimPrefix(req.Path, "/v1/mem/prompts/")
		resp, routeErr = a.handleMemPrompt(ctx, req, id)
	case req.Path == "/v1/mem/stats":
		resp, routeErr = a.handleMemStats(ctx, req)
	case req.Path == "/v1/diary":
		resp, routeErr = a.handleDiaryEntries(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/diary/"):
		id := strings.TrimPrefix(req.Path, "/v1/diary/")
		resp, routeErr = a.handleDiaryEntry(ctx, req, id)
	case req.Path == "/v1/memory":
		resp, routeErr = a.handleMemories(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/memory/"):
		id := strings.TrimPrefix(req.Path, "/v1/memory/")
		resp, routeErr = a.handleMemory(ctx, req, id)
	case req.Path == "/v1/webhooks":
		resp, routeErr = a.handleWebhooks(ctx, req)
	case strings.HasPrefix(req.Path, "/v1/webhooks/"):
		id := strings.TrimPrefix(req.Path, "/v1/webhooks/")
		resp, routeErr = a.handleWebhookEvent(ctx, req, id)
	default:
		resp = jsonResponse(404, `{"error":"not found"}`)
	}

	// Store idempotency record for successful POST responses
	if req.HTTPMethod == "POST" && idempotencyKey != "" && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fullKey := domain.IdempotencyKey(req.Path, idempotencyKey)
		record := domain.IdempotencyRecord{
			ID:         fullKey,
			StatusCode: resp.StatusCode,
			Body:       resp.Body,
			ExpiresAt:  time.Now().Add(24 * time.Hour).Unix(),
			CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		}
		if err := a.service.SetIdempotencyRecord(ctx, record); err != nil {
			slog.WarnContext(ctx, "failed to store idempotency record", "key", fullKey, "error", err)
		}
	}

	// AIDEV-NOTE: Log errors at ERROR level so 5xx responses are easy to find in CloudWatch.
	if routeErr != nil {
		slog.ErrorContext(ctx, "response", "method", req.HTTPMethod, "path", req.Path, "status", resp.StatusCode, "client_ip", clientIP, "error", routeErr)
	} else {
		slog.InfoContext(ctx, "response", "method", req.HTTPMethod, "path", req.Path, "status", resp.StatusCode, "client_ip", clientIP)
	}
	return resp, nil
}

// handleMetrics handles GET /v1/metrics.
func (a *Adapter) handleMetrics(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}

	metrics, err := a.metricsService.GetMetrics(ctx)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	body, err := json.Marshal(metrics)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleStatus routes GET and PUT for /v1/status.
func (a *Adapter) handleStatus(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		status, err := a.service.GetStatus(ctx)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(status)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "PUT":
		var fields map[string]any
		if err := json.Unmarshal([]byte(req.Body), &fields); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.UpdateStatus(ctx, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleProjects routes GET (list) and POST (create) for /v1/projects.
func (a *Adapter) handleProjects(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		projects, err := a.service.GetProjects(ctx)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(projects)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "POST":
		var project domain.Project
		if err := json.Unmarshal([]byte(req.Body), &project); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.CreateProject(ctx, project); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleProject routes GET, PUT, DELETE for /v1/projects/{slug}.
func (a *Adapter) handleProject(ctx context.Context, req events.APIGatewayProxyRequest, slug string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		project, err := a.service.GetProject(ctx, slug)
		if err != nil {
			return errorResponse(err)
		}
		body, err := json.Marshal(project)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "PUT":
		var fields map[string]any
		if err := json.Unmarshal([]byte(req.Body), &fields); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.UpdateProject(ctx, slug, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteProject(ctx, slug); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleLinks routes GET (list) and POST (create) for /v1/links.
func (a *Adapter) handleLinks(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		links, err := a.service.GetLinks(ctx, tag)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(links)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "POST":
		var link domain.Link
		if err := json.Unmarshal([]byte(req.Body), &link); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.CreateLink(ctx, link); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleLink routes GET, PUT, DELETE for /v1/links/{id}.
func (a *Adapter) handleLink(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		link, err := a.service.GetLink(ctx, id)
		if err != nil {
			return errorResponse(err)
		}
		body, err := json.Marshal(link)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "PUT":
		var fields map[string]any
		if err := json.Unmarshal([]byte(req.Body), &fields); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.UpdateLink(ctx, id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteLink(ctx, id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleNotes routes GET (list) and POST (create) for /v1/notes.
func (a *Adapter) handleNotes(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		notes, err := a.service.GetNotes(ctx, tag)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(notes)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "POST":
		var note domain.Note
		if err := json.Unmarshal([]byte(req.Body), &note); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.CreateNote(ctx, note); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleNote routes GET, PUT, DELETE for /v1/notes/{id}.
func (a *Adapter) handleNote(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		note, err := a.service.GetNote(ctx, id)
		if err != nil {
			return errorResponse(err)
		}
		body, err := json.Marshal(note)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "PUT":
		var fields map[string]any
		if err := json.Unmarshal([]byte(req.Body), &fields); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.UpdateNote(ctx, id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteNote(ctx, id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleTILs routes GET (list) and POST (create) for /v1/til.
func (a *Adapter) handleTILs(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		tils, err := a.service.GetTILs(ctx, tag)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(tils)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "POST":
		var til domain.TIL
		if err := json.Unmarshal([]byte(req.Body), &til); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.CreateTIL(ctx, til); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleTIL routes GET, PUT, DELETE for /v1/til/{id}.
func (a *Adapter) handleTIL(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		til, err := a.service.GetTIL(ctx, id)
		if err != nil {
			return errorResponse(err)
		}
		body, err := json.Marshal(til)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "PUT":
		var fields map[string]any
		if err := json.Unmarshal([]byte(req.Body), &fields); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.UpdateTIL(ctx, id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteTIL(ctx, id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleLogEntries routes GET (list) and POST (create) for /v1/log.
func (a *Adapter) handleLogEntries(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		entries, err := a.service.GetLogEntries(ctx, tag)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(entries)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "POST":
		var entry domain.LogEntry
		if err := json.Unmarshal([]byte(req.Body), &entry); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.CreateLogEntry(ctx, entry); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleLogEntry routes GET, PUT, DELETE for /v1/log/{id}.
func (a *Adapter) handleLogEntry(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		entry, err := a.service.GetLogEntry(ctx, id)
		if err != nil {
			return errorResponse(err)
		}
		body, err := json.Marshal(entry)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "PUT":
		var fields map[string]any
		if err := json.Unmarshal([]byte(req.Body), &fields); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.UpdateLogEntry(ctx, id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteLogEntry(ctx, id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleMemObservations handles GET /v1/mem/observations.
func (a *Adapter) handleMemObservations(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	obsType := req.QueryStringParameters["type"]
	project := req.QueryStringParameters["project"]
	observations, err := a.memService.GetObservations(ctx, obsType, project)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	body, err := json.Marshal(observations)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemObservation handles GET /v1/mem/observations/{id}.
func (a *Adapter) handleMemObservation(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	obs, err := a.memService.GetObservation(ctx, id)
	if err != nil {
		return errorResponse(err)
	}
	body, err := json.Marshal(obs)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemSummaries handles GET /v1/mem/summaries.
func (a *Adapter) handleMemSummaries(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	project := req.QueryStringParameters["project"]
	summaries, err := a.memService.GetSummaries(ctx, project)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	body, err := json.Marshal(summaries)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemSummary handles GET /v1/mem/summaries/{id}.
func (a *Adapter) handleMemSummary(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	summary, err := a.memService.GetSummary(ctx, id)
	if err != nil {
		return errorResponse(err)
	}
	body, err := json.Marshal(summary)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemPrompts handles GET /v1/mem/prompts.
func (a *Adapter) handleMemPrompts(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	prompts, err := a.memService.GetPrompts(ctx)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	body, err := json.Marshal(prompts)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemPrompt handles GET /v1/mem/prompts/{id}.
func (a *Adapter) handleMemPrompt(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	prompt, err := a.memService.GetPrompt(ctx, id)
	if err != nil {
		return errorResponse(err)
	}
	body, err := json.Marshal(prompt)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemStats handles GET /v1/mem/stats.
func (a *Adapter) handleMemStats(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	stats, err := a.memService.GetStats(ctx)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	body, err := json.Marshal(stats)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleDiaryEntries routes GET (list) and POST (create) for /v1/diary.
func (a *Adapter) handleDiaryEntries(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		entries, err := a.service.GetDiaryEntries(ctx, tag)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(entries)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "POST":
		var entry domain.DiaryEntry
		if err := json.Unmarshal([]byte(req.Body), &entry); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if a.diaryService != nil {
			result, err := a.diaryService.CreateAndPublish(ctx, entry)
			if err != nil {
				return jsonResponse(500, `{"error":"internal server error"}`), err
			}
			body, err := json.Marshal(result)
			if err != nil {
				return jsonResponse(500, `{"error":"internal server error"}`), err
			}
			return jsonResponse(201, string(body)), nil
		}
		if err := a.service.CreateDiaryEntry(ctx, entry); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleDiaryEntry routes GET, PUT, DELETE for /v1/diary/{id}.
func (a *Adapter) handleDiaryEntry(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		entry, err := a.service.GetDiaryEntry(ctx, id)
		if err != nil {
			return errorResponse(err)
		}
		body, err := json.Marshal(entry)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "PUT":
		var fields map[string]any
		if err := json.Unmarshal([]byte(req.Body), &fields); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.service.UpdateDiaryEntry(ctx, id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteDiaryEntry(ctx, id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleMemories routes GET (list) and POST (create) for /v1/memory.
func (a *Adapter) handleMemories(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		category := req.QueryStringParameters["category"]
		memories, err := a.memService.GetMemories(ctx, category)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(memories)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "POST":
		var memory domain.Memory
		if err := json.Unmarshal([]byte(req.Body), &memory); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.memService.CreateMemory(ctx, memory); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleMemory routes GET, PUT, DELETE for /v1/memory/{id}.
func (a *Adapter) handleMemory(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		memory, err := a.memService.GetMemory(ctx, id)
		if err != nil {
			return errorResponse(err)
		}
		body, err := json.Marshal(memory)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "PUT":
		var fields map[string]any
		if err := json.Unmarshal([]byte(req.Body), &fields); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.memService.UpdateMemory(ctx, id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.memService.DeleteMemory(ctx, id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleWebhooks routes GET (list) and POST (create) for /v1/webhooks.
// AIDEV-NOTE: POST uses HMAC auth (not API key). GET uses normal API key auth.
func (a *Adapter) handleWebhooks(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		eventType := req.QueryStringParameters["type"]
		source := req.QueryStringParameters["source"]
		events, err := a.webhookService.GetWebhookEvents(ctx, eventType, source)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		body, err := json.Marshal(events)
		if err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, string(body)), nil

	case "POST":
		// Reject if webhook secret is not configured
		if a.webhookSecret == "" {
			return jsonResponse(500, `{"error":"webhook secret not configured"}`), nil
		}

		// Validate HMAC signature
		signature := req.Headers["x-webhook-signature"]
		if !domain.ValidateWebhookSignature(req.Body, signature, a.webhookSecret) {
			return jsonResponse(401, `{"error":"invalid webhook signature"}`), nil
		}

		var event domain.WebhookEvent
		dec := json.NewDecoder(strings.NewReader(req.Body))
		dec.UseNumber()
		if err := dec.Decode(&event); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}

		// AIDEV-NOTE: Publish to async queue instead of writing to DynamoDB directly.
		if a.webhookPublisher == nil {
			slog.ErrorContext(ctx, "webhook publisher not configured")
			return jsonResponse(500, `{"error":"webhook publisher not configured"}`), nil
		}
		if err := a.webhookPublisher.Publish(ctx, event); err != nil {
			slog.ErrorContext(ctx, "failed to publish webhook event", "error", err)
			return jsonResponse(500, `{"error":"internal server error"}`), nil
		}
		return jsonResponse(202, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleWebhookEvent handles GET /v1/webhooks/{id}.
func (a *Adapter) handleWebhookEvent(ctx context.Context, req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	event, err := a.webhookService.GetWebhookEvent(ctx, id)
	if err != nil {
		return errorResponse(err)
	}
	body, err := json.Marshal(event)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

func jsonResponse(statusCode int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type, x-api-key, x-webhook-signature",
			"Access-Control-Allow-Methods": "GET, PUT, POST, DELETE, OPTIONS",
		},
	}
}
