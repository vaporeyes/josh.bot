// ABOUTME: This file implements the AWS Lambda adapter for the josh.bot API.
// ABOUTME: It handles API Gateway events, validates API keys, and routes requests to domain services.
package lambda

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jduncan/josh-bot/internal/domain"
)

// Adapter wraps a domain.BotService and handles Lambda API Gateway events.
type Adapter struct {
	service domain.BotService
}

// NewAdapter creates a new Lambda adapter for the given service.
func NewAdapter(service domain.BotService) *Adapter {
	return &Adapter{service: service}
}

// isPublicRoute returns true for routes that don't require API key auth.
func isPublicRoute(method, path string) bool {
	return method == "GET" && path == "/v1/status"
}

// Router handles API Gateway proxy requests with API key validation.
func (a *Adapter) Router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Handle CORS preflight
	if req.HTTPMethod == "OPTIONS" {
		return jsonResponse(204, ""), nil
	}

	// Validate API key from x-api-key header (skip for public routes)
	if !isPublicRoute(req.HTTPMethod, req.Path) {
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
	default:
		return jsonResponse(404, `{"error":"not found"}`), nil
	}
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

func jsonResponse(statusCode int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Headers": "Content-Type, x-api-key",
			"Access-Control-Allow-Methods": "GET, PUT, POST, DELETE, OPTIONS",
		},
	}
}
