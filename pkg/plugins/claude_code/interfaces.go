package claude_code

import (
	"context"
	"encoding/json"
	"time"
)

// Service interfaces that the plugin requires.
// The app layer will provide implementations of these interfaces.

// LogRecord represents a single log event (mirrors app.LogRecord)
type LogRecord struct {
	ID        string
	Timestamp time.Time
	EventType string
	SessionID string
	Payload   json.RawMessage
	Content   string
}

// SessionAnalysis represents a session analysis result (SDK-native type)
type SessionAnalysis struct {
	ID              string
	SessionID       string
	PromptName      string
	ModelUsed       string
	PatternsSummary string
	AnalyzedAt      time.Time
}

// AnalysisService provides access to session analysis data
type AnalysisService interface {
	GetAllSessionIDs(ctx context.Context, limit int) ([]string, error)
	GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*SessionAnalysis, error)
	EstimateTokenCount(ctx context.Context, sessionID string) (int, error)
	GetLastSession(ctx context.Context) (string, error)
	AnalyzeSessionWithPrompt(ctx context.Context, sessionID string, promptName string) (*SessionAnalysis, error)
}

// LogsService provides access to event logs from the event repository.
//
// This interface is critical to the event sourcing pattern: it abstracts access
// to the authoritative event store. Implementations query the event repository
// to fetch events, which are the source of truth for all derived state.
//
// Key characteristics:
//   - ListRecentLogs queries the event repository to retrieve events for a session
//   - Events are ordered chronologically when ordered=true
//   - Sessions are reconstructed from events, not stored directly
//   - The event stream is append-only (events are never modified or deleted)
//
// Implementation note: The app layer provides concrete implementation via
// app.LogsService which wraps domain.EventRepository. See cmd/dw/plugin_registration.go
// for the adapter pattern that wires LogsService into the plugin.
type LogsService interface {
	ListRecentLogs(ctx context.Context, limit, offset int, sessionID string, ordered bool) ([]*LogRecord, error)
}

// SetupService provides setup functionality
type SetupService interface {
	Initialize(ctx context.Context, dbPath string) error
}

// ConfigLoader provides access to configuration
type ConfigLoader interface {
	LoadConfig(path string) (*Config, error)
}

// Config represents the application configuration
type Config struct {
	Analysis AnalysisConfig `yaml:"analysis"`
}

// AnalysisConfig contains analysis-related configuration
type AnalysisConfig struct {
	AutoSummaryEnabled bool   `yaml:"auto_summary_enabled"`
	AutoSummaryPrompt  string `yaml:"auto_summary_prompt"`
}

// HookInputData represents data from Claude Code hooks
// Contains all fields extracted from Claude Code's native hook input format
type HookInputData struct {
	SessionID      string
	TranscriptPath string
	CWD            string
	PermissionMode string
	HookEventName  string
	ToolName       string
	ToolInput      map[string]interface{}
	ToolOutput     interface{}
	Error          interface{}
	UserMessage    string
}

// HookInputParser parses hook input from stdin
type HookInputParser interface {
	Parse(data []byte) (*HookInputData, error)
}

// TimeProvider for session entities
type TimeProvider interface {
	Now() time.Time
}
