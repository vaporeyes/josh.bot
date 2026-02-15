// ABOUTME: This file contains tests for the lift CSV parser.
// ABOUTME: Tests parsing of Strong-app CSV exports into Lift structs.
package domain

import (
	"strings"
	"testing"
)

func TestParseLiftsCSV_Success(t *testing.T) {
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2022-05-11 04:20:50,"josh ppl weighted",55m,"Squat (Barbell)",1,225.0,5.0,0,0.0,
2022-05-11 04:20:50,"josh ppl weighted",55m,"Squat (Barbell)",2,225.0,5.0,0,0.0,
2022-05-11 04:20:50,"josh ppl weighted",55m,"Wide Pull Up",1,35.0,8.0,0,0.0,`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lifts) != 3 {
		t.Fatalf("expected 3 lifts, got %d", len(lifts))
	}

	// Check first lift
	l := lifts[0]
	if l.Date != "2022-05-11 04:20:50" {
		t.Errorf("expected date '2022-05-11 04:20:50', got %q", l.Date)
	}
	if l.WorkoutName != "josh ppl weighted" {
		t.Errorf("expected workout 'josh ppl weighted', got %q", l.WorkoutName)
	}
	if l.Duration != "55m" {
		t.Errorf("expected duration '55m', got %q", l.Duration)
	}
	if l.ExerciseName != "Squat (Barbell)" {
		t.Errorf("expected exercise 'Squat (Barbell)', got %q", l.ExerciseName)
	}
	if l.SetOrder != "1" {
		t.Errorf("expected set order '1', got %q", l.SetOrder)
	}
	if l.Weight != 225.0 {
		t.Errorf("expected weight 225.0, got %f", l.Weight)
	}
	if l.Reps != 5.0 {
		t.Errorf("expected reps 5.0, got %f", l.Reps)
	}

	// Check ID is deterministic
	expectedID := LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "1")
	if l.ID != expectedID {
		t.Errorf("expected ID %q, got %q", expectedID, l.ID)
	}
}

func TestParseLiftsCSV_EmptyRPE(t *testing.T) {
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2022-05-11 04:20:50,"josh ppl weighted",55m,"Squat (Barbell)",1,225.0,5.0,0,0.0,`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lifts) != 1 {
		t.Fatalf("expected 1 lift, got %d", len(lifts))
	}

	// RPE should be zero value when empty
	if lifts[0].RPE != 0 {
		t.Errorf("expected RPE 0 for empty field, got %f", lifts[0].RPE)
	}
}

func TestParseLiftsCSV_WithRPE(t *testing.T) {
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2022-05-11 04:20:50,"josh ppl weighted",55m,"Squat (Barbell)",1,225.0,5.0,0,0.0,8.5`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lifts[0].RPE != 8.5 {
		t.Errorf("expected RPE 8.5, got %f", lifts[0].RPE)
	}
}

func TestParseLiftsCSV_HeaderOnly(t *testing.T) {
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lifts) != 0 {
		t.Errorf("expected 0 lifts for header-only CSV, got %d", len(lifts))
	}
}

func TestParseLiftsCSV_Distance(t *testing.T) {
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2022-05-11 04:20:50,"Cardio",30m,"Farmer's Walk",1,100.0,0.0,50,0.0,`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lifts[0].Distance != 50.0 {
		t.Errorf("expected distance 50.0, got %f", lifts[0].Distance)
	}
}

func TestParseLiftsCSV_Seconds(t *testing.T) {
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2022-05-21 04:56:30,"Week 1 Day 4",50m,"Plank",1,0,0.0,0,120.0,`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lifts[0].Seconds != 120.0 {
		t.Errorf("expected seconds 120.0, got %f", lifts[0].Seconds)
	}
}

func TestParseLiftsCSV_ZeroWeight(t *testing.T) {
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2022-05-11 04:20:50,"josh ppl weighted",55m,"1-Arm Push Up",1,0,8.0,0,0.0,`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lifts[0].Weight != 0 {
		t.Errorf("expected weight 0 for bodyweight exercise, got %f", lifts[0].Weight)
	}
	if lifts[0].Reps != 8.0 {
		t.Errorf("expected reps 8.0, got %f", lifts[0].Reps)
	}
}

func TestParseLiftsCSV_DuplicateWarmupSets(t *testing.T) {
	// Multiple warmup sets for the same exercise all have SetOrder "W".
	// The parser must disambiguate them with a sequence suffix.
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2022-08-15 12:12:05,"Push Pull Legs",60m,"Squat (Barbell)",W,135.0,5.0,0,0.0,
2022-08-15 12:12:05,"Push Pull Legs",60m,"Squat (Barbell)",W,225.0,5.0,0,0.0,
2022-08-15 12:12:05,"Push Pull Legs",60m,"Squat (Barbell)",W,275.0,3.0,0,0.0,`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lifts) != 3 {
		t.Fatalf("expected 3 lifts, got %d", len(lifts))
	}

	// All three must have unique IDs
	ids := make(map[string]bool)
	for _, l := range lifts {
		if ids[l.ID] {
			t.Errorf("duplicate ID: %s", l.ID)
		}
		ids[l.ID] = true
	}

	// First occurrence keeps the base ID, subsequent get a sequence suffix
	if lifts[0].ID != LiftID("2022-08-15 12:12:05", "Squat (Barbell)", "w") {
		t.Errorf("first warmup ID should be base: got %q", lifts[0].ID)
	}
}

func TestParseLiftsCSV_DuplicateExerciseInWorkout(t *testing.T) {
	// Same exercise appears twice in a workout with the same set order number.
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2024-11-11 09:46:56,"Powerbuilding Lower",60m,"Squat (Barbell)",2,45.0,5.0,0,0.0,
2024-11-11 09:46:56,"Powerbuilding Lower",60m,"Squat (Barbell)",2,245.0,3.0,0,0.0,`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lifts) != 2 {
		t.Fatalf("expected 2 lifts, got %d", len(lifts))
	}

	if lifts[0].ID == lifts[1].ID {
		t.Errorf("expected unique IDs, both got %q", lifts[0].ID)
	}
}

func TestParseLiftsCSV_NoDuplicateForUniqueRows(t *testing.T) {
	// Rows with different set orders should NOT get sequence suffixes.
	csv := `Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE
2022-05-11 04:20:50,"josh ppl weighted",55m,"Squat (Barbell)",1,225.0,5.0,0,0.0,
2022-05-11 04:20:50,"josh ppl weighted",55m,"Squat (Barbell)",2,225.0,5.0,0,0.0,`

	lifts, err := ParseLiftsCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// These should use the plain LiftID with no suffix
	if lifts[0].ID != LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "1") {
		t.Errorf("expected base ID for set 1, got %q", lifts[0].ID)
	}
	if lifts[1].ID != LiftID("2022-05-11 04:20:50", "Squat (Barbell)", "2") {
		t.Errorf("expected base ID for set 2, got %q", lifts[1].ID)
	}
}
