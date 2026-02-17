// ABOUTME: This file defines the DiaryService and ObsidianPublisher interfaces.
// ABOUTME: It also contains FormatObsidian, which renders a DiaryEntry as Obsidian-compatible markdown.
package domain

import (
	"context"
	"fmt"
	"strings"
)

// ObsidianPublisher pushes markdown files to an Obsidian vault (e.g. via GitHub).
type ObsidianPublisher interface {
	Publish(ctx context.Context, path string, content []byte, commitMsg string) error
}

// DiaryService orchestrates diary entry creation with optional Obsidian publishing.
type DiaryService interface {
	CreateAndPublish(ctx context.Context, entry DiaryEntry) (DiaryEntry, error)
}

// FormatObsidian renders a DiaryEntry as Obsidian-compatible markdown with YAML frontmatter.
func FormatObsidian(entry DiaryEntry) []byte {
	var b strings.Builder

	// Frontmatter
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("date: %s\n", entry.CreatedAt))
	if entry.Title != "" {
		b.WriteString(fmt.Sprintf("title: %s\n", entry.Title))
	}
	b.WriteString("tags:\n")
	b.WriteString("  - diary\n")
	for _, tag := range entry.Tags {
		if tag == "diary" {
			continue
		}
		b.WriteString(fmt.Sprintf("  - %s\n", tag))
	}
	b.WriteString("---\n\n")

	// Sections
	b.WriteString(fmt.Sprintf("## Context\n\n%s\n\n", entry.Context))
	b.WriteString(fmt.Sprintf("## What Happened\n\n%s\n\n", entry.Body))
	b.WriteString(fmt.Sprintf("## Reaction\n\n%s\n\n", entry.Reaction))
	b.WriteString(fmt.Sprintf("## Takeaway\n\n%s\n", entry.Takeaway))

	return []byte(b.String())
}

// ObsidianFilePath generates the file path for a diary entry in the Obsidian vault.
// Uses the CreatedAt timestamp to produce: diary/YYYY-MM-DD-HHMMSS.md
func ObsidianFilePath(createdAt string) string {
	// Parse "2026-02-17T15:30:45Z" -> "diary/2026-02-17-153045.md"
	s := createdAt
	s = strings.ReplaceAll(s, ":", "")
	s = strings.Replace(s, "T", "-", 1)
	s = strings.TrimSuffix(s, "Z")
	return fmt.Sprintf("diary/%s.md", s)
}
