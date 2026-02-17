// ABOUTME: This file tests the FormatObsidian function that renders diary entries as Obsidian markdown.
// ABOUTME: Covers frontmatter generation, section formatting, and edge cases.
package domain

import (
	"strings"
	"testing"
)

func TestFormatObsidian_BasicEntry(t *testing.T) {
	entry := DiaryEntry{
		ID:        "diary#abc123",
		Title:     "A Good Day",
		Context:   "Monday morning, coffee shop",
		Body:      "Shipped the new API endpoint",
		Reaction:  "Felt accomplished",
		Takeaway:  "Small wins compound",
		Tags:      []string{"work", "wins"},
		CreatedAt: "2026-02-17T15:30:00Z",
	}

	got := string(FormatObsidian(entry))

	// Check frontmatter
	if !strings.Contains(got, "---\n") {
		t.Error("should contain frontmatter delimiters")
	}
	if !strings.Contains(got, "date: 2026-02-17T15:30:00Z") {
		t.Error("should contain date in frontmatter")
	}
	if !strings.Contains(got, "title: A Good Day") {
		t.Error("should contain title in frontmatter")
	}
	if !strings.Contains(got, "  - diary") {
		t.Error("should always include 'diary' tag")
	}
	if !strings.Contains(got, "  - work") {
		t.Error("should include user tags")
	}
	if !strings.Contains(got, "  - wins") {
		t.Error("should include user tags")
	}

	// Check sections
	if !strings.Contains(got, "## Context\n\nMonday morning, coffee shop") {
		t.Error("should contain Context section")
	}
	if !strings.Contains(got, "## What Happened\n\nShipped the new API endpoint") {
		t.Error("should contain What Happened section")
	}
	if !strings.Contains(got, "## Reaction\n\nFelt accomplished") {
		t.Error("should contain Reaction section")
	}
	if !strings.Contains(got, "## Takeaway\n\nSmall wins compound") {
		t.Error("should contain Takeaway section")
	}
}

func TestFormatObsidian_NoTitle(t *testing.T) {
	entry := DiaryEntry{
		Context:   "Tuesday evening",
		Body:      "Debugged a tricky issue",
		Reaction:  "Frustrated then relieved",
		Takeaway:  "Read the error message",
		CreatedAt: "2026-02-18T20:00:00Z",
	}

	got := string(FormatObsidian(entry))

	// Should not contain a title line in frontmatter
	if strings.Contains(got, "title:") {
		t.Error("should not include title line when title is empty")
	}
}

func TestFormatObsidian_NoUserTags(t *testing.T) {
	entry := DiaryEntry{
		Context:   "Wednesday",
		Body:      "Nothing special",
		Reaction:  "Meh",
		Takeaway:  "Rest is important",
		CreatedAt: "2026-02-19T12:00:00Z",
	}

	got := string(FormatObsidian(entry))

	// Should still have the diary tag
	if !strings.Contains(got, "  - diary") {
		t.Error("should always include 'diary' tag even with no user tags")
	}
}

func TestFormatObsidian_DuplicateDiaryTag(t *testing.T) {
	entry := DiaryEntry{
		Context:   "Thursday",
		Body:      "Test",
		Reaction:  "Test",
		Takeaway:  "Test",
		Tags:      []string{"diary", "extra"},
		CreatedAt: "2026-02-20T12:00:00Z",
	}

	got := string(FormatObsidian(entry))

	// Count occurrences of "  - diary" to ensure no duplicate
	count := strings.Count(got, "  - diary")
	if count != 1 {
		t.Errorf("should have exactly 1 'diary' tag, got %d", count)
	}
}

func TestObsidianFilePath(t *testing.T) {
	got := ObsidianFilePath("2026-02-17T15:30:45Z")
	want := "diary/2026-02-17-153045.md"
	if got != want {
		t.Errorf("ObsidianFilePath = %q, want %q", got, want)
	}
}

func TestObsidianFilePath_DifferentTimestamp(t *testing.T) {
	got := ObsidianFilePath("2026-01-01T00:00:00Z")
	want := "diary/2026-01-01-000000.md"
	if got != want {
		t.Errorf("ObsidianFilePath = %q, want %q", got, want)
	}
}
