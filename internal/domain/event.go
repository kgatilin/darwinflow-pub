package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event as a string
// Event types are defined by plugins (e.g., "claude.tool.invoked")
type EventType = string

// Event represents a single logged interaction from Claude Code (domain entity)
// This is the internal event storage format used by the framework.
type Event struct {
	ID        string
	Timestamp time.Time
	Type      string      // Event type (e.g., "claude.tool.invoked")
	SessionID string      // Claude Code session identifier
	Payload   interface{} // Plugin-specific payload structure
	Content   string      // Normalized text for full-text search
	Version   string      // Schema version for event (default: "1.0")
}

// NewEvent creates a new event with generated ID and current timestamp (domain service)
func NewEvent(eventType string, sessionID string, payload interface{}, content string) *Event {
	return &Event{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Type:      eventType,
		SessionID: sessionID,
		Payload:   payload,
		Content:   content,
		Version:   "1.0",
	}
}

// MarshalPayload converts the payload to JSON bytes for storage
func (e *Event) MarshalPayload() ([]byte, error) {
	return json.Marshal(e.Payload)
}
