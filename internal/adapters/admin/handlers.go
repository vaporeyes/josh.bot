// ABOUTME: HTTP handlers for the admin dashboard, rendering templ views.
// ABOUTME: Routes under /admin/ prefix, proxying CRUD operations to the API client.
package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/jduncan/josh-bot/internal/adapters/admin/views"
	"github.com/jduncan/josh-bot/internal/domain"
)

// Handlers holds the admin HTTP handlers and their API client dependency.
type Handlers struct {
	client *Client
}

// NewHandlers creates admin handlers backed by the given API client.
func NewHandlers(client *Client) *Handlers {
	return &Handlers{client: client}
}

// RegisterRoutes registers all admin routes on the given mux.
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/admin/", h.Dashboard)
	mux.HandleFunc("/admin/projects", h.Projects)
	mux.HandleFunc("/admin/projects/new", h.ProjectNew)
	mux.HandleFunc("/admin/projects/", h.ProjectBySlug)
}

// render is a helper that renders a templ component to the response.
func render(w http.ResponseWriter, r *http.Request, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.Render(r.Context(), w); err != nil {
		slog.Error("failed to render template", "error", err)
	}
}

// Dashboard handles GET /admin/ — shows status overview.
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Only handle exact /admin/ path
	if r.URL.Path != "/admin/" && r.URL.Path != "/admin" {
		http.NotFound(w, r)
		return
	}

	status, err := h.client.GetStatus(r.Context())
	if err != nil {
		slog.Error("failed to fetch status", "error", err)
		status = domain.Status{Name: "(unavailable)"}
	}
	render(w, r, views.Dashboard(status))
}

// Projects handles GET/POST /admin/projects.
func (h *Handlers) Projects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.projectsList(w, r)
	case http.MethodPost:
		h.projectsCreate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handlers) projectsList(w http.ResponseWriter, r *http.Request) {
	projects, err := h.client.GetProjects(r.Context())
	if err != nil {
		slog.Error("failed to fetch projects", "error", err)
		render(w, r, views.ProjectsPage(nil))
		return
	}
	render(w, r, views.ProjectsPage(projects))
}

func (h *Handlers) projectsCreate(w http.ResponseWriter, r *http.Request) {
	var project domain.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		render(w, r, views.ProjectForm(project, false, "Invalid form data"))
		return
	}

	if err := h.client.CreateProject(r.Context(), project); err != nil {
		render(w, r, views.ProjectForm(project, false, err.Error()))
		return
	}

	// Redirect to list after successful create
	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

// ProjectNew handles GET /admin/projects/new — shows create form.
func (h *Handlers) ProjectNew(w http.ResponseWriter, r *http.Request) {
	render(w, r, views.ProjectForm(domain.Project{Status: "active"}, false, ""))
}

// ProjectBySlug handles /admin/projects/{slug}/* routes.
func (h *Handlers) ProjectBySlug(w http.ResponseWriter, r *http.Request) {
	// Parse: /admin/projects/{slug} or /admin/projects/{slug}/edit
	path := strings.TrimPrefix(r.URL.Path, "/admin/projects/")
	parts := strings.SplitN(path, "/", 2)
	slug := parts[0]
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch {
	case action == "edit" && r.Method == http.MethodGet:
		h.projectEdit(w, r, slug)
	case action == "" && r.Method == http.MethodPut:
		h.projectUpdate(w, r, slug)
	case action == "" && r.Method == http.MethodDelete:
		h.projectDelete(w, r, slug)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handlers) projectEdit(w http.ResponseWriter, r *http.Request, slug string) {
	project, err := h.client.GetProject(r.Context(), slug)
	if err != nil {
		slog.Error("failed to fetch project", "error", err, "slug", slug)
		render(w, r, views.ProjectForm(domain.Project{Slug: slug}, true, err.Error()))
		return
	}
	render(w, r, views.ProjectForm(project, true, ""))
}

func (h *Handlers) projectUpdate(w http.ResponseWriter, r *http.Request, slug string) {
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		render(w, r, views.ErrorMessage("Invalid form data"))
		return
	}

	if err := h.client.UpdateProject(r.Context(), slug, fields); err != nil {
		render(w, r, views.ErrorMessage(err.Error()))
		return
	}

	// Redirect to list after successful update
	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

func (h *Handlers) projectDelete(w http.ResponseWriter, r *http.Request, slug string) {
	if err := h.client.DeleteProject(r.Context(), slug); err != nil {
		slog.Error("failed to delete project", "error", err, "slug", slug)
		render(w, r, views.ErrorMessage(err.Error()))
		return
	}
	// Return empty response — htmx will remove the row
	w.WriteHeader(http.StatusOK)
}
