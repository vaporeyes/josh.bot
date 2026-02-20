// ABOUTME: This file defines domain types and service interface for claude-mem data.
// ABOUTME: It models observations, summaries, prompts, and memories stored in the josh-bot-mem table.
package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// MemObservation represents a single development observation from claude-mem.
type MemObservation struct {
	ID             string `json:"id" dynamodbav:"id"`
	Type           string `json:"type" dynamodbav:"type"`
	SourceID       int64  `json:"source_id" dynamodbav:"source_id"`
	Project        string `json:"project" dynamodbav:"project"`
	SessionID      string `json:"session_id" dynamodbav:"session_id"`
	Title          string `json:"title,omitempty" dynamodbav:"title,omitempty"`
	Subtitle       string `json:"subtitle,omitempty" dynamodbav:"subtitle,omitempty"`
	Narrative      string `json:"narrative,omitempty" dynamodbav:"narrative,omitempty"`
	Text           string `json:"text,omitempty" dynamodbav:"text,omitempty"`
	Facts          string `json:"facts,omitempty" dynamodbav:"facts,omitempty"`
	Concepts       string `json:"concepts,omitempty" dynamodbav:"concepts,omitempty"`
	FilesRead      string `json:"files_read,omitempty" dynamodbav:"files_read,omitempty"`
	FilesModified  string `json:"files_modified,omitempty" dynamodbav:"files_modified,omitempty"`
	CreatedAt      string `json:"created_at" dynamodbav:"created_at"`
	CreatedAtEpoch int64  `json:"created_at_epoch" dynamodbav:"created_at_epoch"`
}

// MemSummary represents a session summary from claude-mem.
type MemSummary struct {
	ID             string `json:"id" dynamodbav:"id"`
	Type           string `json:"type" dynamodbav:"type"`
	SourceID       int64  `json:"source_id" dynamodbav:"source_id"`
	Project        string `json:"project" dynamodbav:"project"`
	SessionID      string `json:"session_id" dynamodbav:"session_id"`
	Request        string `json:"request,omitempty" dynamodbav:"request,omitempty"`
	Investigated   string `json:"investigated,omitempty" dynamodbav:"investigated,omitempty"`
	Learned        string `json:"learned,omitempty" dynamodbav:"learned,omitempty"`
	Completed      string `json:"completed,omitempty" dynamodbav:"completed,omitempty"`
	NextSteps      string `json:"next_steps,omitempty" dynamodbav:"next_steps,omitempty"`
	Notes          string `json:"notes,omitempty" dynamodbav:"notes,omitempty"`
	CreatedAt      string `json:"created_at" dynamodbav:"created_at"`
	CreatedAtEpoch int64  `json:"created_at_epoch" dynamodbav:"created_at_epoch"`
}

// MemPrompt represents a user prompt from claude-mem.
type MemPrompt struct {
	ID             string `json:"id" dynamodbav:"id"`
	Type           string `json:"type" dynamodbav:"type"`
	SourceID       int64  `json:"source_id" dynamodbav:"source_id"`
	SessionID      string `json:"session_id" dynamodbav:"session_id"`
	PromptNumber   int64  `json:"prompt_number" dynamodbav:"prompt_number"`
	PromptText     string `json:"prompt_text" dynamodbav:"prompt_text"`
	CreatedAt      string `json:"created_at" dynamodbav:"created_at"`
	CreatedAtEpoch int64  `json:"created_at_epoch" dynamodbav:"created_at_epoch"`
}

// MemStats contains aggregate counts of mem data by type and project.
type MemStats struct {
	TotalObservations int            `json:"total_observations"`
	TotalSummaries    int            `json:"total_summaries"`
	TotalPrompts      int            `json:"total_prompts"`
	ByType            map[string]int `json:"by_type"`
	ByProject         map[string]int `json:"by_project"`
}

// Memory represents a k8-one memory entry stored in the josh-bot-mem table.
type Memory struct {
	ID             string   `json:"id" dynamodbav:"id"`
	Type           string   `json:"type" dynamodbav:"type"`
	Content        string   `json:"content" dynamodbav:"content"`
	Category       string   `json:"category" dynamodbav:"category"`
	Tags           []string `json:"tags" dynamodbav:"tags"`
	Source         string   `json:"source" dynamodbav:"source"`
	CreatedAt      string   `json:"created_at" dynamodbav:"created_at"`
	CreatedAtEpoch int64    `json:"created_at_epoch" dynamodbav:"created_at_epoch"`
	UpdatedAt      string   `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
}

// MemoryID generates a random ID with a "mem#" prefix.
func MemoryID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "mem#" + hex.EncodeToString(b)
}

// MemService provides access to claude-mem data and memories stored in DynamoDB.
type MemService interface {
	GetObservations(ctx context.Context, obsType, project string) ([]MemObservation, error)
	GetObservation(ctx context.Context, id string) (MemObservation, error)
	GetSummaries(ctx context.Context, project string) ([]MemSummary, error)
	GetSummary(ctx context.Context, id string) (MemSummary, error)
	GetPrompts(ctx context.Context) ([]MemPrompt, error)
	GetPrompt(ctx context.Context, id string) (MemPrompt, error)
	GetStats(ctx context.Context) (MemStats, error)
	GetMemories(ctx context.Context, category string) ([]Memory, error)
	GetMemory(ctx context.Context, id string) (Memory, error)
	CreateMemory(ctx context.Context, memory Memory) error
	UpdateMemory(ctx context.Context, id string, fields map[string]any) error
	DeleteMemory(ctx context.Context, id string) error
}
