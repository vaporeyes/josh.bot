// ABOUTME: This file tests domain validation for required fields on entities.
// ABOUTME: Verifies Validate() returns ValidationError for empty required fields.
package domain

import (
	"errors"
	"testing"
)

func TestProject_Validate_EmptySlug(t *testing.T) {
	p := Project{Name: "My Project"}
	err := p.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "slug" {
		t.Errorf("field = %q, want %q", ve.Field, "slug")
	}
}

func TestProject_Validate_EmptyName(t *testing.T) {
	p := Project{Slug: "my-project"}
	err := p.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "name" {
		t.Errorf("field = %q, want %q", ve.Field, "name")
	}
}

func TestProject_Validate_OK(t *testing.T) {
	p := Project{Slug: "my-project", Name: "My Project"}
	if err := p.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLink_Validate_EmptyURL(t *testing.T) {
	l := Link{}
	err := l.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "url" {
		t.Errorf("field = %q, want %q", ve.Field, "url")
	}
}

func TestLink_Validate_OK(t *testing.T) {
	l := Link{URL: "https://example.com"}
	if err := l.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNote_Validate_EmptyTitle(t *testing.T) {
	n := Note{Body: "content"}
	err := n.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "title" {
		t.Errorf("field = %q, want %q", ve.Field, "title")
	}
}

func TestNote_Validate_EmptyBody(t *testing.T) {
	n := Note{Title: "Title"}
	err := n.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "body" {
		t.Errorf("field = %q, want %q", ve.Field, "body")
	}
}

func TestNote_Validate_OK(t *testing.T) {
	n := Note{Title: "Title", Body: "content"}
	if err := n.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTIL_Validate_EmptyTitle(t *testing.T) {
	til := TIL{Body: "content"}
	err := til.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "title" {
		t.Errorf("field = %q, want %q", ve.Field, "title")
	}
}

func TestTIL_Validate_EmptyBody(t *testing.T) {
	til := TIL{Title: "Title"}
	err := til.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "body" {
		t.Errorf("field = %q, want %q", ve.Field, "body")
	}
}

func TestLogEntry_Validate_EmptyMessage(t *testing.T) {
	le := LogEntry{}
	err := le.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "message" {
		t.Errorf("field = %q, want %q", ve.Field, "message")
	}
}

func TestDiaryEntry_Validate_EmptyBody(t *testing.T) {
	de := DiaryEntry{}
	err := de.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Field != "body" {
		t.Errorf("field = %q, want %q", ve.Field, "body")
	}
}

func TestDiaryEntry_Validate_OK(t *testing.T) {
	de := DiaryEntry{Body: "Shipped the feature"}
	if err := de.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
