package claude_code

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
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

// AnalysisService provides access to session analysis data
type AnalysisService interface {
	GetAllSessionIDs(ctx context.Context, limit int) ([]string, error)
	GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionAnalysis, error)
	EstimateTokenCount(ctx context.Context, sessionID string) (int, error)
	GetLastSession(ctx context.Context) (string, error)
}

// LogsService provides access to event logs
type LogsService interface {
	ListRecentLogs(ctx context.Context, limit, offset int, sessionID string, ordered bool) ([]*LogRecord, error)
}

// SetupService provides setup functionality
type SetupService interface {
	// Methods needed by setup commands (if any)
}

// ClaudeCommandHandler handles claude-specific commands
type ClaudeCommandHandler interface {
	Init(ctx context.Context, dbPath string) error
	Log(ctx context.Context, eventType string, stdinData []byte, maxParamLength int) error
	AutoSummary(ctx context.Context, stdinData []byte) error
	AutoSummaryExec(ctx context.Context, sessionID string) error
}

// TimeProvider for session entities
type TimeProvider interface {
	Now() time.Time
}
