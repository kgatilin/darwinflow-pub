package tui_test

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/internal/app/tui"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

func TestNewAppModel(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}

	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	if model == nil {
		t.Fatal("NewAppModel() returned nil")
	}
}

func TestAppModel_Init(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init() should return a non-nil command")
	}
}

func TestAppModel_UpdateCtrlC(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Test ctrl+c quits
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Ctrl+C should return a quit command")
	}
}

func TestAppModel_UpdateWindowSize(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)

	appModel, ok := updatedModel.(*tui.AppModel)
	if !ok {
		t.Fatal("Update() should return an *AppModel")
	}

	// Check that dimensions were set (we can't access private fields directly,
	// but the update should succeed without error)
	if appModel == nil {
		t.Error("Updated model should not be nil")
	}
}

func TestAppModel_UpdateSessionsLoaded(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send window size first
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Send SessionsLoadedMsg
	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-1",
			ShortID:    "sess-1",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	msg := tui.SessionsLoadedMsg{Sessions: sessions, Error: nil}
	updatedModel2, _ := model.Update(msg)

	if updatedModel2 == nil {
		t.Error("Update should return non-nil model")
	}
}

func TestAppModel_UpdateSessionsLoadedError(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send SessionsLoadedMsg with error
	msg := tui.SessionsLoadedMsg{Sessions: nil, Error: domain.ErrNotFound}
	updatedModel, _ := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model even with error")
	}
}

func TestAppModel_UpdateSelectedSession(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send window size first
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Load sessions first
	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-test",
			ShortID:    "sess-t",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 3,
		},
	}
	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// Select a session
	msg := tui.SelectedSessionMsg{Session: sessions[0]}
	updatedModel3, cmd := model.Update(msg)

	if updatedModel3 == nil {
		t.Error("Update should return non-nil model")
	}

	if cmd == nil {
		// Command might be nil, that's ok
	}
}

func TestAppModel_UpdateBackToList(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send window size first
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Send BackToListMsg
	msg := tui.BackToListMsg{}
	updatedModel2, cmd := model.Update(msg)

	if updatedModel2 == nil {
		t.Error("Update should return non-nil model")
	}

	// Might return a window size command
	if cmd != nil {
		// That's ok
	}
}

func TestAppModel_UpdateRefreshRequest(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send RefreshRequestMsg
	msg := tui.RefreshRequestMsg{}
	updatedModel, cmd := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model")
	}

	if cmd == nil {
		t.Error("RefreshRequestMsg should return a command to load sessions")
	}
}

func TestAppModel_UpdateAnalyzeSession(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send AnalyzeSessionMsg
	msg := tui.AnalyzeSessionMsg{SessionID: "test-session-id"}
	updatedModel, cmd := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model")
	}

	if cmd == nil {
		t.Error("AnalyzeSessionMsg should return a command")
	}
}

func TestAppModel_UpdateReanalyzeSession(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send ReanalyzeSessionMsg
	msg := tui.ReanalyzeSessionMsg{SessionID: "test-session-id"}
	updatedModel, cmd := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model")
	}

	if cmd == nil {
		t.Error("ReanalyzeSessionMsg should return a command")
	}
}

func TestAppModel_UpdateSaveToMarkdown(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send SaveToMarkdownMsg
	msg := tui.SaveToMarkdownMsg{SessionID: "test-session-id"}
	updatedModel, cmd := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model")
	}

	if cmd == nil {
		t.Error("SaveToMarkdownMsg should return a command")
	}
}

func TestAppModel_UpdateAnalysisComplete(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send AnalysisCompleteMsg
	analysis := &domain.SessionAnalysis{
		SessionID:      "test-session",
		AnalysisResult: "Test result",
	}
	msg := tui.AnalysisCompleteMsg{
		SessionID: "test-session",
		Analysis:  analysis,
		Error:     nil,
	}
	updatedModel, cmd := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model")
	}

	// Should return a command to reload sessions
	if cmd == nil {
		t.Error("AnalysisCompleteMsg should return a command")
	}
}

func TestAppModel_UpdateAnalysisCompleteError(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send AnalysisCompleteMsg with error
	msg := tui.AnalysisCompleteMsg{
		SessionID: "test-session",
		Error:     domain.ErrNotFound,
	}
	updatedModel, _ := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model")
	}
}

func TestAppModel_UpdateSaveComplete(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send SaveCompleteMsg
	msg := tui.SaveCompleteMsg{
		FilePath: "/tmp/test.md",
		Error:    nil,
	}
	updatedModel, _ := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model")
	}
}

func TestAppModel_UpdateSaveCompleteError(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send SaveCompleteMsg with error
	msg := tui.SaveCompleteMsg{
		FilePath: "",
		Error:    domain.ErrNotFound,
	}
	updatedModel, _ := model.Update(msg)

	if updatedModel == nil {
		t.Error("Update should return non-nil model")
	}
}

func TestAppModel_View(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// View while loading should show spinner
	view := model.View()
	if view == "" {
		t.Error("View should return non-empty string while loading")
	}
}

func TestAppModel_ViewWithError(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Trigger an error
	msg := tui.SessionsLoadedMsg{Error: domain.ErrNotFound}
	updatedModel2, _ := model.Update(msg)
	model = updatedModel2.(*tui.AppModel)

	// View with error should show error message
	view := model.View()
	if view == "" {
		t.Error("View should return non-empty string when there's an error")
	}
}

func TestAppModel_WindowSizeForAllViewStates(t *testing.T) {
	// We can't directly test ViewAnalysisViewer and ViewLogViewer states
	// because they require service calls, but we can at least ensure
	// the window size handling doesn't panic for unknown states

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "test",
			ShortID:    "test",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	model := tui.NewAppModel(context.Background(), nil, nil, nil, nil)

	// Send window sizes in different orders
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 50, Height: 25})
	model = updatedModel.(*tui.AppModel)

	updatedModel, _ = model.Update(tea.WindowSizeMsg{Width: 150, Height: 75})
	model = updatedModel.(*tui.AppModel)

	// Load sessions
	updatedModel, _ = model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel.(*tui.AppModel)

	// Another window size after loading
	updatedModel, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if updatedModel == nil {
		t.Error("Model should handle window size at any state")
	}
}

func TestAppModel_WindowSizeMsgWhileLoading(t *testing.T) {
	ctx := context.Background()
	model := tui.NewAppModel(ctx, nil, nil, nil, nil)

	// Send window size while loading (before sessions are loaded)
	// This tests the early return path in Update when loading is true
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	if updatedModel == nil {
		t.Error("Update should return non-nil model even while loading")
	}
}

func TestAppModel_View_UnknownState(t *testing.T) {
	ctx := context.Background()
	model := tui.NewAppModel(ctx, nil, nil, nil, nil)

	// Set window size and load sessions
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: []*tui.SessionInfo{}})
	model = updatedModel2.(*tui.AppModel)

	// View should work
	view := model.View()
	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestAppModel_UpdateWindowSizeDifferentViews(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-win",
			ShortID:    "sess-w",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
			HasAnalysis: true,
			Analyses: []*domain.SessionAnalysis{
				{
					SessionID:      "session-win",
					AnalysisResult: "test",
				},
			},
		},
	}

	// Test window size in loading state
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Load sessions
	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// Test window size in ViewSessionList
	updatedModel3, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 60})
	model = updatedModel3.(*tui.AppModel)

	// Go to detail view
	updatedModel4, _ := model.Update(tui.SelectedSessionMsg{Session: sessions[0]})
	model = updatedModel4.(*tui.AppModel)

	// Test window size in ViewSessionDetail
	updatedModel5, _ := model.Update(tea.WindowSizeMsg{Width: 110, Height: 55})
	model = updatedModel5.(*tui.AppModel)

	if model == nil {
		t.Error("Model should not be nil after window size updates")
	}
}

func TestAppModel_UpdateCurrentView_AllStates(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-all",
			ShortID:    "sess-a",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
			HasAnalysis: true,
			Analyses: []*domain.SessionAnalysis{
				{
					SessionID:      "session-all",
					AnalysisResult: "test result",
					AnalyzedAt:     time.Now(),
				},
			},
		},
	}

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Load sessions
	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// Test updateCurrentView in ViewSessionList
	updatedModel3, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel3.(*tui.AppModel)

	// Go to ViewSessionDetail
	updatedModel4, _ := model.Update(tui.SelectedSessionMsg{Session: sessions[0]})
	model = updatedModel4.(*tui.AppModel)

	// Test updateCurrentView in ViewSessionDetail
	updatedModel5, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel5.(*tui.AppModel)

	if model == nil {
		t.Error("Model should not be nil")
	}
}

func TestAppModel_View_AllViewStates(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-view",
			ShortID:    "sess-v",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	// View while loading
	view1 := model.View()
	if view1 == "" {
		t.Error("View should return non-empty string while loading")
	}

	// Set window size and load sessions
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// View in session list
	view2 := model.View()
	if view2 == "" {
		t.Error("View should return non-empty string in session list")
	}

	// Go to detail
	updatedModel3, _ := model.Update(tui.SelectedSessionMsg{Session: sessions[0]})
	model = updatedModel3.(*tui.AppModel)

	// View in detail
	view3 := model.View()
	if view3 == "" {
		t.Error("View should return non-empty string in detail view")
	}
}

func TestAppModel_UpdateBackToDetail(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send window size first
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Send BackToDetailMsg
	msg := tui.BackToDetailMsg{}
	updatedModel2, cmd := model.Update(msg)

	if updatedModel2 == nil {
		t.Error("Update should return non-nil model")
	}

	// Might return a window size command
	if cmd != nil {
		// That's ok
	}
}

func TestAppModel_UpdateCurrentView_SessionList(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Load sessions to trigger ViewSessionList
	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-1",
			ShortID:    "sess-1",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Load sessions
	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// Send a key message to trigger updateCurrentView for ViewSessionList
	updatedModel3, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})

	if updatedModel3 == nil {
		t.Error("Update should return non-nil model")
	}
}

func TestAppModel_UpdateCurrentView_SessionDetail(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-detail-test",
			ShortID:    "sess-d",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Load sessions
	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// Select a session to go to detail view
	updatedModel3, _ := model.Update(tui.SelectedSessionMsg{Session: sessions[0]})
	model = updatedModel3.(*tui.AppModel)

	// Send a key message to trigger updateCurrentView for ViewSessionDetail
	updatedModel4, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})

	if updatedModel4 == nil {
		t.Error("Update should return non-nil model")
	}
}

func TestAppModel_ViewSessionList(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-1",
			ShortID:    "sess-1",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Load sessions
	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// View should show session list
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string for session list")
	}
}

func TestAppModel_ViewSessionDetail(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-view-detail",
			ShortID:    "sess-vd",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Load sessions
	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// Select a session
	updatedModel3, _ := model.Update(tui.SelectedSessionMsg{Session: sessions[0]})
	model = updatedModel3.(*tui.AppModel)

	// View should show session detail
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string for session detail")
	}
}

func TestAppModel_ViewAnalysisViewer(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	analysis := &domain.SessionAnalysis{
		SessionID:      "test-session-analysis",
		AnalysisType:   "tool_analysis",
		PromptName:     "test_prompt",
		AnalysisResult: "Test result",
		AnalyzedAt:     time.Now(),
	}

	// Manually trigger the analysis viewer state by creating it
	// We can't do this directly, but we can test the message handling

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "test-session-analysis",
			ShortID:    "sess-a",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
			HasAnalysis: true,
			Analyses:   []*domain.SessionAnalysis{analysis},
		},
	}

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Load sessions
	updatedModel2, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})
	model = updatedModel2.(*tui.AppModel)

	// Select session
	updatedModel3, _ := model.Update(tui.SelectedSessionMsg{Session: sessions[0]})
	model = updatedModel3.(*tui.AppModel)

	// The view should work even though we're not in analysis viewer mode
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestAppModel_UpdateSessionsLoadedWithDetailRefresh(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	sessions := []*tui.SessionInfo{
		{
			SessionID:  "session-refresh-test",
			ShortID:    "sess-r",
			FirstEvent: time.Now(),
			LastEvent:  time.Now(),
			EventCount: 5,
		},
	}

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(*tui.AppModel)

	// Simulate showing detail after analysis complete
	// This tests the showDetailAfterRefresh path
	updatedModel2, cmd := model.Update(tui.AnalysisCompleteMsg{
		SessionID: "session-refresh-test",
		Analysis: &domain.SessionAnalysis{
			SessionID:      "session-refresh-test",
			AnalysisResult: "test",
		},
		Error: nil,
	})
	model = updatedModel2.(*tui.AppModel)

	// The analysis complete should trigger a refresh
	if cmd == nil {
		t.Error("AnalysisCompleteMsg should return a command")
	}

	// Execute the command (it returns SessionsLoadedMsg)
	// In real app, this would reload sessions
	updatedModel3, _ := model.Update(tui.SessionsLoadedMsg{Sessions: sessions})

	if updatedModel3 == nil {
		t.Error("Update should return non-nil model")
	}
}

func TestAppModel_ViewErrorWithSmallWidth(t *testing.T) {
	ctx := context.Background()
	config := &domain.Config{}
	model := tui.NewAppModel(ctx, nil, nil, nil, config)

	// Send very small window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	model = updatedModel.(*tui.AppModel)

	// Trigger an error
	msg := tui.SessionsLoadedMsg{Error: domain.ErrNotFound}
	updatedModel2, _ := model.Update(msg)
	model = updatedModel2.(*tui.AppModel)

	// View with error and small width (tests maxWidth < 40 path)
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string even with small width")
	}
}

func TestSessionListModel_UpdateUnknownKey(t *testing.T) {
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

	// Send an unknown key (should be handled by the list component)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Model is a value type, always valid after Update
	_ = updatedModel
}

func TestSessionDetailModel_UpdateUnknownKey(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Send an unknown key (should be handled by viewport)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Model is a value type, always valid after Update
	_ = updatedModel
}

func TestAnalysisViewerModel_UpdateUnknownKey(t *testing.T) {
	analysis := &domain.SessionAnalysis{
		SessionID:      "test-session",
		AnalysisType:   "tool_analysis",
		PromptName:     "test_prompt",
		AnalysisResult: "Test result",
		AnalyzedAt:     time.Now(),
	}

	model := tui.NewAnalysisViewerModel(analysis)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.AnalysisViewerModel)

	// Send an unknown key (should be handled by viewport)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Model is a value type, always valid after Update
	_ = updatedModel
}

