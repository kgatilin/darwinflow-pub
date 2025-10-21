package claude_code_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

func TestNewSessionAnalysis(t *testing.T) {
	sessionID := "test-session-123"
	analysisResult := "This is a test analysis result"
	modelUsed := "claude-sonnet-4"
	promptUsed := "Test prompt used for analysis"

	before := time.Now()
	analysis := claude_code.NewSessionAnalysis(sessionID, analysisResult, modelUsed, promptUsed)
	after := time.Now()

	// Verify all fields are set correctly
	if analysis.ID == "" {
		t.Error("Expected ID to be generated, got empty string")
	}

	if analysis.SessionID != sessionID {
		t.Errorf("Expected SessionID %s, got %s", sessionID, analysis.SessionID)
	}

	if analysis.AnalysisResult != analysisResult {
		t.Errorf("Expected AnalysisResult %s, got %s", analysisResult, analysis.AnalysisResult)
	}

	if analysis.ModelUsed != modelUsed {
		t.Errorf("Expected ModelUsed %s, got %s", modelUsed, analysis.ModelUsed)
	}

	if analysis.PromptUsed != promptUsed {
		t.Errorf("Expected PromptUsed %s, got %s", promptUsed, analysis.PromptUsed)
	}

	// Verify AnalyzedAt is set to a reasonable time (within a few seconds of now)
	if analysis.AnalyzedAt.Before(before) || analysis.AnalyzedAt.After(after) {
		t.Errorf("Expected AnalyzedAt between %v and %v, got %v", before, after, analysis.AnalyzedAt)
	}

	// Verify default values for backward compatibility
	if analysis.AnalysisType != "tool_analysis" {
		t.Errorf("Expected default AnalysisType 'tool_analysis', got %s", analysis.AnalysisType)
	}

	if analysis.PromptName != "analysis" {
		t.Errorf("Expected default PromptName 'analysis', got %s", analysis.PromptName)
	}
}

func TestNewSessionAnalysisWithType(t *testing.T) {
	sessionID := "test-session-456"
	analysisResult := "Custom analysis result"
	modelUsed := "claude-opus-4"
	promptUsed := "Custom prompt"
	analysisType := "session_summary"
	promptName := "custom_prompt"

	before := time.Now()
	analysis := claude_code.NewSessionAnalysisWithType(
		sessionID,
		analysisResult,
		modelUsed,
		promptUsed,
		analysisType,
		promptName,
	)
	after := time.Now()

	// Verify all fields are set correctly
	if analysis.ID == "" {
		t.Error("Expected ID to be generated, got empty string")
	}

	if analysis.SessionID != sessionID {
		t.Errorf("Expected SessionID %s, got %s", sessionID, analysis.SessionID)
	}

	if analysis.AnalysisResult != analysisResult {
		t.Errorf("Expected AnalysisResult %s, got %s", analysisResult, analysis.AnalysisResult)
	}

	if analysis.ModelUsed != modelUsed {
		t.Errorf("Expected ModelUsed %s, got %s", modelUsed, analysis.ModelUsed)
	}

	if analysis.PromptUsed != promptUsed {
		t.Errorf("Expected PromptUsed %s, got %s", promptUsed, analysis.PromptUsed)
	}

	if analysis.AnalysisType != analysisType {
		t.Errorf("Expected AnalysisType %s, got %s", analysisType, analysis.AnalysisType)
	}

	if analysis.PromptName != promptName {
		t.Errorf("Expected PromptName %s, got %s", promptName, analysis.PromptName)
	}

	// Verify AnalyzedAt is set to a reasonable time
	if analysis.AnalyzedAt.Before(before) || analysis.AnalyzedAt.After(after) {
		t.Errorf("Expected AnalyzedAt between %v and %v, got %v", before, after, analysis.AnalyzedAt)
	}
}

func TestNewSessionAnalysis_UniqueIDs(t *testing.T) {
	// Verify that multiple calls generate unique IDs
	analysis1 := claude_code.NewSessionAnalysis("session-1", "result-1", "model-1", "prompt-1")
	analysis2 := claude_code.NewSessionAnalysis("session-2", "result-2", "model-2", "prompt-2")

	if analysis1.ID == analysis2.ID {
		t.Errorf("Expected unique IDs, got duplicate: %s", analysis1.ID)
	}
}

func TestNewSessionAnalysisWithType_UniqueIDs(t *testing.T) {
	// Verify that multiple calls generate unique IDs
	analysis1 := claude_code.NewSessionAnalysisWithType("session-1", "result-1", "model-1", "prompt-1", "type-1", "name-1")
	analysis2 := claude_code.NewSessionAnalysisWithType("session-2", "result-2", "model-2", "prompt-2", "type-2", "name-2")

	if analysis1.ID == analysis2.ID {
		t.Errorf("Expected unique IDs, got duplicate: %s", analysis1.ID)
	}
}
