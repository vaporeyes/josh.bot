// ABOUTME: This file contains tests for the DynamoDB-backed BotService.
// ABOUTME: It uses a mock DynamoDBClient to test without hitting real DynamoDB.
package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

// mockDynamoDBClient implements DynamoDBClient for testing.
type mockDynamoDBClient struct {
	getOutput    *dynamodb.GetItemOutput
	getErr       error
	updateOutput *dynamodb.UpdateItemOutput
	updateErr    error
	updateInput  *dynamodb.UpdateItemInput
	queryOutput  *dynamodb.QueryOutput
	queryErr     error
	putOutput    *dynamodb.PutItemOutput
	putErr       error
	putInput     *dynamodb.PutItemInput
	deleteOutput *dynamodb.DeleteItemOutput
	deleteErr    error
	deleteInput  *dynamodb.DeleteItemInput
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.getOutput, m.getErr
}

func (m *mockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	m.updateInput = params
	return m.updateOutput, m.updateErr
}

func (m *mockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return m.queryOutput, m.queryErr
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	m.putInput = params
	return m.putOutput, m.putErr
}

func (m *mockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	m.deleteInput = params
	return m.deleteOutput, m.deleteErr
}

// --- Status Tests ---

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
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetStatus()
	if err == nil {
		t.Error("expected error for missing item, got nil")
	}
}

func TestGetStatus_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{getErr: context.DeadlineExceeded}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetStatus()
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestUpdateStatus_Success(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(map[string]any{
		"current_activity": "Deploying josh.bot",
		"availability":     "Heads down",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called")
	}
	if *mock.updateInput.TableName != "josh-bot-data" {
		t.Errorf("expected table 'josh-bot-data', got '%s'", *mock.updateInput.TableName)
	}
	expr := *mock.updateInput.UpdateExpression
	if len(expr) == 0 {
		t.Error("expected non-empty update expression")
	}
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
	err := svc.UpdateStatus(map[string]any{"hacker_field": "nope"})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestUpdateStatus_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{updateErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(map[string]any{"status": "busy"})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

// --- Project Tests ---

func TestGetProjects_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":          &types.AttributeValueMemberS{Value: "project#modular-aws-backend"},
					"slug":        &types.AttributeValueMemberS{Value: "modular-aws-backend"},
					"name":        &types.AttributeValueMemberS{Value: "Modular AWS Backend"},
					"stack":       &types.AttributeValueMemberS{Value: "Go, AWS"},
					"description": &types.AttributeValueMemberS{Value: "Read-only S3/DynamoDB access."},
					"url":         &types.AttributeValueMemberS{Value: "https://github.com/vaporeyes/josh-bot"},
					"status":      &types.AttributeValueMemberS{Value: "active"},
				},
				{
					"id":          &types.AttributeValueMemberS{Value: "project#modernist-cookbot"},
					"slug":        &types.AttributeValueMemberS{Value: "modernist-cookbot"},
					"name":        &types.AttributeValueMemberS{Value: "Modernist Cookbot"},
					"stack":       &types.AttributeValueMemberS{Value: "Python, Anthropic"},
					"description": &types.AttributeValueMemberS{Value: "AI sous-chef for sous-vide."},
					"url":         &types.AttributeValueMemberS{Value: "https://github.com/vaporeyes/cookbot"},
					"status":      &types.AttributeValueMemberS{Value: "active"},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	projects, err := svc.GetProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].Slug != "modular-aws-backend" {
		t.Errorf("expected slug 'modular-aws-backend', got '%s'", projects[0].Slug)
	}
	if projects[1].Name != "Modernist Cookbot" {
		t.Errorf("expected name 'Modernist Cookbot', got '%s'", projects[1].Name)
	}
}

func TestGetProjects_Empty(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}},
	}

	svc := NewBotService(mock, "josh-bot-data")
	projects, err := svc.GetProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestGetProject_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":          &types.AttributeValueMemberS{Value: "project#modular-aws-backend"},
				"slug":        &types.AttributeValueMemberS{Value: "modular-aws-backend"},
				"name":        &types.AttributeValueMemberS{Value: "Modular AWS Backend"},
				"stack":       &types.AttributeValueMemberS{Value: "Go, AWS"},
				"description": &types.AttributeValueMemberS{Value: "Read-only S3/DynamoDB access."},
				"url":         &types.AttributeValueMemberS{Value: "https://github.com/vaporeyes/josh-bot"},
				"status":      &types.AttributeValueMemberS{Value: "active"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	project, err := svc.GetProject("modular-aws-backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.Name != "Modular AWS Backend" {
		t.Errorf("expected name 'Modular AWS Backend', got '%s'", project.Name)
	}
	if project.URL != "https://github.com/vaporeyes/josh-bot" {
		t.Errorf("expected url, got '%s'", project.URL)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetProject("nonexistent")
	if err == nil {
		t.Error("expected error for missing project, got nil")
	}
}

func TestCreateProject_Success(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateProject(domain.Project{
		Slug:        "new-project",
		Name:        "New Project",
		Stack:       "Go",
		Description: "A new thing",
		URL:         "https://github.com/vaporeyes/new",
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putInput == nil {
		t.Fatal("expected PutItem to be called")
	}
	if *mock.putInput.TableName != "josh-bot-data" {
		t.Errorf("expected table 'josh-bot-data', got '%s'", *mock.putInput.TableName)
	}
	// Verify the item has the correct id key
	idAttr, ok := mock.putInput.Item["id"]
	if !ok {
		t.Fatal("expected 'id' in put item")
	}
	if idAttr.(*types.AttributeValueMemberS).Value != "project#new-project" {
		t.Errorf("expected id 'project#new-project', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestCreateProject_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{putErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateProject(domain.Project{Slug: "test", Name: "Test"})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestUpdateProject_Success(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateProject("modular-aws-backend", map[string]any{
		"status": "archived",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called")
	}
	// Verify key uses project# prefix
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "project#modular-aws-backend" {
		t.Errorf("expected key 'project#modular-aws-backend', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestUpdateProject_InvalidField(t *testing.T) {
	mock := &mockDynamoDBClient{}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateProject("test", map[string]any{"hacker": "nope"})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestDeleteProject_Success(t *testing.T) {
	mock := &mockDynamoDBClient{deleteOutput: &dynamodb.DeleteItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteProject("modular-aws-backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.deleteInput == nil {
		t.Fatal("expected DeleteItem to be called")
	}
	idAttr := mock.deleteInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "project#modular-aws-backend" {
		t.Errorf("expected key 'project#modular-aws-backend', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestDeleteProject_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{deleteErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteProject("test")
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}
