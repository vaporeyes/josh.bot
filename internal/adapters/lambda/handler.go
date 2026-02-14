// ABOUTME: This file implements the AWS Lambda adapter for the josh.bot API.
// ABOUTME: It handles API Gateway events, validates API keys, and routes requests to domain services.
package lambda

import (
	"encoding/json"
	"os"

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

// Router handles API Gateway proxy requests with API key validation.
func (a *Adapter) Router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Validate API key from x-api-key header
	expectedKey := os.Getenv("API_KEY")
	if expectedKey != "" && req.Headers["x-api-key"] != expectedKey {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       `{"error":"unauthorized"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	switch req.Path {
	case "/v1/status":
		return a.handleStatus(req)
	case "/v1/projects":
		return a.handleProjects(req)
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

// handleProjects routes GET for /v1/projects.
func (a *Adapter) handleProjects(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != "GET" {
		return jsonResponse(405, `{"error":"method not allowed"}`), nil
	}

	projects, err := a.service.GetProjects()
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	body, err := json.Marshal(projects)
	if err != nil {
		return jsonResponse(500, `{"error":"internal server error"}`), err
	}
	return jsonResponse(200, string(body)), nil
}

func jsonResponse(statusCode int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}
