// ABOUTME: This file defines metrics types and pure calculation helpers for the Human Performance API.
// ABOUTME: It computes Epley 1RM, weekly tonnage, and best estimated 1RM from lift data.
package domain

import (
	"context"
	"math"
	"time"
)

// WorkoutSummary contains a snapshot of a single workout session.
type WorkoutSummary struct {
	Date       string   `json:"date"`
	Name       string   `json:"name"`
	Exercises  []string `json:"exercises"`
	Sets       int      `json:"sets"`
	TonnageLbs int      `json:"tonnage_lbs"`
}

// HumanMetrics contains computed fitness metrics from workout data.
type HumanMetrics struct {
	Focus            string          `json:"focus"`
	WeeklyTonnageLbs int             `json:"weekly_tonnage_lbs"`
	Estimated1RM     map[string]int  `json:"estimated_1rm"`
	LastWorkout      *WorkoutSummary `json:"last_workout"`
}

// MetricsResponse is the top-level response for GET /v1/metrics.
type MetricsResponse struct {
	Timestamp string       `json:"timestamp"`
	Human     HumanMetrics `json:"human"`
	Dev       *MemStats    `json:"dev,omitempty"`
}

// MetricsService computes and returns the metrics dashboard.
type MetricsService interface {
	GetMetrics(ctx context.Context) (MetricsResponse, error)
}

// Epley1RM estimates a 1-rep max using the Epley formula: weight * (1 + reps/30).
// Returns 0 for zero reps or zero weight. Truncates to integer.
func Epley1RM(weight, reps float64) int {
	if weight == 0 || reps == 0 {
		return 0
	}
	if reps == 1 {
		return int(weight)
	}
	return int(math.Floor(weight * (1 + reps/30)))
}

// WeeklyTonnage computes total weight moved (weight * reps) for sets within the last 7 days.
// AIDEV-NOTE: Date format from CSV is "2006-01-02 15:04:05".
func WeeklyTonnage(lifts []Lift, now time.Time) int {
	cutoff := now.AddDate(0, 0, -7)
	total := 0.0
	for _, l := range lifts {
		t, err := time.Parse("2006-01-02 15:04:05", l.Date)
		if err != nil {
			continue
		}
		if t.After(cutoff) {
			total += l.Weight * l.Reps
		}
	}
	return int(total)
}

// LastWorkout finds the most recent workout by date and returns a summary.
// Returns nil if lifts is empty. Date is extracted as YYYY-MM-DD.
func LastWorkout(lifts []Lift) *WorkoutSummary {
	if len(lifts) == 0 {
		return nil
	}

	// Find the latest date string (lexicographic max works for YYYY-MM-DD HH:MM:SS)
	latestDate := ""
	for _, l := range lifts {
		if l.Date > latestDate {
			latestDate = l.Date
		}
	}

	// Extract just the date portion (first 10 chars: "2026-02-14")
	dateOnly := latestDate[:10]

	// Gather all sets from that date
	seen := make(map[string]bool)
	var exercises []string
	var sets int
	tonnage := 0.0
	name := ""

	for _, l := range lifts {
		if len(l.Date) < 10 || l.Date[:10] != dateOnly {
			continue
		}
		sets++
		tonnage += l.Weight * l.Reps
		if name == "" {
			name = l.WorkoutName
		}
		if !seen[l.ExerciseName] {
			seen[l.ExerciseName] = true
			exercises = append(exercises, l.ExerciseName)
		}
	}

	return &WorkoutSummary{
		Date:       dateOnly,
		Name:       name,
		Exercises:  exercises,
		Sets:       sets,
		TonnageLbs: int(tonnage),
	}
}

// BestE1RM finds the highest estimated 1RM across all sets matching the given exercise name.
func BestE1RM(lifts []Lift, exerciseName string) int {
	best := 0
	for _, l := range lifts {
		if l.ExerciseName != exerciseName {
			continue
		}
		e := Epley1RM(l.Weight, l.Reps)
		if e > best {
			best = e
		}
	}
	return best
}
