// ABOUTME: This file implements the DiaryService that orchestrates diary entry creation.
// ABOUTME: It stores entries in DynamoDB and publishes formatted markdown to an Obsidian vault via GitHub.
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jduncan/josh-bot/internal/domain"
)

// DiaryServiceImpl implements domain.DiaryService.
type DiaryServiceImpl struct {
	botService domain.BotService
	publisher  domain.ObsidianPublisher
}

// NewDiaryService creates a diary service that stores entries and publishes to Obsidian.
func NewDiaryService(botService domain.BotService, publisher domain.ObsidianPublisher) *DiaryServiceImpl {
	return &DiaryServiceImpl{
		botService: botService,
		publisher:  publisher,
	}
}

// CreateAndPublish generates an ID, stores the entry in DynamoDB, and publishes to GitHub.
// AIDEV-NOTE: GitHub publish is best-effort; DynamoDB is the source of truth.
func (s *DiaryServiceImpl) CreateAndPublish(ctx context.Context, entry domain.DiaryEntry) (domain.DiaryEntry, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	entry.ID = domain.DiaryEntryID()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	if err := s.botService.CreateDiaryEntry(entry); err != nil {
		return domain.DiaryEntry{}, fmt.Errorf("store diary entry: %w", err)
	}

	// Publish to Obsidian via GitHub (best-effort)
	mdContent := domain.FormatObsidian(entry)
	filePath := domain.ObsidianFilePath(entry.CreatedAt)
	commitMsg := fmt.Sprintf("diary: %s", entry.CreatedAt)

	if err := s.publisher.Publish(ctx, filePath, mdContent, commitMsg); err != nil {
		log.Printf("WARNING: failed to publish diary entry %s to GitHub: %v", entry.ID, err)
	}

	return entry, nil
}
