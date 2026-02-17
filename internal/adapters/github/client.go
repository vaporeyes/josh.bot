// ABOUTME: This file implements a GitHub Contents API client for publishing files.
// ABOUTME: It uses net/http directly (no third-party library) to push Obsidian markdown to a repo.
package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client implements domain.ObsidianPublisher using the GitHub Contents API.
type Client struct {
	token   string
	owner   string
	repo    string
	baseURL string
	http    *http.Client
}

// NewClient creates a GitHub Contents API client.
func NewClient(token, owner, repo string) *Client {
	return &Client{
		token:   token,
		owner:   owner,
		repo:    repo,
		baseURL: "https://api.github.com",
		http:    &http.Client{},
	}
}

// contentsRequest is the JSON body for the GitHub Contents API PUT endpoint.
type contentsRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
}

// Publish creates or updates a file in the GitHub repo via the Contents API.
// AIDEV-NOTE: Uses PUT /repos/{owner}/{repo}/contents/{path} with base64-encoded content.
func (c *Client) Publish(ctx context.Context, path string, content []byte, commitMsg string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.baseURL, c.owner, c.repo, path)

	payload := contentsRequest{
		Message: commitMsg,
		Content: base64.StdEncoding.EncodeToString(content),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal github request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create github request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("github api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("github api returned status %d for %s", resp.StatusCode, path)
	}

	return nil
}
