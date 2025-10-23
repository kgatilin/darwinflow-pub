package claude_code_test

import (
	"testing"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

func TestNewSessionView(t *testing.T) {
	sessionID := "test-session-123"
	events := []pluginsdk.Event{
		{
			Type:      "tool.invoked",
			Source:    "claude-code",
			Timestamp: time.Now(),
			Payload:   map[string]interface{}{"tool": "Read", "path": "/workspace/file.go"},
			Metadata:  map[string]string{"session_id": sessionID},
		},
	}

	view := claude_code.NewSessionView(sessionID, events)

	if view.GetID() != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, view.GetID())
	}

	if view.GetType() != "session" {
		t.Errorf("expected type 'session', got %s", view.GetType())
	}

	if len(view.GetEvents()) != len(events) {
		t.Errorf("expected %d events, got %d", len(events), len(view.GetEvents()))
	}
}

func TestSessionView_GetID(t *testing.T) {
	sessionID := "session-abc-123"
	view := claude_code.NewSessionView(sessionID, nil)

	if view.GetID() != sessionID {
		t.Errorf("expected %s, got %s", sessionID, view.GetID())
	}
}

func TestSessionView_GetType(t *testing.T) {
	view := claude_code.NewSessionView("session-1", nil)

	if view.GetType() != "session" {
		t.Errorf("expected 'session', got %s", view.GetType())
	}
}

func TestSessionView_GetEvents(t *testing.T) {
	events := []pluginsdk.Event{
		{
			Type:      "tool.invoked",
			Source:    "claude-code",
			Timestamp: time.Now(),
			Payload:   map[string]interface{}{"tool": "Read"},
			Metadata:  map[string]string{"session_id": "session-1"},
		},
		{
			Type:      "tool.result",
			Source:    "claude-code",
			Timestamp: time.Now().Add(time.Second),
			Payload:   map[string]interface{}{"result": "success"},
			Metadata:  map[string]string{"session_id": "session-1"},
		},
	}

	view := claude_code.NewSessionView("session-1", events)
	retrievedEvents := view.GetEvents()

	if len(retrievedEvents) != len(events) {
		t.Errorf("expected %d events, got %d", len(events), len(retrievedEvents))
	}

	for i, event := range retrievedEvents {
		if event.Type != events[i].Type {
			t.Errorf("event %d: expected type %s, got %s", i, events[i].Type, event.Type)
		}
	}
}

func TestSessionView_FormatForAnalysis_Empty(t *testing.T) {
	view := claude_code.NewSessionView("session-1", []pluginsdk.Event{})
	formatted := view.FormatForAnalysis()

	if formatted == "" {
		t.Error("expected non-empty formatted output")
	}

	if len(formatted) < 20 {
		t.Error("expected reasonable formatted output length")
	}
}

func TestSessionView_FormatForAnalysis_WithEvents(t *testing.T) {
	events := []pluginsdk.Event{
		{
			Type:      "tool.invoked",
			Source:    "claude-code",
			Timestamp: time.Now(),
			Payload:   map[string]interface{}{"tool": "Read", "path": "/workspace/file.go"},
			Metadata:  map[string]string{"session_id": "session-1"},
		},
		{
			Type:      "tool.result",
			Source:    "claude-code",
			Timestamp: time.Now().Add(time.Second),
			Payload:   map[string]interface{}{"result": "success", "lines": 42},
			Metadata:  map[string]string{"session_id": "session-1"},
		},
	}

	view := claude_code.NewSessionView("session-1", events)
	formatted := view.FormatForAnalysis()

	// Check that it contains expected content
	if len(formatted) == 0 {
		t.Error("expected non-empty formatted output")
	}

	if !contains(formatted, "session-1") {
		t.Error("expected session ID in formatted output")
	}

	if !contains(formatted, "tool.invoked") {
		t.Error("expected event type in formatted output")
	}

	if !contains(formatted, "tool.result") {
		t.Error("expected second event type in formatted output")
	}
}

func TestSessionView_GetMetadata(t *testing.T) {
	events := []pluginsdk.Event{
		{
			Type:      "tool.invoked",
			Source:    "claude-code",
			Timestamp: time.Now(),
			Payload:   map[string]interface{}{},
			Metadata:  map[string]string{"session_id": "session-1"},
		},
	}

	view := claude_code.NewSessionView("session-1", events)
	metadata := view.GetMetadata()

	if metadata["session_id"] != "session-1" {
		t.Errorf("expected session_id 'session-1', got %v", metadata["session_id"])
	}

	if metadata["event_count"] != 1 {
		t.Errorf("expected event_count 1, got %v", metadata["event_count"])
	}

	if metadata["view_type"] != "session" {
		t.Errorf("expected view_type 'session', got %v", metadata["view_type"])
	}
}

func TestSessionView_ImplementsAnalysisView(t *testing.T) {
	// This test ensures SessionView implements AnalysisView interface
	var _ pluginsdk.AnalysisView = (*claude_code.SessionView)(nil)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
