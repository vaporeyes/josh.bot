// ABOUTME: This file contains tests for the HTTP handlers.
// ABOUTME: It follows a TDD approach for building the API.
package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jduncan/josh-bot/internal/adapters/mock"
)

func TestStatusHandler(t *testing.T) {
	mockService := mock.NewBotService()
	adapter := NewAdapter(mockService)

	req, err := http.NewRequest("GET", "/v1/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.StatusHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"current_activity":"Architecting josh.bot","location":"Clarksville, TN","status":"ok"}`
	// We trim the newline that Encode appends
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
