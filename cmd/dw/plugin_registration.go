package main

import (
	"context"
	"fmt"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

// loggerAdapter adapts app.Logger to domain.Logger (SDK interface)
type loggerAdapter struct {
	inner app.Logger
}

func (l *loggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	l.inner.Debug(msg, keysAndValues...)
}

func (l *loggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	l.inner.Info(msg, keysAndValues...)
}

func (l *loggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	l.inner.Warn(msg, keysAndValues...)
}

func (l *loggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	l.inner.Error(msg, keysAndValues...)
}

// logsServiceAdapter adapts app.LogsService to claude_code.LogsService
type logsServiceAdapter struct {
	inner *app.LogsService
}

func (a *logsServiceAdapter) ListRecentLogs(ctx context.Context, limit, offset int, sessionID string, ordered bool) ([]*claude_code.LogRecord, error) {
	appLogs, err := a.inner.ListRecentLogs(ctx, limit, offset, sessionID, ordered)
	if err != nil {
		return nil, err
	}

	// Convert app.LogRecord to claude_code.LogRecord
	pluginLogs := make([]*claude_code.LogRecord, len(appLogs))
	for i, log := range appLogs {
		pluginLogs[i] = &claude_code.LogRecord{
			ID:        log.ID,
			Timestamp: log.Timestamp,
			EventType: log.EventType,
			SessionID: log.SessionID,
			Payload:   log.Payload,
			Content:   log.Content,
		}
	}

	return pluginLogs, nil
}

// RegisterBuiltInPlugins registers all built-in plugins with the registry.
// This function lives in cmd layer to avoid app layer importing plugins.
//
// Built-in plugins are constructed here with access to internal services,
// but they implement the SDK Plugin interface and can only return SDK types
// from their public methods.
func RegisterBuiltInPlugins(
	registry *app.PluginRegistry,
	analysisService *app.AnalysisService,
	logsService *app.LogsService,
	logger app.Logger,
	setupService *app.SetupService,
	handler *app.ClaudeCommandHandler,
	dbPath string,
) error {
	// Create plugin context (SDK logger adapter)
	sdkLogger := &loggerAdapter{inner: logger}

	// Create service adapters
	logsAdapter := &logsServiceAdapter{inner: logsService}

	// Register claude-code plugin
	// Note: Built-in plugins can receive internal services during construction,
	// but their public interface uses only SDK types
	claudePlugin := claude_code.NewClaudeCodePlugin(
		analysisService,   // Implements claude_code.AnalysisService
		logsAdapter,       // Adapter to claude_code.LogsService
		sdkLogger,         // SDK logger
		setupService,      // Implements claude_code.SetupService
		handler,           // Implements claude_code.ClaudeCommandHandler
		dbPath,
	)

	if err := registry.RegisterPlugin(claudePlugin); err != nil {
		return fmt.Errorf("failed to register claude-code plugin: %w", err)
	}

	logger.Info("Registered built-in plugin: claude-code")

	// Future: Register other built-in plugins here

	return nil
}
