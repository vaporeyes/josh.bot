// ABOUTME: This file tests the diary service orchestration logic.
// ABOUTME: Verifies CreateAndPublish stores in DynamoDB and publishes to GitHub, handling failures gracefully.
package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jduncan/josh-bot/internal/domain"
)

// stubBotService records CreateDiaryEntry calls for verification.
type stubBotService struct {
	domain.BotService
	createdEntry domain.DiaryEntry
	createErr    error
}

func (s *stubBotService) CreateDiaryEntry(_ context.Context, entry domain.DiaryEntry) error {
	s.createdEntry = entry
	return s.createErr
}

// stubPublisher records Publish calls for verification.
type stubPublisher struct {
	publishedPath    string
	publishedContent []byte
	publishedMsg     string
	publishErr       error
}

func (p *stubPublisher) Publish(_ context.Context, path string, content []byte, commitMsg string) error {
	p.publishedPath = path
	p.publishedContent = content
	p.publishedMsg = commitMsg
	return p.publishErr
}

func TestCreateAndPublish_Success(t *testing.T) {
	bot := &stubBotService{}
	pub := &stubPublisher{}
	svc := NewDiaryService(bot, pub)

	entry := domain.DiaryEntry{
		Context:  "Monday morning",
		Body:     "Shipped the feature",
		Reaction: "Proud",
		Takeaway: "Ship early",
		Tags:     []string{"work"},
	}

	result, err := svc.CreateAndPublish(context.Background(), entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have generated an ID
	if !strings.HasPrefix(result.ID, "diary#") {
		t.Errorf("expected diary# prefix, got %q", result.ID)
	}

	// Should have set timestamps
	if result.CreatedAt == "" {
		t.Error("expected CreatedAt to be set")
	}

	// Should have stored in DynamoDB
	if bot.createdEntry.ID == "" {
		t.Error("expected CreateDiaryEntry to be called")
	}

	// Should have published to GitHub
	if pub.publishedPath == "" {
		t.Error("expected Publish to be called")
	}
	if !strings.HasPrefix(pub.publishedPath, "diary/") {
		t.Errorf("expected path starting with diary/, got %q", pub.publishedPath)
	}
	if len(pub.publishedContent) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestCreateAndPublish_DynamoDBFails(t *testing.T) {
	bot := &stubBotService{createErr: fmt.Errorf("dynamo error")}
	pub := &stubPublisher{}
	svc := NewDiaryService(bot, pub)

	entry := domain.DiaryEntry{
		Context: "Test", Body: "Test", Reaction: "Test", Takeaway: "Test",
	}

	_, err := svc.CreateAndPublish(context.Background(), entry)
	if err == nil {
		t.Fatal("expected error when DynamoDB fails")
	}

	// Should not have attempted GitHub publish
	if pub.publishedPath != "" {
		t.Error("should not publish to GitHub when DynamoDB fails")
	}
}

func TestCreateAndPublish_GitHubFails_StillReturnsEntry(t *testing.T) {
	bot := &stubBotService{}
	pub := &stubPublisher{publishErr: fmt.Errorf("github error")}
	svc := NewDiaryService(bot, pub)

	entry := domain.DiaryEntry{
		Context: "Test", Body: "Test", Reaction: "Test", Takeaway: "Test",
	}

	result, err := svc.CreateAndPublish(context.Background(), entry)
	if err != nil {
		t.Fatalf("should not return error when only GitHub fails: %v", err)
	}

	// Entry should still be returned (DynamoDB is source of truth)
	if result.ID == "" {
		t.Error("expected entry to be returned even when GitHub fails")
	}
}
