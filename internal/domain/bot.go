// ABOUTME: This file defines the core domain models and service interfaces for the josh.bot.
// ABOUTME: It represents the central business logic of the application.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

// Status represents the current state of the bot and its owner.
type Status struct {
	Name            string            `json:"name" dynamodbav:"name"`
	Title           string            `json:"title" dynamodbav:"title"`
	Bio             string            `json:"bio" dynamodbav:"bio"`
	CurrentActivity string            `json:"current_activity" dynamodbav:"current_activity"`
	Location        string            `json:"location" dynamodbav:"location"`
	Availability    string            `json:"availability" dynamodbav:"availability"`
	Status          string            `json:"status" dynamodbav:"status"`
	Links           map[string]string `json:"links" dynamodbav:"links"`
	Interests       []string          `json:"interests" dynamodbav:"interests"`
	UpdatedAt       string            `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
}

// Project represents a software project or effort.
type Project struct {
	Slug        string `json:"slug" dynamodbav:"slug"`
	Name        string `json:"name" dynamodbav:"name"`
	Stack       string `json:"stack" dynamodbav:"stack"`
	Description string `json:"description" dynamodbav:"description"`
	URL         string `json:"url" dynamodbav:"url"`
	Status      string `json:"status" dynamodbav:"status"`
	UpdatedAt   string `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
}

// Link represents a saved bookmark or link.
type Link struct {
	ID        string   `json:"id" dynamodbav:"id"`
	URL       string   `json:"url" dynamodbav:"url"`
	Title     string   `json:"title" dynamodbav:"title"`
	Tags      []string `json:"tags" dynamodbav:"tags"`
	CreatedAt string   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt string   `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
}

// LinkIDFromURL generates a deterministic ID from a URL using SHA256.
// AIDEV-NOTE: 12 hex chars (6 bytes) gives dedup with low collision risk at our scale.
func LinkIDFromURL(rawURL string) string {
	h := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(h[:6])
}

// BotService is the interface that defines the operations for the bot.
type BotService interface {
	GetStatus() (Status, error)
	GetProjects() ([]Project, error)
	GetProject(slug string) (Project, error)
	CreateProject(project Project) error
	UpdateProject(slug string, fields map[string]any) error
	DeleteProject(slug string) error
	UpdateStatus(fields map[string]any) error
	GetLinks(tag string) ([]Link, error)
	GetLink(id string) (Link, error)
	CreateLink(link Link) error
	UpdateLink(id string, fields map[string]any) error
	DeleteLink(id string) error
}
