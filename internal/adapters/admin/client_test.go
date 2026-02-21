// ABOUTME: Tests for the admin API client using httptest to simulate API responses.
// ABOUTME: Verifies correct HTTP method, path, headers, and response deserialization.
package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jduncan/josh-bot/internal/domain"
)

func TestGetProjects(t *testing.T) {
	projects := []domain.Project{
		{Slug: "josh-bot", Name: "josh.bot", Stack: "Go", Status: "active"},
		{Slug: "other", Name: "Other", Stack: "Rust", Status: "archived"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/projects" {
			t.Errorf("expected /v1/projects, got %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected api key header, got %q", r.Header.Get("x-api-key"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(projects)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	got, err := client.GetProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(got))
	}
	if got[0].Slug != "josh-bot" {
		t.Errorf("expected slug josh-bot, got %s", got[0].Slug)
	}
}

func TestGetProject(t *testing.T) {
	project := domain.Project{Slug: "josh-bot", Name: "josh.bot", Stack: "Go"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/josh-bot" {
			t.Errorf("expected /v1/projects/josh-bot, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(project)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	got, err := client.GetProject(context.Background(), "josh-bot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Slug != "josh-bot" {
		t.Errorf("expected slug josh-bot, got %s", got.Slug)
	}
}

func TestCreateProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected json content type, got %q", r.Header.Get("Content-Type"))
		}
		var p domain.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if p.Slug != "test-proj" {
			t.Errorf("expected slug test-proj, got %s", p.Slug)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	err := client.CreateProject(context.Background(), domain.Project{Slug: "test-proj", Name: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v1/projects/test-proj" {
			t.Errorf("expected /v1/projects/test-proj, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	err := client.UpdateProject(context.Background(), "test-proj", map[string]any{"name": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	err := client.DeleteProject(context.Background(), "test-proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"project not found"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	_, err := client.GetProject(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "project not found" {
		t.Errorf("expected 'project not found', got %q", apiErr.Message)
	}
}

func TestGetStatus(t *testing.T) {
	status := domain.Status{Name: "Josh", Status: "online", Location: "Austin"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/status" {
			t.Errorf("expected /v1/status, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(status)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-key")
	got, err := client.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Josh" {
		t.Errorf("expected name Josh, got %s", got.Name)
	}
}
