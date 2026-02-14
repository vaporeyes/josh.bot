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

	var body []byte
	var err error

	switch req.Path {
	case "/v1/status":
		var status domain.Status
		status, err = a.service.GetStatus()
		if err != nil {
			return errorResponse(500), err
		}
		body, err = json.Marshal(status)

	case "/v1/projects":
		var projects []domain.Project
		projects, err = a.service.GetProjects()
		if err != nil {
			return errorResponse(500), err
		}
		body, err = json.Marshal(projects)

	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error":"not found"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	if err != nil {
		return errorResponse(500), err
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func errorResponse(statusCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       `{"error":"internal server error"}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}
