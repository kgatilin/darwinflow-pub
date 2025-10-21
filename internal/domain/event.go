package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

// EventType represents the type of event captured from Claude Code
// NOTE: EventType constants are defined in the claude-code plugin
// These are re-exported here for backwards compatibility with internal domain code
type EventType = string

// Re-export event type constants from claude-code plugin
const (
	// Chat events
	ChatStarted          = claude_code.ChatStarted
	ChatMessageUser      = claude_code.ChatMessageUser
	ChatMessageAssistant = claude_code.ChatMessageAssistant

	// Tool events
	ToolInvoked = claude_code.ToolInvoked
	ToolResult  = claude_code.ToolResult

	// File events
	FileRead    = claude_code.FileRead
	FileWritten = claude_code.FileWritten

	// Context events
	ContextChanged = claude_code.ContextChanged

	// Error events
	Error = claude_code.Error
)

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

// Payload types for different events (value objects - re-exported from plugin)

// ChatPayload contains data for chat-related events
type ChatPayload = claude_code.ChatPayload

// ToolPayload contains data for tool invocation and result events
type ToolPayload = claude_code.ToolPayload

// FilePayload contains data for file access events
type FilePayload = claude_code.FilePayload

// ContextPayload contains data for context change events
type ContextPayload = claude_code.ContextPayload

// ErrorPayload contains data for error events
type ErrorPayload = claude_code.ErrorPayload
