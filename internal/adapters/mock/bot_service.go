// ABOUTME: This file provides a mock implementation of the BotService for testing purposes.
// ABOUTME: It returns hardcoded data to decouple tests from real data sources.
package mock

import (
	"context"
	"time"

	"github.com/jduncan/josh-bot/internal/domain"
)

// BotService is a mock implementation of the domain.BotService interface.
type BotService struct{}

// NewBotService creates a new mock BotService.
func NewBotService() *BotService {
	return &BotService{}
}

// GetStatus returns a hardcoded status for testing.
func (s *BotService) GetStatus(_ context.Context) (domain.Status, error) {
	return domain.Status{
		Name:            "Josh Duncan",
		Title:           "Platform Engineer",
		Bio:             "Builder of systems, lifter of heavy things, cooker of sous vide.",
		CurrentActivity: "Refining Go backends for josh.bot",
		Location:        "Clarksville, TN",
		Availability:    "Open to interesting projects",
		Status:          "ok",
		Links: map[string]string{
			"github":   "https://github.com/jduncan",
			"linkedin": "https://linkedin.com/in/jduncan",
		},
		Interests: []string{"Go", "AWS", "Sous vide", "Powerlifting", "Art Nouveau"},
	}, nil
}

// UpdateStatus is a no-op in the mock adapter.
func (s *BotService) UpdateStatus(_ context.Context, fields map[string]any) error {
	return nil
}

// GetProjects returns a hardcoded list of projects.
func (s *BotService) GetProjects(_ context.Context) ([]domain.Project, error) {
	return []domain.Project{
		{Slug: "modular-aws-backend", Name: "Modular AWS Backend", Stack: "Go, AWS", Description: "Read-only S3/DynamoDB access.", URL: "https://github.com/vaporeyes/josh-bot", Status: "active"},
		{Slug: "modernist-cookbot", Name: "Modernist Cookbot", Stack: "Python, Anthropic", Description: "AI sous-chef for sous-vide.", URL: "https://github.com/vaporeyes/cookbot", Status: "active"},
	}, nil
}

// GetProject returns a hardcoded project by slug.
func (s *BotService) GetProject(ctx context.Context, slug string) (domain.Project, error) {
	projects, _ := s.GetProjects(ctx)
	for _, p := range projects {
		if p.Slug == slug {
			return p, nil
		}
	}
	return domain.Project{}, &domain.NotFoundError{Resource: "project", ID: slug}
}

// CreateProject is a no-op in the mock adapter.
func (s *BotService) CreateProject(_ context.Context, project domain.Project) error {
	return nil
}

// UpdateProject is a no-op in the mock adapter.
func (s *BotService) UpdateProject(_ context.Context, slug string, fields map[string]any) error {
	return nil
}

// DeleteProject is a no-op in the mock adapter.
func (s *BotService) DeleteProject(_ context.Context, slug string) error {
	return nil
}

// GetLinks returns hardcoded links, optionally filtered by tag.
func (s *BotService) GetLinks(_ context.Context, tag string) ([]domain.Link, error) {
	links := []domain.Link{
		{ID: "a1b2c3d4e5f6", URL: "https://go.dev/blog/", Title: "The Go Blog", Tags: []string{"go", "programming"}},
		{ID: "b2c3d4e5f6a1", URL: "https://aws.amazon.com/dynamodb/", Title: "Amazon DynamoDB", Tags: []string{"aws", "dynamodb", "databases"}},
	}
	if tag == "" {
		return links, nil
	}
	var filtered []domain.Link
	for _, l := range links {
		for _, t := range l.Tags {
			if t == tag {
				filtered = append(filtered, l)
				break
			}
		}
	}
	return filtered, nil
}

// GetLink returns a hardcoded link by ID.
func (s *BotService) GetLink(ctx context.Context, id string) (domain.Link, error) {
	links, _ := s.GetLinks(ctx, "")
	for _, l := range links {
		if l.ID == id {
			return l, nil
		}
	}
	return domain.Link{}, &domain.NotFoundError{Resource: "link", ID: id}
}

// CreateLink is a no-op in the mock adapter.
func (s *BotService) CreateLink(_ context.Context, link domain.Link) error {
	return nil
}

// UpdateLink is a no-op in the mock adapter.
func (s *BotService) UpdateLink(_ context.Context, id string, fields map[string]any) error {
	return nil
}

// DeleteLink is a no-op in the mock adapter.
func (s *BotService) DeleteLink(_ context.Context, id string) error {
	return nil
}

// GetNotes returns hardcoded notes, optionally filtered by tag.
func (s *BotService) GetNotes(_ context.Context, tag string) ([]domain.Note, error) {
	notes := []domain.Note{
		{ID: "note#abc123", Title: "Meeting notes", Body: "Discussed API design", Tags: []string{"work"}},
		{ID: "note#def456", Title: "Grocery list", Body: "Eggs, milk, bread", Tags: []string{"personal"}},
	}
	if tag == "" {
		return notes, nil
	}
	var filtered []domain.Note
	for _, n := range notes {
		for _, t := range n.Tags {
			if t == tag {
				filtered = append(filtered, n)
				break
			}
		}
	}
	return filtered, nil
}

// GetNote returns a hardcoded note by ID.
func (s *BotService) GetNote(ctx context.Context, id string) (domain.Note, error) {
	notes, _ := s.GetNotes(ctx, "")
	for _, n := range notes {
		if n.ID == id {
			return n, nil
		}
	}
	return domain.Note{}, &domain.NotFoundError{Resource: "note", ID: id}
}

// CreateNote is a no-op in the mock adapter.
func (s *BotService) CreateNote(_ context.Context, note domain.Note) error {
	return nil
}

// UpdateNote is a no-op in the mock adapter.
func (s *BotService) UpdateNote(_ context.Context, id string, fields map[string]any) error {
	return nil
}

// DeleteNote is a no-op in the mock adapter.
func (s *BotService) DeleteNote(_ context.Context, id string) error {
	return nil
}

// GetTILs returns hardcoded TIL entries, optionally filtered by tag.
func (s *BotService) GetTILs(_ context.Context, tag string) ([]domain.TIL, error) {
	tils := []domain.TIL{
		{ID: "til#abc123", Title: "Go slices grow by 2x", Body: "When a slice exceeds capacity, Go doubles it", Tags: []string{"go"}},
		{ID: "til#def456", Title: "DynamoDB scan is O(n)", Body: "Scans read the entire table", Tags: []string{"aws", "dynamodb"}},
	}
	if tag == "" {
		return tils, nil
	}
	var filtered []domain.TIL
	for _, t := range tils {
		for _, tg := range t.Tags {
			if tg == tag {
				filtered = append(filtered, t)
				break
			}
		}
	}
	return filtered, nil
}

// GetTIL returns a hardcoded TIL by ID.
func (s *BotService) GetTIL(ctx context.Context, id string) (domain.TIL, error) {
	tils, _ := s.GetTILs(ctx, "")
	for _, t := range tils {
		if t.ID == id {
			return t, nil
		}
	}
	return domain.TIL{}, &domain.NotFoundError{Resource: "til", ID: id}
}

// CreateTIL is a no-op in the mock adapter.
func (s *BotService) CreateTIL(_ context.Context, til domain.TIL) error {
	return nil
}

// UpdateTIL is a no-op in the mock adapter.
func (s *BotService) UpdateTIL(_ context.Context, id string, fields map[string]any) error {
	return nil
}

// DeleteTIL is a no-op in the mock adapter.
func (s *BotService) DeleteTIL(_ context.Context, id string) error {
	return nil
}

// GetLogEntries returns hardcoded log entries, optionally filtered by tag.
func (s *BotService) GetLogEntries(_ context.Context, tag string) ([]domain.LogEntry, error) {
	entries := []domain.LogEntry{
		{ID: "log#abc123", Message: "deployed josh-bot v1.2", Tags: []string{"deploy"}},
		{ID: "log#def456", Message: "updated DNS for josh.bot", Tags: []string{"infra"}},
	}
	if tag == "" {
		return entries, nil
	}
	var filtered []domain.LogEntry
	for _, e := range entries {
		for _, t := range e.Tags {
			if t == tag {
				filtered = append(filtered, e)
				break
			}
		}
	}
	return filtered, nil
}

// GetLogEntry returns a hardcoded log entry by ID.
func (s *BotService) GetLogEntry(ctx context.Context, id string) (domain.LogEntry, error) {
	entries, _ := s.GetLogEntries(ctx, "")
	for _, e := range entries {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.LogEntry{}, &domain.NotFoundError{Resource: "log entry", ID: id}
}

// CreateLogEntry is a no-op in the mock adapter.
func (s *BotService) CreateLogEntry(_ context.Context, entry domain.LogEntry) error {
	return nil
}

// UpdateLogEntry is a no-op in the mock adapter.
func (s *BotService) UpdateLogEntry(_ context.Context, id string, fields map[string]any) error {
	return nil
}

// DeleteLogEntry is a no-op in the mock adapter.
func (s *BotService) DeleteLogEntry(_ context.Context, id string) error {
	return nil
}

// GetDiaryEntries returns hardcoded diary entries, optionally filtered by tag.
func (s *BotService) GetDiaryEntries(_ context.Context, tag string) ([]domain.DiaryEntry, error) {
	entries := []domain.DiaryEntry{
		{
			ID: "diary#abc123", Title: "A Good Day", Context: "Monday morning",
			Body: "Shipped the API", Reaction: "Proud", Takeaway: "Ship early",
			Tags: []string{"work"}, CreatedAt: "2026-02-17T15:00:00Z",
		},
	}
	if tag == "" {
		return entries, nil
	}
	var filtered []domain.DiaryEntry
	for _, e := range entries {
		for _, t := range e.Tags {
			if t == tag {
				filtered = append(filtered, e)
				break
			}
		}
	}
	return filtered, nil
}

// GetDiaryEntry returns a hardcoded diary entry by ID.
// AIDEV-NOTE: Matches by "diary#"+id to mirror DynamoDB adapter key construction.
func (s *BotService) GetDiaryEntry(ctx context.Context, id string) (domain.DiaryEntry, error) {
	fullID := "diary#" + id
	entries, _ := s.GetDiaryEntries(ctx, "")
	for _, e := range entries {
		if e.ID == fullID {
			return e, nil
		}
	}
	return domain.DiaryEntry{}, &domain.NotFoundError{Resource: "diary entry", ID: id}
}

// CreateDiaryEntry is a no-op in the mock adapter.
func (s *BotService) CreateDiaryEntry(_ context.Context, entry domain.DiaryEntry) error {
	return nil
}

// UpdateDiaryEntry is a no-op in the mock adapter.
func (s *BotService) UpdateDiaryEntry(_ context.Context, id string, fields map[string]any) error {
	return nil
}

// DeleteDiaryEntry is a no-op in the mock adapter.
func (s *BotService) DeleteDiaryEntry(_ context.Context, id string) error {
	return nil
}

// GetIdempotencyRecord returns nil (no cached record) in the mock.
func (s *BotService) GetIdempotencyRecord(_ context.Context, key string) (*domain.IdempotencyRecord, error) {
	return nil, nil
}

// SetIdempotencyRecord is a no-op in the mock.
func (s *BotService) SetIdempotencyRecord(_ context.Context, record domain.IdempotencyRecord) error {
	return nil
}

// MetricsService is a mock implementation of domain.MetricsService.
type MetricsService struct{}

// NewMetricsService creates a new mock MetricsService.
func NewMetricsService() *MetricsService {
	return &MetricsService{}
}

// GetMetrics returns hardcoded metrics for testing.
func (s *MetricsService) GetMetrics(_ context.Context) (domain.MetricsResponse, error) {
	devStats := &domain.MemStats{
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
	}
	return domain.MetricsResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Human: domain.HumanMetrics{
			Focus:            "Powerlifting / Hypertrophy",
			WeeklyTonnageLbs: 42500,
			Estimated1RM: map[string]int{
				"deadlift": 525,
				"squat":    455,
				"bench":    315,
			},
			LastWorkout: &domain.WorkoutSummary{
				Date:       "2026-02-14",
				Name:       "Pull Day",
				Exercises:  []string{"Deadlift (Barbell)", "Barbell Row", "Lat Pulldown (Cable)"},
				Sets:       18,
				TonnageLbs: 12500,
			},
		},
		Dev: devStats,
	}, nil
}
