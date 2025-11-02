package task_manager

import (
	"encoding/json"
	"time"
)

// TaskEntity represents a task and implements SDK capability interfaces.
// It implements IExtensible and ITrackable interfaces.
type TaskEntity struct {
	ID          string    `json:"id"`
	TrackID     string    `json:"track_id"` // Parent track ID
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // todo, in-progress, done
	Rank        int       `json:"rank"` // 1-1000 (lower = higher priority)
	Branch      string    `json:"branch"` // Git branch name (optional)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewTaskEntity creates a new task entity
func NewTaskEntity(id, trackID, title, description, status string, rank int, branch string, createdAt, updatedAt time.Time) *TaskEntity {
	return &TaskEntity{
		ID:          id,
		TrackID:     trackID,
		Title:       title,
		Description: description,
		Status:      status,
		Rank:        rank,
		Branch:      branch,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// IExtensible implementation

// GetID returns the unique identifier for this entity
func (t *TaskEntity) GetID() string {
	return t.ID
}

// GetType returns the entity type
func (t *TaskEntity) GetType() string {
	return "task"
}

// GetCapabilities returns list of capability names this entity supports
func (t *TaskEntity) GetCapabilities() []string {
	return []string{"IExtensible", "ITrackable"}
}

// GetField retrieves a named field value
func (t *TaskEntity) GetField(name string) interface{} {
	fields := t.GetAllFields()
	return fields[name]
}

// GetAllFields returns all fields as a map
func (t *TaskEntity) GetAllFields() map[string]interface{} {
	return map[string]interface{}{
		"id":          t.ID,
		"track_id":    t.TrackID,
		"title":       t.Title,
		"description": t.Description,
		"status":      t.Status,
		"rank":        t.Rank,
		"branch":      t.Branch,
		"created_at":  t.CreatedAt,
		"updated_at":  t.UpdatedAt,
		"progress":    t.GetProgress(),
		"is_blocked":  t.IsBlocked(),
	}
}

// ITrackable implementation

// GetStatus returns the current status (todo, in-progress, done)
func (t *TaskEntity) GetStatus() string {
	return t.Status
}

// GetProgress returns completion progress as a value between 0.0 and 1.0
func (t *TaskEntity) GetProgress() float64 {
	switch t.Status {
	case "done":
		return 1.0
	case "in-progress":
		return 0.5
	default: // todo
		return 0.0
	}
}

// IsBlocked returns true if the entity is blocked from progressing
func (t *TaskEntity) IsBlocked() bool {
	// Tasks are never blocked in this implementation
	return false
}

// GetBlockReason returns the reason for blocking, or empty string if not blocked
func (t *TaskEntity) GetBlockReason() string {
	return ""
}

// MarshalTask serializes a task to JSON bytes with indentation
func MarshalTask(t *TaskEntity) ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}
