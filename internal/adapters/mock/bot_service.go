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
		CurrentActivity: "Refining Go backends for josh.bot",
		Location:        "Clarksville, TN",
		Status:          "ok",
	}, nil
}

// GetProjects returns a hardcoded list of projects.
func (s *BotService) GetProjects() ([]domain.Project, error) {
	return []domain.Project{
		{Name: "Modular AWS Backend", Stack: "Go, AWS", Description: "Read-only S3/DynamoDB access."},
		{Name: "Modernist Cookbot", Stack: "Python, Anthropic", Description: "AI sous-chef for sous-vide."},
	}, nil
}
