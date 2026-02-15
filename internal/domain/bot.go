// ABOUTME: This file defines the core domain models and service interfaces for the josh.bot.
// ABOUTME: It represents the central business logic of the application.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
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

// Lift represents a single set within a workout.
type Lift struct {
	ID           string  `json:"id" dynamodbav:"id"`
	Date         string  `json:"date" dynamodbav:"date"`
	WorkoutName  string  `json:"workout_name" dynamodbav:"workout_name"`
	Duration     string  `json:"duration" dynamodbav:"duration"`
	ExerciseName string  `json:"exercise_name" dynamodbav:"exercise_name"`
	SetOrder     string  `json:"set_order" dynamodbav:"set_order"`
	Weight       float64 `json:"weight" dynamodbav:"weight"`
	Reps         float64 `json:"reps" dynamodbav:"reps"`
	Distance     float64 `json:"distance" dynamodbav:"distance"`
	Seconds      float64 `json:"seconds" dynamodbav:"seconds"`
	RPE          float64 `json:"rpe,omitempty" dynamodbav:"rpe,omitempty"`
}

// slugRegexp matches one or more non-alphanumeric characters for slug generation.
var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

// ExerciseSlug normalizes an exercise name to a URL-safe slug.
// "Squat (Barbell)" -> "squat-barbell"
func ExerciseSlug(name string) string {
	lower := strings.ToLower(name)
	slug := slugRegexp.ReplaceAllString(lower, "-")
	return strings.Trim(slug, "-")
}

// CompactDate converts a CSV date "2022-05-11 04:20:50" to compact "20220511T042050".
func CompactDate(csvDate string) string {
	d := strings.ReplaceAll(csvDate, "-", "")
	d = strings.Replace(d, " ", "T", 1)
	d = strings.ReplaceAll(d, ":", "")
	return d
}

// LiftID generates a deterministic DynamoDB key from workout set data.
// AIDEV-NOTE: Deterministic IDs make CSV re-imports idempotent (PutItem overwrites).
// SetOrder can be numeric ("1") or a marker like "W" (warmup) or "F" (failure).
func LiftID(date string, exerciseName string, setOrder string) string {
	return fmt.Sprintf("lift#%s#%s#%s", CompactDate(date), ExerciseSlug(exerciseName), strings.ToLower(setOrder))
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
