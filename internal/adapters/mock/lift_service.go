// ABOUTME: This file provides a mock implementation of the LiftService for local development.
// ABOUTME: It returns hardcoded workout data to test the lift API endpoints without DynamoDB.
package mock

import (
	"context"
	"io"

	"github.com/jduncan/josh-bot/internal/domain"
)

// LiftService is a mock implementation of the domain.LiftService interface.
type LiftService struct{}

// NewLiftService creates a new mock LiftService.
func NewLiftService() *LiftService {
	return &LiftService{}
}

// GetRecentWorkouts returns hardcoded recent workouts for testing.
func (s *LiftService) GetRecentWorkouts(_ context.Context, limit int) ([]domain.WorkoutResponse, error) {
	workouts := []domain.WorkoutResponse{
		{
			Date:     "2026-03-25",
			Name:     "Day 4",
			Duration: "1h 13m",
			Exercises: []domain.ExerciseGroup{
				{
					Name: "Squat (Barbell)",
					Sets: []domain.SetDetail{
						{SetOrder: "1", Weight: 225, Reps: 5},
						{SetOrder: "2", Weight: 225, Reps: 5},
						{SetOrder: "3", Weight: 225, Reps: 5},
					},
				},
				{
					Name: "Bench Press (Barbell)",
					Sets: []domain.SetDetail{
						{SetOrder: "1", Weight: 185, Reps: 8},
						{SetOrder: "2", Weight: 185, Reps: 8},
					},
				},
			},
		},
		{
			Date:     "2026-03-23",
			Name:     "Day 3",
			Duration: "55m",
			Exercises: []domain.ExerciseGroup{
				{
					Name: "Deadlift (Barbell)",
					Sets: []domain.SetDetail{
						{SetOrder: "1", Weight: 315, Reps: 3},
						{SetOrder: "2", Weight: 315, Reps: 3},
					},
				},
			},
		},
	}
	if limit > 0 && limit < len(workouts) {
		workouts = workouts[:limit]
	}
	return workouts, nil
}

// GetLiftsByExercise returns hardcoded lifts for testing.
func (s *LiftService) GetLiftsByExercise(_ context.Context, exerciseName string) ([]domain.Lift, error) {
	allLifts := []domain.Lift{
		{ID: "lift#20260325#squat-barbell#1", Date: "2026-03-25 10:06:11", WorkoutName: "Day 4", Duration: "1h 13m", ExerciseName: "Squat (Barbell)", SetOrder: "1", Weight: 225, Reps: 5},
		{ID: "lift#20260323#squat-barbell#1", Date: "2026-03-23 09:30:00", WorkoutName: "Day 3", Duration: "55m", ExerciseName: "Squat (Barbell)", SetOrder: "1", Weight: 225, Reps: 5},
		{ID: "lift#20260325#bench-press-barbell#1", Date: "2026-03-25 10:06:11", WorkoutName: "Day 4", Duration: "1h 13m", ExerciseName: "Bench Press (Barbell)", SetOrder: "1", Weight: 185, Reps: 8},
	}
	var filtered []domain.Lift
	for _, l := range allLifts {
		if l.ExerciseName == exerciseName {
			filtered = append(filtered, l)
		}
	}
	return filtered, nil
}

// ImportLifts is a no-op in the mock adapter. It parses the CSV but does not persist.
func (s *LiftService) ImportLifts(_ context.Context, csvBody io.Reader) (domain.ImportSummary, error) {
	lifts, err := domain.ParseLiftsCSV(csvBody)
	if err != nil {
		return domain.ImportSummary{}, err
	}
	workouts := make(map[string]bool)
	exercises := make(map[string]bool)
	for _, l := range lifts {
		workouts[l.Date] = true
		exercises[l.ExerciseName] = true
	}
	return domain.ImportSummary{
		SetsImported: len(lifts),
		Workouts:     len(workouts),
		Exercises:    len(exercises),
	}, nil
}
