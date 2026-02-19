// ABOUTME: This file implements the AWS Lambda adapter for the josh.bot API.
// ABOUTME: It handles API Gateway events, validates API keys, and routes requests to domain services.
package lambda

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jduncan/josh-bot/internal/domain"
)

// Adapter wraps domain services and handles Lambda API Gateway events.
type Adapter struct {
	service        domain.BotService
	metricsService domain.MetricsService
	memService     domain.MemService
	diaryService   domain.DiaryService
	webhookService domain.WebhookService
	webhookSecret  string
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
func (a *Adapter) Router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	switch {
	case req.Path == "/v1/status":
		return a.handleStatus(req)
	case req.Path == "/v1/projects":
		return a.handleProjects(req)
	case strings.HasPrefix(req.Path, "/v1/projects/"):
		slug := strings.TrimPrefix(req.Path, "/v1/projects/")
		return a.handleProject(req, slug)
	case req.Path == "/v1/metrics":
		return a.handleMetrics(req)
	case req.Path == "/v1/links":
		return a.handleLinks(req)
	case strings.HasPrefix(req.Path, "/v1/links/"):
		id := strings.TrimPrefix(req.Path, "/v1/links/")
		return a.handleLink(req, id)
	case req.Path == "/v1/notes":
		return a.handleNotes(req)
	case strings.HasPrefix(req.Path, "/v1/notes/"):
		id := strings.TrimPrefix(req.Path, "/v1/notes/")
		return a.handleNote(req, id)
	case req.Path == "/v1/til":
		return a.handleTILs(req)
	case strings.HasPrefix(req.Path, "/v1/til/"):
		id := strings.TrimPrefix(req.Path, "/v1/til/")
		return a.handleTIL(req, id)
	case req.Path == "/v1/log":
		return a.handleLogEntries(req)
	case strings.HasPrefix(req.Path, "/v1/log/"):
		id := strings.TrimPrefix(req.Path, "/v1/log/")
		return a.handleLogEntry(req, id)
	case req.Path == "/v1/mem/observations":
		return a.handleMemObservations(req)
	case strings.HasPrefix(req.Path, "/v1/mem/observations/"):
		id := strings.TrimPrefix(req.Path, "/v1/mem/observations/")
		return a.handleMemObservation(req, id)
	case req.Path == "/v1/mem/summaries":
		return a.handleMemSummaries(req)
	case strings.HasPrefix(req.Path, "/v1/mem/summaries/"):
		id := strings.TrimPrefix(req.Path, "/v1/mem/summaries/")
		return a.handleMemSummary(req, id)
	case req.Path == "/v1/mem/prompts":
		return a.handleMemPrompts(req)
	case strings.HasPrefix(req.Path, "/v1/mem/prompts/"):
		id := strings.TrimPrefix(req.Path, "/v1/mem/prompts/")
		return a.handleMemPrompt(req, id)
	case req.Path == "/v1/mem/stats":
		return a.handleMemStats(req)
	case req.Path == "/v1/diary":
		return a.handleDiaryEntries(req)
	case strings.HasPrefix(req.Path, "/v1/diary/"):
		id := strings.TrimPrefix(req.Path, "/v1/diary/")
		return a.handleDiaryEntry(req, id)
	case req.Path == "/v1/memory":
		return a.handleMemories(req)
	case strings.HasPrefix(req.Path, "/v1/memory/"):
		id := strings.TrimPrefix(req.Path, "/v1/memory/")
		return a.handleMemory(req, id)
	case req.Path == "/v1/webhooks":
		return a.handleWebhooks(req)
	case strings.HasPrefix(req.Path, "/v1/webhooks/"):
		id := strings.TrimPrefix(req.Path, "/v1/webhooks/")
		return a.handleWebhookEvent(req, id)
	default:
		return jsonResponse(404, `{"error":"not found"}`), nil
	}
}

// handleMetrics handles GET /v1/metrics.
func (a *Adapter) handleMetrics(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}

	metrics, err := a.metricsService.GetMetrics()
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
func (a *Adapter) handleStatus(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		status, err := a.service.GetStatus()
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
		if err := a.service.UpdateStatus(fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleProjects routes GET (list) and POST (create) for /v1/projects.
func (a *Adapter) handleProjects(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		projects, err := a.service.GetProjects()
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
		if err := a.service.CreateProject(project); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleProject routes GET, PUT, DELETE for /v1/projects/{slug}.
func (a *Adapter) handleProject(req events.APIGatewayProxyRequest, slug string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		project, err := a.service.GetProject(slug)
		if err != nil {
			return jsonResponse(404, `{"error":"project not found"}`), nil
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
		if err := a.service.UpdateProject(slug, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteProject(slug); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleLinks routes GET (list) and POST (create) for /v1/links.
func (a *Adapter) handleLinks(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		links, err := a.service.GetLinks(tag)
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
		if err := a.service.CreateLink(link); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleLink routes GET, PUT, DELETE for /v1/links/{id}.
func (a *Adapter) handleLink(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		link, err := a.service.GetLink(id)
		if err != nil {
			return jsonResponse(404, `{"error":"link not found"}`), nil
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
		if err := a.service.UpdateLink(id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteLink(id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleNotes routes GET (list) and POST (create) for /v1/notes.
func (a *Adapter) handleNotes(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		notes, err := a.service.GetNotes(tag)
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
		if err := a.service.CreateNote(note); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleNote routes GET, PUT, DELETE for /v1/notes/{id}.
func (a *Adapter) handleNote(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		note, err := a.service.GetNote(id)
		if err != nil {
			return jsonResponse(404, `{"error":"note not found"}`), nil
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
		if err := a.service.UpdateNote(id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteNote(id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleTILs routes GET (list) and POST (create) for /v1/til.
func (a *Adapter) handleTILs(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		tils, err := a.service.GetTILs(tag)
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
		if err := a.service.CreateTIL(til); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleTIL routes GET, PUT, DELETE for /v1/til/{id}.
func (a *Adapter) handleTIL(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		til, err := a.service.GetTIL(id)
		if err != nil {
			return jsonResponse(404, `{"error":"til not found"}`), nil
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
		if err := a.service.UpdateTIL(id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteTIL(id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleLogEntries routes GET (list) and POST (create) for /v1/log.
func (a *Adapter) handleLogEntries(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		entries, err := a.service.GetLogEntries(tag)
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
		if err := a.service.CreateLogEntry(entry); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleLogEntry routes GET, PUT, DELETE for /v1/log/{id}.
func (a *Adapter) handleLogEntry(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		entry, err := a.service.GetLogEntry(id)
		if err != nil {
			return jsonResponse(404, `{"error":"log entry not found"}`), nil
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
		if err := a.service.UpdateLogEntry(id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteLogEntry(id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleMemObservations handles GET /v1/mem/observations.
func (a *Adapter) handleMemObservations(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	obsType := req.QueryStringParameters["type"]
	project := req.QueryStringParameters["project"]
	observations, err := a.memService.GetObservations(obsType, project)
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
func (a *Adapter) handleMemObservation(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	obs, err := a.memService.GetObservation(id)
	if err != nil {
		return jsonResponse(404, `{"error":"observation not found"}`), nil
	}
	body, err := json.Marshal(obs)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemSummaries handles GET /v1/mem/summaries.
func (a *Adapter) handleMemSummaries(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	project := req.QueryStringParameters["project"]
	summaries, err := a.memService.GetSummaries(project)
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
func (a *Adapter) handleMemSummary(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	summary, err := a.memService.GetSummary(id)
	if err != nil {
		return jsonResponse(404, `{"error":"summary not found"}`), nil
	}
	body, err := json.Marshal(summary)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemPrompts handles GET /v1/mem/prompts.
func (a *Adapter) handleMemPrompts(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	prompts, err := a.memService.GetPrompts()
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
func (a *Adapter) handleMemPrompt(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	prompt, err := a.memService.GetPrompt(id)
	if err != nil {
		return jsonResponse(404, `{"error":"prompt not found"}`), nil
	}
	body, err := json.Marshal(prompt)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

// handleMemStats handles GET /v1/mem/stats.
func (a *Adapter) handleMemStats(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	stats, err := a.memService.GetStats()
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
func (a *Adapter) handleDiaryEntries(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		tag := req.QueryStringParameters["tag"]
		entries, err := a.service.GetDiaryEntries(tag)
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
			result, err := a.diaryService.CreateAndPublish(context.Background(), entry)
			if err != nil {
				return jsonResponse(500, `{"error":"internal server error"}`), err
			}
			body, err := json.Marshal(result)
			if err != nil {
				return jsonResponse(500, `{"error":"internal server error"}`), err
			}
			return jsonResponse(201, string(body)), nil
		}
		if err := a.service.CreateDiaryEntry(entry); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleDiaryEntry routes GET, PUT, DELETE for /v1/diary/{id}.
func (a *Adapter) handleDiaryEntry(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		entry, err := a.service.GetDiaryEntry(id)
		if err != nil {
			return jsonResponse(404, `{"error":"diary entry not found"}`), nil
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
		if err := a.service.UpdateDiaryEntry(id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.service.DeleteDiaryEntry(id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleMemories routes GET (list) and POST (create) for /v1/memory.
func (a *Adapter) handleMemories(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		category := req.QueryStringParameters["category"]
		memories, err := a.memService.GetMemories(category)
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
		if err := a.memService.CreateMemory(memory); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleMemory routes GET, PUT, DELETE for /v1/memory/{id}.
func (a *Adapter) handleMemory(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		memory, err := a.memService.GetMemory(id)
		if err != nil {
			return jsonResponse(404, `{"error":"memory not found"}`), nil
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
		if err := a.memService.UpdateMemory(id, fields); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	case "DELETE":
		if err := a.memService.DeleteMemory(id); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(200, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleWebhooks routes GET (list) and POST (create) for /v1/webhooks.
// AIDEV-NOTE: POST uses HMAC auth (not API key). GET uses normal API key auth.
func (a *Adapter) handleWebhooks(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
	case "GET":
		eventType := req.QueryStringParameters["type"]
		source := req.QueryStringParameters["source"]
		events, err := a.webhookService.GetWebhookEvents(eventType, source)
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
		if err := json.Unmarshal([]byte(req.Body), &event); err != nil {
			return jsonResponse(400, `{"error":"invalid JSON body"}`), nil
		}
		if err := a.webhookService.CreateWebhookEvent(event); err != nil {
			return jsonResponse(500, `{"error":"internal server error"}`), err
		}
		return jsonResponse(201, `{"ok":true}`), nil

	default:
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
}

// handleWebhookEvent handles GET /v1/webhooks/{id}.
func (a *Adapter) handleWebhookEvent(req events.APIGatewayProxyRequest, id string) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}
	event, err := a.webhookService.GetWebhookEvent(id)
	if err != nil {
		return jsonResponse(404, `{"error":"webhook event not found"}`), nil
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
