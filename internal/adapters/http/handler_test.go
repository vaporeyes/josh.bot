// ABOUTME: This file contains tests for the HTTP handlers.
// ABOUTME: It follows a TDD approach for building the API.
package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestUpdateStatusHandler(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	body := strings.NewReader(`{"current_activity": "Testing", "availability": "Busy"}`)
	req, err := http.NewRequest("PUT", "/v1/status", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.UpdateStatusHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("expected ok:true, got %v", result)
	}
}

func TestUpdateStatusHandler_InvalidJSON(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	body := strings.NewReader(`{not valid`)
	req, err := http.NewRequest("PUT", "/v1/status", body)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.UpdateStatusHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
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

func TestCreateProjectHandler(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	body := strings.NewReader(`{"slug":"new-proj","name":"New Project","stack":"Go","description":"A thing","url":"https://github.com/vaporeyes/new","status":"active"}`)
	req, err := http.NewRequest("POST", "/v1/projects", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.CreateProjectHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCreateProjectHandler_InvalidJSON(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	body := strings.NewReader(`{bad json`)
	req, err := http.NewRequest("POST", "/v1/projects", body)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.CreateProjectHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestProjectHandler(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	req, err := http.NewRequest("GET", "/v1/projects/modular-aws-backend", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.ProjectHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var project domain.Project
	if err := json.Unmarshal(rr.Body.Bytes(), &project); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if project.Name != "Modular AWS Backend" {
		t.Errorf("expected name 'Modular AWS Backend', got '%s'", project.Name)
	}
}

func TestProjectHandler_NotFound(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	req, err := http.NewRequest("GET", "/v1/projects/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.ProjectHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateProjectHandler(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	body := strings.NewReader(`{"status":"archived"}`)
	req, err := http.NewRequest("PUT", "/v1/projects/modular-aws-backend", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.UpdateProjectHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteProjectHandler(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	req, err := http.NewRequest("DELETE", "/v1/projects/modular-aws-backend", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.DeleteProjectHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}
