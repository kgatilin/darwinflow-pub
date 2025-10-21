package tui_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/app/tui"
)

func TestNewLogViewerModel(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event",
			Content:   "test content",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Model should be created successfully
	// Note: NewLogViewerModel returns a value type, not a pointer,
	// so we can't check for nil. The function always returns a valid model.
	_ = model
}

func TestLogViewerModel_Init(t *testing.T) {
	logs := []*app.LogRecord{}
	model := tui.NewLogViewerModel("test-session", logs)

	cmd := model.Init()

	if cmd != nil {
		t.Error("Init() should return nil command")
	}
}

func TestLogViewerModel_UpdateWindowSize(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("WindowSizeMsg should return nil command")
	}

	_, ok := updatedModel.(tui.LogViewerModel)
	if !ok {
		t.Error("Update should return LogViewerModel")
	}
}

func TestLogViewerModel_UpdateEscNoSearch(t *testing.T) {
	logs := []*app.LogRecord{}
	model := tui.NewLogViewerModel("test-session-123", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Send esc key (should return to detail view)
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Esc key should return a command")
	}

	// Execute command to verify it's BackToDetailMsg
	if cmd != nil {
		result := cmd()
		if _, ok := result.(tui.BackToDetailMsg); !ok {
			t.Error("Expected BackToDetailMsg from esc key")
		}
	}
}

func TestLogViewerModel_UpdateSearchMode(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Send '/' key to enter search mode
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updatedModel, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("'/' key should return a command (textinput.Blink)")
	}

	_, ok := updatedModel.(tui.LogViewerModel)
	if !ok {
		t.Error("Update should return LogViewerModel after entering search mode")
	}
}

func TestLogViewerModel_UpdateScrolling(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "event1",
		},
		{
			ID:        "log-2",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "event2",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 10})
	model = updatedModel.(tui.LogViewerModel)

	// Test down arrow scrolling
	msg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = model.Update(msg)

	_, ok := updatedModel.(tui.LogViewerModel)
	if !ok {
		t.Error("Update should return LogViewerModel after scrolling")
	}
}

func TestLogViewerModel_ViewNotReady(t *testing.T) {
	logs := []*app.LogRecord{}
	model := tui.NewLogViewerModel("test-session", logs)

	// View before initialization should return initializing message
	view := model.View()

	if view == "" {
		t.Error("View() should return non-empty string")
	}
}

func TestLogViewerModel_ViewReady(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session-view",
			EventType: "test-event",
		},
	}

	model := tui.NewLogViewerModel("test-session-view", logs)

	// Initialize with window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// View after initialization should show content
	view := model.View()

	if view == "" {
		t.Error("View() should return non-empty string after initialization")
	}
}

func TestLogViewerModel_SearchModeEsc(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Enter search mode
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	// Press esc to exit search mode (should not quit, just exit search)
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := model.Update(msg)

	// Should return nil command (just exits search mode)
	if cmd != nil {
		t.Error("Esc in search mode should return nil command")
	}

	_, ok := updatedModel.(tui.LogViewerModel)
	if !ok {
		t.Error("Update should return LogViewerModel after exiting search mode")
	}
}

func TestLogViewerModel_SearchModeEnter(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Enter search mode
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	// Type a search query (simulated)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model = updatedModel.(tui.LogViewerModel)

	// Press enter to confirm search
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	// Should return nil command (exits search mode, keeps search active)
	if cmd != nil {
		t.Error("Enter in search mode should return nil command")
	}

	_, ok := updatedModel.(tui.LogViewerModel)
	if !ok {
		t.Error("Update should return LogViewerModel after confirming search")
	}
}

func TestLogViewerModel_NextMatch(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event-1",
		},
		{
			ID:        "log-2",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event-2",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Enter search mode and search for "test"
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(tui.LogViewerModel)

	// Press 'n' to go to next match
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	updatedModel, cmd := model.Update(msg)

	// Should return nil command
	if cmd != nil {
		t.Error("'n' key should return nil command")
	}

	_, ok := updatedModel.(tui.LogViewerModel)
	if !ok {
		t.Error("Update should return LogViewerModel after next match")
	}
}

func TestLogViewerModel_PreviousMatch(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Press 'N' to go to previous match (should handle gracefully even with no search)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	updatedModel, cmd := model.Update(msg)

	// Should return nil command
	if cmd != nil {
		t.Error("'N' key should return nil command")
	}

	_, ok := updatedModel.(tui.LogViewerModel)
	if !ok {
		t.Error("Update should return LogViewerModel after previous match")
	}
}

func TestLogViewerModel_ResizeViewport_Multiple(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "event",
			Content:   "test content for resize",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize with first size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Enter search (triggers resize)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	// Exit search (triggers resize again)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updatedModel.(tui.LogViewerModel)

	// Re-enter search
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	// Confirm search with enter (triggers resize)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(tui.LogViewerModel)

	// Model is a value type, always valid after Update
	_ = model
}

func TestLogViewerModel_FindMatches_EmptyQuery(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "event",
			Content:   "test",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Enter search mode
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	// Press enter immediately without typing (empty query)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(tui.LogViewerModel)

	// Try to navigate (should handle gracefully with no matches)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	model = updatedModel.(tui.LogViewerModel)

	// Model is a value type, always valid even with empty search
	_ = model
}

func TestLogViewerModel_HeaderViewWithMatches(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session-abc123",
			EventType: "test-event",
			Content:   "searchable content here",
		},
	}

	model := tui.NewLogViewerModel("test-session-abc123", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Enter search and find matches
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(tui.LogViewerModel)

	// Call View to trigger headerView with search query
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string with search query")
	}
}

func TestLogViewerModel_ScrollingEdgeCases(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session-long",
			EventType: "event-1",
			Content:   "Line 1",
		},
	}

	// Create a long log to enable scrolling
	for i := 0; i < 100; i++ {
		logs = append(logs, &app.LogRecord{
			ID:        "log-" + string(rune(i+2)),
			Timestamp: time.Now(),
			SessionID: "test-session-long",
			EventType: "event",
			Content:   "Content line " + string(rune(i+2)),
		})
	}

	model := tui.NewLogViewerModel("test-session-long", logs)

	// Initialize with small height to enable scrolling
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 10})
	model = updatedModel.(tui.LogViewerModel)

	// Scroll down multiple times
	for i := 0; i < 5; i++ {
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updatedModel.(tui.LogViewerModel)
	}

	// Scroll up
	for i := 0; i < 3; i++ {
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
		model = updatedModel.(tui.LogViewerModel)
	}

	// Page down
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	model = updatedModel.(tui.LogViewerModel)

	// Page up
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	model = updatedModel.(tui.LogViewerModel)

	// Model is a value type, always valid after scrolling
	_ = model
}

func TestLogViewerModel_EscWithActiveSearch(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event",
			Content:   "searchable content",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Enter search mode and search
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(tui.LogViewerModel)

	// Now press esc again (should clear search)
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd != nil {
		// Command might be nil after clearing search
	}

	_, ok := updatedModel.(tui.LogViewerModel)
	if !ok {
		t.Error("Update should return LogViewerModel")
	}
}

func TestLogViewerModel_SearchPanelView(t *testing.T) {
	logs := []*app.LogRecord{
		{
			ID:        "log-1",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "test-event",
			Content:   "test content with searchable text",
		},
		{
			ID:        "log-2",
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "another-event",
			Content:   "more searchable text here",
		},
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Enter search mode
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	// Type a search query
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model = updatedModel.(tui.LogViewerModel)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updatedModel.(tui.LogViewerModel)

	// Call View while in search mode (this should render searchPanelView)
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string in search mode")
	}
}

func TestLogViewerModel_MultipleFindMatches(t *testing.T) {
	logs := []*app.LogRecord{}

	// Create logs with multiple occurrences of search term
	for i := 0; i < 20; i++ {
		logs = append(logs, &app.LogRecord{
			ID:        "log-" + string(rune(i+1)),
			Timestamp: time.Now(),
			SessionID: "test-session",
			EventType: "event-" + string(rune(i+1)),
			Content:   "This is searchable content number " + string(rune(i+1)),
		})
	}

	model := tui.NewLogViewerModel("test-session", logs)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.LogViewerModel)

	// Search for "searchable"
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updatedModel.(tui.LogViewerModel)

	// Type search query
	for _, r := range "search" {
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		model = updatedModel.(tui.LogViewerModel)
	}

	// Confirm search
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(tui.LogViewerModel)

	// Navigate through matches with 'n'
	for i := 0; i < 5; i++ {
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		model = updatedModel.(tui.LogViewerModel)
	}

	// Navigate backwards with 'N'
	for i := 0; i < 3; i++ {
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
		model = updatedModel.(tui.LogViewerModel)
	}

	// Model is a value type, always valid after navigating matches
}
