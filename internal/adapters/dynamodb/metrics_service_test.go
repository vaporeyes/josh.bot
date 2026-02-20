// ABOUTME: This file contains tests for the DynamoDB-backed MetricsService.
// ABOUTME: Tests use a mock DynamoDB client to verify metrics computation from lift data.
package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

// mockMetricsClient implements DynamoDBClient for metrics tests.
type mockMetricsClient struct {
	scanOutput    *dynamodb.ScanOutput
	getItemOutput *dynamodb.GetItemOutput
	scanErr       error
	getItemErr    error
}

func (m *mockMetricsClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	if m.scanErr != nil {
		return nil, m.scanErr
	}
	return m.scanOutput, nil
}

func (m *mockMetricsClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if m.getItemErr != nil {
		return nil, m.getItemErr
	}
	return m.getItemOutput, nil
}

func (m *mockMetricsClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return &dynamodb.UpdateItemOutput{}, nil
}

func (m *mockMetricsClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return &dynamodb.PutItemOutput{}, nil
}

func (m *mockMetricsClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return &dynamodb.DeleteItemOutput{}, nil
}

func (m *mockMetricsClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return &dynamodb.QueryOutput{}, nil
}

func marshalLift(t *testing.T, l domain.Lift) map[string]types.AttributeValue {
	t.Helper()
	item, err := attributevalue.MarshalMap(l)
	if err != nil {
		t.Fatalf("marshal lift: %v", err)
	}
	return item
}

func TestMetricsService_GetMetrics(t *testing.T) {
	statusItem := map[string]types.AttributeValue{
		"id":    &types.AttributeValueMemberS{Value: "status"},
		"focus": &types.AttributeValueMemberS{Value: "Powerlifting / Hypertrophy"},
	}

	lifts := []domain.Lift{
		{ID: "lift#1", Date: "2026-02-14 10:00:00", ExerciseName: "Deadlift (Barbell)", Weight: 455, Reps: 1},
		{ID: "lift#2", Date: "2026-02-14 10:05:00", ExerciseName: "Squat (Barbell)", Weight: 315, Reps: 5},
		{ID: "lift#3", Date: "2026-02-14 10:10:00", ExerciseName: "Bench Press (Barbell)", Weight: 225, Reps: 3},
		{ID: "lift#4", Date: "2025-01-01 08:00:00", ExerciseName: "Deadlift (Barbell)", Weight: 500, Reps: 1}, // old but best E1RM
	}

	var items []map[string]types.AttributeValue
	for _, l := range lifts {
		items = append(items, marshalLift(t, l))
	}

	mock := &mockMetricsClient{
		scanOutput:    &dynamodb.ScanOutput{Items: items},
		getItemOutput: &dynamodb.GetItemOutput{Item: statusItem},
	}

	svc := NewMetricsService(mock, "lifts-table", "data-table", nil)
	resp, err := svc.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetMetrics error: %v", err)
	}

	if resp.Human.Focus != "Powerlifting / Hypertrophy" {
		t.Errorf("focus = %q, want %q", resp.Human.Focus, "Powerlifting / Hypertrophy")
	}

	if resp.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}

	// E1RM: deadlift should be 500 (from old set, 1 rep)
	if resp.Human.Estimated1RM["deadlift"] != 500 {
		t.Errorf("deadlift E1RM = %d, want 500", resp.Human.Estimated1RM["deadlift"])
	}

	// E1RM: squat 315*(1+5/30) = 367
	if resp.Human.Estimated1RM["squat"] != 367 {
		t.Errorf("squat E1RM = %d, want 367", resp.Human.Estimated1RM["squat"])
	}

	// E1RM: bench 225*(1+3/30) = 247
	if resp.Human.Estimated1RM["bench"] != 247 {
		t.Errorf("bench E1RM = %d, want 247", resp.Human.Estimated1RM["bench"])
	}

	// LastWorkout: most recent date is 2026-02-14 (3 sets from that date)
	if resp.Human.LastWorkout == nil {
		t.Fatal("expected LastWorkout, got nil")
	}
	if resp.Human.LastWorkout.Date != "2026-02-14" {
		t.Errorf("LastWorkout.Date = %q, want %q", resp.Human.LastWorkout.Date, "2026-02-14")
	}
	if resp.Human.LastWorkout.Sets != 3 {
		t.Errorf("LastWorkout.Sets = %d, want 3", resp.Human.LastWorkout.Sets)
	}
}

func TestMetricsService_EmptyLifts(t *testing.T) {
	statusItem := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "status"},
	}

	mock := &mockMetricsClient{
		scanOutput:    &dynamodb.ScanOutput{Items: nil},
		getItemOutput: &dynamodb.GetItemOutput{Item: statusItem},
	}

	svc := NewMetricsService(mock, "lifts-table", "data-table", nil)
	resp, err := svc.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetMetrics error: %v", err)
	}

	if resp.Human.WeeklyTonnageLbs != 0 {
		t.Errorf("tonnage = %d, want 0", resp.Human.WeeklyTonnageLbs)
	}
	if resp.Human.Estimated1RM["deadlift"] != 0 {
		t.Errorf("deadlift E1RM = %d, want 0", resp.Human.Estimated1RM["deadlift"])
	}
	if resp.Human.LastWorkout != nil {
		t.Errorf("LastWorkout = %v, want nil for empty lifts", resp.Human.LastWorkout)
	}
}

func TestMetricsService_ScanError(t *testing.T) {
	mock := &mockMetricsClient{
		scanErr:       context.DeadlineExceeded,
		getItemOutput: &dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{}},
	}

	svc := NewMetricsService(mock, "lifts-table", "data-table", nil)
	_, err := svc.GetMetrics(context.Background())
	if err == nil {
		t.Fatal("expected error from scan failure, got nil")
	}
}

func TestMetricsService_FocusMissing(t *testing.T) {
	// Status exists but has no focus field — should return empty string
	statusItem := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: "status"},
	}

	mock := &mockMetricsClient{
		scanOutput:    &dynamodb.ScanOutput{Items: nil},
		getItemOutput: &dynamodb.GetItemOutput{Item: statusItem},
	}

	svc := NewMetricsService(mock, "lifts-table", "data-table", nil)
	resp, err := svc.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetMetrics error: %v", err)
	}

	if resp.Human.Focus != "" {
		t.Errorf("focus = %q, want empty", resp.Human.Focus)
	}
}
