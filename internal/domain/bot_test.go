// ABOUTME: This file contains tests for domain helper functions.
// ABOUTME: Tests ExerciseSlug, CompactDate, Lift ID generation, and Note ID generation.
package domain

import (
	"strings"
	"testing"
)

func TestExerciseSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Squat (Barbell)", "squat-barbell"},
		{"Bench Press (Barbell)", "bench-press-barbell"},
		{"Wide Pull Up", "wide-pull-up"},
		{"1-Arm Push Up", "1-arm-push-up"},
		{"Triceps Pushdown (Cable - Straight Bar)", "triceps-pushdown-cable-straight-bar"},
		{"Crunch", "crunch"},
		{"Flat Leg Raise", "flat-leg-raise"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExerciseSlug(tt.input)
			if got != tt.want {
				t.Errorf("ExerciseSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompactDate(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2022-05-11 04:20:50", "20220511T042050"},
		{"2026-02-06 03:31:15", "20260206T033115"},
		{"2023-12-25 23:59:59", "20231225T235959"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CompactDate(tt.input)
			if got != tt.want {
				t.Errorf("CompactDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLiftID(t *testing.T) {
	tests := []struct {
		date     string
		exercise string
		setOrder string
		want     string
	}{
		{
			"2022-05-11 04:20:50", "Squat (Barbell)", "1",
			"lift#20220511T042050#squat-barbell#1",
		},
		{
			"2022-05-11 04:20:50", "Wide Pull Up", "3",
			"lift#20220511T042050#wide-pull-up#3",
		},
		{
			"2026-02-06 03:31:15", "Bench Press (Barbell)", "4",
			"lift#20260206T033115#bench-press-barbell#4",
		},
		{
			"2022-08-15 12:12:05", "Squat (Barbell)", "W",
			"lift#20220815T121205#squat-barbell#w",
		},
		{
			"2022-08-15 12:12:05", "Squat (Barbell)", "F",
			"lift#20220815T121205#squat-barbell#f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := LiftID(tt.date, tt.exercise, tt.setOrder)
			if got != tt.want {
				t.Errorf("LiftID(%q, %q, %q) = %q, want %q",
					tt.date, tt.exercise, tt.setOrder, got, tt.want)
			}
		})
	}
}

func TestExerciseSlug_ConsecutiveSpecialChars(t *testing.T) {
	// Ensure multiple consecutive non-alphanumeric chars collapse to single hyphen
	got := ExerciseSlug("Bench Press - Close Grip (Barbell)")
	want := "bench-press-close-grip-barbell"
	if got != want {
		t.Errorf("ExerciseSlug with consecutive special chars = %q, want %q", got, want)
	}
}

func TestExerciseSlug_NoTrailingHyphen(t *testing.T) {
	// Exercise names ending with special chars shouldn't produce trailing hyphens
	got := ExerciseSlug("Chest Fly (Dumbbell)")
	if got[len(got)-1] == '-' {
		t.Errorf("ExerciseSlug(%q) has trailing hyphen: %q", "Chest Fly (Dumbbell)", got)
	}
	want := "chest-fly-dumbbell"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLiftID_Format(t *testing.T) {
	id := LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "1")
	// Must start with lift# prefix
	if id[:5] != "lift#" {
		t.Errorf("LiftID should start with 'lift#', got %q", id)
	}
	// Must have exactly 4 parts separated by #
	parts := 0
	for _, c := range id {
		if c == '#' {
			parts++
		}
	}
	if parts != 3 {
		t.Errorf("LiftID should have 3 '#' separators, got %d in %q", parts, id)
	}
}

func TestCompactDate_Roundtrip(t *testing.T) {
	// Verify the compact date preserves all time info
	input := "2022-05-11 04:20:50"
	compact := CompactDate(input)
	if len(compact) != 15 { // "20220511T042050"
		t.Errorf("CompactDate length should be 15, got %d for %q", len(compact), compact)
	}
	// T separator should be at position 8
	if compact[8] != 'T' {
		t.Errorf("CompactDate should have 'T' at position 8, got %q", compact)
	}
}

func TestExerciseSlug_Deterministic(t *testing.T) {
	name := "Bent Over Row (Barbell)"
	a := ExerciseSlug(name)
	b := ExerciseSlug(name)
	if a != b {
		t.Errorf("ExerciseSlug should be deterministic: %q != %q", a, b)
	}
}

func TestLiftID_Deterministic(t *testing.T) {
	a := LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "1")
	b := LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "1")
	if a != b {
		t.Errorf("LiftID should be deterministic for idempotent imports: %q != %q", a, b)
	}
}

func TestLiftID_Unique(t *testing.T) {
	// Same workout, different exercises should produce different IDs
	a := LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "1")
	b := LiftID("2022-05-11 04:20:50", "Bench Press (Barbell)", "1")
	if a == b {
		t.Error("Different exercises should produce different IDs")
	}

	// Same exercise, different sets should produce different IDs
	c := LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "1")
	d := LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "2")
	if c == d {
		t.Error("Different set orders should produce different IDs")
	}

	// Same exercise+set, different dates should produce different IDs
	e := LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "1")
	f := LiftID("2022-06-11 04:20:50", "Squat (Barbell)", "1")
	if e == f {
		t.Error("Different dates should produce different IDs")
	}
}

func TestNoteID_Format(t *testing.T) {
	id := NoteID()
	if !strings.HasPrefix(id, "note#") {
		t.Errorf("NoteID should start with 'note#', got %q", id)
	}
	// "note#" (5 chars) + 16 hex chars (8 bytes) = 21 chars total
	if len(id) != 21 {
		t.Errorf("NoteID length should be 21, got %d for %q", len(id), id)
	}
}

func TestNoteID_Unique(t *testing.T) {
	a := NoteID()
	b := NoteID()
	if a == b {
		t.Errorf("two NoteID calls should produce different IDs: %q == %q", a, b)
	}
}
