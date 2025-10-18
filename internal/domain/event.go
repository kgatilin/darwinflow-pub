package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event captured from Claude Code
type EventType string

const (
	// Chat events
	ChatStarted          EventType = "chat.started"
	ChatMessageUser      EventType = "chat.message.user"
	ChatMessageAssistant EventType = "chat.message.assistant"

	// Tool events
	ToolInvoked EventType = "tool.invoked"
	ToolResult  EventType = "tool.result"

	// File events
	FileRead    EventType = "file.read"
	FileWritten EventType = "file.written"

	// Context events
	ContextChanged EventType = "context.changed"

	// Error events
	Error EventType = "error"
)

// Event represents a single logged interaction from Claude Code (domain entity)
type Event struct {
	ID        string
	Timestamp time.Time
	Type      EventType
	SessionID string      // Claude Code session identifier
	Payload   interface{}
	Content   string // Normalized text for full-text search
}

// NewEvent creates a new event with generated ID and current timestamp (domain service)
func NewEvent(eventType EventType, sessionID string, payload interface{}, content string) *Event {
	return &Event{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Type:      eventType,
		SessionID: sessionID,
		Payload:   payload,
		Content:   content,
	}
}

// MarshalPayload converts the payload to JSON bytes for storage
func (e *Event) MarshalPayload() ([]byte, error) {
	return json.Marshal(e.Payload)
}

// Payload types for different events (value objects)

// ChatPayload contains data for chat-related events
type ChatPayload struct {
	Message string `json:"message,omitempty"`
	Context string `json:"context,omitempty"`
}

// ToolPayload contains data for tool invocation and result events
type ToolPayload struct {
	Tool       string      `json:"tool"`
	Parameters interface{} `json:"parameters,omitempty"` // Can be object, array, or string
	Result     interface{} `json:"result,omitempty"`     // Can be object, array, or string
	DurationMs int64       `json:"duration_ms,omitempty"`
	Context    string      `json:"context,omitempty"`
}

// FilePayload contains data for file access events
type FilePayload struct {
	FilePath   string `json:"file_path"`
	Changes    string `json:"changes,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Context    string `json:"context,omitempty"`
}

// ContextPayload contains data for context change events
type ContextPayload struct {
	Context     string `json:"context"`
	Description string `json:"description,omitempty"`
}

// ErrorPayload contains data for error events
type ErrorPayload struct {
	Error      string `json:"error"`
	StackTrace string `json:"stack_trace,omitempty"`
	Context    string `json:"context,omitempty"`
}
