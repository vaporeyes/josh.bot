// ABOUTME: This file is the main entrypoint for the josh.bot local development server.
// ABOUTME: It initializes dependencies and starts the HTTP server without API key auth.
package main

import (
	"log"
	"net/http"

	httpadapter "github.com/jduncan/josh-bot/internal/adapters/http"
	"github.com/jduncan/josh-bot/internal/adapters/mock"
)

func main() {
	// Initialize the services
	service := mock.NewBotService()
	metricsService := mock.NewMetricsService()

	// Initialize the HTTP adapter
	adapter := httpadapter.NewAdapter(service, metricsService)

	// Register the handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", adapter.StatusHandler)
	mux.HandleFunc("/v1/metrics", adapter.MetricsHandler)
	mux.HandleFunc("/v1/projects", adapter.ProjectsHandler)
	mux.HandleFunc("/v1/notes", adapter.NotesHandler)
	mux.HandleFunc("/v1/til", adapter.TILsHandler)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
