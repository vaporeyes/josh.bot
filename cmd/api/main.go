// ABOUTME: This file is the main entrypoint for the josh.bot local development server.
// ABOUTME: It initializes dependencies and starts the HTTP server without API key auth.
package main

import (
	"log/slog"
	"net/http"
	"os"

	httpadapter "github.com/jduncan/josh-bot/internal/adapters/http"
	"github.com/jduncan/josh-bot/internal/adapters/mock"
)

func main() {
	// Initialize the services
	service := mock.NewBotService()
	metricsService := mock.NewMetricsService()
	memService := mock.NewMemService()
	liftService := mock.NewLiftService()

	// Initialize the HTTP adapter
	adapter := httpadapter.NewAdapter(service, metricsService, memService)
	adapter.SetLiftService(liftService)

	// Register the handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", adapter.StatusHandler)
	mux.HandleFunc("/v1/metrics", adapter.MetricsHandler)
	mux.HandleFunc("/v1/projects", adapter.ProjectsHandler)
	mux.HandleFunc("/v1/notes", adapter.NotesHandler)
	mux.HandleFunc("/v1/til", adapter.TILsHandler)
	mux.HandleFunc("/v1/log", adapter.LogEntriesHandler)
	mux.HandleFunc("/v1/books", adapter.BooksHandler)
	mux.HandleFunc("/v1/books/", adapter.BookHandler)
	mux.HandleFunc("/v1/diary", adapter.DiaryEntriesHandler)
	mux.HandleFunc("/v1/diary/", adapter.DiaryEntryHandler)
	mux.HandleFunc("/v1/lifts/recent", adapter.LiftsRecentHandler)
	mux.HandleFunc("/v1/lifts/import", adapter.LiftsImportHandler)
	mux.HandleFunc("/v1/lifts/exercise/", adapter.LiftsExerciseHandler)
	mux.HandleFunc("/v1/mem/observations", adapter.MemObservationsHandler)
	mux.HandleFunc("/v1/mem/observations/", adapter.MemObservationHandler)
	mux.HandleFunc("/v1/mem/summaries", adapter.MemSummariesHandler)
	mux.HandleFunc("/v1/mem/summaries/", adapter.MemSummaryHandler)
	mux.HandleFunc("/v1/mem/prompts", adapter.MemPromptsHandler)
	mux.HandleFunc("/v1/mem/prompts/", adapter.MemPromptHandler)
	mux.HandleFunc("/v1/mem/stats", adapter.MemStatsHandler)
	mux.HandleFunc("/v1/memory", adapter.MemoriesHandler)
	mux.HandleFunc("/v1/memory/", adapter.MemoryHandler)

	// Start the server
	slog.Info("starting server", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("could not start server", "error", err)
		os.Exit(1)
	}
}
