// ABOUTME: This file tests the mem domain types for correct JSON serialization.
// ABOUTME: It verifies struct tags and basic field behavior for observations, summaries, and prompts.
package domain

import (
	"encoding/json"
	"testing"
)

func TestMemObservation_JSONTags(t *testing.T) {
	obs := MemObservation{
		ID:             "obs#42",
		Type:           "decision",
		SourceID:       42,
		Project:        "josh.bot",
		SessionID:      "sess-abc",
		Title:          "Chose DynamoDB",
		CreatedAt:      "2026-02-15T12:00:00Z",
		CreatedAtEpoch: 1739620800,
	}

	data, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m["id"] != "obs#42" {
		t.Errorf("expected id 'obs#42', got %v", m["id"])
	}
	if m["type"] != "decision" {
		t.Errorf("expected type 'decision', got %v", m["type"])
	}
	if m["source_id"].(float64) != 42 {
		t.Errorf("expected source_id 42, got %v", m["source_id"])
	}
	if m["project"] != "josh.bot" {
		t.Errorf("expected project 'josh.bot', got %v", m["project"])
	}
	if m["created_at_epoch"].(float64) != 1739620800 {
		t.Errorf("expected created_at_epoch 1739620800, got %v", m["created_at_epoch"])
	}
}

func TestMemObservation_OmitsEmpty(t *testing.T) {
	obs := MemObservation{
		ID:             "obs#1",
		Type:           "feature",
		CreatedAt:      "2026-02-15T12:00:00Z",
		CreatedAtEpoch: 1739620800,
	}

	data, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// These fields have omitempty and should be absent
	for _, key := range []string{"title", "subtitle", "narrative", "text", "facts", "concepts", "files_read", "files_modified"} {
		if _, ok := m[key]; ok {
			t.Errorf("expected %q to be omitted when empty", key)
		}
	}
}

func TestMemSummary_JSONTags(t *testing.T) {
	s := MemSummary{
		ID:             "summary#10",
		Type:           "summary",
		SourceID:       10,
		Project:        "josh.bot",
		SessionID:      "sess-xyz",
		Request:        "Add mem endpoints",
		CreatedAt:      "2026-02-15T12:00:00Z",
		CreatedAtEpoch: 1739620800,
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m["id"] != "summary#10" {
		t.Errorf("expected id 'summary#10', got %v", m["id"])
	}
	if m["request"] != "Add mem endpoints" {
		t.Errorf("expected request, got %v", m["request"])
	}
}

func TestMemPrompt_JSONTags(t *testing.T) {
	p := MemPrompt{
		ID:             "prompt#5",
		Type:           "prompt",
		SourceID:       5,
		SessionID:      "sess-abc",
		PromptNumber:   3,
		PromptText:     "Show me the plan",
		CreatedAt:      "2026-02-15T12:00:00Z",
		CreatedAtEpoch: 1739620800,
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m["prompt_number"].(float64) != 3 {
		t.Errorf("expected prompt_number 3, got %v", m["prompt_number"])
	}
	if m["prompt_text"] != "Show me the plan" {
		t.Errorf("expected prompt_text, got %v", m["prompt_text"])
	}
}

func TestMemStats_JSONTags(t *testing.T) {
	stats := MemStats{
		TotalObservations: 100,
		TotalSummaries:    20,
		TotalPrompts:      50,
		ByType:            map[string]int{"decision": 30, "feature": 70},
		ByProject:         map[string]int{"josh.bot": 80, "other": 20},
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m["total_observations"].(float64) != 100 {
		t.Errorf("expected total_observations 100, got %v", m["total_observations"])
	}
	if m["total_summaries"].(float64) != 20 {
		t.Errorf("expected total_summaries 20, got %v", m["total_summaries"])
	}
	if m["total_prompts"].(float64) != 50 {
		t.Errorf("expected total_prompts 50, got %v", m["total_prompts"])
	}
}
