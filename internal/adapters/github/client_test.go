// ABOUTME: This file tests the GitHub Contents API client used for publishing diary entries.
// ABOUTME: Uses httptest to verify request construction without hitting the real GitHub API.
package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublish_SendsCorrectRequest(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")

		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"content":{"sha":"abc123"}}`))
	}))
	defer server.Close()

	client := NewClient("test-token", "vaporeyes", "obsidian-diary")
	client.baseURL = server.URL

	err := client.Publish(context.Background(), "diary/2026-02-17-153045.md", []byte("# Test"), "add diary entry")
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if gotMethod != "PUT" {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if gotPath != "/repos/vaporeyes/obsidian-diary/contents/diary/2026-02-17-153045.md" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("unexpected auth header: %s", gotAuth)
	}
	if gotBody["message"] != "add diary entry" {
		t.Errorf("unexpected commit message: %s", gotBody["message"])
	}
	if gotBody["content"] == "" {
		t.Error("content should be base64-encoded, got empty string")
	}
}

func TestPublish_HandlesErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Invalid request"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", "vaporeyes", "obsidian-diary")
	client.baseURL = server.URL

	err := client.Publish(context.Background(), "diary/test.md", []byte("# Test"), "test commit")
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestPublish_HandlesNetworkError(t *testing.T) {
	client := NewClient("test-token", "vaporeyes", "obsidian-diary")
	client.baseURL = "http://localhost:0" // unreachable

	err := client.Publish(context.Background(), "diary/test.md", []byte("# Test"), "test commit")
	if err == nil {
		t.Fatal("expected error for network failure")
	}
}
