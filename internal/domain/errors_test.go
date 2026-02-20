// ABOUTME: This file tests custom error types for the domain layer.
// ABOUTME: Verifies NotFoundError and ValidationError work correctly with errors.As and wrapping.
package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestNotFoundError_Message(t *testing.T) {
	err := &NotFoundError{Resource: "project", ID: "my-slug"}
	want := `project "my-slug" not found`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestNotFoundError_MatchesThroughWrapping(t *testing.T) {
	inner := &NotFoundError{Resource: "link", ID: "abc"}
	wrapped := fmt.Errorf("service layer: %w", inner)

	var target *NotFoundError
	if !errors.As(wrapped, &target) {
		t.Fatal("expected errors.As to match NotFoundError through wrapping")
	}
	if target.Resource != "link" {
		t.Errorf("resource = %q, want %q", target.Resource, "link")
	}
}

func TestValidationError_Message(t *testing.T) {
	err := &ValidationError{Field: "slug", Message: "cannot be empty"}
	want := "validation: slug cannot be empty"
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestValidationError_MatchesThroughWrapping(t *testing.T) {
	inner := &ValidationError{Field: "url", Message: "required"}
	wrapped := fmt.Errorf("create link: %w", inner)

	var target *ValidationError
	if !errors.As(wrapped, &target) {
		t.Fatal("expected errors.As to match ValidationError through wrapping")
	}
	if target.Field != "url" {
		t.Errorf("field = %q, want %q", target.Field, "url")
	}
}
