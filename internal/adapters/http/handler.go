// ABOUTME: This file implements the HTTP handlers for the API endpoints.
// ABOUTME: It adapts incoming HTTP requests to calls on the domain service.
package http

import (
	"encoding/json"
	"log"
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

	writeJSON(w, http.StatusOK, status)
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

	writeOK(w, http.StatusOK)
}

func (a *Adapter) ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	projects, err := a.service.GetProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, projects)
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

	writeOK(w, http.StatusCreated)
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

	writeJSON(w, http.StatusOK, project)
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

	writeOK(w, http.StatusOK)
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

	writeOK(w, http.StatusOK)
}

func (a *Adapter) LinksHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	links, err := a.service.GetLinks(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, links)
}

func (a *Adapter) CreateLinkHandler(w http.ResponseWriter, r *http.Request) {
	var link domain.Link
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.CreateLink(link); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusCreated)
}

// LinkHandler handles GET /v1/links/{id}.
func (a *Adapter) LinkHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/links/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	link, err := a.service.GetLink(id)
	if err != nil {
		http.Error(w, `{"error":"link not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, link)
}

func (a *Adapter) UpdateLinkHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/links/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.UpdateLink(id, fields); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusOK)
}

func (a *Adapter) DeleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/links/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.DeleteLink(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusOK)
}

// writeJSON encodes val as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, statusCode int, val any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(val); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// writeOK writes a standard {"ok":true} JSON response.
func writeOK(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
