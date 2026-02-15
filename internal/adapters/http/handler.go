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
	service        domain.BotService
	metricsService domain.MetricsService
}

func NewAdapter(service domain.BotService, metricsService domain.MetricsService) *Adapter {
	return &Adapter{service: service, metricsService: metricsService}
}

// MetricsHandler handles GET /v1/metrics.
func (a *Adapter) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	metrics, err := a.metricsService.GetMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, metrics)
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

func (a *Adapter) NotesHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	notes, err := a.service.GetNotes(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, notes)
}

func (a *Adapter) CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	var note domain.Note
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.CreateNote(note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusCreated)
}

// NoteHandler handles GET /v1/notes/{id}.
func (a *Adapter) NoteHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/notes/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	note, err := a.service.GetNote(id)
	if err != nil {
		http.Error(w, `{"error":"note not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, note)
}

func (a *Adapter) UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/notes/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.UpdateNote(id, fields); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusOK)
}

func (a *Adapter) DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/notes/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.DeleteNote(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusOK)
}

func (a *Adapter) TILsHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	tils, err := a.service.GetTILs(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tils)
}

func (a *Adapter) CreateTILHandler(w http.ResponseWriter, r *http.Request) {
	var til domain.TIL
	if err := json.NewDecoder(r.Body).Decode(&til); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.CreateTIL(til); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusCreated)
}

// TILHandler handles GET /v1/til/{id}.
func (a *Adapter) TILHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/til/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	til, err := a.service.GetTIL(id)
	if err != nil {
		http.Error(w, `{"error":"til not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, til)
}

func (a *Adapter) UpdateTILHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/til/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.UpdateTIL(id, fields); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusOK)
}

func (a *Adapter) DeleteTILHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/til/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.DeleteTIL(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusOK)
}

func (a *Adapter) LogEntriesHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	entries, err := a.service.GetLogEntries(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, entries)
}

func (a *Adapter) CreateLogEntryHandler(w http.ResponseWriter, r *http.Request) {
	var entry domain.LogEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.CreateLogEntry(entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusCreated)
}

// LogEntryHandler handles GET /v1/log/{id}.
func (a *Adapter) LogEntryHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/log/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	entry, err := a.service.GetLogEntry(id)
	if err != nil {
		http.Error(w, `{"error":"log entry not found"}`, http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

func (a *Adapter) UpdateLogEntryHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/log/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.UpdateLogEntry(id, fields); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeOK(w, http.StatusOK)
}

func (a *Adapter) DeleteLogEntryHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/log/")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	if err := a.service.DeleteLogEntry(id); err != nil {
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
