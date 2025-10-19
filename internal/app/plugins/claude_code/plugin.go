package claude_code

import (
	"context"
	"fmt"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// ClaudeCodePlugin provides Claude Code sessions as entities.
// This is a core plugin that ships with DarwinFlow.
type ClaudeCodePlugin struct {
	analysisService *app.AnalysisService
	logsService     *app.LogsService
	logger          app.Logger
}

// NewClaudeCodePlugin creates a new Claude Code plugin
func NewClaudeCodePlugin(
	analysisService *app.AnalysisService,
	logsService *app.LogsService,
	logger app.Logger,
) *ClaudeCodePlugin {
	return &ClaudeCodePlugin{
		analysisService: analysisService,
		logsService:     logsService,
		logger:          logger,
	}
}

// GetInfo returns metadata about this plugin
func (p *ClaudeCodePlugin) GetInfo() domain.PluginInfo {
	return domain.PluginInfo{
		Name:        "claude-code",
		Version:     "1.0.0",
		Description: "Claude Code session tracking and analysis",
		IsCore:      true,
	}
}

// GetEntityTypes returns the entity types this plugin provides
func (p *ClaudeCodePlugin) GetEntityTypes() []domain.EntityTypeInfo {
	return []domain.EntityTypeInfo{
		{
			Type:              "session",
			DisplayName:       "Claude Session",
			DisplayNamePlural: "Claude Sessions",
			Capabilities:      []string{"IExtensible", "IHasContext", "ITrackable"},
			Icon:              "💬",
		},
	}
}

// Query returns entities matching the given query
func (p *ClaudeCodePlugin) Query(ctx context.Context, query domain.EntityQuery) ([]domain.IExtensible, error) {
	// Get all session IDs
	sessionIDs, err := p.analysisService.GetAllSessionIDs(ctx, query.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get session IDs: %w", err)
	}

	// Apply offset if specified
	if query.Offset > 0 {
		if query.Offset >= len(sessionIDs) {
			return []domain.IExtensible{}, nil
		}
		sessionIDs = sessionIDs[query.Offset:]
	}

	// Build entities
	entities := make([]domain.IExtensible, 0, len(sessionIDs))

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

// GetEntity retrieves a single entity by ID
func (p *ClaudeCodePlugin) GetEntity(ctx context.Context, entityID string) (domain.IExtensible, error) {
	return p.buildSessionEntity(ctx, entityID)
}

// UpdateEntity updates an entity's fields
// For sessions, this is currently not supported (read-only)
func (p *ClaudeCodePlugin) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (domain.IExtensible, error) {
	return nil, fmt.Errorf("sessions are read-only")
}

// buildSessionEntity constructs a SessionEntity from a session ID
func (p *ClaudeCodePlugin) buildSessionEntity(ctx context.Context, sessionID string) (*SessionEntity, error) {
	// Get session logs to extract metadata
	logs, err := p.logsService.ListRecentLogs(ctx, 0, 0, sessionID, true)
	if err != nil || len(logs) == 0 {
		return nil, fmt.Errorf("failed to get logs for session %s: %w", sessionID, err)
	}

	// Get analyses for this session
	analyses, err := p.analysisService.GetAnalysesBySessionID(ctx, sessionID)
	if err != nil {
		analyses = []*domain.SessionAnalysis{}
	}

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

// matchesFilters checks if an entity matches the given filters
func (p *ClaudeCodePlugin) matchesFilters(entity domain.IExtensible, filters map[string]interface{}) bool {
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
