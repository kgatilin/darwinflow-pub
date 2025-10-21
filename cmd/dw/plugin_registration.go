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

// analysisServiceAdapter adapts app.AnalysisService to claude_code.AnalysisService
type analysisServiceAdapter struct {
	inner *app.AnalysisService
}

func (a *analysisServiceAdapter) GetAllSessionIDs(ctx context.Context, limit int) ([]string, error) {
	return a.inner.GetAllSessionIDs(ctx, limit)
}

func (a *analysisServiceAdapter) GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*claude_code.SessionAnalysis, error) {
	domainAnalyses, err := a.inner.GetAnalysesBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Convert domain.SessionAnalysis to claude_code.SessionAnalysis
	pluginAnalyses := make([]*claude_code.SessionAnalysis, len(domainAnalyses))
	for i, da := range domainAnalyses {
		pluginAnalyses[i] = &claude_code.SessionAnalysis{
			ID:              da.ID,
			SessionID:       da.SessionID,
			PromptName:      da.PromptName,
			ModelUsed:       da.ModelUsed,
			PatternsSummary: da.PatternsSummary,
			AnalyzedAt:      da.AnalyzedAt,
		}
	}

	return pluginAnalyses, nil
}

func (a *analysisServiceAdapter) EstimateTokenCount(ctx context.Context, sessionID string) (int, error) {
	return a.inner.EstimateTokenCount(ctx, sessionID)
}

func (a *analysisServiceAdapter) GetLastSession(ctx context.Context) (string, error) {
	return a.inner.GetLastSession(ctx)
}

func (a *analysisServiceAdapter) AnalyzeSessionWithPrompt(ctx context.Context, sessionID string, promptName string) (*claude_code.SessionAnalysis, error) {
	domainAnalysis, err := a.inner.AnalyzeSessionWithPrompt(ctx, sessionID, promptName)
	if err != nil {
		return nil, err
	}

	return &claude_code.SessionAnalysis{
		ID:              domainAnalysis.ID,
		SessionID:       domainAnalysis.SessionID,
		PromptName:      domainAnalysis.PromptName,
		ModelUsed:       domainAnalysis.ModelUsed,
		PatternsSummary: domainAnalysis.PatternsSummary,
		AnalyzedAt:      domainAnalysis.AnalyzedAt,
	}, nil
}

// configLoaderAdapter adapts app.ConfigLoader to claude_code.ConfigLoader
type configLoaderAdapter struct {
	inner app.ConfigLoader
}

func (a *configLoaderAdapter) LoadConfig(path string) (*claude_code.Config, error) {
	domainConfig, err := a.inner.LoadConfig(path)
	if err != nil {
		return nil, err
	}

	return &claude_code.Config{
		Analysis: claude_code.AnalysisConfig{
			AutoSummaryEnabled: domainConfig.Analysis.AutoSummaryEnabled,
			AutoSummaryPrompt:  domainConfig.Analysis.AutoSummaryPrompt,
		},
		Logging: claude_code.LoggingConfig{
			FileLogLevel: domainConfig.Logging.FileLogLevel,
		},
	}, nil
}

// setupServiceAdapter adapts app.SetupService to claude_code.SetupService
type setupServiceAdapter struct {
	inner *app.SetupService
}

func (a *setupServiceAdapter) Initialize(ctx context.Context, dbPath string) error {
	return a.inner.Initialize(ctx, dbPath)
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
	configLoader app.ConfigLoader,
	dbPath string,
) error {
	// Create plugin context (SDK logger adapter)
	sdkLogger := &loggerAdapter{inner: logger}

	// Create service adapters
	logsAdapter := &logsServiceAdapter{inner: logsService}
	analysisAdapter := &analysisServiceAdapter{inner: analysisService}
	setupAdapter := &setupServiceAdapter{inner: setupService}
	configAdapter := &configLoaderAdapter{inner: configLoader}

	// Register claude-code plugin
	// Note: Built-in plugins can receive internal services during construction,
	// but their public interface uses only SDK types
	claudePlugin := claude_code.NewClaudeCodePlugin(
		analysisAdapter, // Adapter to claude_code.AnalysisService
		logsAdapter,     // Adapter to claude_code.LogsService
		sdkLogger,       // SDK logger
		setupAdapter,    // Adapter to claude_code.SetupService
		configAdapter,   // Adapter to claude_code.ConfigLoader
		dbPath,
	)

	if err := registry.RegisterPlugin(claudePlugin); err != nil {
		return fmt.Errorf("failed to register claude-code plugin: %w", err)
	}

	// Note: Plugin registration is logged by PluginRegistry.RegisterPlugin()
	// which provides detailed info (version, capabilities). No need to log again here.

	// Future: Register other built-in plugins here

	return nil
}
