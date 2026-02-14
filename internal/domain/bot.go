// ABOUTME: This file defines the core domain models and service interfaces for the josh.bot.
// ABOUTME: It represents the central business logic of the application.
package domain

// Status represents the current state of the bot.
type Status struct {
	CurrentActivity string `json:"current_activity"`
	Location        string `json:"location"`
	Status          string `json:"status"`
}

// BotService is the interface that defines the operations for the bot.
type BotService interface {
	GetStatus() (Status, error)
}
