// ABOUTME: This file provides a mock implementation of the BotService for testing purposes.
// ABOUTME: It returns hardcoded data to decouple tests from real data sources.
package mock

import "github.com/jduncan/josh-bot/internal/domain"

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

// GetProjects returns a hardcoded list of projects.
func (s *BotService) GetProjects() ([]domain.Project, error) {
	return []domain.Project{
		{Name: "Modular AWS Backend", Stack: "Go, AWS", Description: "Read-only S3/DynamoDB access."},
		{Name: "Modernist Cookbot", Stack: "Python, Anthropic", Description: "AI sous-chef for sous-vide."},
	}, nil
}
