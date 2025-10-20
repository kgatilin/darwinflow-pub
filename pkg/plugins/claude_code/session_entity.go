package claude_code

import (
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// SessionAnalysisData holds analysis data without depending on internal types
type SessionAnalysisData struct {
	ID              string
	SessionID       string
	PromptName      string
	ModelUsed       string
	PatternsSummary string
	CreatedAt       time.Time
}

// SessionEntity wraps a Claude Code session and implements SDK capability interfaces.
// It adapts the existing session data structure to the plugin system.
type SessionEntity struct {
	sessionID      string
	shortID        string
	firstEvent     time.Time
	lastEvent      time.Time
	eventCount     int
	analysisCount  int
	analyses       []SessionAnalysisData
	analysisTypes  []string
	tokenCount     int
	context        *pluginsdk.EntityContext // Cached context
}

// NewSessionEntity creates a new session entity from session data
func NewSessionEntity(
	sessionID string,
	firstEvent, lastEvent time.Time,
	eventCount int,
	analyses []SessionAnalysisData,
	tokenCount int,
) *SessionEntity {
	shortID := sessionID
	if len(sessionID) > 8 {
		shortID = sessionID[:8]
	}

	analysisTypes := make([]string, 0, len(analyses))
	for _, a := range analyses {
		analysisTypes = append(analysisTypes, a.PromptName)
	}

	return &SessionEntity{
		sessionID:     sessionID,
		shortID:       shortID,
		firstEvent:    firstEvent,
		lastEvent:     lastEvent,
		eventCount:    eventCount,
		analysisCount: len(analyses),
		analyses:      analyses,
		analysisTypes: analysisTypes,
		tokenCount:    tokenCount,
	}
}

// IExtensible implementation

func (s *SessionEntity) GetID() string {
	return s.sessionID
}

func (s *SessionEntity) GetType() string {
	return "session"
}

func (s *SessionEntity) GetCapabilities() []string {
	return []string{"IExtensible", "IHasContext", "ITrackable"}
}

func (s *SessionEntity) GetField(name string) interface{} {
	fields := s.GetAllFields()
	return fields[name]
}

func (s *SessionEntity) GetAllFields() map[string]interface{} {
	return map[string]interface{}{
		"session_id":     s.sessionID,
		"short_id":       s.shortID,
		"first_event":    s.firstEvent,
		"last_event":     s.lastEvent,
		"event_count":    s.eventCount,
		"analysis_count": s.analysisCount,
		"analysis_types": s.analysisTypes,
		"token_count":    s.tokenCount,
		"status":         s.GetStatus(),
		"has_analysis":   s.analysisCount > 0,
	}
}

// IHasContext implementation

func (s *SessionEntity) GetContext() *pluginsdk.EntityContext {
	if s.context != nil {
		return s.context
	}

	// Build context from session data
	context := &pluginsdk.EntityContext{
		RelatedEntities: make(map[string][]string),
		LinkedFiles:     []string{},
		RecentActivity:  []pluginsdk.ActivityRecord{},
		Metadata:        make(map[string]interface{}),
	}

	// Add analyses as related entities
	if len(s.analyses) > 0 {
		analysisIDs := make([]string, 0, len(s.analyses))
		for _, a := range s.analyses {
			analysisIDs = append(analysisIDs, a.ID)
		}
		context.RelatedEntities["analysis"] = analysisIDs
	}

	// Add metadata
	context.Metadata["session_duration"] = s.lastEvent.Sub(s.firstEvent)
	context.Metadata["events_per_minute"] = float64(s.eventCount) / s.lastEvent.Sub(s.firstEvent).Minutes()

	s.context = context
	return s.context
}

// ITrackable implementation

func (s *SessionEntity) GetStatus() string {
	// A session is "completed" if it has analysis, otherwise "active"
	if s.analysisCount > 0 {
		return "analyzed"
	}
	return "active"
}

func (s *SessionEntity) GetProgress() float64 {
	// Progress is based on whether session has been analyzed
	if s.analysisCount > 0 {
		return 1.0
	}
	return 0.0
}

func (s *SessionEntity) IsBlocked() bool {
	// Sessions are never blocked
	return false
}

func (s *SessionEntity) GetBlockReason() string {
	return ""
}

// Additional helper methods

// GetAnalyses returns all analyses for this session
func (s *SessionEntity) GetAnalyses() []SessionAnalysisData {
	return s.analyses
}

// GetLatestAnalysis returns the most recent analysis
func (s *SessionEntity) GetLatestAnalysis() *SessionAnalysisData {
	if len(s.analyses) > 0 {
		return &s.analyses[0]
	}
	return nil
}
