// ABOUTME: This file defines the core domain models and service interfaces for the josh.bot.
// ABOUTME: It represents the central business logic of the application.
package domain

// Status represents the current state of the bot and its owner.
type Status struct {
	Name            string            `json:"name"`
	Title           string            `json:"title"`
	Bio             string            `json:"bio"`
	CurrentActivity string            `json:"current_activity"`
	Location        string            `json:"location"`
	Availability    string            `json:"availability"`
	Status          string            `json:"status"`
	Links           map[string]string `json:"links"`
	Interests       []string          `json:"interests"`
}

// Project represents a software project or effort.
type Project struct {
	Name        string `json:"name"`
	Stack       string `json:"stack"`
	Description string `json:"description"`
}

// BotService is the interface that defines the operations for the bot.
type BotService interface {
	GetStatus() (Status, error)
	GetProjects() ([]Project, error)
}
