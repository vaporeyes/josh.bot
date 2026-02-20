// ABOUTME: This file defines custom error types for the domain layer.
// ABOUTME: NotFoundError and ValidationError enable handlers to return correct HTTP status codes.
package domain

import "fmt"

// NotFoundError indicates a requested resource does not exist.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %q not found", e.Resource, e.ID)
}

// ValidationError indicates invalid input for a domain operation.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s %s", e.Field, e.Message)
}
