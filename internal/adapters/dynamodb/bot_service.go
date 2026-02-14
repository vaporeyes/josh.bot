// ABOUTME: This file implements a DynamoDB-backed BotService.
// ABOUTME: It provides CRUD for status and projects using a single-table design.
package dynamodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

// DynamoDBClient is the interface for DynamoDB operations used by this adapter.
type DynamoDBClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// allowedStatusFields defines which status fields can be updated via PUT.
var allowedStatusFields = map[string]bool{
	"name": true, "title": true, "bio": true,
	"current_activity": true, "location": true,
	"availability": true, "status": true,
	"links": true, "interests": true,
}

// allowedProjectFields defines which project fields can be updated via PUT.
var allowedProjectFields = map[string]bool{
	"name": true, "stack": true, "description": true,
	"url": true, "status": true,
}

// BotService implements domain.BotService using DynamoDB.
type BotService struct {
	client    DynamoDBClient
	tableName string
}

// NewBotService creates a DynamoDB-backed BotService.
func NewBotService(client DynamoDBClient, tableName string) *BotService {
	return &BotService{client: client, tableName: tableName}
}

// --- Status Operations ---

// GetStatus fetches the status item from DynamoDB.
func (s *BotService) GetStatus() (domain.Status, error) {
	output, err := s.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "status"},
		},
	})
	if err != nil {
		return domain.Status{}, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if output.Item == nil {
		return domain.Status{}, fmt.Errorf("status item not found in table %s", s.tableName)
	}

	var status domain.Status
	if err := attributevalue.UnmarshalMap(output.Item, &status); err != nil {
		return domain.Status{}, fmt.Errorf("unmarshal status: %w", err)
	}

	return status, nil
}

// UpdateStatus updates specific fields on the status item in DynamoDB.
// Only fields in the allowlist are accepted. updated_at is set automatically.
func (s *BotService) UpdateStatus(fields map[string]any) error {
	if len(fields) == 0 {
		return fmt.Errorf("no fields provided for update")
	}

	for key := range fields {
		if !allowedStatusFields[key] {
			return fmt.Errorf("field %q is not an updatable status field", key)
		}
	}

	return s.updateItem("status", fields)
}

// --- Project Operations ---

// GetProjects fetches all projects from DynamoDB using a Scan with a filter.
// AIDEV-NOTE: Uses Scan instead of Query because the table has no sort key.
func (s *BotService) GetProjects() ([]domain.Project, error) {
	filterExpr := "begins_with(id, :prefix)"
	output, err := s.client.Scan(context.Background(), &dynamodb.ScanInput{
		TableName:        &s.tableName,
		FilterExpression: &filterExpr,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":prefix": &types.AttributeValueMemberS{Value: "project#"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb Scan: %w", err)
	}

	projects := make([]domain.Project, 0, len(output.Items))
	for _, item := range output.Items {
		var p domain.Project
		if err := attributevalue.UnmarshalMap(item, &p); err != nil {
			return nil, fmt.Errorf("unmarshal project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// GetProject fetches a single project by slug from DynamoDB.
func (s *BotService) GetProject(slug string) (domain.Project, error) {
	output, err := s.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "project#" + slug},
		},
	})
	if err != nil {
		return domain.Project{}, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if output.Item == nil {
		return domain.Project{}, fmt.Errorf("project %q not found", slug)
	}

	var project domain.Project
	if err := attributevalue.UnmarshalMap(output.Item, &project); err != nil {
		return domain.Project{}, fmt.Errorf("unmarshal project: %w", err)
	}

	return project, nil
}

// CreateProject adds a new project to DynamoDB.
func (s *BotService) CreateProject(project domain.Project) error {
	project.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	item, err := attributevalue.MarshalMap(project)
	if err != nil {
		return fmt.Errorf("marshal project: %w", err)
	}
	// Set the partition key using the slug
	item["id"] = &types.AttributeValueMemberS{Value: "project#" + project.Slug}

	_, err = s.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("dynamodb PutItem: %w", err)
	}

	return nil
}

// UpdateProject updates specific fields on a project in DynamoDB.
// Only fields in the allowlist are accepted. updated_at is set automatically.
func (s *BotService) UpdateProject(slug string, fields map[string]any) error {
	if len(fields) == 0 {
		return fmt.Errorf("no fields provided for update")
	}

	for key := range fields {
		if !allowedProjectFields[key] {
			return fmt.Errorf("field %q is not an updatable project field", key)
		}
	}

	return s.updateItem("project#"+slug, fields)
}

// DeleteProject removes a project from DynamoDB.
func (s *BotService) DeleteProject(slug string) error {
	_, err := s.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "project#" + slug},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb DeleteItem: %w", err)
	}
	return nil
}

// --- Shared Helpers ---

// updateItem builds and executes a DynamoDB UpdateItem with SET expression.
// Automatically adds updated_at timestamp.
func (s *BotService) updateItem(id string, fields map[string]any) error {
	fields["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	var setParts []string
	exprNames := make(map[string]string)
	exprValues := make(map[string]types.AttributeValue)

	for key, val := range fields {
		placeholder := "#" + key
		valuePlaceholder := ":" + key
		setParts = append(setParts, fmt.Sprintf("%s = %s", placeholder, valuePlaceholder))
		exprNames[placeholder] = key

		av, err := attributevalue.Marshal(val)
		if err != nil {
			return fmt.Errorf("marshal field %q: %w", key, err)
		}
		exprValues[valuePlaceholder] = av
	}

	updateExpr := "SET " + strings.Join(setParts, ", ")

	_, err := s.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          &updateExpr,
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		return fmt.Errorf("dynamodb UpdateItem: %w", err)
	}

	return nil
}
