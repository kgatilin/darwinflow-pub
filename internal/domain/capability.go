package domain

import "time"

// IExtensible is the base capability all entities must implement.
// It provides core identity and introspection for any entity in the system.
type IExtensible interface {
	// GetID returns the unique identifier for this entity
	GetID() string

	// GetType returns the entity type (e.g., "session", "task", "roadmap")
	GetType() string

	// GetCapabilities returns list of capability names this entity supports
	// (e.g., ["IExtensible", "IHasContext", "ITrackable"])
	GetCapabilities() []string

	// GetField retrieves a named field value
	// Returns nil if field doesn't exist
	GetField(name string) interface{}

	// GetAllFields returns all fields as a map
	GetAllFields() map[string]interface{}
}

// IHasContext is a capability for entities that have related data,
// such as linked files, related entities, or activity history.
type IHasContext interface {
	IExtensible

	// GetContext returns contextual information about this entity
	GetContext() *EntityContext
}

// EntityContext contains contextual information about an entity
type EntityContext struct {
	// RelatedEntities are other entities connected to this one
	// Key is entity type, value is list of entity IDs
	RelatedEntities map[string][]string

	// LinkedFiles are file paths referenced by this entity
	LinkedFiles []string

	// RecentActivity is a log of recent actions on this entity
	RecentActivity []ActivityRecord

	// Metadata for any additional context
	Metadata map[string]interface{}
}

// ActivityRecord represents a single activity event related to an entity
type ActivityRecord struct {
	// Timestamp is when the activity occurred
	Timestamp time.Time `json:"timestamp"`

	// Type is the kind of activity (e.g., "created", "updated", "analyzed")
	Type string `json:"type"`

	// Description is a human-readable description of the activity
	Description string `json:"description"`

	// Actor is who/what performed the activity (user, system, etc.)
	Actor string `json:"actor"`
}

// ITrackable is a capability for entities that have status and progress tracking.
type ITrackable interface {
	IExtensible

	// GetStatus returns the current status (e.g., "active", "completed", "blocked")
	GetStatus() string

	// GetProgress returns completion progress as a value between 0.0 and 1.0
	GetProgress() float64

	// IsBlocked returns true if the entity is blocked from progressing
	IsBlocked() bool

	// GetBlockReason returns the reason for blocking, or empty string if not blocked
	GetBlockReason() string
}

// ISchedulable is a capability for entities that have time-based scheduling.
type ISchedulable interface {
	IExtensible

	// GetStartDate returns when the entity should/did start
	GetStartDate() *time.Time

	// GetDueDate returns when the entity should be completed
	GetDueDate() *time.Time

	// IsOverdue returns true if past due date and not complete
	IsOverdue() bool
}

// IRelatable is a capability for entities that have explicit relationships with other entities.
type IRelatable interface {
	IExtensible

	// GetRelated returns IDs of related entities of the specified type
	GetRelated(entityType string) []string

	// GetAllRelations returns all relationships grouped by entity type
	GetAllRelations() map[string][]string
}
