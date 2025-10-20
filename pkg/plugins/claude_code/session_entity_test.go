package claude_code_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

// Helper to create test SessionAnalysisData
func makeTestAnalyses(count int) []claude_code.SessionAnalysisData {
	analyses := make([]claude_code.SessionAnalysisData, count)
	for i := 0; i < count; i++ {
		analyses[i] = claude_code.SessionAnalysisData{
			ID:              "a" + string(rune('1'+i)),
			SessionID:       "s1",
			PromptName:      "test_prompt",
			ModelUsed:       "claude-3",
			PatternsSummary: "Test summary",
			CreatedAt:       time.Now(),
		}
	}
	return analyses
}

func TestNewSessionEntity(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)

	analyses := makeTestAnalyses(1)

	entity := claude_code.NewSessionEntity("full-session-id", now, later, 10, analyses, 1000)

	if entity == nil {
		t.Fatal("NewSessionEntity returned nil")
	}

	if entity.GetID() != "full-session-id" {
		t.Errorf("Expected ID 'full-session-id', got %q", entity.GetID())
	}
}

func TestSessionEntity_GetType(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 5, nil, 0)

	if entity.GetType() != "session" {
		t.Errorf("Expected type 'session', got %q", entity.GetType())
	}
}

func TestSessionEntity_GetCapabilities(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 5, nil, 0)

	capabilities := entity.GetCapabilities()

	expectedCaps := map[string]bool{
		"IExtensible":  true,
		"IHasContext":  true,
		"ITrackable":   true,
	}

	if len(capabilities) != len(expectedCaps) {
		t.Errorf("Expected %d capabilities, got %d", len(expectedCaps), len(capabilities))
	}

	for _, cap := range capabilities {
		if !expectedCaps[cap] {
			t.Errorf("Unexpected capability: %s", cap)
		}
	}
}

func TestSessionEntity_GetField(t *testing.T) {
	now := time.Now()
	entity := claude_code.NewSessionEntity("test-session", now, now, 10, nil, 500)

	tests := []struct {
		fieldName string
		checkFn   func(interface{}) bool
	}{
		{"session_id", func(v interface{}) bool { return v.(string) == "test-session" }},
		{"event_count", func(v interface{}) bool { return v.(int) == 10 }},
		{"token_count", func(v interface{}) bool { return v.(int) == 500 }},
		{"status", func(v interface{}) bool { return v.(string) == "active" }},
		{"has_analysis", func(v interface{}) bool { return v.(bool) == false }},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			value := entity.GetField(tt.fieldName)
			if value == nil {
				t.Fatalf("GetField(%q) returned nil", tt.fieldName)
			}
			if !tt.checkFn(value) {
				t.Errorf("GetField(%q) returned unexpected value: %v", tt.fieldName, value)
			}
		})
	}
}

func TestSessionEntity_GetAllFields(t *testing.T) {
	now := time.Now()
	entity := claude_code.NewSessionEntity("s1", now, now, 15, nil, 750)

	fields := entity.GetAllFields()

	expectedFields := []string{
		"session_id", "short_id", "first_event", "last_event",
		"event_count", "analysis_count", "analysis_types", "token_count",
		"status", "has_analysis",
	}

	for _, fieldName := range expectedFields {
		if _, exists := fields[fieldName]; !exists {
			t.Errorf("Expected field %q not found in GetAllFields", fieldName)
		}
	}
}

func TestSessionEntity_ShortID(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantShort string
	}{
		{
			name:      "long ID",
			sessionID: "very-long-session-identifier-here",
			wantShort: "very-lon",
		},
		{
			name:      "short ID",
			sessionID: "short",
			wantShort: "short",
		},
		{
			name:      "exact 8 chars",
			sessionID: "exactly8",
			wantShort: "exactly8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := claude_code.NewSessionEntity(tt.sessionID, time.Now(), time.Now(), 1, nil, 0)
			shortID := entity.GetField("short_id").(string)
			if shortID != tt.wantShort {
				t.Errorf("Expected short_id %q, got %q", tt.wantShort, shortID)
			}
		})
	}
}

func TestSessionEntity_GetContext(t *testing.T) {
	now := time.Now()
	analyses := makeTestAnalyses(2)

	entity := claude_code.NewSessionEntity("s1", now, now.Add(10*time.Minute), 20, analyses, 1000)

	context := entity.GetContext()

	if context == nil {
		t.Fatal("GetContext returned nil")
	}

	// Check related entities
	if context.RelatedEntities == nil {
		t.Fatal("RelatedEntities is nil")
	}

	analysisIDs, ok := context.RelatedEntities["analysis"]
	if !ok {
		t.Fatal("Expected 'analysis' in RelatedEntities")
	}

	if len(analysisIDs) != 2 {
		t.Errorf("Expected 2 analysis IDs, got %d", len(analysisIDs))
	}

	// Check metadata
	if context.Metadata == nil {
		t.Fatal("Metadata is nil")
	}

	if _, ok := context.Metadata["session_duration"]; !ok {
		t.Error("Expected 'session_duration' in metadata")
	}

	if _, ok := context.Metadata["events_per_minute"]; !ok {
		t.Error("Expected 'events_per_minute' in metadata")
	}
}

func TestSessionEntity_GetContext_Cached(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 1, nil, 0)

	// Call twice to test caching
	context1 := entity.GetContext()
	context2 := entity.GetContext()

	// Should return the same cached instance
	if context1 != context2 {
		t.Error("GetContext should return cached instance")
	}
}

func TestSessionEntity_GetContext_NoAnalyses(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 5, nil, 0)

	context := entity.GetContext()

	if context == nil {
		t.Fatal("GetContext returned nil")
	}

	// Should have empty related entities
	if len(context.RelatedEntities) != 0 {
		t.Errorf("Expected no related entities, got %d", len(context.RelatedEntities))
	}
}

func TestSessionEntity_GetStatus_Active(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 10, nil, 0)

	status := entity.GetStatus()
	if status != "active" {
		t.Errorf("Expected status 'active' for session without analysis, got %q", status)
	}
}

func TestSessionEntity_GetStatus_Analyzed(t *testing.T) {
	analyses := makeTestAnalyses(1)

	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 10, analyses, 0)

	status := entity.GetStatus()
	if status != "analyzed" {
		t.Errorf("Expected status 'analyzed' for session with analysis, got %q", status)
	}
}

func TestSessionEntity_GetProgress(t *testing.T) {
	tests := []struct {
		name            string
		analysisCount   int
		expectedProgress float64
	}{
		{name: "no analysis", analysisCount: 0, expectedProgress: 0.0},
		{name: "with analysis", analysisCount: 1, expectedProgress: 1.0},
		{name: "multiple analyses", analysisCount: 3, expectedProgress: 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var analyses []claude_code.SessionAnalysisData
			for i := 0; i < tt.analysisCount; i++ {
				analyses = append(analyses, claude_code.SessionAnalysisData{
					ID:         "a" + string(rune('1'+i)),
					SessionID:  "s1",
					PromptName: "test",
					CreatedAt:  time.Now(),
				})
			}

			entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 10, analyses, 0)

			progress := entity.GetProgress()
			if progress != tt.expectedProgress {
				t.Errorf("Expected progress %v, got %v", tt.expectedProgress, progress)
			}
		})
	}
}

func TestSessionEntity_IsBlocked(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 5, nil, 0)

	if entity.IsBlocked() {
		t.Error("Sessions should never be blocked")
	}
}

func TestSessionEntity_GetBlockReason(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 5, nil, 0)

	reason := entity.GetBlockReason()
	if reason != "" {
		t.Errorf("Expected empty block reason, got %q", reason)
	}
}

func TestSessionEntity_GetAnalyses(t *testing.T) {
	analyses := makeTestAnalyses(2)

	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 10, analyses, 0)

	returnedAnalyses := entity.GetAnalyses()

	if len(returnedAnalyses) != 2 {
		t.Errorf("Expected 2 analyses, got %d", len(returnedAnalyses))
	}

	if returnedAnalyses[0].ID != "a1" || returnedAnalyses[1].ID != "a2" {
		t.Error("Analyses not returned in correct order")
	}
}

func TestSessionEntity_GetAnalyses_Empty(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 10, nil, 0)

	analyses := entity.GetAnalyses()

	if analyses != nil {
		t.Errorf("Expected nil analyses, got %v", analyses)
	}
}

func TestSessionEntity_GetLatestAnalysis(t *testing.T) {
	analyses := makeTestAnalyses(2)

	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 10, analyses, 0)

	latest := entity.GetLatestAnalysis()

	if latest == nil {
		t.Fatal("GetLatestAnalysis returned nil")
	}

	if latest.ID != "a1" {
		t.Errorf("Expected latest analysis ID 'a1', got %q", latest.ID)
	}
}

func TestSessionEntity_GetLatestAnalysis_NoAnalyses(t *testing.T) {
	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 10, nil, 0)

	latest := entity.GetLatestAnalysis()

	if latest != nil {
		t.Errorf("Expected nil for session without analyses, got %v", latest)
	}
}

func TestSessionEntity_AnalysisTypes(t *testing.T) {
	analyses := []claude_code.SessionAnalysisData{
		{ID: "a1", SessionID: "s1", PromptName: "tool_analysis", CreatedAt: time.Now()},
		{ID: "a2", SessionID: "s1", PromptName: "session_summary", CreatedAt: time.Now()},
	}

	entity := claude_code.NewSessionEntity("s1", time.Now(), time.Now(), 10, analyses, 0)

	analysisTypes := entity.GetField("analysis_types").([]string)

	if len(analysisTypes) != 2 {
		t.Fatalf("Expected 2 analysis types, got %d", len(analysisTypes))
	}

	if analysisTypes[0] != "tool_analysis" || analysisTypes[1] != "session_summary" {
		t.Errorf("Analysis types not correct: %v", analysisTypes)
	}
}

func TestSessionEntity_Fields_ConsistentWithGetField(t *testing.T) {
	now := time.Now()
	analyses := makeTestAnalyses(1)
	entity := claude_code.NewSessionEntity("test-session", now, now, 10, analyses, 500)

	allFields := entity.GetAllFields()

	for fieldName, expectedValue := range allFields {
		actualValue := entity.GetField(fieldName)
		// For complex types (slices, structs), we just check non-nil
		if actualValue == nil && expectedValue != nil {
			t.Errorf("GetField(%q) inconsistent with GetAllFields", fieldName)
		}
	}
}
