// ABOUTME: This file provides a mock implementation of MemService for local dev.
// ABOUTME: It returns hardcoded observation, summary, and prompt data without DynamoDB.
package mock

import (
	"context"

	"github.com/jduncan/josh-bot/internal/domain"
)

// MemService is a mock implementation of domain.MemService.
type MemService struct{}

// NewMemService creates a new mock MemService.
func NewMemService() *MemService {
	return &MemService{}
}

// GetObservations returns hardcoded observations, optionally filtered by type and project.
func (s *MemService) GetObservations(_ context.Context, obsType, project string) ([]domain.MemObservation, error) {
	observations := []domain.MemObservation{
		{
			ID:             "obs#42",
			Type:           "decision",
			SourceID:       42,
			Project:        "josh.bot",
			SessionID:      "sess-abc",
			Title:          "Chose DynamoDB single-table design",
			Narrative:      "Evaluated DynamoDB vs PostgreSQL for the bot data layer",
			CreatedAt:      "2026-02-15T12:00:00Z",
			CreatedAtEpoch: 1739620800,
		},
		{
			ID:             "obs#43",
			Type:           "feature",
			SourceID:       43,
			Project:        "josh.bot",
			SessionID:      "sess-abc",
			Title:          "Implemented /v1/links endpoint",
			CreatedAt:      "2026-02-15T13:00:00Z",
			CreatedAtEpoch: 1739624400,
		},
	}

	var filtered []domain.MemObservation
	for _, o := range observations {
		if obsType != "" && o.Type != obsType {
			continue
		}
		if project != "" && o.Project != project {
			continue
		}
		filtered = append(filtered, o)
	}
	return filtered, nil
}

// GetObservation returns a hardcoded observation by ID.
func (s *MemService) GetObservation(ctx context.Context, id string) (domain.MemObservation, error) {
	obs, _ := s.GetObservations(ctx, "", "")
	for _, o := range obs {
		if o.ID == id || o.ID == "obs#"+id {
			return o, nil
		}
	}
	return domain.MemObservation{}, &domain.NotFoundError{Resource: "observation", ID: id}
}

// GetSummaries returns hardcoded summaries, optionally filtered by project.
func (s *MemService) GetSummaries(_ context.Context, project string) ([]domain.MemSummary, error) {
	summaries := []domain.MemSummary{
		{
			ID:             "summary#10",
			Type:           "summary",
			SourceID:       10,
			Project:        "josh.bot",
			SessionID:      "sess-xyz",
			Request:        "Add mem endpoints to josh.bot API",
			Completed:      "Built domain types and DynamoDB adapter",
			CreatedAt:      "2026-02-15T14:00:00Z",
			CreatedAtEpoch: 1739628000,
		},
	}

	if project == "" {
		return summaries, nil
	}
	var filtered []domain.MemSummary
	for _, sm := range summaries {
		if sm.Project == project {
			filtered = append(filtered, sm)
		}
	}
	return filtered, nil
}

// GetSummary returns a hardcoded summary by ID.
func (s *MemService) GetSummary(ctx context.Context, id string) (domain.MemSummary, error) {
	summaries, _ := s.GetSummaries(ctx, "")
	for _, sm := range summaries {
		if sm.ID == id || sm.ID == "summary#"+id {
			return sm, nil
		}
	}
	return domain.MemSummary{}, &domain.NotFoundError{Resource: "summary", ID: id}
}

// GetPrompts returns hardcoded prompts.
func (s *MemService) GetPrompts(_ context.Context) ([]domain.MemPrompt, error) {
	return []domain.MemPrompt{
		{
			ID:             "prompt#5",
			Type:           "prompt",
			SourceID:       5,
			SessionID:      "sess-abc",
			PromptNumber:   3,
			PromptText:     "Implement the mem API endpoints",
			CreatedAt:      "2026-02-15T12:00:00Z",
			CreatedAtEpoch: 1739620800,
		},
	}, nil
}

// GetPrompt returns a hardcoded prompt by ID.
func (s *MemService) GetPrompt(ctx context.Context, id string) (domain.MemPrompt, error) {
	prompts, _ := s.GetPrompts(ctx)
	for _, p := range prompts {
		if p.ID == id || p.ID == "prompt#"+id {
			return p, nil
		}
	}
	return domain.MemPrompt{}, &domain.NotFoundError{Resource: "prompt", ID: id}
}

// GetMemories returns hardcoded memories, optionally filtered by category.
func (s *MemService) GetMemories(_ context.Context, category string) ([]domain.Memory, error) {
	memories := []domain.Memory{
		{
			ID:             "mem#abc12345",
			Type:           "memory",
			Content:        "k8-one prefers Go for backend services",
			Category:       "preference",
			Tags:           []string{"coding", "go"},
			Source:         "manual",
			CreatedAt:      "2026-02-15T12:00:00Z",
			CreatedAtEpoch: 1739620800,
		},
	}

	if category == "" {
		return memories, nil
	}
	var filtered []domain.Memory
	for _, m := range memories {
		if m.Category == category {
			filtered = append(filtered, m)
		}
	}
	return filtered, nil
}

// GetMemory returns a hardcoded memory by ID.
func (s *MemService) GetMemory(ctx context.Context, id string) (domain.Memory, error) {
	memories, _ := s.GetMemories(ctx, "")
	for _, m := range memories {
		if m.ID == id || m.ID == "mem#"+id {
			return m, nil
		}
	}
	return domain.Memory{}, &domain.NotFoundError{Resource: "memory", ID: id}
}

// CreateMemory is a no-op mock for creating memories.
func (s *MemService) CreateMemory(_ context.Context, memory domain.Memory) error {
	return nil
}

// UpdateMemory is a no-op mock for updating memories.
func (s *MemService) UpdateMemory(_ context.Context, id string, fields map[string]any) error {
	return nil
}

// DeleteMemory is a no-op mock for deleting memories.
func (s *MemService) DeleteMemory(_ context.Context, id string) error {
	return nil
}

// GetStats returns hardcoded aggregate stats.
func (s *MemService) GetStats(_ context.Context) (domain.MemStats, error) {
	return domain.MemStats{
		TotalObservations: 150,
		TotalSummaries:    30,
		TotalPrompts:      75,
		ByType: map[string]int{
			"decision": 45,
			"feature":  60,
			"bugfix":   25,
			"change":   20,
		},
		ByProject: map[string]int{
			"josh.bot": 120,
			"other":    30,
		},
	}, nil
}
