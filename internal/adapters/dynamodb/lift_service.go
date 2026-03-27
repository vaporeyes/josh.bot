// ABOUTME: This file implements LiftService using DynamoDB.
// ABOUTME: It scans the lifts table and groups results into workout sessions for the API.
package dynamodb

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

// LiftService implements domain.LiftService using DynamoDB.
type LiftService struct {
	client    DynamoDBClient
	tableName string
}

// NewLiftService creates a DynamoDB-backed LiftService.
func NewLiftService(client DynamoDBClient, tableName string) *LiftService {
	return &LiftService{client: client, tableName: tableName}
}

// GetRecentWorkouts returns the N most recent workout sessions grouped by date and name.
func (s *LiftService) GetRecentWorkouts(ctx context.Context, limit int) ([]domain.WorkoutResponse, error) {
	lifts, err := s.scanLifts(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan lifts: %w", err)
	}
	return groupWorkouts(lifts, limit), nil
}

// GetLiftsByExercise returns all sets for a specific exercise sorted by date descending.
func (s *LiftService) GetLiftsByExercise(ctx context.Context, exerciseName string) ([]domain.Lift, error) {
	lifts, err := s.scanLifts(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan lifts: %w", err)
	}

	var filtered []domain.Lift
	for _, l := range lifts {
		if l.ExerciseName == exerciseName {
			filtered = append(filtered, l)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date > filtered[j].Date
	})

	return filtered, nil
}

// ImportLifts parses a Strong-format CSV and writes all sets to DynamoDB.
func (s *LiftService) ImportLifts(ctx context.Context, csvBody io.Reader) (domain.ImportSummary, error) {
	lifts, err := domain.ParseLiftsCSV(csvBody)
	if err != nil {
		return domain.ImportSummary{}, fmt.Errorf("parse CSV: %w", err)
	}

	if len(lifts) == 0 {
		return domain.ImportSummary{}, nil
	}

	if err := domain.BatchWriteLifts(ctx, s.client, s.tableName, lifts); err != nil {
		return domain.ImportSummary{}, fmt.Errorf("batch write: %w", err)
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

// scanLifts retrieves all lift records from the lifts table.
func (s *LiftService) scanLifts(ctx context.Context) ([]domain.Lift, error) {
	var lifts []domain.Lift
	var lastKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:         &s.tableName,
			ExclusiveStartKey: lastKey,
		}

		output, err := s.client.Scan(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("dynamodb Scan: %w", err)
		}

		for _, item := range output.Items {
			var l domain.Lift
			if err := attributevalue.UnmarshalMap(item, &l); err != nil {
				return nil, fmt.Errorf("unmarshal lift: %w", err)
			}
			lifts = append(lifts, l)
		}

		lastKey = output.LastEvaluatedKey
		if lastKey == nil {
			break
		}
	}

	return lifts, nil
}

// groupWorkouts groups flat lift sets into WorkoutResponse structs.
// Workouts are sorted by date descending and limited to the specified count.
func groupWorkouts(lifts []domain.Lift, limit int) []domain.WorkoutResponse {
	if len(lifts) == 0 {
		return []domain.WorkoutResponse{}
	}

	// Sort lifts by date descending so we encounter recent workouts first.
	sort.Slice(lifts, func(i, j int) bool {
		return lifts[i].Date > lifts[j].Date
	})

	// Group by (date prefix, workout name). Date prefix is the first 10 chars.
	type workoutKey struct {
		date string
		name string
	}

	var orderedKeys []workoutKey
	keyIndex := make(map[workoutKey]int)
	var workouts []domain.WorkoutResponse

	for _, l := range lifts {
		datePrefix := l.Date
		if len(datePrefix) >= 10 {
			datePrefix = datePrefix[:10]
		}
		key := workoutKey{date: datePrefix, name: l.WorkoutName}

		idx, exists := keyIndex[key]
		if !exists {
			if limit > 0 && len(orderedKeys) >= limit {
				continue
			}
			idx = len(orderedKeys)
			keyIndex[key] = idx
			orderedKeys = append(orderedKeys, key)
			workouts = append(workouts, domain.WorkoutResponse{
				Date:      datePrefix,
				Name:      l.WorkoutName,
				Duration:  l.Duration,
				Exercises: []domain.ExerciseGroup{},
			})
		}

		// Find or create exercise group within the workout.
		w := &workouts[idx]
		exerciseIdx := -1
		for i, eg := range w.Exercises {
			if eg.Name == l.ExerciseName {
				exerciseIdx = i
				break
			}
		}
		if exerciseIdx == -1 {
			w.Exercises = append(w.Exercises, domain.ExerciseGroup{
				Name: l.ExerciseName,
				Sets: []domain.SetDetail{},
			})
			exerciseIdx = len(w.Exercises) - 1
		}

		w.Exercises[exerciseIdx].Sets = append(w.Exercises[exerciseIdx].Sets, domain.SetDetail{
			SetOrder: l.SetOrder,
			Weight:   l.Weight,
			Reps:     l.Reps,
			Distance: l.Distance,
			Seconds:  l.Seconds,
			RPE:      l.RPE,
		})
	}

	return workouts
}
