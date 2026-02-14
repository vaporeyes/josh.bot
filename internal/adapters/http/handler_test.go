// ABOUTME: This file contains tests for the HTTP handlers.
// ABOUTME: It follows a TDD approach for building the API.
package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jduncan/josh-bot/internal/adapters/mock"
	"github.com/jduncan/josh-bot/internal/domain"
)

func TestStatusHandler(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	req, err := http.NewRequest("GET", "/v1/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.StatusHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var status domain.Status
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Core fields
	if status.Name == "" {
		t.Error("expected name to be set")
	}
	if status.Title == "" {
		t.Error("expected title to be set")
	}
	if status.Bio == "" {
		t.Error("expected bio to be set")
	}
	if status.CurrentActivity == "" {
		t.Error("expected current_activity to be set")
	}
	if status.Location == "" {
		t.Error("expected location to be set")
	}
	if status.Availability == "" {
		t.Error("expected availability to be set")
	}
	if status.Status == "" {
		t.Error("expected status to be set")
	}

	// Structured fields
	if len(status.Links) == 0 {
		t.Error("expected links to be non-empty")
	}
	if _, ok := status.Links["github"]; !ok {
		t.Error("expected links to contain github")
	}
	if len(status.Interests) == 0 {
		t.Error("expected interests to be non-empty")
	}
}

func TestProjectsHandler(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	req, err := http.NewRequest("GET", "/v1/projects", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.ProjectsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var projects []domain.Project
	if err := json.Unmarshal(rr.Body.Bytes(), &projects); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(projects) < 2 {
		t.Errorf("expected at least 2 projects, got %d", len(projects))
	}
}
