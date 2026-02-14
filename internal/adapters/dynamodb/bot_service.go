// ABOUTME: This file implements a DynamoDB-backed BotService.
// ABOUTME: It fetches status data from a DynamoDB table, allowing updates without redeployment.
package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

// ItemGetter is the minimal interface for DynamoDB GetItem operations.
type ItemGetter interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

// BotService implements domain.BotService using DynamoDB.
type BotService struct {
	client    ItemGetter
	tableName string
}

// NewBotService creates a DynamoDB-backed BotService.
func NewBotService(client ItemGetter, tableName string) *BotService {
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

// GetProjects returns hardcoded projects for now.
// AIDEV-TODO: move projects to DynamoDB when implementing /v1/projects with real data
func (s *BotService) GetProjects() ([]domain.Project, error) {
	return []domain.Project{
		{Name: "Modular AWS Backend", Stack: "Go, AWS", Description: "Read-only S3/DynamoDB access."},
		{Name: "Modernist Cookbot", Stack: "Python, Anthropic", Description: "AI sous-chef for sous-vide."},
	}, nil
}
