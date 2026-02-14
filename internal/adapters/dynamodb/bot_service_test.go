// ABOUTME: This file contains tests for the DynamoDB-backed BotService.
// ABOUTME: It uses a mock ItemGetter to test without hitting real DynamoDB.
package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDynamoDBClient implements DynamoDBClient for testing.
type mockDynamoDBClient struct {
	getOutput    *dynamodb.GetItemOutput
	getErr       error
	updateOutput *dynamodb.UpdateItemOutput
	updateErr    error
	updateInput  *dynamodb.UpdateItemInput // captured for assertions
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.getOutput, m.getErr
}

func (m *mockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	m.updateInput = params
	return m.updateOutput, m.updateErr
}

func TestGetStatus_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
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
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
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
	mock := &mockDynamoDBClient{
		getErr: context.DeadlineExceeded,
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetStatus()
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestGetProjects_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
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

func TestUpdateStatus_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		updateOutput: &dynamodb.UpdateItemOutput{},
	}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(map[string]any{
		"current_activity": "Deploying josh.bot",
		"availability":     "Heads down",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify UpdateItem was called with the right table and key
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called")
	}
	if *mock.updateInput.TableName != "josh-bot-data" {
		t.Errorf("expected table 'josh-bot-data', got '%s'", *mock.updateInput.TableName)
	}
	// Verify the expression contains our fields
	expr := *mock.updateInput.UpdateExpression
	if len(expr) == 0 {
		t.Error("expected non-empty update expression")
	}
	// Verify updated_at is always included
	if _, ok := mock.updateInput.ExpressionAttributeNames["#updated_at"]; !ok {
		t.Error("expected updated_at in expression attribute names")
	}
}

func TestUpdateStatus_EmptyFields(t *testing.T) {
	mock := &mockDynamoDBClient{}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(map[string]any{})
	if err == nil {
		t.Error("expected error for empty fields, got nil")
	}
}

func TestUpdateStatus_InvalidField(t *testing.T) {
	mock := &mockDynamoDBClient{}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(map[string]any{
		"hacker_field": "nope",
	})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestUpdateStatus_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{
		updateErr: context.DeadlineExceeded,
	}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(map[string]any{
		"status": "busy",
	})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}
