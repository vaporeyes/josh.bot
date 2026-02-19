// ABOUTME: This file implements a DynamoDB-backed WebhookService for storing inbound webhook events.
// ABOUTME: Events are immutable once created; no update or delete operations are provided.
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

// WebhookService implements domain.WebhookService using DynamoDB.
type WebhookService struct {
	client    DynamoDBClient
	tableName string
}

// NewWebhookService creates a DynamoDB-backed WebhookService.
func NewWebhookService(client DynamoDBClient, tableName string) *WebhookService {
	return &WebhookService{client: client, tableName: tableName}
}

// CreateWebhookEvent stores a webhook event in DynamoDB.
// AIDEV-NOTE: ID and CreatedAt are set automatically; events are immutable after creation.
func (s *WebhookService) CreateWebhookEvent(event domain.WebhookEvent) error {
	event.ID = domain.WebhookEventID()
	event.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return fmt.Errorf("marshal webhook event: %w", err)
	}

	_, err = s.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("dynamodb PutItem: %w", err)
	}

	return nil
}

// GetWebhookEvents fetches all webhook events, optionally filtered by type and/or source.
func (s *WebhookService) GetWebhookEvents(eventType, source string) ([]domain.WebhookEvent, error) {
	filterExpr := "begins_with(id, :prefix)"
	exprValues := map[string]types.AttributeValue{
		":prefix": &types.AttributeValueMemberS{Value: "webhook#"},
	}
	// AIDEV-NOTE: "type" is a DynamoDB reserved word, so we alias it with ExpressionAttributeNames.
	exprNames := map[string]string{}

	if eventType != "" {
		filterExpr += " AND #t = :eventType"
		exprValues[":eventType"] = &types.AttributeValueMemberS{Value: eventType}
		exprNames["#t"] = "type"
	}

	if source != "" {
		filterExpr += " AND #s = :source"
		exprValues[":source"] = &types.AttributeValueMemberS{Value: source}
		exprNames["#s"] = "source"
	}

	input := &dynamodb.ScanInput{
		TableName:                 &s.tableName,
		FilterExpression:          &filterExpr,
		ExpressionAttributeValues: exprValues,
	}
	if len(exprNames) > 0 {
		input.ExpressionAttributeNames = exprNames
	}

	output, err := s.client.Scan(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("dynamodb Scan: %w", err)
	}

	events := make([]domain.WebhookEvent, 0, len(output.Items))
	for _, item := range output.Items {
		var e domain.WebhookEvent
		if err := attributevalue.UnmarshalMap(item, &e); err != nil {
			return nil, fmt.Errorf("unmarshal webhook event: %w", err)
		}
		events = append(events, e)
	}

	return events, nil
}

// GetWebhookEvent fetches a single webhook event by ID.
func (s *WebhookService) GetWebhookEvent(id string) (domain.WebhookEvent, error) {
	// Support both full ID ("webhook#abc") and short ID ("abc")
	fullID := id
	if !strings.HasPrefix(id, "webhook#") {
		fullID = "webhook#" + id
	}

	output, err := s.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: fullID},
		},
	})
	if err != nil {
		return domain.WebhookEvent{}, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if output.Item == nil {
		return domain.WebhookEvent{}, fmt.Errorf("webhook event %q not found", id)
	}

	var event domain.WebhookEvent
	if err := attributevalue.UnmarshalMap(output.Item, &event); err != nil {
		return domain.WebhookEvent{}, fmt.Errorf("unmarshal webhook event: %w", err)
	}

	return event, nil
}
