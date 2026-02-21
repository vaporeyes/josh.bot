// ABOUTME: This file defines the core domain models and service interfaces for the josh.bot.
// ABOUTME: It represents the central business logic of the application.
package domain

import (
	"context"
	"crypto/rand"
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

// Note represents a captured text note for later consumption.
type Note struct {
	ID        string   `json:"id" dynamodbav:"id"`
	Title     string   `json:"title" dynamodbav:"title"`
	Body      string   `json:"body" dynamodbav:"body"`
	Tags      []string `json:"tags" dynamodbav:"tags"`
	CreatedAt string   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt string   `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
}

// NoteID generates a random ID with a "note#" prefix.
// AIDEV-NOTE: Random IDs (not deterministic) since notes have no natural dedup key.
func NoteID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "note#" + hex.EncodeToString(b)
}

// TIL represents a "Today I Learned" entry.
type TIL struct {
	ID        string   `json:"id" dynamodbav:"id"`
	Title     string   `json:"title" dynamodbav:"title"`
	Body      string   `json:"body" dynamodbav:"body"`
	Tags      []string `json:"tags" dynamodbav:"tags"`
	CreatedAt string   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt string   `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
}

// TILID generates a random ID with a "til#" prefix.
func TILID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "til#" + hex.EncodeToString(b)
}

// LogEntry represents a timestamped activity/event log entry.
type LogEntry struct {
	ID        string   `json:"id" dynamodbav:"id"`
	Message   string   `json:"message" dynamodbav:"message"`
	Tags      []string `json:"tags" dynamodbav:"tags"`
	CreatedAt string   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt string   `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
}

// LogEntryID generates a random ID with a "log#" prefix.
func LogEntryID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "log#" + hex.EncodeToString(b)
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

// DiaryEntry represents a structured journal entry with context, reaction, and takeaway.
// AIDEV-NOTE: Four fields match the journaling spec: Context, Body, Reaction, Takeaway.
type DiaryEntry struct {
	ID        string   `json:"id" dynamodbav:"id"`
	Title     string   `json:"title,omitempty" dynamodbav:"title,omitempty"`
	Context   string   `json:"context" dynamodbav:"context"`
	Body      string   `json:"body" dynamodbav:"body"`
	Reaction  string   `json:"reaction" dynamodbav:"reaction"`
	Takeaway  string   `json:"takeaway" dynamodbav:"takeaway"`
	Tags      []string `json:"tags" dynamodbav:"tags"`
	CreatedAt string   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt string   `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
}

// DiaryEntryID generates a random ID with a "diary#" prefix.
// AIDEV-NOTE: Random IDs (not deterministic) since diary entries have no natural dedup key.
func DiaryEntryID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "diary#" + hex.EncodeToString(b)
}

// IdempotencyRecord stores the result of a POST request for deduplication.
// AIDEV-NOTE: Stored as idem#<path>#<key> in josh-bot-data with DynamoDB TTL on expires_at.
type IdempotencyRecord struct {
	ID         string `json:"id" dynamodbav:"id"`
	StatusCode int    `json:"status_code" dynamodbav:"status_code"`
	Body       string `json:"body" dynamodbav:"body"`
	ExpiresAt  int64  `json:"expires_at" dynamodbav:"expires_at"`
	CreatedAt  string `json:"created_at" dynamodbav:"created_at"`
}

// IdempotencyKey builds a deterministic DynamoDB key from the request path and client-provided key.
func IdempotencyKey(path, key string) string {
	return fmt.Sprintf("idem#%s#%s", path, key)
}

// --- Validation ---

// Validate checks required fields on a Project.
func (p Project) Validate() error {
	if p.Slug == "" {
		return &ValidationError{Field: "slug", Message: "cannot be empty"}
	}
	if p.Name == "" {
		return &ValidationError{Field: "name", Message: "cannot be empty"}
	}
	return nil
}

// Validate checks required fields on a Link.
func (l Link) Validate() error {
	if l.URL == "" {
		return &ValidationError{Field: "url", Message: "cannot be empty"}
	}
	return nil
}

// Validate checks required fields on a Note.
func (n Note) Validate() error {
	if n.Title == "" {
		return &ValidationError{Field: "title", Message: "cannot be empty"}
	}
	if n.Body == "" {
		return &ValidationError{Field: "body", Message: "cannot be empty"}
	}
	return nil
}

// Validate checks required fields on a TIL.
func (t TIL) Validate() error {
	if t.Title == "" {
		return &ValidationError{Field: "title", Message: "cannot be empty"}
	}
	if t.Body == "" {
		return &ValidationError{Field: "body", Message: "cannot be empty"}
	}
	return nil
}

// Validate checks required fields on a LogEntry.
func (le LogEntry) Validate() error {
	if le.Message == "" {
		return &ValidationError{Field: "message", Message: "cannot be empty"}
	}
	return nil
}

// Validate checks required fields on a DiaryEntry.
func (de DiaryEntry) Validate() error {
	if de.Body == "" {
		return &ValidationError{Field: "body", Message: "cannot be empty"}
	}
	return nil
}

// BotService is the interface that defines the operations for the bot.
type BotService interface {
	GetStatus(ctx context.Context) (Status, error)
	GetProjects(ctx context.Context) ([]Project, error)
	GetProject(ctx context.Context, slug string) (Project, error)
	CreateProject(ctx context.Context, project Project) error
	UpdateProject(ctx context.Context, slug string, fields map[string]any) error
	DeleteProject(ctx context.Context, slug string) error
	UpdateStatus(ctx context.Context, fields map[string]any) error
	GetLinks(ctx context.Context, tag string) ([]Link, error)
	GetLink(ctx context.Context, id string) (Link, error)
	CreateLink(ctx context.Context, link Link) error
	UpdateLink(ctx context.Context, id string, fields map[string]any) error
	DeleteLink(ctx context.Context, id string) error
	GetNotes(ctx context.Context, tag string) ([]Note, error)
	GetNote(ctx context.Context, id string) (Note, error)
	CreateNote(ctx context.Context, note Note) error
	UpdateNote(ctx context.Context, id string, fields map[string]any) error
	DeleteNote(ctx context.Context, id string) error
	GetTILs(ctx context.Context, tag string) ([]TIL, error)
	GetTIL(ctx context.Context, id string) (TIL, error)
	CreateTIL(ctx context.Context, til TIL) error
	UpdateTIL(ctx context.Context, id string, fields map[string]any) error
	DeleteTIL(ctx context.Context, id string) error
	GetLogEntries(ctx context.Context, tag string) ([]LogEntry, error)
	GetLogEntry(ctx context.Context, id string) (LogEntry, error)
	CreateLogEntry(ctx context.Context, entry LogEntry) error
	UpdateLogEntry(ctx context.Context, id string, fields map[string]any) error
	DeleteLogEntry(ctx context.Context, id string) error
	GetDiaryEntries(ctx context.Context, tag string) ([]DiaryEntry, error)
	GetDiaryEntry(ctx context.Context, id string) (DiaryEntry, error)
	CreateDiaryEntry(ctx context.Context, entry DiaryEntry) error
	UpdateDiaryEntry(ctx context.Context, id string, fields map[string]any) error
	DeleteDiaryEntry(ctx context.Context, id string) error
	GetIdempotencyRecord(ctx context.Context, key string) (*IdempotencyRecord, error)
	SetIdempotencyRecord(ctx context.Context, record IdempotencyRecord) error
}
