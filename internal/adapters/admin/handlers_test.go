// ABOUTME: Tests for admin HTTP handlers using a fake API backend.
// ABOUTME: Verifies correct routing, rendering, and error handling.
package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jduncan/josh-bot/internal/domain"
)

// fakeAPI creates a test HTTP server that mimics the josh.bot API.
func fakeAPI(t *testing.T) *httptest.Server {
	t.Helper()
	projects := []domain.Project{
		{Slug: "josh-bot", Name: "josh.bot", Stack: "Go", Status: "active"},
	}
	status := domain.Status{Name: "Josh", Status: "online"}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/v1/status" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(status)

		case r.URL.Path == "/v1/projects" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(projects)

		case r.URL.Path == "/v1/projects" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ok":true}`))

		case strings.HasPrefix(r.URL.Path, "/v1/projects/") && r.Method == http.MethodGet:
			slug := strings.TrimPrefix(r.URL.Path, "/v1/projects/")
			for _, p := range projects {
				if p.Slug == slug {
					_ = json.NewEncoder(w).Encode(p)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"project not found"}`))

		case strings.HasPrefix(r.URL.Path, "/v1/projects/") && r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"ok":true}`))

		case strings.HasPrefix(r.URL.Path, "/v1/projects/") && r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"ok":true}`))

		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		}
	}))
}

func TestDashboard(t *testing.T) {
	api := fakeAPI(t)
	defer api.Close()

	client := NewClient(api.URL, "test-key")
	handlers := NewHandlers(client)

	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	rec := httptest.NewRecorder()
	handlers.Dashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Dashboard") {
		t.Error("expected Dashboard in response body")
	}
	if !strings.Contains(body, "Josh") {
		t.Error("expected status name 'Josh' in response body")
	}
}

func TestProjectsList(t *testing.T) {
	api := fakeAPI(t)
	defer api.Close()

	client := NewClient(api.URL, "test-key")
	handlers := NewHandlers(client)

	req := httptest.NewRequest(http.MethodGet, "/admin/projects", nil)
	rec := httptest.NewRecorder()
	handlers.Projects(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "josh-bot") {
		t.Error("expected project slug 'josh-bot' in response body")
	}
	if !strings.Contains(body, "josh.bot") {
		t.Error("expected project name 'josh.bot' in response body")
	}
}

func TestProjectNew(t *testing.T) {
	api := fakeAPI(t)
	defer api.Close()

	client := NewClient(api.URL, "test-key")
	handlers := NewHandlers(client)

	req := httptest.NewRequest(http.MethodGet, "/admin/projects/new", nil)
	rec := httptest.NewRecorder()
	handlers.ProjectNew(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Add Project") {
		t.Error("expected 'Add Project' in response body")
	}
}

func TestProjectEdit(t *testing.T) {
	api := fakeAPI(t)
	defer api.Close()

	client := NewClient(api.URL, "test-key")
	handlers := NewHandlers(client)

	req := httptest.NewRequest(http.MethodGet, "/admin/projects/josh-bot/edit", nil)
	rec := httptest.NewRecorder()
	handlers.ProjectBySlug(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Edit Project") {
		t.Error("expected 'Edit Project' in response body")
	}
	if !strings.Contains(body, "josh.bot") {
		t.Error("expected project name in form")
	}
}

func TestProjectDelete(t *testing.T) {
	api := fakeAPI(t)
	defer api.Close()

	client := NewClient(api.URL, "test-key")
	handlers := NewHandlers(client)

	req := httptest.NewRequest(http.MethodDelete, "/admin/projects/josh-bot", nil)
	rec := httptest.NewRecorder()
	handlers.ProjectBySlug(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestProjectCreate(t *testing.T) {
	api := fakeAPI(t)
	defer api.Close()

	client := NewClient(api.URL, "test-key")
	handlers := NewHandlers(client)

	body := `{"slug":"new-proj","name":"New Project","stack":"Go","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handlers.Projects(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/admin/projects" {
		t.Errorf("expected redirect to /admin/projects, got %q", loc)
	}
}

func TestProjectUpdate(t *testing.T) {
	api := fakeAPI(t)
	defer api.Close()

	client := NewClient(api.URL, "test-key")
	handlers := NewHandlers(client)

	body := `{"name":"Updated Name"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/projects/josh-bot", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handlers.ProjectBySlug(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect, got %d", rec.Code)
	}
}

func TestDashboardNonRoot(t *testing.T) {
	api := fakeAPI(t)
	defer api.Close()

	client := NewClient(api.URL, "test-key")
	handlers := NewHandlers(client)

	req := httptest.NewRequest(http.MethodGet, "/admin/unknown", nil)
	rec := httptest.NewRecorder()
	handlers.Dashboard(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
