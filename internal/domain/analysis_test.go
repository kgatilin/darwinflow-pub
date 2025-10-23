package domain_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

func TestNewSessionAnalysis(t *testing.T) {
	sessionID := "test-session-123"
	analysisResult := "This is the analysis result from LLM"
	modelUsed := "claude-sonnet-4"
	promptUsed := "analyze this session"

	analysis := domain.NewSessionAnalysis(sessionID, analysisResult, modelUsed, promptUsed)

	// Verify required fields
	if analysis.ID == "" {
		t.Error("Expected ID to be generated, got empty string")
	}
	if analysis.SessionID != sessionID {
		t.Errorf("Expected SessionID = %q, got %q", sessionID, analysis.SessionID)
	}
	if analysis.AnalysisResult != analysisResult {
		t.Errorf("Expected AnalysisResult = %q, got %q", analysisResult, analysis.AnalysisResult)
	}
	if analysis.ModelUsed != modelUsed {
		t.Errorf("Expected ModelUsed = %q, got %q", modelUsed, analysis.ModelUsed)
	}
	if analysis.PromptUsed != promptUsed {
		t.Errorf("Expected PromptUsed = %q, got %q", promptUsed, analysis.PromptUsed)
	}

	// Verify defaults for backward compatibility
	if analysis.AnalysisType != "tool_analysis" {
		t.Errorf("Expected AnalysisType = %q, got %q", "tool_analysis", analysis.AnalysisType)
	}
	if analysis.PromptName != "analysis" {
		t.Errorf("Expected PromptName = %q, got %q", "analysis", analysis.PromptName)
	}

	// Verify timestamp is recent
	if time.Since(analysis.AnalyzedAt) > time.Second {
		t.Errorf("Expected recent timestamp, got %v", analysis.AnalyzedAt)
	}

	// Verify PatternsSummary is empty by default
	if analysis.PatternsSummary != "" {
		t.Errorf("Expected empty PatternsSummary, got %q", analysis.PatternsSummary)
	}
}

func TestNewSessionAnalysisWithType(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		analysisResult string
		modelUsed      string
		promptUsed     string
		analysisType   string
		promptName     string
	}{
		{
			name:           "creates session summary analysis",
			sessionID:      "session-1",
			analysisResult: "Session summary text",
			modelUsed:      "claude-opus-4",
			promptUsed:     "summarize this session",
			analysisType:   "session_summary",
			promptName:     "session_summary",
		},
		{
			name:           "creates tool analysis",
			sessionID:      "session-2",
			analysisResult: "Tool analysis text",
			modelUsed:      "claude-sonnet-4",
			promptUsed:     "analyze tool usage",
			analysisType:   "tool_analysis",
			promptName:     "tool_analysis",
		},
		{
			name:           "creates custom analysis type",
			sessionID:      "session-3",
			analysisResult: "Custom analysis",
			modelUsed:      "claude-haiku-4",
			promptUsed:     "custom prompt",
			analysisType:   "custom_type",
			promptName:     "custom_prompt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := domain.NewSessionAnalysisWithType(
				tt.sessionID,
				tt.analysisResult,
				tt.modelUsed,
				tt.promptUsed,
				tt.analysisType,
				tt.promptName,
			)

			if analysis.ID == "" {
				t.Error("Expected ID to be generated, got empty string")
			}
			if analysis.SessionID != tt.sessionID {
				t.Errorf("Expected SessionID = %q, got %q", tt.sessionID, analysis.SessionID)
			}
			if analysis.AnalysisResult != tt.analysisResult {
				t.Errorf("Expected AnalysisResult = %q, got %q", tt.analysisResult, analysis.AnalysisResult)
			}
			if analysis.ModelUsed != tt.modelUsed {
				t.Errorf("Expected ModelUsed = %q, got %q", tt.modelUsed, analysis.ModelUsed)
			}
			if analysis.PromptUsed != tt.promptUsed {
				t.Errorf("Expected PromptUsed = %q, got %q", tt.promptUsed, analysis.PromptUsed)
			}
			if analysis.AnalysisType != tt.analysisType {
				t.Errorf("Expected AnalysisType = %q, got %q", tt.analysisType, analysis.AnalysisType)
			}
			if analysis.PromptName != tt.promptName {
				t.Errorf("Expected PromptName = %q, got %q", tt.promptName, analysis.PromptName)
			}

			// Verify timestamp is recent
			if time.Since(analysis.AnalyzedAt) > time.Second {
				t.Errorf("Expected recent timestamp, got %v", analysis.AnalyzedAt)
			}
		})
	}
}

func TestSessionAnalysis_AllFields(t *testing.T) {
	// Test that all fields can be set and retrieved
	analysis := &domain.SessionAnalysis{
		ID:              "test-id",
		SessionID:       "test-session",
		AnalyzedAt:      time.Now(),
		AnalysisResult:  "full analysis text",
		ModelUsed:       "claude-sonnet-4",
		PromptUsed:      "test prompt",
		PatternsSummary: "summary of patterns",
		AnalysisType:    "tool_analysis",
		PromptName:      "custom_prompt",
	}

	// Verify all fields are accessible
	if analysis.ID != "test-id" {
		t.Errorf("ID mismatch")
	}
	if analysis.SessionID != "test-session" {
		t.Errorf("SessionID mismatch")
	}
	if analysis.AnalysisResult != "full analysis text" {
		t.Errorf("AnalysisResult mismatch")
	}
	if analysis.ModelUsed != "claude-sonnet-4" {
		t.Errorf("ModelUsed mismatch")
	}
	if analysis.PromptUsed != "test prompt" {
		t.Errorf("PromptUsed mismatch")
	}
	if analysis.PatternsSummary != "summary of patterns" {
		t.Errorf("PatternsSummary mismatch")
	}
	if analysis.AnalysisType != "tool_analysis" {
		t.Errorf("AnalysisType mismatch")
	}
	if analysis.PromptName != "custom_prompt" {
		t.Errorf("PromptName mismatch")
	}
}

func TestToolSuggestion_Fields(t *testing.T) {
	// Test that ToolSuggestion struct can be created and fields accessed
	suggestion := domain.ToolSuggestion{
		Name:        "CodeAnalyzer",
		Description: "Analyzes code patterns",
		Rationale:   "Would speed up analysis by 50%",
		Examples:    []string{"example 1", "example 2"},
	}

	if suggestion.Name != "CodeAnalyzer" {
		t.Errorf("Name mismatch")
	}
	if suggestion.Description != "Analyzes code patterns" {
		t.Errorf("Description mismatch")
	}
	if suggestion.Rationale != "Would speed up analysis by 50%" {
		t.Errorf("Rationale mismatch")
	}
	if len(suggestion.Examples) != 2 {
		t.Errorf("Expected 2 examples, got %d", len(suggestion.Examples))
	}
	if suggestion.Examples[0] != "example 1" {
		t.Errorf("Examples[0] mismatch")
	}
}

func TestNewSessionAnalysis_UniqueIDs(t *testing.T) {
	// Verify that multiple calls generate unique IDs
	analysis1 := domain.NewSessionAnalysis("session-1", "result1", "model1", "prompt1")
	analysis2 := domain.NewSessionAnalysis("session-1", "result2", "model2", "prompt2")

	if analysis1.ID == analysis2.ID {
		t.Error("Expected unique IDs for different analyses, got same ID")
	}
}

func TestNewSessionAnalysisWithType_UniqueIDs(t *testing.T) {
	// Verify that multiple calls generate unique IDs
	analysis1 := domain.NewSessionAnalysisWithType("session-1", "result1", "model1", "prompt1", "type1", "name1")
	analysis2 := domain.NewSessionAnalysisWithType("session-1", "result2", "model2", "prompt2", "type2", "name2")

	if analysis1.ID == analysis2.ID {
		t.Error("Expected unique IDs for different analyses, got same ID")
	}
}

func TestSessionAnalysis_EmptyFields(t *testing.T) {
	// Test that analysis can be created with empty optional fields
	analysis := domain.NewSessionAnalysisWithType("session-1", "", "", "", "", "")

	if analysis.ID == "" {
		t.Error("ID should be generated even with empty fields")
	}
	if analysis.SessionID != "session-1" {
		t.Error("SessionID should be preserved")
	}
	if analysis.AnalysisResult != "" {
		t.Error("Empty AnalysisResult should be preserved")
	}
}

// Tests for generic Analysis type

func TestNewAnalysis(t *testing.T) {
	viewID := "view-123"
	viewType := "session"
	result := "This is the analysis result"
	modelUsed := "claude-sonnet-4"
	promptUsed := "test_prompt"

	analysis := domain.NewAnalysis(viewID, viewType, result, modelUsed, promptUsed)

	// Verify required fields
	if analysis.ID == "" {
		t.Error("Expected ID to be generated, got empty string")
	}
	if analysis.ViewID != viewID {
		t.Errorf("Expected ViewID = %q, got %q", viewID, analysis.ViewID)
	}
	if analysis.ViewType != viewType {
		t.Errorf("Expected ViewType = %q, got %q", viewType, analysis.ViewType)
	}
	if analysis.Result != result {
		t.Errorf("Expected Result = %q, got %q", result, analysis.Result)
	}
	if analysis.ModelUsed != modelUsed {
		t.Errorf("Expected ModelUsed = %q, got %q", modelUsed, analysis.ModelUsed)
	}
	if analysis.PromptUsed != promptUsed {
		t.Errorf("Expected PromptUsed = %q, got %q", promptUsed, analysis.PromptUsed)
	}

	// Verify metadata is initialized
	if analysis.Metadata == nil {
		t.Error("Expected Metadata to be initialized, got nil")
	}

	// Verify timestamp is recent
	if time.Since(analysis.Timestamp) > time.Second {
		t.Errorf("Expected recent timestamp, got %v", analysis.Timestamp)
	}
}

func TestNewAnalysis_UniqueIDs(t *testing.T) {
	// Verify that multiple calls generate unique IDs
	analysis1 := domain.NewAnalysis("view-1", "session", "result1", "model1", "prompt1")
	analysis2 := domain.NewAnalysis("view-1", "session", "result2", "model2", "prompt2")

	if analysis1.ID == analysis2.ID {
		t.Error("Expected unique IDs for different analyses, got same ID")
	}
}

func TestAnalysis_MarshalMetadata(t *testing.T) {
	analysis := domain.NewAnalysis("view-1", "session", "result", "model", "prompt")
	analysis.Metadata["key1"] = "value1"
	analysis.Metadata["key2"] = 42
	analysis.Metadata["key3"] = true

	jsonData, err := analysis.MarshalMetadata()
	if err != nil {
		t.Fatalf("MarshalMetadata failed: %v", err)
	}

	// Should contain the keys
	jsonStr := string(jsonData)
	if !contains(jsonStr, "key1") {
		t.Error("Expected JSON to contain key1")
	}
	if !contains(jsonStr, "value1") {
		t.Error("Expected JSON to contain value1")
	}
}

func TestAnalysis_MarshalMetadata_NilMetadata(t *testing.T) {
	analysis := domain.NewAnalysis("view-1", "session", "result", "model", "prompt")
	analysis.Metadata = nil

	jsonData, err := analysis.MarshalMetadata()
	if err != nil {
		t.Fatalf("MarshalMetadata failed: %v", err)
	}

	// Should return empty object
	if string(jsonData) != "{}" {
		t.Errorf("Expected {}, got %s", string(jsonData))
	}
}

func TestAnalysis_UnmarshalMetadata(t *testing.T) {
	analysis := domain.NewAnalysis("view-1", "session", "result", "model", "prompt")

	jsonData := []byte(`{"key1":"value1","key2":42,"key3":true}`)
	err := analysis.UnmarshalMetadata(jsonData)
	if err != nil {
		t.Fatalf("UnmarshalMetadata failed: %v", err)
	}

	// Verify metadata was unmarshaled
	if analysis.Metadata == nil {
		t.Fatal("Expected Metadata to be set")
	}

	if v, ok := analysis.Metadata["key1"].(string); !ok || v != "value1" {
		t.Error("Expected key1 to be 'value1'")
	}

	if v, ok := analysis.Metadata["key2"].(float64); !ok || v != 42 {
		t.Error("Expected key2 to be 42")
	}

	if v, ok := analysis.Metadata["key3"].(bool); !ok || v != true {
		t.Error("Expected key3 to be true")
	}
}

func TestAnalysis_UnmarshalMetadata_EmptyData(t *testing.T) {
	analysis := domain.NewAnalysis("view-1", "session", "result", "model", "prompt")

	err := analysis.UnmarshalMetadata([]byte{})
	if err != nil {
		t.Fatalf("UnmarshalMetadata failed: %v", err)
	}

	// Should initialize empty metadata
	if analysis.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
}

func TestAnalysis_RoundtripMetadata(t *testing.T) {
	// Test marshaling and unmarshaling preserves data
	original := domain.NewAnalysis("view-1", "session", "result", "model", "prompt")
	original.Metadata["string_key"] = "string_value"
	original.Metadata["int_key"] = 123
	original.Metadata["bool_key"] = false
	original.Metadata["nested"] = map[string]interface{}{
		"nested_key": "nested_value",
	}

	// Marshal
	jsonData, err := original.MarshalMetadata()
	if err != nil {
		t.Fatalf("MarshalMetadata failed: %v", err)
	}

	// Unmarshal into new analysis
	restored := domain.NewAnalysis("view-2", "task", "result2", "model2", "prompt2")
	err = restored.UnmarshalMetadata(jsonData)
	if err != nil {
		t.Fatalf("UnmarshalMetadata failed: %v", err)
	}

	// Verify data preserved
	if v, ok := restored.Metadata["string_key"].(string); !ok || v != "string_value" {
		t.Error("string_key not preserved")
	}

	// Note: JSON unmarshaling converts numbers to float64
	if v, ok := restored.Metadata["int_key"].(float64); !ok || v != 123 {
		t.Error("int_key not preserved")
	}

	if v, ok := restored.Metadata["bool_key"].(bool); !ok || v != false {
		t.Error("bool_key not preserved")
	}

	if nested, ok := restored.Metadata["nested"].(map[string]interface{}); !ok {
		t.Error("nested not preserved as map")
	} else if nestedValue, ok := nested["nested_key"].(string); !ok || nestedValue != "nested_value" {
		t.Error("nested_key not preserved")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
