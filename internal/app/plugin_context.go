package app

import (
	"context"
	"fmt"
	"io"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// ProjectContext provides access to app-layer services for plugin tools.
// This is the internal context used by the tool registry (not SDK).
// Renamed from PluginContext to avoid confusion with SDK PluginContext.
type ProjectContext struct {
	// EventRepo provides access to logged events
	EventRepo domain.EventRepository

	// AnalysisRepo provides access to session analyses
	AnalysisRepo domain.AnalysisRepository

	// Config is the project's configuration
	Config *domain.Config

	// CWD is the current working directory
	CWD string

	// DBPath is the path to the database
	DBPath string
}

// pluginContextAdapter adapts internal services to SDK PluginContext interface.
// This allows plugins to access system capabilities without depending on internal types.
type pluginContextAdapter struct {
	logger     Logger
	dbPath     string
	workingDir string
	eventRepo  domain.EventRepository
}

// NewPluginContext creates a new plugin context adapter
func NewPluginContext(logger Logger, dbPath, workingDir string, eventRepo domain.EventRepository) domain.PluginContext {
	return &pluginContextAdapter{
		logger:     logger,
		dbPath:     dbPath,
		workingDir: workingDir,
		eventRepo:  eventRepo,
	}
}

func (p *pluginContextAdapter) GetLogger() domain.Logger {
	return &loggerAdapter{inner: p.logger}
}

func (p *pluginContextAdapter) GetWorkingDir() string {
	return p.workingDir
}

func (p *pluginContextAdapter) EmitEvent(ctx context.Context, event domain.PluginEvent) error {
	// Convert SDK event to domain event
	// SDK Event has: Type, Source, Timestamp, Payload (map[string]interface{}), Metadata (map[string]string)
	// Domain Event has: ID, Timestamp, Type (EventType), SessionID, Payload (interface{}), Content (string)

	// Extract session ID from metadata if present
	sessionID := ""
	if event.Metadata != nil {
		sessionID = event.Metadata["session_id"]
	}

	// Combine type and source for event type
	// Use the event.Type directly as it's already in dot notation
	eventType := domain.EventType(event.Type)

	// Build payload that includes both the event payload and source information
	payload := map[string]interface{}{
		"source": event.Source,
		"data":   event.Payload,
	}
	if event.Metadata != nil {
		payload["metadata"] = event.Metadata
	}

	// Create normalized content for full-text search
	// Combine type, source, and payload fields
	contentParts := []string{string(eventType), event.Source}
	for _, v := range event.Payload {
		contentParts = append(contentParts, fmt.Sprintf("%v", v))
	}
	content := fmt.Sprintf("%s", contentParts)

	// Create domain event
	domainEvent := domain.NewEvent(eventType, sessionID, payload, content)

	// Save to repository
	if err := p.eventRepo.Save(ctx, domainEvent); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

// loggerAdapter adapts app.Logger to domain.Logger
type loggerAdapter struct {
	inner Logger
}

func (l *loggerAdapter) Debug(format string, args ...interface{}) {
	l.inner.Debug(format, args...)
}

func (l *loggerAdapter) Info(format string, args ...interface{}) {
	l.inner.Info(format, args...)
}

func (l *loggerAdapter) Warn(format string, args ...interface{}) {
	l.inner.Warn(format, args...)
}

func (l *loggerAdapter) Error(format string, args ...interface{}) {
	l.inner.Error(format, args...)
}

// commandContextAdapter adapts internal services to SDK CommandContext interface
type commandContextAdapter struct {
	pluginContextAdapter
	output io.Writer
	input  io.Reader
}

// NewCommandContext creates a new command context adapter
func NewCommandContext(logger Logger, dbPath, workingDir string, eventRepo interface{}, output io.Writer, input io.Reader) domain.CommandContext {
	return &commandContextAdapter{
		pluginContextAdapter: pluginContextAdapter{
			logger:     logger,
			dbPath:     dbPath,
			workingDir: workingDir,
			eventRepo:  eventRepo.(domain.EventRepository),
		},
		output: output,
		input:  input,
	}
}

func (c *commandContextAdapter) GetStdout() io.Writer {
	return c.output
}

func (c *commandContextAdapter) GetStdin() io.Reader {
	return c.input
}

// Note: ToolContext removed - tools now use regular context
// Tools are executed via the Tool interface which receives context.Context and args
