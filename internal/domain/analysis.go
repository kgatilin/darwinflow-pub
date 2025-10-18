package domain

import (
	"time"

	"github.com/google/uuid"
)

// SessionAnalysis represents an AI-generated analysis of a session
type SessionAnalysis struct {
	ID              string
	SessionID       string
	AnalyzedAt      time.Time
	AnalysisResult  string // Full analysis text from LLM
	ModelUsed       string
	PromptUsed      string
	PatternsSummary string // Brief summary extracted from analysis
}

// NewSessionAnalysis creates a new session analysis
func NewSessionAnalysis(sessionID, analysisResult, modelUsed, promptUsed string) *SessionAnalysis {
	return &SessionAnalysis{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		AnalyzedAt:     time.Now(),
		AnalysisResult: analysisResult,
		ModelUsed:      modelUsed,
		PromptUsed:     promptUsed,
	}
}

// ToolSuggestion represents a suggested tool from analysis
type ToolSuggestion struct {
	Name        string
	Description string
	Rationale   string
	Examples    []string
}
