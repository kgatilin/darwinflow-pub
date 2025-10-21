package claude_code

import (
	"time"

	"github.com/google/uuid"
)

// SessionAnalysis represents an AI-generated analysis of a Claude Code session.
// This type is plugin-specific and owned by the claude-code plugin.
type SessionAnalysis struct {
	ID              string
	SessionID       string
	AnalyzedAt      time.Time
	AnalysisResult  string // Full analysis text from LLM
	ModelUsed       string
	PromptUsed      string
	PatternsSummary string // Brief summary extracted from analysis
	AnalysisType    string // Type of analysis: session_summary, tool_analysis, etc.
	PromptName      string // Name of the prompt from config
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
		AnalysisType:   "tool_analysis", // default for backward compatibility
		PromptName:     "analysis",      // default for backward compatibility
	}
}

// NewSessionAnalysisWithType creates a new session analysis with specific type
func NewSessionAnalysisWithType(sessionID, analysisResult, modelUsed, promptUsed, analysisType, promptName string) *SessionAnalysis {
	return &SessionAnalysis{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		AnalyzedAt:     time.Now(),
		AnalysisResult: analysisResult,
		ModelUsed:      modelUsed,
		PromptUsed:     promptUsed,
		AnalysisType:   analysisType,
		PromptName:     promptName,
	}
}

// ToolSuggestion represents a suggested tool from analysis
type ToolSuggestion struct {
	Name        string
	Description string
	Rationale   string
	Examples    []string
}
