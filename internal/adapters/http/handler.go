// ABOUTME: This file implements the HTTP handlers for the API endpoints.
// ABOUTME: It adapts incoming HTTP requests to calls on the domain service.
package http

import (
	"encoding/json"
	"net/http"

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
