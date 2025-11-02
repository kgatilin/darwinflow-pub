package task_manager

import (
	"fmt"
	"regexp"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// TrackEntity represents a major work area/track and implements SDK capability interfaces.
// It implements IExtensible and ITrackable interfaces.
type TrackEntity struct {
	ID           string    `json:"id"`
	RoadmapID    string    `json:"roadmap_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"` // not-started, in-progress, complete, blocked, waiting
	Rank         int       `json:"rank"` // 1-1000 (lower = higher priority)
	Dependencies []string  `json:"dependencies"` // Track IDs this depends on
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Valid status values for tracks
var validTrackStatuses = map[string]bool{
	"not-started": true,
	"in-progress": true,
	"complete":    true,
	"blocked":     true,
	"waiting":     true,
}

// NewTrackEntity creates a new track entity with validation
func NewTrackEntity(id, roadmapID, title, description, status string, rank int, dependencies []string, createdAt, updatedAt time.Time) (*TrackEntity, error) {
	// Validate track ID format
	if !isValidTrackID(id) {
		return nil, fmt.Errorf("%w: track ID must follow convention: track-<slug>", pluginsdk.ErrInvalidArgument)
	}

	// Validate status
	if !validTrackStatuses[status] {
		return nil, fmt.Errorf("%w: invalid track status: must be one of not-started, in-progress, complete, blocked, waiting", pluginsdk.ErrInvalidArgument)
	}

	// Validate rank
	if rank < 1 || rank > 1000 {
		return nil, fmt.Errorf("%w: invalid track rank: must be between 1 and 1000", pluginsdk.ErrInvalidArgument)
	}

	// Check for self-dependency
	for _, dep := range dependencies {
		if dep == id {
			return nil, fmt.Errorf("%w: track cannot depend on itself", pluginsdk.ErrInvalidArgument)
		}
	}

	if dependencies == nil {
		dependencies = []string{}
	}

	return &TrackEntity{
		ID:           id,
		RoadmapID:    roadmapID,
		Title:        title,
		Description:  description,
		Status:       status,
		Rank:         rank,
		Dependencies: dependencies,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}

// isValidTrackID validates track ID format
// Accepts both old format (track-<slug>) and new format (<CODE>-track-<number>)
func isValidTrackID(id string) bool {
	// New format: <CODE>-track-<number> (e.g., DW-track-1, PROD-track-123)
	newPattern := `^[A-Z0-9]+-track-[0-9]+$`
	newRegex := regexp.MustCompile(newPattern)
	if newRegex.MatchString(id) {
		return true
	}

	// Old format: track-<slug> (for backward compatibility)
	oldPattern := `^track-[a-z0-9]+(-[a-z0-9]+)*$`
	oldRegex := regexp.MustCompile(oldPattern)
	return oldRegex.MatchString(id)
}

// IExtensible implementation

// GetID returns the unique identifier for this entity
func (t *TrackEntity) GetID() string {
	return t.ID
}

// GetType returns the entity type
func (t *TrackEntity) GetType() string {
	return "track"
}

// GetCapabilities returns list of capability names this entity supports
func (t *TrackEntity) GetCapabilities() []string {
	return []string{"IExtensible", "ITrackable"}
}

// GetField retrieves a named field value
func (t *TrackEntity) GetField(name string) interface{} {
	fields := t.GetAllFields()
	return fields[name]
}

// GetAllFields returns all fields as a map
func (t *TrackEntity) GetAllFields() map[string]interface{} {
	return map[string]interface{}{
		"id":           t.ID,
		"roadmap_id":   t.RoadmapID,
		"title":        t.Title,
		"description":  t.Description,
		"status":       t.Status,
		"rank":         t.Rank,
		"dependencies": t.Dependencies,
		"created_at":   t.CreatedAt,
		"updated_at":   t.UpdatedAt,
		"progress":     t.GetProgress(),
		"is_blocked":   t.IsBlocked(),
	}
}

// ITrackable implementation

// GetStatus returns the current status
func (t *TrackEntity) GetStatus() string {
	return t.Status
}

// GetProgress returns completion progress as a value between 0.0 and 1.0
// For now, returns 0.0 as we don't have child tasks yet
func (t *TrackEntity) GetProgress() float64 {
	// This will be calculated based on child tasks in Phase 2
	switch t.Status {
	case "complete":
		return 1.0
	case "in-progress":
		return 0.5
	default: // not-started, blocked, waiting
		return 0.0
	}
}

// IsBlocked returns true if the entity is blocked from progressing
func (t *TrackEntity) IsBlocked() bool {
	return t.Status == "blocked"
}

// GetBlockReason returns the reason for blocking, or empty string if not blocked
func (t *TrackEntity) GetBlockReason() string {
	if t.IsBlocked() {
		return fmt.Sprintf("Track %s is blocked", t.ID)
	}
	return ""
}

// AddDependency adds a track dependency with validation
func (t *TrackEntity) AddDependency(trackID string) error {
	// Check for self-dependency
	if trackID == t.ID {
		return fmt.Errorf("%w: track cannot depend on itself", pluginsdk.ErrInvalidArgument)
	}

	// Check if dependency already exists
	for _, dep := range t.Dependencies {
		if dep == trackID {
			return fmt.Errorf("%w: track %s already depends on %s", pluginsdk.ErrAlreadyExists, t.ID, trackID)
		}
	}

	t.Dependencies = append(t.Dependencies, trackID)
	return nil
}

// RemoveDependency removes a track dependency
func (t *TrackEntity) RemoveDependency(trackID string) error {
	newDeps := []string{}
	found := false
	for _, dep := range t.Dependencies {
		if dep != trackID {
			newDeps = append(newDeps, dep)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("%w: track %s does not depend on %s", pluginsdk.ErrNotFound, t.ID, trackID)
	}

	t.Dependencies = newDeps
	return nil
}

// HasDependency checks if this track depends on another
func (t *TrackEntity) HasDependency(trackID string) bool {
	for _, dep := range t.Dependencies {
		if dep == trackID {
			return true
		}
	}
	return false
}
