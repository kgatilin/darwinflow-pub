package app

import (
	"context"
	"fmt"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

// logsServiceAdapter adapts app.LogsService to claude_code.LogsService
type logsServiceAdapter struct {
	inner *LogsService
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
// This function lives in app layer so cmd doesn't need to import plugins.
//
// Built-in plugins are constructed here with access to internal services,
// but they implement the SDK Plugin interface and can only return SDK types
// from their public methods.
func RegisterBuiltInPlugins(
	registry *PluginRegistry,
	analysisService *AnalysisService,
	logsService *LogsService,
	logger Logger,
	setupService *SetupService,
	handler *ClaudeCommandHandler,
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
