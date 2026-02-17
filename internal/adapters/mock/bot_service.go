// ABOUTME: This file provides a mock implementation of the BotService for testing purposes.
// ABOUTME: It returns hardcoded data to decouple tests from real data sources.
package mock

import (
	"fmt"
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
func (s *BotService) GetStatus() (domain.Status, error) {
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
func (s *BotService) UpdateStatus(fields map[string]any) error {
	return nil
}

// GetProjects returns a hardcoded list of projects.
func (s *BotService) GetProjects() ([]domain.Project, error) {
	return []domain.Project{
		{Slug: "modular-aws-backend", Name: "Modular AWS Backend", Stack: "Go, AWS", Description: "Read-only S3/DynamoDB access.", URL: "https://github.com/vaporeyes/josh-bot", Status: "active"},
		{Slug: "modernist-cookbot", Name: "Modernist Cookbot", Stack: "Python, Anthropic", Description: "AI sous-chef for sous-vide.", URL: "https://github.com/vaporeyes/cookbot", Status: "active"},
	}, nil
}

// GetProject returns a hardcoded project by slug.
func (s *BotService) GetProject(slug string) (domain.Project, error) {
	projects, _ := s.GetProjects()
	for _, p := range projects {
		if p.Slug == slug {
			return p, nil
		}
	}
	return domain.Project{}, fmt.Errorf("project %q not found", slug)
}

// CreateProject is a no-op in the mock adapter.
func (s *BotService) CreateProject(project domain.Project) error {
	return nil
}

// UpdateProject is a no-op in the mock adapter.
func (s *BotService) UpdateProject(slug string, fields map[string]any) error {
	return nil
}

// DeleteProject is a no-op in the mock adapter.
func (s *BotService) DeleteProject(slug string) error {
	return nil
}

// GetLinks returns hardcoded links, optionally filtered by tag.
func (s *BotService) GetLinks(tag string) ([]domain.Link, error) {
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
func (s *BotService) GetLink(id string) (domain.Link, error) {
	links, _ := s.GetLinks("")
	for _, l := range links {
		if l.ID == id {
			return l, nil
		}
	}
	return domain.Link{}, fmt.Errorf("link %q not found", id)
}

// CreateLink is a no-op in the mock adapter.
func (s *BotService) CreateLink(link domain.Link) error {
	return nil
}

// UpdateLink is a no-op in the mock adapter.
func (s *BotService) UpdateLink(id string, fields map[string]any) error {
	return nil
}

// DeleteLink is a no-op in the mock adapter.
func (s *BotService) DeleteLink(id string) error {
	return nil
}

// GetNotes returns hardcoded notes, optionally filtered by tag.
func (s *BotService) GetNotes(tag string) ([]domain.Note, error) {
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
func (s *BotService) GetNote(id string) (domain.Note, error) {
	notes, _ := s.GetNotes("")
	for _, n := range notes {
		if n.ID == id {
			return n, nil
		}
	}
	return domain.Note{}, fmt.Errorf("note %q not found", id)
}

// CreateNote is a no-op in the mock adapter.
func (s *BotService) CreateNote(note domain.Note) error {
	return nil
}

// UpdateNote is a no-op in the mock adapter.
func (s *BotService) UpdateNote(id string, fields map[string]any) error {
	return nil
}

// DeleteNote is a no-op in the mock adapter.
func (s *BotService) DeleteNote(id string) error {
	return nil
}

// GetTILs returns hardcoded TIL entries, optionally filtered by tag.
func (s *BotService) GetTILs(tag string) ([]domain.TIL, error) {
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
func (s *BotService) GetTIL(id string) (domain.TIL, error) {
	tils, _ := s.GetTILs("")
	for _, t := range tils {
		if t.ID == id {
			return t, nil
		}
	}
	return domain.TIL{}, fmt.Errorf("til %q not found", id)
}

// CreateTIL is a no-op in the mock adapter.
func (s *BotService) CreateTIL(til domain.TIL) error {
	return nil
}

// UpdateTIL is a no-op in the mock adapter.
func (s *BotService) UpdateTIL(id string, fields map[string]any) error {
	return nil
}

// DeleteTIL is a no-op in the mock adapter.
func (s *BotService) DeleteTIL(id string) error {
	return nil
}

// GetLogEntries returns hardcoded log entries, optionally filtered by tag.
func (s *BotService) GetLogEntries(tag string) ([]domain.LogEntry, error) {
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
func (s *BotService) GetLogEntry(id string) (domain.LogEntry, error) {
	entries, _ := s.GetLogEntries("")
	for _, e := range entries {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.LogEntry{}, fmt.Errorf("log entry %q not found", id)
}

// CreateLogEntry is a no-op in the mock adapter.
func (s *BotService) CreateLogEntry(entry domain.LogEntry) error {
	return nil
}

// UpdateLogEntry is a no-op in the mock adapter.
func (s *BotService) UpdateLogEntry(id string, fields map[string]any) error {
	return nil
}

// DeleteLogEntry is a no-op in the mock adapter.
func (s *BotService) DeleteLogEntry(id string) error {
	return nil
}

// GetDiaryEntries returns hardcoded diary entries, optionally filtered by tag.
func (s *BotService) GetDiaryEntries(tag string) ([]domain.DiaryEntry, error) {
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
func (s *BotService) GetDiaryEntry(id string) (domain.DiaryEntry, error) {
	fullID := "diary#" + id
	entries, _ := s.GetDiaryEntries("")
	for _, e := range entries {
		if e.ID == fullID {
			return e, nil
		}
	}
	return domain.DiaryEntry{}, fmt.Errorf("diary entry %q not found", id)
}

// CreateDiaryEntry is a no-op in the mock adapter.
func (s *BotService) CreateDiaryEntry(entry domain.DiaryEntry) error {
	return nil
}

// UpdateDiaryEntry is a no-op in the mock adapter.
func (s *BotService) UpdateDiaryEntry(id string, fields map[string]any) error {
	return nil
}

// DeleteDiaryEntry is a no-op in the mock adapter.
func (s *BotService) DeleteDiaryEntry(id string) error {
	return nil
}

// MetricsService is a mock implementation of domain.MetricsService.
type MetricsService struct{}

// NewMetricsService creates a new mock MetricsService.
func NewMetricsService() *MetricsService {
	return &MetricsService{}
}

// GetMetrics returns hardcoded metrics for testing.
func (s *MetricsService) GetMetrics() (domain.MetricsResponse, error) {
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
