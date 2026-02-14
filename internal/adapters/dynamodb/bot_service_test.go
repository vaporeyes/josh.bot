// ABOUTME: This file contains tests for the DynamoDB-backed BotService.
// ABOUTME: It uses a mock ItemGetter to test without hitting real DynamoDB.
package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockItemGetter implements ItemGetter for testing.
type mockItemGetter struct {
	output *dynamodb.GetItemOutput
	err    error
}

func (m *mockItemGetter) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.output, m.err
}

func TestGetStatus_Success(t *testing.T) {
	mock := &mockItemGetter{
		output: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":               &types.AttributeValueMemberS{Value: "status"},
				"name":             &types.AttributeValueMemberS{Value: "Josh Duncan"},
				"title":            &types.AttributeValueMemberS{Value: "Platform Engineer"},
				"bio":              &types.AttributeValueMemberS{Value: "Builder of systems."},
				"current_activity": &types.AttributeValueMemberS{Value: "Building josh.bot"},
				"location":         &types.AttributeValueMemberS{Value: "Clarksville, TN"},
				"availability":     &types.AttributeValueMemberS{Value: "Open to projects"},
				"status":           &types.AttributeValueMemberS{Value: "ok"},
				"links": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
					"github": &types.AttributeValueMemberS{Value: "https://github.com/jduncan"},
				}},
				"interests": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "Go"},
					&types.AttributeValueMemberS{Value: "AWS"},
				}},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	status, err := svc.GetStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Name != "Josh Duncan" {
		t.Errorf("expected name 'Josh Duncan', got '%s'", status.Name)
	}
	if status.Title != "Platform Engineer" {
		t.Errorf("expected title 'Platform Engineer', got '%s'", status.Title)
	}
	if status.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", status.Status)
	}
	if status.Links["github"] != "https://github.com/jduncan" {
		t.Errorf("expected github link, got %v", status.Links)
	}
	if len(status.Interests) != 2 {
		t.Errorf("expected 2 interests, got %d", len(status.Interests))
	}
}

func TestGetStatus_ItemNotFound(t *testing.T) {
	mock := &mockItemGetter{
		output: &dynamodb.GetItemOutput{
			Item: nil,
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetStatus()
	if err == nil {
		t.Error("expected error for missing item, got nil")
	}
}

func TestGetStatus_DynamoDBError(t *testing.T) {
	mock := &mockItemGetter{
		err: context.DeadlineExceeded,
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetStatus()
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestGetProjects_Success(t *testing.T) {
	mock := &mockItemGetter{
		output: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: "status"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	projects, err := svc.GetProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Projects are still hardcoded for now
	if len(projects) < 2 {
		t.Errorf("expected at least 2 projects, got %d", len(projects))
	}
}
