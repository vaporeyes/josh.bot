// ABOUTME: This file defines metrics types and pure calculation helpers for the Human Performance API.
// ABOUTME: It computes Epley 1RM, weekly tonnage, and best estimated 1RM from lift data.
package domain

import (
	"math"
	"time"
)

// HumanMetrics contains computed fitness metrics from workout data.
type HumanMetrics struct {
	Focus            string         `json:"focus"`
	WeeklyTonnageLbs int            `json:"weekly_tonnage_lbs"`
	Estimated1RM     map[string]int `json:"estimated_1rm"`
}

// MetricsResponse is the top-level response for GET /v1/metrics.
type MetricsResponse struct {
	Timestamp string       `json:"timestamp"`
	Human     HumanMetrics `json:"human"`
}

// MetricsService computes and returns the metrics dashboard.
type MetricsService interface {
	GetMetrics() (MetricsResponse, error)
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
