// ABOUTME: This file contains tests for the DynamoDB LiftService implementation.
// ABOUTME: It tests workout grouping, exercise filtering, and CSV import logic.
package dynamodb

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

func marshalLiftItem(t *testing.T, l domain.Lift) map[string]types.AttributeValue {
	t.Helper()
	item, err := attributevalue.MarshalMap(l)
	if err != nil {
		t.Fatalf("marshal lift: %v", err)
	}
	return item
}

func TestGetRecentWorkouts_Empty(t *testing.T) {
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{Items: []map[string]types.AttributeValue{}},
	}
	svc := NewLiftService(mock, "test-lifts")
	workouts, err := svc.GetRecentWorkouts(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workouts) != 0 {
		t.Errorf("expected 0 workouts, got %d", len(workouts))
	}
}

func TestGetRecentWorkouts_SingleWorkout(t *testing.T) {
	items := []map[string]types.AttributeValue{
		marshalLiftItem(t, domain.Lift{ID: "lift#1", Date: "2026-03-25 10:00:00", WorkoutName: "Day 1", Duration: "1h", ExerciseName: "Squat (Barbell)", SetOrder: "1", Weight: 225, Reps: 5}),
		marshalLiftItem(t, domain.Lift{ID: "lift#2", Date: "2026-03-25 10:00:00", WorkoutName: "Day 1", Duration: "1h", ExerciseName: "Squat (Barbell)", SetOrder: "2", Weight: 225, Reps: 5}),
		marshalLiftItem(t, domain.Lift{ID: "lift#3", Date: "2026-03-25 10:00:00", WorkoutName: "Day 1", Duration: "1h", ExerciseName: "Bench Press (Barbell)", SetOrder: "1", Weight: 185, Reps: 8}),
	}
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{Items: items},
	}
	svc := NewLiftService(mock, "test-lifts")
	workouts, err := svc.GetRecentWorkouts(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workouts) != 1 {
		t.Fatalf("expected 1 workout, got %d", len(workouts))
	}
	if workouts[0].Date != "2026-03-25" {
		t.Errorf("expected date 2026-03-25, got %s", workouts[0].Date)
	}
	if len(workouts[0].Exercises) != 2 {
		t.Errorf("expected 2 exercises, got %d", len(workouts[0].Exercises))
	}
	// Squat should have 2 sets
	if len(workouts[0].Exercises[0].Sets) != 2 {
		t.Errorf("expected 2 sets for first exercise, got %d", len(workouts[0].Exercises[0].Sets))
	}
}

func TestGetRecentWorkouts_SortedDescending(t *testing.T) {
	items := []map[string]types.AttributeValue{
		marshalLiftItem(t, domain.Lift{ID: "lift#1", Date: "2026-03-20 10:00:00", WorkoutName: "Old", Duration: "1h", ExerciseName: "Squat", SetOrder: "1", Weight: 200, Reps: 5}),
		marshalLiftItem(t, domain.Lift{ID: "lift#2", Date: "2026-03-25 10:00:00", WorkoutName: "Recent", Duration: "1h", ExerciseName: "Squat", SetOrder: "1", Weight: 225, Reps: 5}),
		marshalLiftItem(t, domain.Lift{ID: "lift#3", Date: "2026-03-22 10:00:00", WorkoutName: "Middle", Duration: "1h", ExerciseName: "Squat", SetOrder: "1", Weight: 215, Reps: 5}),
	}
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{Items: items},
	}
	svc := NewLiftService(mock, "test-lifts")
	workouts, err := svc.GetRecentWorkouts(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workouts) != 3 {
		t.Fatalf("expected 3 workouts, got %d", len(workouts))
	}
	if workouts[0].Date != "2026-03-25" {
		t.Errorf("expected first workout date 2026-03-25, got %s", workouts[0].Date)
	}
	if workouts[1].Date != "2026-03-22" {
		t.Errorf("expected second workout date 2026-03-22, got %s", workouts[1].Date)
	}
	if workouts[2].Date != "2026-03-20" {
		t.Errorf("expected third workout date 2026-03-20, got %s", workouts[2].Date)
	}
}

func TestGetRecentWorkouts_LimitParameter(t *testing.T) {
	items := []map[string]types.AttributeValue{
		marshalLiftItem(t, domain.Lift{ID: "lift#1", Date: "2026-03-25 10:00:00", WorkoutName: "W3", Duration: "1h", ExerciseName: "Squat", SetOrder: "1", Weight: 225, Reps: 5}),
		marshalLiftItem(t, domain.Lift{ID: "lift#2", Date: "2026-03-22 10:00:00", WorkoutName: "W2", Duration: "1h", ExerciseName: "Squat", SetOrder: "1", Weight: 215, Reps: 5}),
		marshalLiftItem(t, domain.Lift{ID: "lift#3", Date: "2026-03-20 10:00:00", WorkoutName: "W1", Duration: "1h", ExerciseName: "Squat", SetOrder: "1", Weight: 200, Reps: 5}),
	}
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{Items: items},
	}
	svc := NewLiftService(mock, "test-lifts")
	workouts, err := svc.GetRecentWorkouts(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workouts) != 2 {
		t.Errorf("expected 2 workouts with limit=2, got %d", len(workouts))
	}
}

func TestGetLiftsByExercise_Match(t *testing.T) {
	items := []map[string]types.AttributeValue{
		marshalLiftItem(t, domain.Lift{ID: "lift#1", Date: "2026-03-25 10:00:00", ExerciseName: "Squat (Barbell)", SetOrder: "1", Weight: 225, Reps: 5}),
		marshalLiftItem(t, domain.Lift{ID: "lift#2", Date: "2026-03-25 10:00:00", ExerciseName: "Bench Press (Barbell)", SetOrder: "1", Weight: 185, Reps: 8}),
		marshalLiftItem(t, domain.Lift{ID: "lift#3", Date: "2026-03-20 10:00:00", ExerciseName: "Squat (Barbell)", SetOrder: "1", Weight: 200, Reps: 5}),
	}
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{Items: items},
	}
	svc := NewLiftService(mock, "test-lifts")
	lifts, err := svc.GetLiftsByExercise(context.Background(), "Squat (Barbell)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lifts) != 2 {
		t.Fatalf("expected 2 squat sets, got %d", len(lifts))
	}
	// Should be sorted by date descending
	if lifts[0].Date != "2026-03-25 10:00:00" {
		t.Errorf("expected first set date 2026-03-25, got %s", lifts[0].Date)
	}
}

func TestGetLiftsByExercise_NoMatch(t *testing.T) {
	items := []map[string]types.AttributeValue{
		marshalLiftItem(t, domain.Lift{ID: "lift#1", Date: "2026-03-25 10:00:00", ExerciseName: "Squat (Barbell)", SetOrder: "1", Weight: 225, Reps: 5}),
	}
	mock := &mockDynamoDBClient{
		scanOutput: &dynamodb.ScanOutput{Items: items},
	}
	svc := NewLiftService(mock, "test-lifts")
	lifts, err := svc.GetLiftsByExercise(context.Background(), "Deadlift (Barbell)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lifts) != 0 {
		t.Errorf("expected 0 sets, got %d", len(lifts))
	}
}

func TestImportLifts_ValidCSV(t *testing.T) {
	csv := "Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE\n" +
		"2026-03-25 10:00:00,Day 1,1h,Squat (Barbell),1,225,5,0,0,\n" +
		"2026-03-25 10:00:00,Day 1,1h,Squat (Barbell),2,225,5,0,0,\n"

	mock := &mockDynamoDBClient{}
	svc := NewLiftService(mock, "test-lifts")
	summary, err := svc.ImportLifts(context.Background(), strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.SetsImported != 2 {
		t.Errorf("expected 2 sets imported, got %d", summary.SetsImported)
	}
	if summary.Workouts != 1 {
		t.Errorf("expected 1 workout, got %d", summary.Workouts)
	}
	if summary.Exercises != 1 {
		t.Errorf("expected 1 exercise, got %d", summary.Exercises)
	}
}

func TestImportLifts_EmptyCSV(t *testing.T) {
	csv := "Date,Workout Name,Duration,Exercise Name,Set Order,Weight,Reps,Distance,Seconds,RPE\n"
	mock := &mockDynamoDBClient{}
	svc := NewLiftService(mock, "test-lifts")
	summary, err := svc.ImportLifts(context.Background(), strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.SetsImported != 0 {
		t.Errorf("expected 0 sets imported, got %d", summary.SetsImported)
	}
}

func TestImportLifts_MissingColumns(t *testing.T) {
	csv := "Date,Workout Name\n2026-03-25,Day 1\n"
	mock := &mockDynamoDBClient{}
	svc := NewLiftService(mock, "test-lifts")
	_, err := svc.ImportLifts(context.Background(), strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing columns, got nil")
	}
}
