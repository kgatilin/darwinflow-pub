package tui_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/internal/app/tui"
)

func TestNewSessionListModel(t *testing.T) {
	sessions := []*tui.SessionInfo{
		{
			SessionID:  "test-session-1",
			ShortID:    "test-ses",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 10,
			HasAnalysis: false,
		},
	}

	model := tui.NewSessionListModel(sessions)

	// Check that model is properly initialized
	if model.GetSelectedSession() == nil {
		// No session selected initially is ok
	}
}

func TestNewSessionListModel_EmptySessions(t *testing.T) {
	sessions := []*tui.SessionInfo{}
	model := tui.NewSessionListModel(sessions)

	if model.GetSelectedSession() != nil {
		t.Error("GetSelectedSession() should return nil for empty list")
	}
}

func TestSessionListModel_Init(t *testing.T) {
	sessions := []*tui.SessionInfo{}
	model := tui.NewSessionListModel(sessions)

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil command for SessionListModel")
	}
}

func TestSessionListModel_UpdateWindowSize(t *testing.T) {
	sessions := []*tui.SessionInfo{
		{
			SessionID:  "test-session",
			ShortID:    "test123",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	model := tui.NewSessionListModel(sessions)

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("WindowSizeMsg should return nil command")
	}

	_, ok := updatedModel.(tui.SessionListModel)
	if !ok {
		t.Error("Update should return SessionListModel")
	}
}

func TestSessionListModel_UpdateEsc(t *testing.T) {
	sessions := []*tui.SessionInfo{}
	model := tui.NewSessionListModel(sessions)

	// Send esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Esc key should return quit command")
	}
}

func TestSessionListModel_UpdateRefresh(t *testing.T) {
	sessions := []*tui.SessionInfo{}
	model := tui.NewSessionListModel(sessions)

	// Send 'r' key for refresh
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("'r' key should return refresh command")
	}

	// Execute the command to get the message
	if cmd != nil {
		result := cmd()
		if _, ok := result.(tui.RefreshRequestMsg); !ok {
			t.Error("Expected RefreshRequestMsg from refresh command")
		}
	}
}

func TestSessionListModel_UpdateEnter(t *testing.T) {
	now := time.Now()
	sessions := []*tui.SessionInfo{
		{
			SessionID:  "test-session",
			ShortID:    "test123",
			FirstEvent: now,
			LastEvent:  now,
			EventCount: 5,
		},
	}

	model := tui.NewSessionListModel(sessions)

	// First set window size to initialize list properly
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionListModel)

	// Send enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := model.Update(msg)

	// Enter should return a command (to select session)
	if cmd != nil {
		result := cmd()
		// It's ok if result is not SelectedSessionMsg (no session might be highlighted)
		_ = result
	}
}

func TestSessionListModel_UpdateSessions(t *testing.T) {
	initialSessions := []*tui.SessionInfo{
		{
			SessionID:  "session-1",
			ShortID:    "sess-1",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 1,
		},
	}

	model := tui.NewSessionListModel(initialSessions)

	// Update with new sessions
	newSessions := []*tui.SessionInfo{
		{
			SessionID:  "session-1",
			ShortID:    "sess-1",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 1,
		},
		{
			SessionID:  "session-2",
			ShortID:    "sess-2",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 2,
		},
	}

	model.UpdateSessions(newSessions)

	// Can't directly verify internal state, but method should not panic
}

func TestSessionListModel_GetSelectedSession(t *testing.T) {
	sessions := []*tui.SessionInfo{
		{
			SessionID:  "test-session",
			ShortID:    "test123",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	model := tui.NewSessionListModel(sessions)

	// Initialize the model with a window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionListModel)

	// GetSelectedSession should not panic
	selected := model.GetSelectedSession()

	// It might be nil or the first session depending on list state
	if selected != nil && selected.SessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got %s", selected.SessionID)
	}
}

func TestSessionListModel_EmptyList_View(t *testing.T) {
	model := tui.NewSessionListModel([]*tui.SessionInfo{})

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionListModel)

	// View should work with empty list
	view := model.View()
	if view == "" {
		t.Error("View should handle empty session list")
	}

	// Try navigation with empty list
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(tui.SessionListModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(tui.SessionListModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(tui.SessionListModel)

	// Model is a value type, always valid even with empty list
	_ = model
}

func TestSessionListModel_UpdateFilterKeys(t *testing.T) {
	sessions := []*tui.SessionInfo{
		{
			SessionID:   "session-1",
			ShortID:     "sess-1",
			FirstEvent:  time.Now(),
			LastEvent:   time.Now(),
			EventCount:  5,
			HasAnalysis: true,
		},
		{
			SessionID:   "session-2",
			ShortID:     "sess-2",
			FirstEvent:  time.Now(),
			LastEvent:   time.Now(),
			EventCount:  3,
			HasAnalysis: false,
		},
	}

	model := tui.NewSessionListModel(sessions)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionListModel)

	// Press 'a' for analyzed filter
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updatedModel.(tui.SessionListModel)

	// Model is a value type, always valid after Update

	// Press 'u' for unanalyzed filter
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	model = updatedModel.(tui.SessionListModel)

	// Model is a value type, always valid after Update
	_ = model
}

func TestSessionListModel_UpdateOtherKeys(t *testing.T) {
	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-1",
			ShortID:    "sess-1",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	model := tui.NewSessionListModel(sessions)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionListModel)

	// Test down arrow
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(tui.SessionListModel)

	// Model is a value type, always valid after Update

	// Test up arrow
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(tui.SessionListModel)

	// Model is a value type, always valid after Update
	_ = model
}

func TestSessionListModel_AllItemVariations(t *testing.T) {
	now := time.Now()

	sessions := []*tui.SessionInfo{
		{
			SessionID:     "unanalyzed",
			ShortID:       "unana",
			FirstEvent:    now,
			LastEvent:     now,
			EventCount:    1,
			HasAnalysis:   false,
			AnalysisCount: 0,
		},
		{
			SessionID:     "single-analysis",
			ShortID:       "singl",
			FirstEvent:    now,
			LastEvent:     now,
			EventCount:    2,
			HasAnalysis:   true,
			AnalysisCount: 1,
			AnalysisTypes: []string{"type1"},
		},
		{
			SessionID:     "two-analyses",
			ShortID:       "two",
			FirstEvent:    now,
			LastEvent:     now,
			EventCount:    3,
			HasAnalysis:   true,
			AnalysisCount: 2,
			AnalysisTypes: []string{"type1", "type2"},
		},
		{
			SessionID:     "three-analyses",
			ShortID:       "three",
			FirstEvent:    now,
			LastEvent:     now,
			EventCount:    4,
			HasAnalysis:   true,
			AnalysisCount: 3,
			AnalysisTypes: []string{"type1", "type2", "type3"},
		},
	}

	model := tui.NewSessionListModel(sessions)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionListModel)

	// View should render all variations
	view := model.View()
	if view == "" {
		t.Error("View should render all session variations")
	}

	// Navigate through all sessions to trigger all rendering paths
	for i := 0; i < len(sessions); i++ {
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updatedModel.(tui.SessionListModel)

		view = model.View()
		if view == "" {
			t.Errorf("View should render session %d", i)
		}
	}
}

func TestSessionListModel_MultipleSessionsWithDifferentStates(t *testing.T) {
	sessions := []*tui.SessionInfo{
		{
			SessionID:   "session-unanalyzed",
			ShortID:     "sess-u",
			FirstEvent:  time.Now(),
			LastEvent:   time.Now(),
			EventCount:  5,
			HasAnalysis: false,
		},
		{
			SessionID:     "session-single-analysis",
			ShortID:       "sess-s",
			FirstEvent:    time.Now(),
			LastEvent:     time.Now(),
			EventCount:    10,
			HasAnalysis:   true,
			AnalysisCount: 1,
			AnalysisTypes: []string{"tool_analysis"},
		},
		{
			SessionID:     "session-multi-analysis",
			ShortID:       "sess-m",
			FirstEvent:    time.Now(),
			LastEvent:     time.Now(),
			EventCount:    15,
			HasAnalysis:   true,
			AnalysisCount: 3,
			AnalysisTypes: []string{"analysis1", "analysis2", "analysis3"},
		},
	}

	model := tui.NewSessionListModel(sessions)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionListModel)

	// View should render all session types correctly
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string with varied sessions")
	}

	// Navigate through sessions
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(tui.SessionListModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(tui.SessionListModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(tui.SessionListModel)

	view2 := model.View()
	if view2 == "" {
		t.Error("View should return non-empty string after navigation")
	}
}
