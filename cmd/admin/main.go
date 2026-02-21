// ABOUTME: Entrypoint for the admin dashboard HTTP server.
// ABOUTME: Proxies to api.josh.bot (or configured API_BASE_URL) with API key auth.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/jduncan/josh-bot/internal/adapters/admin"
)

func main() {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.josh.bot"
	}
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		slog.Error("API_KEY environment variable is required")
		os.Exit(1)
	}

	addr := os.Getenv("ADMIN_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	client := admin.NewClient(baseURL, apiKey)
	handlers := admin.NewHandlers(client)

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)

	slog.Info("starting admin dashboard", "addr", addr, "api", baseURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("could not start server", "error", err)
		os.Exit(1)
	}
}
