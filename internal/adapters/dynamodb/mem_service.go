// ABOUTME: This file implements MemService using DynamoDB for reading claude-mem data and memory CRUD.
// ABOUTME: It queries the josh-bot-mem table using the type-index GSI and prefix-based scans.
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

// MemService implements domain.MemService using DynamoDB.
type MemService struct {
	client    DynamoDBClient
	tableName string
}

// NewMemService creates a DynamoDB-backed MemService.
func NewMemService(client DynamoDBClient, tableName string) *MemService {
	return &MemService{client: client, tableName: tableName}
}

// AIDEV-NOTE: type-index GSI has partition key "type" and sort key "created_at_epoch".
const typeIndexName = "type-index"

// GetObservations returns observations, optionally filtered by type and project.
// When obsType is provided, uses Query on type-index GSI. Otherwise uses Scan with prefix filter.
func (s *MemService) GetObservations(ctx context.Context, obsType, project string) ([]domain.MemObservation, error) {
	var items []map[string]types.AttributeValue

	if obsType != "" {
		// Query the type-index GSI for the specific observation type
		queryItems, err := s.queryByType(ctx, obsType, project)
		if err != nil {
			return nil, fmt.Errorf("query observations: %w", err)
		}
		items = queryItems
	} else {
		// Scan with obs# prefix filter
		scanItems, err := s.scanByPrefix(ctx, "obs#", project)
		if err != nil {
			return nil, fmt.Errorf("scan observations: %w", err)
		}
		items = scanItems
	}

	observations := make([]domain.MemObservation, 0, len(items))
	for _, item := range items {
		var obs domain.MemObservation
		if err := attributevalue.UnmarshalMap(item, &obs); err != nil {
			return nil, fmt.Errorf("unmarshal observation: %w", err)
		}
		observations = append(observations, obs)
	}

	return observations, nil
}

// GetObservation returns a single observation by ID.
func (s *MemService) GetObservation(ctx context.Context, id string) (domain.MemObservation, error) {
	key := id
	if !strings.HasPrefix(id, "obs#") {
		key = "obs#" + id
	}

	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return domain.MemObservation{}, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if output.Item == nil {
		return domain.MemObservation{}, &domain.NotFoundError{Resource: "observation", ID: id}
	}

	var obs domain.MemObservation
	if err := attributevalue.UnmarshalMap(output.Item, &obs); err != nil {
		return domain.MemObservation{}, fmt.Errorf("unmarshal observation: %w", err)
	}

	return obs, nil
}

// GetSummaries returns summaries, optionally filtered by project.
func (s *MemService) GetSummaries(ctx context.Context, project string) ([]domain.MemSummary, error) {
	items, err := s.queryByType(ctx, "summary", project)
	if err != nil {
		return nil, fmt.Errorf("query summaries: %w", err)
	}

	summaries := make([]domain.MemSummary, 0, len(items))
	for _, item := range items {
		var summary domain.MemSummary
		if err := attributevalue.UnmarshalMap(item, &summary); err != nil {
			return nil, fmt.Errorf("unmarshal summary: %w", err)
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetSummary returns a single summary by ID.
func (s *MemService) GetSummary(ctx context.Context, id string) (domain.MemSummary, error) {
	key := id
	if !strings.HasPrefix(id, "summary#") {
		key = "summary#" + id
	}

	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return domain.MemSummary{}, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if output.Item == nil {
		return domain.MemSummary{}, &domain.NotFoundError{Resource: "summary", ID: id}
	}

	var summary domain.MemSummary
	if err := attributevalue.UnmarshalMap(output.Item, &summary); err != nil {
		return domain.MemSummary{}, fmt.Errorf("unmarshal summary: %w", err)
	}

	return summary, nil
}

// GetPrompts returns all prompts.
func (s *MemService) GetPrompts(ctx context.Context) ([]domain.MemPrompt, error) {
	items, err := s.queryByType(ctx, "prompt", "")
	if err != nil {
		return nil, fmt.Errorf("query prompts: %w", err)
	}

	prompts := make([]domain.MemPrompt, 0, len(items))
	for _, item := range items {
		var prompt domain.MemPrompt
		if err := attributevalue.UnmarshalMap(item, &prompt); err != nil {
			return nil, fmt.Errorf("unmarshal prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// GetPrompt returns a single prompt by ID.
func (s *MemService) GetPrompt(ctx context.Context, id string) (domain.MemPrompt, error) {
	key := id
	if !strings.HasPrefix(id, "prompt#") {
		key = "prompt#" + id
	}

	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return domain.MemPrompt{}, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if output.Item == nil {
		return domain.MemPrompt{}, &domain.NotFoundError{Resource: "prompt", ID: id}
	}

	var prompt domain.MemPrompt
	if err := attributevalue.UnmarshalMap(output.Item, &prompt); err != nil {
		return domain.MemPrompt{}, fmt.Errorf("unmarshal prompt: %w", err)
	}

	return prompt, nil
}

// GetStats scans the full table and aggregates counts by type and project.
// AIDEV-NOTE: Full table scan is acceptable for mem table sizes (~hundreds to low thousands of items).
// AIDEV-NOTE: Paginates using LastEvaluatedKey to handle tables exceeding 1MB per scan.
func (s *MemService) GetStats(ctx context.Context) (domain.MemStats, error) {
	stats := domain.MemStats{
		ByType:    make(map[string]int),
		ByProject: make(map[string]int),
	}

	var startKey map[string]types.AttributeValue
	for {
		input := &dynamodb.ScanInput{
			TableName:         &s.tableName,
			ExclusiveStartKey: startKey,
		}

		output, err := s.client.Scan(ctx, input)
		if err != nil {
			return domain.MemStats{}, fmt.Errorf("dynamodb Scan: %w", err)
		}

		for _, item := range output.Items {
			itemType := ""
			if typeAttr, ok := item["type"]; ok {
				if s, ok := typeAttr.(*types.AttributeValueMemberS); ok {
					itemType = s.Value
				}
			}

			project := ""
			if projAttr, ok := item["project"]; ok {
				if s, ok := projAttr.(*types.AttributeValueMemberS); ok {
					project = s.Value
				}
			}

			stats.ByType[itemType]++
			if project != "" {
				stats.ByProject[project]++
			}

			// Count by category based on id prefix
			idStr := ""
			if idAttr, ok := item["id"]; ok {
				if s, ok := idAttr.(*types.AttributeValueMemberS); ok {
					idStr = s.Value
				}
			}
			switch {
			case strings.HasPrefix(idStr, "obs#"):
				stats.TotalObservations++
			case strings.HasPrefix(idStr, "summary#"):
				stats.TotalSummaries++
			case strings.HasPrefix(idStr, "prompt#"):
				stats.TotalPrompts++
			}
		}

		if output.LastEvaluatedKey == nil {
			break
		}
		startKey = output.LastEvaluatedKey
	}

	return stats, nil
}

// queryByType queries the type-index GSI for items of a given type, with optional project filter.
// AIDEV-NOTE: Paginates using LastEvaluatedKey to handle result sets exceeding 1MB.
func (s *MemService) queryByType(ctx context.Context, typeName, project string) ([]map[string]types.AttributeValue, error) {
	indexName := typeIndexName
	keyCondExpr := "#t = :type"
	exprNames := map[string]string{
		"#t": "type",
	}
	exprValues := map[string]types.AttributeValue{
		":type": &types.AttributeValueMemberS{Value: typeName},
	}

	if project != "" {
		exprValues[":project"] = &types.AttributeValueMemberS{Value: project}
	}

	var allItems []map[string]types.AttributeValue
	var startKey map[string]types.AttributeValue
	for {
		input := &dynamodb.QueryInput{
			TableName:                 &s.tableName,
			IndexName:                 &indexName,
			KeyConditionExpression:    &keyCondExpr,
			ExpressionAttributeNames:  exprNames,
			ExpressionAttributeValues: exprValues,
			ScanIndexForward:          boolPtr(false),
			ExclusiveStartKey:         startKey,
		}

		if project != "" {
			filterExpr := "project = :project"
			input.FilterExpression = &filterExpr
		}

		output, err := s.client.Query(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("dynamodb Query: %w", err)
		}

		allItems = append(allItems, output.Items...)

		if output.LastEvaluatedKey == nil {
			break
		}
		startKey = output.LastEvaluatedKey
	}

	return allItems, nil
}

// scanByPrefix scans the table for items whose id begins with the given prefix.
// AIDEV-NOTE: Paginates using LastEvaluatedKey to handle tables exceeding 1MB per scan.
func (s *MemService) scanByPrefix(ctx context.Context, prefix, project string) ([]map[string]types.AttributeValue, error) {
	filterExpr := "begins_with(id, :prefix)"
	exprValues := map[string]types.AttributeValue{
		":prefix": &types.AttributeValueMemberS{Value: prefix},
	}

	if project != "" {
		filterExpr += " AND project = :project"
		exprValues[":project"] = &types.AttributeValueMemberS{Value: project}
	}

	var allItems []map[string]types.AttributeValue
	var startKey map[string]types.AttributeValue
	for {
		output, err := s.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:                 &s.tableName,
			FilterExpression:          &filterExpr,
			ExpressionAttributeValues: exprValues,
			ExclusiveStartKey:         startKey,
		})
		if err != nil {
			return nil, fmt.Errorf("dynamodb Scan: %w", err)
		}

		allItems = append(allItems, output.Items...)

		if output.LastEvaluatedKey == nil {
			break
		}
		startKey = output.LastEvaluatedKey
	}

	return allItems, nil
}

// allowedMemoryFields defines which memory fields can be updated via PUT.
var allowedMemoryFields = map[string]bool{
	"content": true, "category": true, "tags": true, "source": true,
}

// GetMemories returns memories, optionally filtered by category.
func (s *MemService) GetMemories(ctx context.Context, category string) ([]domain.Memory, error) {
	items, err := s.queryByType(ctx, "memory", "")
	if err != nil {
		return nil, fmt.Errorf("query memories: %w", err)
	}

	memories := make([]domain.Memory, 0, len(items))
	for _, item := range items {
		var mem domain.Memory
		if err := attributevalue.UnmarshalMap(item, &mem); err != nil {
			return nil, fmt.Errorf("unmarshal memory: %w", err)
		}
		if category != "" && mem.Category != category {
			continue
		}
		memories = append(memories, mem)
	}

	return memories, nil
}

// GetMemory returns a single memory by ID.
func (s *MemService) GetMemory(ctx context.Context, id string) (domain.Memory, error) {
	key := id
	if !strings.HasPrefix(id, "mem#") {
		key = "mem#" + id
	}

	output, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return domain.Memory{}, fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if output.Item == nil {
		return domain.Memory{}, &domain.NotFoundError{Resource: "memory", ID: id}
	}

	var mem domain.Memory
	if err := attributevalue.UnmarshalMap(output.Item, &mem); err != nil {
		return domain.Memory{}, fmt.Errorf("unmarshal memory: %w", err)
	}

	return mem, nil
}

// CreateMemory adds a new memory to DynamoDB with a generated random ID.
func (s *MemService) CreateMemory(ctx context.Context, memory domain.Memory) error {
	now := time.Now().UTC()
	memory.ID = domain.MemoryID()
	memory.Type = "memory"
	memory.CreatedAt = now.Format(time.RFC3339)
	memory.CreatedAtEpoch = now.Unix()

	item, err := attributevalue.MarshalMap(memory)
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("dynamodb PutItem: %w", err)
	}

	return nil
}

// UpdateMemory updates specific fields on a memory in DynamoDB.
func (s *MemService) UpdateMemory(ctx context.Context, id string, fields map[string]any) error {
	if len(fields) == 0 {
		return fmt.Errorf("no fields provided for update")
	}

	for key := range fields {
		if !allowedMemoryFields[key] {
			return fmt.Errorf("field %q is not an updatable memory field", key)
		}
	}

	key := id
	if !strings.HasPrefix(id, "mem#") {
		key = "mem#" + id
	}

	// Add updated_at timestamp
	fields["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	updateExpr := "SET "
	exprNames := map[string]string{}
	exprValues := map[string]types.AttributeValue{}

	i := 0
	for k, v := range fields {
		if i > 0 {
			updateExpr += ", "
		}
		placeholder := fmt.Sprintf("#f%d", i)
		valuePlaceholder := fmt.Sprintf(":v%d", i)
		updateExpr += placeholder + " = " + valuePlaceholder
		exprNames[placeholder] = k

		av, err := attributevalue.Marshal(v)
		if err != nil {
			return fmt.Errorf("marshal field %q: %w", k, err)
		}
		exprValues[valuePlaceholder] = av
		i++
	}

	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: key},
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

// DeleteMemory removes a memory from DynamoDB.
func (s *MemService) DeleteMemory(ctx context.Context, id string) error {
	key := id
	if !strings.HasPrefix(id, "mem#") {
		key = "mem#" + id
	}

	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb DeleteItem: %w", err)
	}
	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
