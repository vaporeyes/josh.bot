// ABOUTME: This file contains tests for metrics calculation helpers.
// ABOUTME: Tests Epley 1RM formula, weekly tonnage, and best E1RM extraction.
package domain

import (
	"testing"
	"time"
)

// --- Epley1RM ---

func TestEpley1RM_SingleRep(t *testing.T) {
	// 1 rep = just the weight (no multiplier)
	got := Epley1RM(315, 1)
	if got != 315 {
		t.Errorf("Epley1RM(315, 1) = %d, want 315", got)
	}
}

func TestEpley1RM_FiveReps(t *testing.T) {
	// 225 * (1 + 5/30) = 225 * 1.1667 = 262.5 -> 262
	got := Epley1RM(225, 5)
	if got != 262 {
		t.Errorf("Epley1RM(225, 5) = %d, want 262", got)
	}
}

func TestEpley1RM_ZeroReps(t *testing.T) {
	got := Epley1RM(225, 0)
	if got != 0 {
		t.Errorf("Epley1RM(225, 0) = %d, want 0", got)
	}
}

func TestEpley1RM_ZeroWeight(t *testing.T) {
	got := Epley1RM(0, 10)
	if got != 0 {
		t.Errorf("Epley1RM(0, 10) = %d, want 0", got)
	}
}

func TestEpley1RM_TenReps(t *testing.T) {
	// 200 * (1 + 10/30) = 200 * 1.3333 = 266.67 -> 266
	got := Epley1RM(200, 10)
	if got != 266 {
		t.Errorf("Epley1RM(200, 10) = %d, want 266", got)
	}
}

// --- WeeklyTonnage ---

func TestWeeklyTonnage_WithinWindow(t *testing.T) {
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	lifts := []Lift{
		{Date: "2026-02-14 10:00:00", Weight: 225, Reps: 5}, // 1125
		{Date: "2026-02-13 08:00:00", Weight: 315, Reps: 3}, // 945
	}
	got := WeeklyTonnage(lifts, now)
	want := 2070
	if got != want {
		t.Errorf("WeeklyTonnage = %d, want %d", got, want)
	}
}

func TestWeeklyTonnage_ExcludesOldSets(t *testing.T) {
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	lifts := []Lift{
		{Date: "2026-02-14 10:00:00", Weight: 225, Reps: 5}, // 1125 (recent)
		{Date: "2026-02-01 08:00:00", Weight: 315, Reps: 3}, // excluded (>7 days)
	}
	got := WeeklyTonnage(lifts, now)
	want := 1125
	if got != want {
		t.Errorf("WeeklyTonnage = %d, want %d", got, want)
	}
}

func TestWeeklyTonnage_Empty(t *testing.T) {
	now := time.Now()
	got := WeeklyTonnage(nil, now)
	if got != 0 {
		t.Errorf("WeeklyTonnage(nil) = %d, want 0", got)
	}
}

func TestWeeklyTonnage_IncludesBodyweight(t *testing.T) {
	// Zero weight sets contribute 0 tonnage
	now := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	lifts := []Lift{
		{Date: "2026-02-14 10:00:00", Weight: 0, Reps: 10},
	}
	got := WeeklyTonnage(lifts, now)
	if got != 0 {
		t.Errorf("WeeklyTonnage bodyweight = %d, want 0", got)
	}
}

// --- BestE1RM ---

func TestBestE1RM_PicksHighest(t *testing.T) {
	lifts := []Lift{
		{ExerciseName: "Deadlift (Barbell)", Weight: 405, Reps: 3}, // 405 * (1+3/30) = 445.5 -> 445
		{ExerciseName: "Deadlift (Barbell)", Weight: 455, Reps: 1}, // 455
		{ExerciseName: "Deadlift (Barbell)", Weight: 365, Reps: 5}, // 365 * (1+5/30) = 425.8 -> 425
	}
	got := BestE1RM(lifts, "Deadlift (Barbell)")
	if got != 455 {
		t.Errorf("BestE1RM = %d, want 455", got)
	}
}

func TestBestE1RM_FiltersExercise(t *testing.T) {
	lifts := []Lift{
		{ExerciseName: "Deadlift (Barbell)", Weight: 405, Reps: 1},
		{ExerciseName: "Squat (Barbell)", Weight: 500, Reps: 1}, // heavier but wrong exercise
	}
	got := BestE1RM(lifts, "Deadlift (Barbell)")
	if got != 405 {
		t.Errorf("BestE1RM = %d, want 405", got)
	}
}

func TestBestE1RM_NoMatches(t *testing.T) {
	lifts := []Lift{
		{ExerciseName: "Squat (Barbell)", Weight: 315, Reps: 5},
	}
	got := BestE1RM(lifts, "Deadlift (Barbell)")
	if got != 0 {
		t.Errorf("BestE1RM no matches = %d, want 0", got)
	}
}

func TestBestE1RM_Empty(t *testing.T) {
	got := BestE1RM(nil, "Deadlift (Barbell)")
	if got != 0 {
		t.Errorf("BestE1RM(nil) = %d, want 0", got)
	}
}

func TestBestE1RM_HighRepSetBeatsHeavySingle(t *testing.T) {
	lifts := []Lift{
		{ExerciseName: "Squat (Barbell)", Weight: 315, Reps: 8}, // 315 * (1+8/30) = 399 -> 399
		{ExerciseName: "Squat (Barbell)", Weight: 385, Reps: 1}, // 385
	}
	got := BestE1RM(lifts, "Squat (Barbell)")
	if got != 399 {
		t.Errorf("BestE1RM = %d, want 399", got)
	}
}
