// ABOUTME: This file is the main entrypoint for the josh.bot API server.
// ABOUTME: It initializes dependencies and starts the HTTP server.
package main

import (
	"log"
	"net/http"

	"github.com/jduncan/josh-bot/internal/adapters/mock"
	httpadapter "github.com/jduncan/josh-bot/internal/adapters/http"
)

func main() {
	// Initialize the service
	service := mock.NewBotService()

	// Initialize the HTTP adapter
	adapter := httpadapter.NewAdapter(service)

	// Register the handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", adapter.StatusHandler)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
