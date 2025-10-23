package claude_code

import (
	"context"
	"fmt"

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
	logger          pluginsdk.Logger
	setupService    SetupService
	configLoader    ConfigLoader
	hookInputParser HookInputParser
	eventBus        pluginsdk.EventBus
	dbPath          string
}

// NewClaudeCodePlugin creates a new Claude Code plugin
//
// For built-in plugins, we inject service implementations.
// External plugins would receive only PluginContext.
// eventBus is passed as interface{} to allow cmd package to avoid importing pluginsdk.
func NewClaudeCodePlugin(
	analysisService AnalysisService,
	logsService LogsService,
	logger pluginsdk.Logger,
	setupService SetupService,
	configLoader ConfigLoader,
	dbPath string,
	eventBus interface{},
) *ClaudeCodePlugin {
	// Type assert eventBus to pluginsdk.EventBus
	var eb pluginsdk.EventBus
	if eventBus != nil {
		if bus, ok := eventBus.(pluginsdk.EventBus); ok {
			eb = bus
		}
	}

	return &ClaudeCodePlugin{
		analysisService: analysisService,
		logsService:     logsService,
		logger:          logger,
		setupService:    setupService,
		configLoader:    configLoader,
		hookInputParser: newHookInputParser(), // Plugin creates its own parser
		eventBus:        eb,
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

// GetCapabilities returns the capability interfaces this plugin implements (SDK interface)
func (p *ClaudeCodePlugin) GetCapabilities() []string {
	return []string{"IEntityProvider", "IEntityUpdater", "ICommandProvider"}
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

// Query returns entities matching the given query (SDK interface).
//
// This method implements event sourcing principles: sessions are derived entities
// reconstructed from historical events stored in the event repository. Each call
// to Query rebuilds the current session state by:
//   1. Fetching all session IDs from the analysis service
//   2. For each session ID, calling buildSessionEntity to reconstruct the session
//      from its event history
//   3. Applying filters and pagination to the reconstructed sessions
//
// The query never returns stale data because sessions are always rebuilt from
// the event store, ensuring consistency with the authoritative event history.
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

// buildSessionEntity constructs a SessionEntity from a session ID.
//
// This is the core of event sourcing in the plugin: it reconstructs session state
// from the authoritative event repository. The method:
//
//   1. Queries the event repository via LogsService.ListRecentLogs() to fetch all
//      events for this session, ordered chronologically
//   2. Extracts session metadata from the event stream:
//      - Session ID (first and last events bound the session lifetime)
//      - Event timestamps (first and last event times)
//      - Event count (total number of events)
//   3. Fetches any analyses associated with this session
//   4. Constructs a SessionEntity representing the current derived state
//
// The session is read-only from the plugin perspective - all state flows from events.
// Any attempt to update a session is rejected (see UpdateEntity).
func (p *ClaudeCodePlugin) buildSessionEntity(ctx context.Context, sessionID string) (*SessionEntity, error) {
	// Query the event repository to get all events for this session
	// This is the defining characteristic of event sourcing: the session state
	// is reconstructed from historical events
	logs, err := p.logsService.ListRecentLogs(ctx, 0, 0, sessionID, true)
	if err != nil || len(logs) == 0 {
		return nil, fmt.Errorf("failed to get logs for session %s: %w", sessionID, err)
	}

	// Get analyses for this session
	analyses, err := p.analysisService.GetAnalysesBySessionID(ctx, sessionID)
	if err != nil {
		analyses = []*SessionAnalysis{}
	}

	// Convert to SessionAnalysisData
	analysisData := convertAnalyses(analyses)

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
		analysisData,
		tokenCount,
	), nil
}

// convertAnalyses converts SessionAnalysis to SessionAnalysisData
func convertAnalyses(analyses []*SessionAnalysis) []SessionAnalysisData {
	if analyses == nil {
		return []SessionAnalysisData{}
	}

	data := make([]SessionAnalysisData, len(analyses))
	for i, a := range analyses {
		data[i] = SessionAnalysisData{
			ID:              a.ID,
			SessionID:       a.SessionID,
			PromptName:      a.PromptName,
			ModelUsed:       a.ModelUsed,
			PatternsSummary: a.PatternsSummary,
			CreatedAt:       a.AnalyzedAt,
		}
	}
	return data
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
