// ABOUTME: HTTP client that proxies requests to the josh.bot API (api.josh.bot).
// ABOUTME: Each method maps 1:1 to an API endpoint, returning domain types directly.
package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jduncan/josh-bot/internal/domain"
)

// Client wraps an HTTP client configured to talk to the josh.bot API.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient creates a Client pointing at the given base URL with the given API key.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{},
	}
}

// APIError represents a non-2xx response from the API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d: %s", e.StatusCode, e.Message)
}

// do executes an HTTP request with the API key header and returns the response body.
func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := string(respBody)
		// Try to extract error message from JSON response
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			msg = errResp.Error
		}
		return nil, resp.StatusCode, &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	return respBody, resp.StatusCode, nil
}

// --- Projects ---

// GetProjects returns all projects.
func (c *Client) GetProjects(ctx context.Context) ([]domain.Project, error) {
	body, _, err := c.do(ctx, http.MethodGet, "/v1/projects", nil)
	if err != nil {
		return nil, err
	}
	var projects []domain.Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("decode projects: %w", err)
	}
	return projects, nil
}

// GetProject returns a single project by slug.
func (c *Client) GetProject(ctx context.Context, slug string) (domain.Project, error) {
	body, _, err := c.do(ctx, http.MethodGet, "/v1/projects/"+slug, nil)
	if err != nil {
		return domain.Project{}, err
	}
	var project domain.Project
	if err := json.Unmarshal(body, &project); err != nil {
		return domain.Project{}, fmt.Errorf("decode project: %w", err)
	}
	return project, nil
}

// CreateProject creates a project via POST.
func (c *Client) CreateProject(ctx context.Context, project domain.Project) error {
	_, _, err := c.do(ctx, http.MethodPost, "/v1/projects", project)
	return err
}

// UpdateProject updates a project via PUT with partial fields.
func (c *Client) UpdateProject(ctx context.Context, slug string, fields map[string]any) error {
	_, _, err := c.do(ctx, http.MethodPut, "/v1/projects/"+slug, fields)
	return err
}

// DeleteProject deletes a project by slug.
func (c *Client) DeleteProject(ctx context.Context, slug string) error {
	_, _, err := c.do(ctx, http.MethodDelete, "/v1/projects/"+slug, nil)
	return err
}

// --- Status ---

// GetStatus returns the current status.
func (c *Client) GetStatus(ctx context.Context) (domain.Status, error) {
	body, _, err := c.do(ctx, http.MethodGet, "/v1/status", nil)
	if err != nil {
		return domain.Status{}, err
	}
	var status domain.Status
	if err := json.Unmarshal(body, &status); err != nil {
		return domain.Status{}, fmt.Errorf("decode status: %w", err)
	}
	return status, nil
}
