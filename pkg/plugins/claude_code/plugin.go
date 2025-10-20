package claude_code

import (
	"context"
	"fmt"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ClaudeCodePlugin provides Claude Code sessions as entities.
// This is a core plugin that ships with DarwinFlow.
//
// NOTE: This plugin is built-in and has access to service interfaces.
// External plugins would only have access to PluginContext from the SDK.
type ClaudeCodePlugin struct {
	// Service interfaces (injected by app layer for built-in plugins)
	analysisService AnalysisService
	logsService     LogsService
	logger          pluginsdk.Logger // Use SDK logger
	setupService    SetupService
	handler         ClaudeCommandHandler
	dbPath          string
}

// NewClaudeCodePlugin creates a new Claude Code plugin
//
// For built-in plugins, we inject service implementations.
// External plugins would receive only PluginContext.
func NewClaudeCodePlugin(
	analysisService AnalysisService,
	logsService LogsService,
	logger pluginsdk.Logger,
	setupService SetupService,
	handler ClaudeCommandHandler,
	dbPath string,
) *ClaudeCodePlugin {
	return &ClaudeCodePlugin{
		analysisService: analysisService,
		logsService:     logsService,
		logger:          logger,
		setupService:    setupService,
		handler:         handler,
		dbPath:          dbPath,
	}
}

// GetInfo returns metadata about this plugin (SDK interface)
func (p *ClaudeCodePlugin) GetInfo() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:        "claude-code",
		Version:     "1.0.0",
		Description: "Claude Code session tracking and analysis",
		IsCore:      true,
	}
}

// GetEntityTypes returns the entity types this plugin provides (SDK interface)
func (p *ClaudeCodePlugin) GetEntityTypes() []pluginsdk.EntityTypeInfo {
	return []pluginsdk.EntityTypeInfo{
		{
			Type:              "session",
			DisplayName:       "Claude Session",
			DisplayNamePlural: "Claude Sessions",
			Capabilities:      []string{"IExtensible", "IHasContext", "ITrackable"},
			Icon:              "💬",
		},
	}
}

// Query returns entities matching the given query (SDK interface)
func (p *ClaudeCodePlugin) Query(ctx context.Context, query pluginsdk.EntityQuery) ([]pluginsdk.IExtensible, error) {
	// Get all session IDs
	sessionIDs, err := p.analysisService.GetAllSessionIDs(ctx, query.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get session IDs: %w", err)
	}

	// Apply offset if specified
	if query.Offset > 0 {
		if query.Offset >= len(sessionIDs) {
			return []pluginsdk.IExtensible{}, nil
		}
		sessionIDs = sessionIDs[query.Offset:]
	}

	// Build entities
	entities := make([]pluginsdk.IExtensible, 0, len(sessionIDs))

	for _, sessionID := range sessionIDs {
		entity, err := p.buildSessionEntity(ctx, sessionID)
		if err != nil {
			p.logger.Warn("Failed to build session entity %s: %v", sessionID, err)
			continue
		}

		// Apply filters if specified
		if !p.matchesFilters(entity, query.Filters) {
			continue
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

// GetEntity retrieves a single entity by ID (SDK interface)
func (p *ClaudeCodePlugin) GetEntity(ctx context.Context, entityID string) (pluginsdk.IExtensible, error) {
	return p.buildSessionEntity(ctx, entityID)
}

// UpdateEntity updates an entity's fields (SDK interface)
// For sessions, this is currently not supported (read-only)
func (p *ClaudeCodePlugin) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (pluginsdk.IExtensible, error) {
	return nil, pluginsdk.ErrReadOnly
}

// buildSessionEntity constructs a SessionEntity from a session ID
func (p *ClaudeCodePlugin) buildSessionEntity(ctx context.Context, sessionID string) (*SessionEntity, error) {
	// Get session logs to extract metadata
	logs, err := p.logsService.ListRecentLogs(ctx, 0, 0, sessionID, true)
	if err != nil || len(logs) == 0 {
		return nil, fmt.Errorf("failed to get logs for session %s: %w", sessionID, err)
	}

	// Get analyses for this session
	domainAnalyses, err := p.analysisService.GetAnalysesBySessionID(ctx, sessionID)
	if err != nil {
		domainAnalyses = []*domain.SessionAnalysis{}
	}

	// Convert domain analyses to SDK analyses
	analyses := convertAnalyses(domainAnalyses)

	// Estimate token count for the session
	tokenCount, err := p.analysisService.EstimateTokenCount(ctx, sessionID)
	if err != nil {
		tokenCount = 0
	}

	return NewSessionEntity(
		sessionID,
		logs[0].Timestamp,
		logs[len(logs)-1].Timestamp,
		len(logs),
		analyses,
		tokenCount,
	), nil
}

// convertAnalyses converts domain SessionAnalysis to SDK SessionAnalysisData
func convertAnalyses(domainAnalyses []*domain.SessionAnalysis) []SessionAnalysisData {
	if domainAnalyses == nil {
		return []SessionAnalysisData{}
	}

	analyses := make([]SessionAnalysisData, len(domainAnalyses))
	for i, da := range domainAnalyses {
		analyses[i] = SessionAnalysisData{
			ID:              da.ID,
			SessionID:       da.SessionID,
			PromptName:      da.PromptName,
			ModelUsed:       da.ModelUsed,
			PatternsSummary: da.PatternsSummary,
			CreatedAt:       da.AnalyzedAt, // Field is called AnalyzedAt in domain
		}
	}
	return analyses
}

// matchesFilters checks if an entity matches the given filters
func (p *ClaudeCodePlugin) matchesFilters(entity pluginsdk.IExtensible, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}

	for key, expectedValue := range filters {
		actualValue := entity.GetField(key)
		if actualValue != expectedValue {
			return false
		}
	}

	return true
}
