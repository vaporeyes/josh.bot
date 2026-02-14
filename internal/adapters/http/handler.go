// ABOUTME: This file implements the HTTP handlers for the API endpoints.
// ABOUTME: It adapts incoming HTTP requests to calls on the domain service.
package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jduncan/josh-bot/internal/domain"
)

type Adapter struct {
	service domain.BotService
}

func NewAdapter(service domain.BotService) *Adapter {
	return &Adapter{service: service}
}

func (a *Adapter) StatusHandler(w http.ResponseWriter, r *http.Request) {
	status, err := a.service.GetStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (a *Adapter) UpdateStatusHandler(w http.ResponseWriter, r *http.Request) {
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.UpdateStatus(fields); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func (a *Adapter) ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	projects, err := a.service.GetProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(projects)
}

func (a *Adapter) CreateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var project domain.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.CreateProject(project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"ok":true}`))
}

// ProjectHandler handles GET /v1/projects/{slug}.
func (a *Adapter) ProjectHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/v1/projects/")
	if slug == "" {
		http.Error(w, `{"error":"slug required"}`, http.StatusBadRequest)
		return
	}

	project, err := a.service.GetProject(slug)
	if err != nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(project)
}

func (a *Adapter) UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/v1/projects/")
	if slug == "" {
		http.Error(w, `{"error":"slug required"}`, http.StatusBadRequest)
		return
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.UpdateProject(slug, fields); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

func (a *Adapter) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/v1/projects/")
	if slug == "" {
		http.Error(w, `{"error":"slug required"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.DeleteProject(slug); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}
