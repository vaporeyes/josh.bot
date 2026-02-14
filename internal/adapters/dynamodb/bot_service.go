// ABOUTME: This file implements a DynamoDB-backed BotService.
// ABOUTME: It fetches and updates status data in a DynamoDB table, allowing changes without redeployment.
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
}

// allowedFields defines which status fields can be updated via PUT.
var allowedFields = map[string]bool{
	"name": true, "title": true, "bio": true,
	"current_activity": true, "location": true,
	"availability": true, "status": true,
	"links": true, "interests": true,
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

	// Validate all fields before building the expression
	for key := range fields {
		if !allowedFields[key] {
			return fmt.Errorf("field %q is not an updatable status field", key)
		}
	}

	// Always set updated_at
	fields["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	// Build SET expression: SET #field1 = :field1, #field2 = :field2, ...
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
			"id": &types.AttributeValueMemberS{Value: "status"},
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

// GetProjects returns hardcoded projects for now.
// AIDEV-TODO: move projects to DynamoDB when implementing /v1/projects with real data
func (s *BotService) GetProjects() ([]domain.Project, error) {
	return []domain.Project{
		{Name: "Modular AWS Backend", Stack: "Go, AWS", Description: "Read-only S3/DynamoDB access."},
		{Name: "Modernist Cookbot", Stack: "Python, Anthropic", Description: "AI sous-chef for sous-vide."},
	}, nil
}
