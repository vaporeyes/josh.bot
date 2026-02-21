// ABOUTME: This file contains tests for the DynamoDB-backed BotService.
// ABOUTME: It uses a mock DynamoDBClient to test without hitting real DynamoDB.
package dynamodb

import (
	"context"
	"errors"
	"strings"
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
	scanOutput   *dynamodb.ScanOutput
	scanErr      error
	scanInput    *dynamodb.ScanInput
	queryOutput  *dynamodb.QueryOutput
	queryErr     error
	queryInput   *dynamodb.QueryInput
	putOutput    *dynamodb.PutItemOutput
	putErr       error
	putInput     *dynamodb.PutItemInput
	deleteOutput *dynamodb.DeleteItemOutput
	deleteErr    error
	deleteInput  *dynamodb.DeleteItemInput

	// AIDEV-NOTE: Multi-page support for pagination tests. When set, these take priority over single outputs.
	scanOutputs  []*dynamodb.ScanOutput
	scanCallNum  int
	queryOutputs []*dynamodb.QueryOutput
	queryCallNum int
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.getOutput, m.getErr
}

func (m *mockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	m.updateInput = params
	return m.updateOutput, m.updateErr
}

func (m *mockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	m.scanInput = params
	if len(m.scanOutputs) > 0 {
		idx := m.scanCallNum
		m.scanCallNum++
		if idx < len(m.scanOutputs) {
			return m.scanOutputs[idx], m.scanErr
		}
	}
	return m.scanOutput, m.scanErr
}

func (m *mockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	m.queryInput = params
	if len(m.queryOutputs) > 0 {
		idx := m.queryCallNum
		m.queryCallNum++
		if idx < len(m.queryOutputs) {
			return m.queryOutputs[idx], m.queryErr
		}
	}
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
	status, err := svc.GetStatus(context.Background())
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
	_, err := svc.GetStatus(context.Background())
	if err == nil {
		t.Error("expected error for missing item, got nil")
	}
}

func TestGetStatus_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{getErr: context.DeadlineExceeded}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetStatus(context.Background())
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestUpdateStatus_Success(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(context.Background(), map[string]any{
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
	err := svc.UpdateStatus(context.Background(), map[string]any{})
	if err == nil {
		t.Error("expected error for empty fields, got nil")
	}
}

func TestUpdateStatus_InvalidField(t *testing.T) {
	mock := &mockDynamoDBClient{}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(context.Background(), map[string]any{"hacker_field": "nope"})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestUpdateStatus_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{updateErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateStatus(context.Background(), map[string]any{"status": "busy"})
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
					"item_type":   &types.AttributeValueMemberS{Value: "project"},
				},
				{
					"id":          &types.AttributeValueMemberS{Value: "project#modernist-cookbot"},
					"slug":        &types.AttributeValueMemberS{Value: "modernist-cookbot"},
					"name":        &types.AttributeValueMemberS{Value: "Modernist Cookbot"},
					"stack":       &types.AttributeValueMemberS{Value: "Python, Anthropic"},
					"description": &types.AttributeValueMemberS{Value: "AI sous-chef for sous-vide."},
					"url":         &types.AttributeValueMemberS{Value: "https://github.com/vaporeyes/cookbot"},
					"status":      &types.AttributeValueMemberS{Value: "active"},
					"item_type":   &types.AttributeValueMemberS{Value: "project"},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	projects, err := svc.GetProjects(context.Background())
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
	// Verify Query was used (not Scan) on item-type-index GSI
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	if mock.scanInput != nil {
		t.Error("expected Scan NOT to be called")
	}
	if *mock.queryInput.IndexName != "item-type-index" {
		t.Errorf("expected index 'item-type-index', got '%s'", *mock.queryInput.IndexName)
	}
}

func TestGetProjects_Empty(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}},
	}

	svc := NewBotService(mock, "josh-bot-data")
	projects, err := svc.GetProjects(context.Background())
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
	project, err := svc.GetProject(context.Background(), "modular-aws-backend")
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
	_, err := svc.GetProject(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for missing project, got nil")
	}
}

func TestCreateProject_Success(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateProject(context.Background(), domain.Project{
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
	// Verify item_type is set for GSI
	typeAttr, ok := mock.putInput.Item["item_type"]
	if !ok {
		t.Fatal("expected 'item_type' in put item")
	}
	if typeAttr.(*types.AttributeValueMemberS).Value != "project" {
		t.Errorf("expected item_type 'project', got '%s'", typeAttr.(*types.AttributeValueMemberS).Value)
	}
	// Verify created_at is set
	if _, ok := mock.putInput.Item["created_at"]; !ok {
		t.Fatal("expected 'created_at' in put item")
	}
}

func TestCreateProject_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{putErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateProject(context.Background(), domain.Project{Slug: "test", Name: "Test"})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestUpdateProject_Success(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateProject(context.Background(), "modular-aws-backend", map[string]any{
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
	err := svc.UpdateProject(context.Background(), "test", map[string]any{"hacker": "nope"})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestDeleteProject_SoftDelete(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteProject(context.Background(), "modular-aws-backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// AIDEV-NOTE: Soft delete uses UpdateItem to set deleted_at, not DeleteItem.
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called for soft delete")
	}
	if mock.deleteInput != nil {
		t.Error("expected DeleteItem NOT to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "project#modular-aws-backend" {
		t.Errorf("expected key 'project#modular-aws-backend', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
	expr := *mock.updateInput.UpdateExpression
	if !strings.Contains(expr, "deleted_at") {
		t.Errorf("expected update expression to set deleted_at, got '%s'", expr)
	}
}

func TestDeleteProject_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{updateErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteProject(context.Background(), "test")
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestGetProject_SoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":         &types.AttributeValueMemberS{Value: "project#old-project"},
				"slug":       &types.AttributeValueMemberS{Value: "old-project"},
				"name":       &types.AttributeValueMemberS{Value: "Old Project"},
				"deleted_at": &types.AttributeValueMemberS{Value: "2026-02-21T00:00:00Z"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetProject(context.Background(), "old-project")
	if err == nil {
		t.Fatal("expected NotFoundError for soft-deleted project, got nil")
	}
	var nfe *domain.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestGetProjects_ExcludesSoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "project#active"},
					"slug":      &types.AttributeValueMemberS{Value: "active"},
					"name":      &types.AttributeValueMemberS{Value: "Active Project"},
					"item_type": &types.AttributeValueMemberS{Value: "project"},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify the query includes a filter to exclude soft-deleted items
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	if mock.queryInput.FilterExpression == nil {
		t.Fatal("expected FilterExpression to exclude soft-deleted items")
	}
	if !strings.Contains(*mock.queryInput.FilterExpression, "attribute_not_exists(deleted_at)") {
		t.Errorf("expected filter 'attribute_not_exists(deleted_at)', got '%s'", *mock.queryInput.FilterExpression)
	}
}

// --- Link Tests ---

func TestGetLinks_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "link#a1b2c3d4e5f6"},
					"url":       &types.AttributeValueMemberS{Value: "https://go.dev/blog/"},
					"title":     &types.AttributeValueMemberS{Value: "The Go Blog"},
					"item_type": &types.AttributeValueMemberS{Value: "link"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "go"},
						&types.AttributeValueMemberS{Value: "programming"},
					}},
				},
				{
					"id":        &types.AttributeValueMemberS{Value: "link#b2c3d4e5f6a1"},
					"url":       &types.AttributeValueMemberS{Value: "https://aws.amazon.com/dynamodb/"},
					"title":     &types.AttributeValueMemberS{Value: "Amazon DynamoDB"},
					"item_type": &types.AttributeValueMemberS{Value: "link"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "aws"},
					}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	links, err := svc.GetLinks(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0].Title != "The Go Blog" {
		t.Errorf("expected title 'The Go Blog', got '%s'", links[0].Title)
	}
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	if *mock.queryInput.IndexName != "item-type-index" {
		t.Errorf("expected index 'item-type-index', got '%s'", *mock.queryInput.IndexName)
	}
}

func TestGetLinks_FilterByTag(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "link#b2c3d4e5f6a1"},
					"url":       &types.AttributeValueMemberS{Value: "https://aws.amazon.com/dynamodb/"},
					"title":     &types.AttributeValueMemberS{Value: "Amazon DynamoDB"},
					"item_type": &types.AttributeValueMemberS{Value: "link"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "aws"},
					}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	links, err := svc.GetLinks(context.Background(), "aws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	// Verify Query with tag filter
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	filterExpr := *mock.queryInput.FilterExpression
	if !strings.Contains(filterExpr, "contains") {
		t.Errorf("expected filter expression to contain 'contains', got '%s'", filterExpr)
	}
}

func TestGetLinks_Empty(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}},
	}

	svc := NewBotService(mock, "josh-bot-data")
	links, err := svc.GetLinks(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}

func TestGetLink_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "link#a1b2c3d4e5f6"},
				"url":   &types.AttributeValueMemberS{Value: "https://go.dev/blog/"},
				"title": &types.AttributeValueMemberS{Value: "The Go Blog"},
				"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "go"},
				}},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	link, err := svc.GetLink(context.Background(), "a1b2c3d4e5f6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link.Title != "The Go Blog" {
		t.Errorf("expected title 'The Go Blog', got '%s'", link.Title)
	}
	if link.URL != "https://go.dev/blog/" {
		t.Errorf("expected url, got '%s'", link.URL)
	}
}

func TestGetLink_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetLink(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for missing link, got nil")
	}
}

func TestCreateLink_Success(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateLink(context.Background(), domain.Link{
		URL:   "https://go.dev/blog/",
		Title: "The Go Blog",
		Tags:  []string{"go", "programming"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putInput == nil {
		t.Fatal("expected PutItem to be called")
	}
	// Verify the item has the correct link# prefixed id
	idAttr, ok := mock.putInput.Item["id"]
	if !ok {
		t.Fatal("expected 'id' in put item")
	}
	idVal := idAttr.(*types.AttributeValueMemberS).Value
	expectedID := "link#" + domain.LinkIDFromURL("https://go.dev/blog/")
	if idVal != expectedID {
		t.Errorf("expected id '%s', got '%s'", expectedID, idVal)
	}
	// Verify item_type is set for GSI
	typeAttr, ok := mock.putInput.Item["item_type"]
	if !ok {
		t.Fatal("expected 'item_type' in put item")
	}
	if typeAttr.(*types.AttributeValueMemberS).Value != "link" {
		t.Errorf("expected item_type 'link', got '%s'", typeAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestCreateLink_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{putErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateLink(context.Background(), domain.Link{URL: "https://example.com", Title: "Test"})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestUpdateLink_Success(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateLink(context.Background(), "a1b2c3d4e5f6", map[string]any{
		"title": "Updated Title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "link#a1b2c3d4e5f6" {
		t.Errorf("expected key 'link#a1b2c3d4e5f6', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestUpdateLink_InvalidField(t *testing.T) {
	mock := &mockDynamoDBClient{}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateLink(context.Background(), "test", map[string]any{"hacker": "nope"})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestDeleteLink_SoftDelete(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteLink(context.Background(), "a1b2c3d4e5f6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called for soft delete")
	}
	if mock.deleteInput != nil {
		t.Error("expected DeleteItem NOT to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "link#a1b2c3d4e5f6" {
		t.Errorf("expected key 'link#a1b2c3d4e5f6', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
	expr := *mock.updateInput.UpdateExpression
	if !strings.Contains(expr, "deleted_at") {
		t.Errorf("expected update expression to set deleted_at, got '%s'", expr)
	}
}

func TestDeleteLink_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{updateErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteLink(context.Background(), "test")
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestGetLink_SoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":         &types.AttributeValueMemberS{Value: "link#a1b2c3d4e5f6"},
				"url":        &types.AttributeValueMemberS{Value: "https://example.com"},
				"deleted_at": &types.AttributeValueMemberS{Value: "2026-02-21T00:00:00Z"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetLink(context.Background(), "a1b2c3d4e5f6")
	if err == nil {
		t.Fatal("expected NotFoundError for soft-deleted link, got nil")
	}
	var nfe *domain.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestGetLinks_ExcludesSoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetLinks(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.queryInput.FilterExpression == nil {
		t.Fatal("expected FilterExpression to exclude soft-deleted items")
	}
	if !strings.Contains(*mock.queryInput.FilterExpression, "attribute_not_exists(deleted_at)") {
		t.Errorf("expected filter 'attribute_not_exists(deleted_at)', got '%s'", *mock.queryInput.FilterExpression)
	}
}

// --- Note Tests ---

func TestGetNotes_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "note#abc123"},
					"title":     &types.AttributeValueMemberS{Value: "Meeting notes"},
					"body":      &types.AttributeValueMemberS{Value: "Discussed API design"},
					"item_type": &types.AttributeValueMemberS{Value: "note"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "work"},
					}},
				},
				{
					"id":        &types.AttributeValueMemberS{Value: "note#def456"},
					"title":     &types.AttributeValueMemberS{Value: "Grocery list"},
					"body":      &types.AttributeValueMemberS{Value: "Eggs, milk, bread"},
					"item_type": &types.AttributeValueMemberS{Value: "note"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "personal"},
					}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	notes, err := svc.GetNotes(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(notes))
	}
	if notes[0].Title != "Meeting notes" {
		t.Errorf("expected title 'Meeting notes', got '%s'", notes[0].Title)
	}
}

func TestGetNotes_FilterByTag(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "note#abc123"},
					"title":     &types.AttributeValueMemberS{Value: "Meeting notes"},
					"body":      &types.AttributeValueMemberS{Value: "Discussed API design"},
					"item_type": &types.AttributeValueMemberS{Value: "note"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "work"},
					}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	notes, err := svc.GetNotes(context.Background(), "work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	filterExpr := *mock.queryInput.FilterExpression
	if !strings.Contains(filterExpr, "contains") {
		t.Errorf("expected filter expression to contain 'contains', got '%s'", filterExpr)
	}
}

func TestGetNote_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "note#abc123"},
				"title": &types.AttributeValueMemberS{Value: "Meeting notes"},
				"body":  &types.AttributeValueMemberS{Value: "Discussed API design"},
				"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "work"},
				}},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	note, err := svc.GetNote(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.Title != "Meeting notes" {
		t.Errorf("expected title 'Meeting notes', got '%s'", note.Title)
	}
	if note.Body != "Discussed API design" {
		t.Errorf("expected body 'Discussed API design', got '%s'", note.Body)
	}
}

func TestGetNote_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetNote(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for missing note, got nil")
	}
}

func TestCreateNote_Success(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateNote(context.Background(), domain.Note{
		Title: "Meeting notes",
		Body:  "Discussed API design",
		Tags:  []string{"work"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putInput == nil {
		t.Fatal("expected PutItem to be called")
	}
	idAttr, ok := mock.putInput.Item["id"]
	if !ok {
		t.Fatal("expected 'id' in put item")
	}
	idVal := idAttr.(*types.AttributeValueMemberS).Value
	if !strings.HasPrefix(idVal, "note#") {
		t.Errorf("expected id to start with 'note#', got '%s'", idVal)
	}
	// Verify item_type is set for GSI
	typeAttr, ok := mock.putInput.Item["item_type"]
	if !ok {
		t.Fatal("expected 'item_type' in put item")
	}
	if typeAttr.(*types.AttributeValueMemberS).Value != "note" {
		t.Errorf("expected item_type 'note', got '%s'", typeAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestCreateNote_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{putErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateNote(context.Background(), domain.Note{Title: "Test", Body: "body"})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestUpdateNote_Success(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateNote(context.Background(), "abc123", map[string]any{
		"title": "Updated Title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "note#abc123" {
		t.Errorf("expected key 'note#abc123', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestUpdateNote_InvalidField(t *testing.T) {
	mock := &mockDynamoDBClient{}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateNote(context.Background(), "test", map[string]any{"hacker": "nope"})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestDeleteNote_SoftDelete(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteNote(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called for soft delete")
	}
	if mock.deleteInput != nil {
		t.Error("expected DeleteItem NOT to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "note#abc123" {
		t.Errorf("expected key 'note#abc123', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
	expr := *mock.updateInput.UpdateExpression
	if !strings.Contains(expr, "deleted_at") {
		t.Errorf("expected update expression to set deleted_at, got '%s'", expr)
	}
}

func TestDeleteNote_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{updateErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteNote(context.Background(), "test")
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestGetNote_SoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":         &types.AttributeValueMemberS{Value: "note#abc123"},
				"title":      &types.AttributeValueMemberS{Value: "Old Note"},
				"body":       &types.AttributeValueMemberS{Value: "Deleted content"},
				"deleted_at": &types.AttributeValueMemberS{Value: "2026-02-21T00:00:00Z"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetNote(context.Background(), "abc123")
	if err == nil {
		t.Fatal("expected NotFoundError for soft-deleted note, got nil")
	}
	var nfe *domain.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestGetNotes_ExcludesSoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetNotes(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.queryInput.FilterExpression == nil {
		t.Fatal("expected FilterExpression to exclude soft-deleted items")
	}
	if !strings.Contains(*mock.queryInput.FilterExpression, "attribute_not_exists(deleted_at)") {
		t.Errorf("expected filter 'attribute_not_exists(deleted_at)', got '%s'", *mock.queryInput.FilterExpression)
	}
}

// --- TIL Tests ---

func TestGetTILs_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "til#abc123"},
					"title":     &types.AttributeValueMemberS{Value: "Go slices grow by 2x"},
					"body":      &types.AttributeValueMemberS{Value: "When a slice exceeds capacity, Go doubles it"},
					"item_type": &types.AttributeValueMemberS{Value: "til"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "go"},
					}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	tils, err := svc.GetTILs(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tils) != 1 {
		t.Fatalf("expected 1 til, got %d", len(tils))
	}
	if tils[0].Title != "Go slices grow by 2x" {
		t.Errorf("expected title 'Go slices grow by 2x', got '%s'", tils[0].Title)
	}
}

func TestGetTILs_FilterByTag(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "til#abc123"},
					"title":     &types.AttributeValueMemberS{Value: "Go slices grow by 2x"},
					"body":      &types.AttributeValueMemberS{Value: "Capacity doubles"},
					"item_type": &types.AttributeValueMemberS{Value: "til"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "go"},
					}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	tils, err := svc.GetTILs(context.Background(), "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tils) != 1 {
		t.Fatalf("expected 1 til, got %d", len(tils))
	}
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	filterExpr := *mock.queryInput.FilterExpression
	if !strings.Contains(filterExpr, "contains") {
		t.Errorf("expected filter expression to contain 'contains', got '%s'", filterExpr)
	}
}

func TestGetTIL_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":    &types.AttributeValueMemberS{Value: "til#abc123"},
				"title": &types.AttributeValueMemberS{Value: "Go slices grow by 2x"},
				"body":  &types.AttributeValueMemberS{Value: "Capacity doubles"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	til, err := svc.GetTIL(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if til.Title != "Go slices grow by 2x" {
		t.Errorf("expected title 'Go slices grow by 2x', got '%s'", til.Title)
	}
}

func TestGetTIL_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetTIL(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for missing til, got nil")
	}
}

func TestCreateTIL_Success(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateTIL(context.Background(), domain.TIL{
		Title: "Go slices grow by 2x",
		Body:  "Capacity doubles",
		Tags:  []string{"go"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putInput == nil {
		t.Fatal("expected PutItem to be called")
	}
	idAttr, ok := mock.putInput.Item["id"]
	if !ok {
		t.Fatal("expected 'id' in put item")
	}
	idVal := idAttr.(*types.AttributeValueMemberS).Value
	if !strings.HasPrefix(idVal, "til#") {
		t.Errorf("expected id to start with 'til#', got '%s'", idVal)
	}
	// Verify item_type is set for GSI
	typeAttr, ok := mock.putInput.Item["item_type"]
	if !ok {
		t.Fatal("expected 'item_type' in put item")
	}
	if typeAttr.(*types.AttributeValueMemberS).Value != "til" {
		t.Errorf("expected item_type 'til', got '%s'", typeAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestCreateTIL_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{putErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateTIL(context.Background(), domain.TIL{Title: "Test", Body: "body"})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestUpdateTIL_Success(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateTIL(context.Background(), "abc123", map[string]any{
		"title": "Updated Title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "til#abc123" {
		t.Errorf("expected key 'til#abc123', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestUpdateTIL_InvalidField(t *testing.T) {
	mock := &mockDynamoDBClient{}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateTIL(context.Background(), "test", map[string]any{"hacker": "nope"})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestDeleteTIL_SoftDelete(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteTIL(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called for soft delete")
	}
	if mock.deleteInput != nil {
		t.Error("expected DeleteItem NOT to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "til#abc123" {
		t.Errorf("expected key 'til#abc123', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
	expr := *mock.updateInput.UpdateExpression
	if !strings.Contains(expr, "deleted_at") {
		t.Errorf("expected update expression to set deleted_at, got '%s'", expr)
	}
}

func TestDeleteTIL_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{updateErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteTIL(context.Background(), "test")
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestGetTIL_SoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":         &types.AttributeValueMemberS{Value: "til#abc123"},
				"title":      &types.AttributeValueMemberS{Value: "Old TIL"},
				"body":       &types.AttributeValueMemberS{Value: "Deleted content"},
				"deleted_at": &types.AttributeValueMemberS{Value: "2026-02-21T00:00:00Z"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetTIL(context.Background(), "abc123")
	if err == nil {
		t.Fatal("expected NotFoundError for soft-deleted TIL, got nil")
	}
	var nfe *domain.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- Log Entry Tests ---

func TestGetLogEntries_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "log#abc123"},
					"message":   &types.AttributeValueMemberS{Value: "deployed josh-bot v1.2"},
					"item_type": &types.AttributeValueMemberS{Value: "log"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "deploy"},
					}},
				},
				{
					"id":        &types.AttributeValueMemberS{Value: "log#def456"},
					"message":   &types.AttributeValueMemberS{Value: "updated DNS for josh.bot"},
					"item_type": &types.AttributeValueMemberS{Value: "log"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "infra"},
					}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	entries, err := svc.GetLogEntries(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Message != "deployed josh-bot v1.2" {
		t.Errorf("expected message 'deployed josh-bot v1.2', got '%s'", entries[0].Message)
	}
}

func TestGetLogEntries_FilterByTag(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"id":        &types.AttributeValueMemberS{Value: "log#abc123"},
					"message":   &types.AttributeValueMemberS{Value: "deployed josh-bot v1.2"},
					"item_type": &types.AttributeValueMemberS{Value: "log"},
					"tags": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: "deploy"},
					}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	entries, err := svc.GetLogEntries(context.Background(), "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if mock.queryInput == nil {
		t.Fatal("expected Query to be called")
	}
	filterExpr := *mock.queryInput.FilterExpression
	if !strings.Contains(filterExpr, "contains") {
		t.Errorf("expected filter expression to contain 'contains', got '%s'", filterExpr)
	}
}

func TestGetLogEntry_Success(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":      &types.AttributeValueMemberS{Value: "log#abc123"},
				"message": &types.AttributeValueMemberS{Value: "deployed josh-bot v1.2"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	entry, err := svc.GetLogEntry(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Message != "deployed josh-bot v1.2" {
		t.Errorf("expected message 'deployed josh-bot v1.2', got '%s'", entry.Message)
	}
}

func TestGetLogEntry_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetLogEntry(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for missing log entry, got nil")
	}
}

func TestCreateLogEntry_Success(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateLogEntry(context.Background(), domain.LogEntry{
		Message: "deployed josh-bot v1.2",
		Tags:    []string{"deploy"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putInput == nil {
		t.Fatal("expected PutItem to be called")
	}
	idAttr, ok := mock.putInput.Item["id"]
	if !ok {
		t.Fatal("expected 'id' in put item")
	}
	idVal := idAttr.(*types.AttributeValueMemberS).Value
	if !strings.HasPrefix(idVal, "log#") {
		t.Errorf("expected id to start with 'log#', got '%s'", idVal)
	}
	// Verify item_type is set for GSI
	typeAttr, ok := mock.putInput.Item["item_type"]
	if !ok {
		t.Fatal("expected 'item_type' in put item")
	}
	if typeAttr.(*types.AttributeValueMemberS).Value != "log" {
		t.Errorf("expected item_type 'log', got '%s'", typeAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestCreateLogEntry_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{putErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateLogEntry(context.Background(), domain.LogEntry{Message: "test"})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestUpdateLogEntry_Success(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateLogEntry(context.Background(), "abc123", map[string]any{
		"message": "deployed josh-bot v1.3",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "log#abc123" {
		t.Errorf("expected key 'log#abc123', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestUpdateLogEntry_InvalidField(t *testing.T) {
	mock := &mockDynamoDBClient{}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.UpdateLogEntry(context.Background(), "test", map[string]any{"hacker": "nope"})
	if err == nil {
		t.Error("expected error for invalid field, got nil")
	}
}

func TestDeleteLogEntry_SoftDelete(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteLogEntry(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called for soft delete")
	}
	if mock.deleteInput != nil {
		t.Error("expected DeleteItem NOT to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "log#abc123" {
		t.Errorf("expected key 'log#abc123', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
	expr := *mock.updateInput.UpdateExpression
	if !strings.Contains(expr, "deleted_at") {
		t.Errorf("expected update expression to set deleted_at, got '%s'", expr)
	}
}

func TestDeleteLogEntry_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{updateErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteLogEntry(context.Background(), "test")
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestGetLogEntry_SoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":         &types.AttributeValueMemberS{Value: "log#abc123"},
				"message":    &types.AttributeValueMemberS{Value: "old log"},
				"deleted_at": &types.AttributeValueMemberS{Value: "2026-02-21T00:00:00Z"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetLogEntry(context.Background(), "abc123")
	if err == nil {
		t.Fatal("expected NotFoundError for soft-deleted log entry, got nil")
	}
	var nfe *domain.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestCreateDiaryEntry_SetsCreatedAtWhenEmpty(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}
	svc := NewBotService(mock, "josh-bot-data")

	err := svc.CreateDiaryEntry(context.Background(), domain.DiaryEntry{Body: "test entry"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putInput == nil {
		t.Fatal("expected PutItem to be called")
	}

	idAttr, ok := mock.putInput.Item["id"]
	if !ok {
		t.Fatal("expected 'id' in put item")
	}
	idVal := idAttr.(*types.AttributeValueMemberS).Value
	if !strings.HasPrefix(idVal, "diary#") {
		t.Errorf("expected id to start with 'diary#', got '%s'", idVal)
	}

	createdAtAttr, ok := mock.putInput.Item["created_at"]
	if !ok {
		t.Fatal("expected 'created_at' in put item")
	}
	createdAtVal := createdAtAttr.(*types.AttributeValueMemberS).Value
	if createdAtVal == "" {
		t.Error("expected created_at to be non-empty")
	}

	updatedAtAttr, ok := mock.putInput.Item["updated_at"]
	if !ok {
		t.Fatal("expected 'updated_at' in put item")
	}
	updatedAtVal := updatedAtAttr.(*types.AttributeValueMemberS).Value
	if updatedAtVal == "" {
		t.Error("expected updated_at to be non-empty")
	}

	// Verify item_type is set for GSI
	typeAttr, ok := mock.putInput.Item["item_type"]
	if !ok {
		t.Fatal("expected 'item_type' in put item")
	}
	if typeAttr.(*types.AttributeValueMemberS).Value != "diary" {
		t.Errorf("expected item_type 'diary', got '%s'", typeAttr.(*types.AttributeValueMemberS).Value)
	}
}

func TestCreateDiaryEntry_PreservesCallerValues(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}
	svc := NewBotService(mock, "josh-bot-data")

	err := svc.CreateDiaryEntry(context.Background(), domain.DiaryEntry{
		ID:        "diary#caller-set-id",
		Body:      "test entry",
		CreatedAt: "2026-01-01T00:00:00Z",
		UpdatedAt: "2026-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	idVal := mock.putInput.Item["id"].(*types.AttributeValueMemberS).Value
	if idVal != "diary#caller-set-id" {
		t.Errorf("expected caller-set ID to be preserved, got '%s'", idVal)
	}

	createdAtVal := mock.putInput.Item["created_at"].(*types.AttributeValueMemberS).Value
	if createdAtVal != "2026-01-01T00:00:00Z" {
		t.Errorf("expected caller-set created_at to be preserved, got '%s'", createdAtVal)
	}
}

func TestCreateDiaryEntry_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{putErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.CreateDiaryEntry(context.Background(), domain.DiaryEntry{Body: "test"})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestDeleteDiaryEntry_SoftDelete(t *testing.T) {
	mock := &mockDynamoDBClient{updateOutput: &dynamodb.UpdateItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteDiaryEntry(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updateInput == nil {
		t.Fatal("expected UpdateItem to be called for soft delete")
	}
	if mock.deleteInput != nil {
		t.Error("expected DeleteItem NOT to be called")
	}
	idAttr := mock.updateInput.Key["id"]
	if idAttr.(*types.AttributeValueMemberS).Value != "diary#abc123" {
		t.Errorf("expected key 'diary#abc123', got '%s'", idAttr.(*types.AttributeValueMemberS).Value)
	}
	expr := *mock.updateInput.UpdateExpression
	if !strings.Contains(expr, "deleted_at") {
		t.Errorf("expected update expression to set deleted_at, got '%s'", expr)
	}
}

func TestDeleteDiaryEntry_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{updateErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.DeleteDiaryEntry(context.Background(), "test")
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}

func TestGetDiaryEntry_SoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":         &types.AttributeValueMemberS{Value: "diary#abc123"},
				"body":       &types.AttributeValueMemberS{Value: "old entry"},
				"deleted_at": &types.AttributeValueMemberS{Value: "2026-02-21T00:00:00Z"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetDiaryEntry(context.Background(), "abc123")
	if err == nil {
		t.Fatal("expected NotFoundError for soft-deleted diary entry, got nil")
	}
	var nfe *domain.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- Pagination Tests ---

func TestGetProjects_Paginated(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutputs: []*dynamodb.QueryOutput{
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "project#a"}, "slug": &types.AttributeValueMemberS{Value: "a"}, "name": &types.AttributeValueMemberS{Value: "Project A"}, "item_type": &types.AttributeValueMemberS{Value: "project"}},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "project#a"}},
			},
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "project#b"}, "slug": &types.AttributeValueMemberS{Value: "b"}, "name": &types.AttributeValueMemberS{Value: "Project B"}, "item_type": &types.AttributeValueMemberS{Value: "project"}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	projects, err := svc.GetProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects across pages, got %d", len(projects))
	}
	if mock.queryCallNum != 2 {
		t.Errorf("expected 2 query calls, got %d", mock.queryCallNum)
	}
}

func TestGetLinks_Paginated(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutputs: []*dynamodb.QueryOutput{
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "link#aaa"}, "url": &types.AttributeValueMemberS{Value: "https://a.com"}, "title": &types.AttributeValueMemberS{Value: "A"}, "item_type": &types.AttributeValueMemberS{Value: "link"}},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "link#aaa"}},
			},
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "link#bbb"}, "url": &types.AttributeValueMemberS{Value: "https://b.com"}, "title": &types.AttributeValueMemberS{Value: "B"}, "item_type": &types.AttributeValueMemberS{Value: "link"}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	links, err := svc.GetLinks(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links across pages, got %d", len(links))
	}
	if mock.queryCallNum != 2 {
		t.Errorf("expected 2 query calls, got %d", mock.queryCallNum)
	}
}

func TestGetNotes_Paginated(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutputs: []*dynamodb.QueryOutput{
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "note#a"}, "title": &types.AttributeValueMemberS{Value: "Note A"}, "body": &types.AttributeValueMemberS{Value: "body a"}, "item_type": &types.AttributeValueMemberS{Value: "note"}},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "note#a"}},
			},
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "note#b"}, "title": &types.AttributeValueMemberS{Value: "Note B"}, "body": &types.AttributeValueMemberS{Value: "body b"}, "item_type": &types.AttributeValueMemberS{Value: "note"}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	notes, err := svc.GetNotes(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("expected 2 notes across pages, got %d", len(notes))
	}
	if mock.queryCallNum != 2 {
		t.Errorf("expected 2 query calls, got %d", mock.queryCallNum)
	}
}

func TestGetTILs_Paginated(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutputs: []*dynamodb.QueryOutput{
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "til#a"}, "title": &types.AttributeValueMemberS{Value: "TIL A"}, "body": &types.AttributeValueMemberS{Value: "a"}, "item_type": &types.AttributeValueMemberS{Value: "til"}},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "til#a"}},
			},
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "til#b"}, "title": &types.AttributeValueMemberS{Value: "TIL B"}, "body": &types.AttributeValueMemberS{Value: "b"}, "item_type": &types.AttributeValueMemberS{Value: "til"}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	tils, err := svc.GetTILs(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tils) != 2 {
		t.Errorf("expected 2 TILs across pages, got %d", len(tils))
	}
	if mock.queryCallNum != 2 {
		t.Errorf("expected 2 query calls, got %d", mock.queryCallNum)
	}
}

func TestGetLogEntries_Paginated(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutputs: []*dynamodb.QueryOutput{
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "log#a"}, "message": &types.AttributeValueMemberS{Value: "log a"}, "item_type": &types.AttributeValueMemberS{Value: "log"}},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "log#a"}},
			},
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "log#b"}, "message": &types.AttributeValueMemberS{Value: "log b"}, "item_type": &types.AttributeValueMemberS{Value: "log"}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	entries, err := svc.GetLogEntries(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 log entries across pages, got %d", len(entries))
	}
	if mock.queryCallNum != 2 {
		t.Errorf("expected 2 query calls, got %d", mock.queryCallNum)
	}
}

func TestGetDiaryEntries_Paginated(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutputs: []*dynamodb.QueryOutput{
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "diary#a"}, "body": &types.AttributeValueMemberS{Value: "entry a"}, "item_type": &types.AttributeValueMemberS{Value: "diary"}},
				},
				LastEvaluatedKey: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: "diary#a"}},
			},
			{
				Items: []map[string]types.AttributeValue{
					{"id": &types.AttributeValueMemberS{Value: "diary#b"}, "body": &types.AttributeValueMemberS{Value: "entry b"}, "item_type": &types.AttributeValueMemberS{Value: "diary"}},
				},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	entries, err := svc.GetDiaryEntries(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 diary entries across pages, got %d", len(entries))
	}
	if mock.queryCallNum != 2 {
		t.Errorf("expected 2 query calls, got %d", mock.queryCallNum)
	}
}

func TestGetDiaryEntries_ExcludesSoftDeleted(t *testing.T) {
	mock := &mockDynamoDBClient{
		queryOutput: &dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	_, err := svc.GetDiaryEntries(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.queryInput.FilterExpression == nil {
		t.Fatal("expected FilterExpression to exclude soft-deleted items")
	}
	if !strings.Contains(*mock.queryInput.FilterExpression, "attribute_not_exists(deleted_at)") {
		t.Errorf("expected filter 'attribute_not_exists(deleted_at)', got '%s'", *mock.queryInput.FilterExpression)
	}
}

// --- Idempotency Tests ---

func TestGetIdempotencyRecord_Found(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{
			Item: map[string]types.AttributeValue{
				"id":          &types.AttributeValueMemberS{Value: "idem#/v1/notes#abc123"},
				"status_code": &types.AttributeValueMemberN{Value: "201"},
				"body":        &types.AttributeValueMemberS{Value: `{"ok":true}`},
				"expires_at":  &types.AttributeValueMemberN{Value: "1740000000"},
				"created_at":  &types.AttributeValueMemberS{Value: "2026-02-20T10:00:00Z"},
			},
		},
	}

	svc := NewBotService(mock, "josh-bot-data")
	record, err := svc.GetIdempotencyRecord(context.Background(), "idem#/v1/notes#abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record == nil {
		t.Fatal("expected record, got nil")
	}
	if record.StatusCode != 201 {
		t.Errorf("expected status_code 201, got %d", record.StatusCode)
	}
	if record.Body != `{"ok":true}` {
		t.Errorf("expected body '{\"ok\":true}', got '%s'", record.Body)
	}
}

func TestGetIdempotencyRecord_NotFound(t *testing.T) {
	mock := &mockDynamoDBClient{
		getOutput: &dynamodb.GetItemOutput{Item: nil},
	}

	svc := NewBotService(mock, "josh-bot-data")
	record, err := svc.GetIdempotencyRecord(context.Background(), "idem#/v1/notes#nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record != nil {
		t.Errorf("expected nil record for missing key, got %+v", record)
	}
}

func TestSetIdempotencyRecord_Success(t *testing.T) {
	mock := &mockDynamoDBClient{putOutput: &dynamodb.PutItemOutput{}}

	svc := NewBotService(mock, "josh-bot-data")
	err := svc.SetIdempotencyRecord(context.Background(), domain.IdempotencyRecord{
		ID:         "idem#/v1/notes#abc123",
		StatusCode: 201,
		Body:       `{"ok":true}`,
		ExpiresAt:  1740000000,
		CreatedAt:  "2026-02-20T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.putInput == nil {
		t.Fatal("expected PutItem to be called")
	}
	idAttr := mock.putInput.Item["id"].(*types.AttributeValueMemberS).Value
	if idAttr != "idem#/v1/notes#abc123" {
		t.Errorf("expected id 'idem#/v1/notes#abc123', got '%s'", idAttr)
	}
}

func TestSetIdempotencyRecord_DynamoDBError(t *testing.T) {
	mock := &mockDynamoDBClient{putErr: context.DeadlineExceeded}
	svc := NewBotService(mock, "josh-bot-data")
	err := svc.SetIdempotencyRecord(context.Background(), domain.IdempotencyRecord{
		ID:         "idem#/v1/notes#abc123",
		StatusCode: 201,
		Body:       `{"ok":true}`,
		ExpiresAt:  1740000000,
	})
	if err == nil {
		t.Error("expected error from DynamoDB failure, got nil")
	}
}
