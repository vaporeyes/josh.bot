// ABOUTME: This file implements MetricsService using DynamoDB.
// ABOUTME: It scans the lifts table and computes tonnage, E1RM, and reads focus from status.
package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jduncan/josh-bot/internal/domain"
)

// MetricsService implements domain.MetricsService using DynamoDB.
type MetricsService struct {
	client         DynamoDBClient
	liftsTableName string
	dataTableName  string
}

// NewMetricsService creates a DynamoDB-backed MetricsService.
func NewMetricsService(client DynamoDBClient, liftsTableName, dataTableName string) *MetricsService {
	return &MetricsService{
		client:         client,
		liftsTableName: liftsTableName,
		dataTableName:  dataTableName,
	}
}

// GetMetrics computes the metrics dashboard from lift data and status.
func (s *MetricsService) GetMetrics() (domain.MetricsResponse, error) {
	lifts, err := s.scanLifts()
	if err != nil {
		return domain.MetricsResponse{}, fmt.Errorf("scan lifts: %w", err)
	}

	focus, err := s.getFocus()
	if err != nil {
		return domain.MetricsResponse{}, fmt.Errorf("get focus: %w", err)
	}

	now := time.Now().UTC()

	return domain.MetricsResponse{
		Timestamp: now.Format(time.RFC3339),
		Human: domain.HumanMetrics{
			Focus:            focus,
			WeeklyTonnageLbs: domain.WeeklyTonnage(lifts, now),
			Estimated1RM: map[string]int{
				"deadlift": domain.BestE1RM(lifts, "Deadlift (Barbell)"),
				"squat":    domain.BestE1RM(lifts, "Squat (Barbell)"),
				"bench":    domain.BestE1RM(lifts, "Bench Press (Barbell)"),
			},
			LastWorkout: domain.LastWorkout(lifts),
		},
	}, nil
}

// scanLifts retrieves all lift records from the lifts table.
// AIDEV-NOTE: Full table scan is fine at ~5K items. Revisit if data grows significantly.
func (s *MetricsService) scanLifts() ([]domain.Lift, error) {
	var lifts []domain.Lift
	var lastKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:         &s.liftsTableName,
			ExclusiveStartKey: lastKey,
		}

		output, err := s.client.Scan(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("dynamodb Scan: %w", err)
		}

		for _, item := range output.Items {
			var l domain.Lift
			if err := attributevalue.UnmarshalMap(item, &l); err != nil {
				return nil, fmt.Errorf("unmarshal lift: %w", err)
			}
			lifts = append(lifts, l)
		}

		lastKey = output.LastEvaluatedKey
		if lastKey == nil {
			break
		}
	}

	return lifts, nil
}

// getFocus reads the focus field from the status item in the data table.
func (s *MetricsService) getFocus() (string, error) {
	output, err := s.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: &s.dataTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "status"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("dynamodb GetItem: %w", err)
	}
	if output.Item == nil {
		return "", nil
	}

	// Extract focus field directly
	if focusAttr, ok := output.Item["focus"]; ok {
		if s, ok := focusAttr.(*types.AttributeValueMemberS); ok {
			return s.Value, nil
		}
	}

	return "", nil
}
