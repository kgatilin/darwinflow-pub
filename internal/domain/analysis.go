package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Analysis represents a generic analysis result for any view type.
// This is a plugin-agnostic type that can store analysis results from any plugin.
type Analysis struct {
	ID         string                 // Unique analysis ID
	ViewID     string                 // ID of the analyzed view
	ViewType   string                 // Type of view ("session", "task-list", "date-range", etc.)
	Timestamp  time.Time              // When analysis was performed
	Result     string                 // LLM analysis output
	ModelUsed  string                 // LLM model used
	PromptUsed string                 // Prompt name/template used
	Metadata   map[string]interface{} // View-specific metadata (JSON in DB)
}

// NewAnalysis creates a new generic analysis
func NewAnalysis(viewID, viewType, result, modelUsed, promptUsed string) *Analysis {
	return &Analysis{
		ID:         uuid.New().String(),
		ViewID:     viewID,
		ViewType:   viewType,
		Timestamp:  time.Now(),
		Result:     result,
		ModelUsed:  modelUsed,
		PromptUsed: promptUsed,
		Metadata:   make(map[string]interface{}),
	}
}

// MarshalMetadata marshals the metadata to JSON
func (a *Analysis) MarshalMetadata() ([]byte, error) {
	if a.Metadata == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(a.Metadata)
}

// UnmarshalMetadata unmarshals the metadata from JSON
func (a *Analysis) UnmarshalMetadata(data []byte) error {
	if len(data) == 0 {
		a.Metadata = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal(data, &a.Metadata)
}

// SessionAnalysis represents an AI-generated analysis of a Claude Code session.
// NOTE: This type exists in domain for backward compatibility with internal code.
// Analysis is semantically owned by the claude-code plugin.
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
