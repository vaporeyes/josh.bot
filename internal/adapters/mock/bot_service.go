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

// MetricsService is a mock implementation of domain.MetricsService.
type MetricsService struct{}

// NewMetricsService creates a new mock MetricsService.
func NewMetricsService() *MetricsService {
	return &MetricsService{}
}

// GetMetrics returns hardcoded metrics for testing.
func (s *MetricsService) GetMetrics() (domain.MetricsResponse, error) {
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
		},
	}, nil
}
