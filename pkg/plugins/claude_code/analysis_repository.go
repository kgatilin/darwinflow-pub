package claude_code

import (
	"context"
)

// AnalysisRepository defines the interface for persisting and retrieving session analyses.
// This interface is plugin-specific and owned by the claude-code plugin.
type AnalysisRepository interface {
	// SaveAnalysis persists a session analysis
	SaveAnalysis(ctx context.Context, analysis *SessionAnalysis) error

	// GetAnalysisBySessionID retrieves the most recent analysis for a session
	GetAnalysisBySessionID(ctx context.Context, sessionID string) (*SessionAnalysis, error)

	// GetAnalysesBySessionID retrieves all analyses for a session, ordered by analyzed_at DESC
	GetAnalysesBySessionID(ctx context.Context, sessionID string) ([]*SessionAnalysis, error)

	// GetUnanalyzedSessionIDs retrieves session IDs that have not been analyzed
	GetUnanalyzedSessionIDs(ctx context.Context) ([]string, error)

	// GetAllAnalyses retrieves all analyses, ordered by analyzed_at DESC
	GetAllAnalyses(ctx context.Context, limit int) ([]*SessionAnalysis, error)

	// GetAllSessionIDs retrieves all session IDs, ordered by most recent first
	// If limit > 0, returns only the latest N sessions
	GetAllSessionIDs(ctx context.Context, limit int) ([]string, error)
}
