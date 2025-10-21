package tui_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/internal/app/tui"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

func TestFormatTokenCount(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "small number",
			count: 123,
			want:  "~123 tokens",
		},
		{
			name:  "one thousand",
			count: 1000,
			want:  "~1,000 tokens",
		},
		{
			name:  "five thousand",
			count: 5432,
			want:  "~5,432 tokens",
		},
		{
			name:  "fifty thousand",
			count: 50000,
			want:  "~50,000 tokens",
		},
		{
			name:  "one hundred thousand",
			count: 123456,
			want:  "~123,456 tokens",
		},
		{
			name:  "one million",
			count: 1000000,
			want:  "~1,000,000 tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tui.FormatTokenCount(tt.count)
			if got != tt.want {
				t.Errorf("FormatTokenCount(%d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}

func TestNewSessionDetailModel(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 10,
		HasAnalysis: false,
	}

	model := tui.NewSessionDetailModel(session)

	// Model should be created successfully
	// Note: NewSessionDetailModel returns a value type, not a pointer,
	// so we can't check for nil. The function always returns a valid model.
	_ = model
}

func TestSessionDetailModel_Init(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)
	cmd := model.Init()

	if cmd != nil {
		t.Error("Init() should return nil command")
	}
}

func TestSessionDetailModel_UpdateWindowSize(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, cmd := model.Update(msg)

	if cmd != nil {
		t.Error("WindowSizeMsg should return nil command")
	}

	_, ok := updatedModel.(tui.SessionDetailModel)
	if !ok {
		t.Error("Update should return SessionDetailModel")
	}
}

func TestSessionDetailModel_UpdateEsc(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize viewport with window size first
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Send esc key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Esc key should return a command")
	}

	// Execute command to verify it's BackToListMsg
	if cmd != nil {
		result := cmd()
		if _, ok := result.(tui.BackToListMsg); !ok {
			t.Error("Expected BackToListMsg from esc key")
		}
	}
}

func TestSessionDetailModel_UpdateAnalyze(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session-123",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
		HasAnalysis: false,
	}

	model := tui.NewSessionDetailModel(session)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Send 'a' key to analyze
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("'a' key should return analyze command")
	}

	// Execute command to verify it's AnalyzeSessionMsg
	if cmd != nil {
		result := cmd()
		if analyzeMsg, ok := result.(tui.AnalyzeSessionMsg); ok {
			if analyzeMsg.SessionID != "test-session-123" {
				t.Errorf("Expected session ID 'test-session-123', got %s", analyzeMsg.SessionID)
			}
		} else {
			t.Error("Expected AnalyzeSessionMsg from 'a' key")
		}
	}
}

func TestSessionDetailModel_UpdateReanalyze(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session-456",
		ShortID:    "test456",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-session-456",
				AnalysisType:   "tool_analysis",
				PromptName:     "test_prompt",
				AnalysisResult: "test result",
			},
		},
	}

	model := tui.NewSessionDetailModel(session)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Send 'r' key to re-analyze
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("'r' key should return re-analyze command")
	}

	// Execute command to verify it's ReanalyzeSessionMsg
	if cmd != nil {
		result := cmd()
		if reanalyzeMsg, ok := result.(tui.ReanalyzeSessionMsg); ok {
			if reanalyzeMsg.SessionID != "test-session-456" {
				t.Errorf("Expected session ID 'test-session-456', got %s", reanalyzeMsg.SessionID)
			}
		} else {
			t.Error("Expected ReanalyzeSessionMsg from 'r' key")
		}
	}
}

func TestSessionDetailModel_UpdateSave(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session-789",
		ShortID:    "test789",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Send 's' key to save
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("'s' key should return save command")
	}

	// Execute command to verify it's SaveToMarkdownMsg
	if cmd != nil {
		result := cmd()
		if saveMsg, ok := result.(tui.SaveToMarkdownMsg); ok {
			if saveMsg.SessionID != "test-session-789" {
				t.Errorf("Expected session ID 'test-session-789', got %s", saveMsg.SessionID)
			}
		} else {
			t.Error("Expected SaveToMarkdownMsg from 's' key")
		}
	}
}

func TestSessionDetailModel_UpdateViewAnalysis(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session-view",
		ShortID:    "testview",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Send 'v' key to view analysis
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("'v' key should return view analysis command")
	}

	// Execute command to verify it's ViewAnalysisMsg
	if cmd != nil {
		result := cmd()
		if viewMsg, ok := result.(tui.ViewAnalysisMsg); ok {
			if viewMsg.SessionID != "test-session-view" {
				t.Errorf("Expected session ID 'test-session-view', got %s", viewMsg.SessionID)
			}
		} else {
			t.Error("Expected ViewAnalysisMsg from 'v' key")
		}
	}
}

func TestSessionDetailModel_UpdateViewLog(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session-log",
		ShortID:    "testlog",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Send 'l' key to view log
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("'l' key should return view log command")
	}

	// Execute command to verify it's ViewLogMsg
	if cmd != nil {
		result := cmd()
		if logMsg, ok := result.(tui.ViewLogMsg); ok {
			if logMsg.SessionID != "test-session-log" {
				t.Errorf("Expected session ID 'test-session-log', got %s", logMsg.SessionID)
			}
		} else {
			t.Error("Expected ViewLogMsg from 'l' key")
		}
	}
}

func TestSessionDetailModel_ViewNotReady(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)

	// View before initialization should return initializing message
	view := model.View()

	if view == "" {
		t.Error("View() should return non-empty string")
	}
}

func TestSessionDetailModel_FooterView_AllPaths(t *testing.T) {
	// Session with no analysis (different footer actions)
	session1 := &tui.SessionInfo{
		SessionID:   "test-no-analysis",
		ShortID:     "test-na",
		FirstEvent:  time.Now(),
		LastEvent:   time.Now(),
		EventCount:  5,
		HasAnalysis: false,
	}

	model1 := tui.NewSessionDetailModel(session1)
	updatedModel1, _ := model1.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model1 = updatedModel1.(tui.SessionDetailModel)

	view1 := model1.View()
	if view1 == "" {
		t.Error("View should render footer without analysis actions")
	}

	// Session with analysis (full footer actions)
	session2 := &tui.SessionInfo{
		SessionID:  "test-with-analysis",
		ShortID:    "test-wa",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-with-analysis",
				AnalysisResult: "test",
			},
		},
	}

	model2 := tui.NewSessionDetailModel(session2)
	updatedModel2, _ := model2.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model2 = updatedModel2.(tui.SessionDetailModel)

	view2 := model2.View()
	if view2 == "" {
		t.Error("View should render footer with all analysis actions")
	}

	// Test with different viewport scroll percentages
	longSession := &tui.SessionInfo{
		SessionID:  "test-long",
		ShortID:    "test-l",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 100,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-long",
				AnalysisResult: string(make([]byte, 5000)), // Very long
			},
		},
	}

	model3 := tui.NewSessionDetailModel(longSession)
	updatedModel3, _ := model3.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	model3 = updatedModel3.(tui.SessionDetailModel)

	// Scroll to trigger different scroll percentages
	for i := 0; i < 10; i++ {
		updatedModel3, _ = model3.Update(tea.KeyMsg{Type: tea.KeyDown})
		model3 = updatedModel3.(tui.SessionDetailModel)

		view := model3.View()
		if view == "" {
			t.Error("View should render at different scroll positions")
		}
	}
}

func TestSessionDetailModel_RenderContent_CompleteCoverage(t *testing.T) {
	// Multiple analyses with exactly 300-char preview
	preview := string(make([]byte, 300))

	session := &tui.SessionInfo{
		SessionID:     "test-multi",
		ShortID:       "test-m",
		FirstEvent:    time.Now(),
		LastEvent:     time.Now(),
		EventCount:    10,
		TokenCount:    50000, // With token count
		HasAnalysis:   true,
		AnalysisCount: 2,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-multi",
				AnalysisType:   "analysis_type_1",
				PromptName:     "prompt_1",
				ModelUsed:      "model_1",
				AnalysisResult: preview, // Exactly 300 chars
				AnalyzedAt:     time.Now().Add(-1 * time.Hour),
			},
			{
				SessionID:      "test-multi",
				AnalysisType:   "analysis_type_2",
				PromptName:     "prompt_2",
				ModelUsed:      "model_2",
				AnalysisResult: string(make([]byte, 500)), // > 300, will truncate
				AnalyzedAt:     time.Now(),
			},
		},
	}

	model := tui.NewSessionDetailModel(session)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	view := model.View()
	if view == "" {
		t.Error("View should render all content branches")
	}
}

func TestSessionDetailModel_RenderDifferentConfigurations(t *testing.T) {
	// Test 1: Session with analysis but empty AnalysisResult
	session1 := &tui.SessionInfo{
		SessionID:  "test-empty-result",
		ShortID:    "test-er",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
		HasAnalysis: false, // No analysis
	}

	model1 := tui.NewSessionDetailModel(session1)
	updatedModel1, _ := model1.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model1 = updatedModel1.(tui.SessionDetailModel)

	view1 := model1.View()
	if view1 == "" {
		t.Error("View should work with no analysis")
	}

	// Test 2: Session with exactly 300 chars preview (boundary test)
	preview300 := ""
	for i := 0; i < 30; i++ {
		preview300 += "0123456789"
	}

	session2 := &tui.SessionInfo{
		SessionID:   "test-300",
		ShortID:     "test-3",
		FirstEvent:  time.Now(),
		LastEvent:   time.Now(),
		EventCount:  5,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-300",
				AnalysisResult: preview300, // Exactly 300 chars
			},
		},
	}

	model2 := tui.NewSessionDetailModel(session2)
	updatedModel2, _ := model2.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model2 = updatedModel2.(tui.SessionDetailModel)

	view2 := model2.View()
	if view2 == "" {
		t.Error("View should work with 300 char preview")
	}

	// Test 3: Session with 301 chars preview (triggers truncation)
	preview301 := preview300 + "1"

	session3 := &tui.SessionInfo{
		SessionID:   "test-301",
		ShortID:     "test-3",
		FirstEvent:  time.Now(),
		LastEvent:   time.Now(),
		EventCount:  5,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-301",
				AnalysisResult: preview301, // 301 chars - should truncate
			},
		},
	}

	model3 := tui.NewSessionDetailModel(session3)
	updatedModel3, _ := model3.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model3 = updatedModel3.(tui.SessionDetailModel)

	view3 := model3.View()
	if view3 == "" {
		t.Error("View should work with truncated preview")
	}
}

func TestSessionDetailModel_Max_BothBranches(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test",
		ShortID:    "test",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)

	// Test where a > b (normal case)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	view1 := model.View()
	if view1 == "" {
		t.Error("View should work when width is large")
	}

	// Test where b > a (returns b)
	updatedModel, _ = model.Update(tea.WindowSizeMsg{Width: 3, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	view2 := model.View()
	if view2 == "" {
		t.Error("View should work when width is tiny")
	}
}

func TestSessionDetailModel_UpdateOtherKeys(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 10,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-session",
				AnalysisType:   "tool_analysis",
				PromptName:     "test_prompt",
				AnalysisResult: "Long analysis result\n" + string(make([]byte, 1000)),
				AnalyzedAt:     time.Now(),
			},
		},
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	model = updatedModel.(tui.SessionDetailModel)

	// Test up arrow for scrolling
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(tui.SessionDetailModel)

	// Model is a value type, always valid after Update

	// Test down arrow for scrolling
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(tui.SessionDetailModel)

	// Model is a value type, always valid after Update
	_ = model
}

func TestSessionDetailModel_FooterVariations(t *testing.T) {
	// Test footer with session that has no analysis
	session := &tui.SessionInfo{
		SessionID:   "test-session",
		ShortID:     "test123",
		FirstEvent:  time.Now(),
		LastEvent:   time.Now(),
		EventCount:  5,
		HasAnalysis: false,
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize viewport
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	view := model.View()
	if view == "" {
		t.Error("View should return non-empty string for unanalyzed session")
	}
}

func TestSessionDetailModel_MaxWithNegativeResult(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123456789", // Long ID
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize with very small width to force max(0, negative)
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 1, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	view := model.View()
	if view == "" {
		t.Error("View should handle very small width")
	}
}

func TestSessionDetailModel_ScrollPercent(t *testing.T) {
	longContent := ""
	for i := 0; i < 1000; i++ {
		longContent += "Very long content line " + string(rune(i)) + "\n"
	}

	session := &tui.SessionInfo{
		SessionID:  "test-scroll",
		ShortID:    "test-s",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-scroll",
				AnalysisResult: longContent,
			},
		},
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize with small height to enable scrolling
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 10})
	model = updatedModel.(tui.SessionDetailModel)

	// Scroll to different positions
	for i := 0; i < 10; i++ {
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updatedModel.(tui.SessionDetailModel)

		// Call View to trigger ScrollPercent calculation in footer
		view := model.View()
		if view == "" {
			t.Error("View should return non-empty string while scrolling")
		}
	}

	// Scroll to bottom
	for i := 0; i < 100; i++ {
		updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updatedModel.(tui.SessionDetailModel)
	}

	view := model.View()
	if view == "" {
		t.Error("View should return non-empty string at bottom")
	}
}

func TestSessionDetailModel_RenderContent_AllBranches(t *testing.T) {
	// Session with zero token count (skips token count display)
	session1 := &tui.SessionInfo{
		SessionID:   "test-session-no-tokens",
		ShortID:     "test-nt",
		FirstEvent:  time.Now(),
		LastEvent:   time.Now(),
		EventCount:  5,
		TokenCount:  0,
		HasAnalysis: false,
	}

	model1 := tui.NewSessionDetailModel(session1)
	updatedModel1, _ := model1.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model1 = updatedModel1.(tui.SessionDetailModel)

	view1 := model1.View()
	if view1 == "" {
		t.Error("View should return non-empty string without token count")
	}

	// Session with analysis
	session2 := &tui.SessionInfo{
		SessionID:     "test-session-with-analysis",
		ShortID:       "test-wa",
		FirstEvent:    time.Now(),
		LastEvent:     time.Now(),
		EventCount:    5,
		TokenCount:    5000,
		HasAnalysis:   true,
		AnalysisCount: 2,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-session-with-analysis",
				AnalysisType:   "type1",
				PromptName:     "prompt1",
				ModelUsed:      "model1",
				AnalysisResult: "Short result",
				AnalyzedAt:     time.Now(),
			},
			{
				SessionID:      "test-session-with-analysis",
				AnalysisType:   "type2",
				PromptName:     "prompt2",
				ModelUsed:      "model2",
				AnalysisResult: "Another short result",
				AnalyzedAt:     time.Now(),
			},
		},
	}

	model2 := tui.NewSessionDetailModel(session2)
	updatedModel2, _ := model2.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model2 = updatedModel2.(tui.SessionDetailModel)

	view2 := model2.View()
	if view2 == "" {
		t.Error("View should return non-empty string with analysis")
	}
}

func TestSessionDetailModel_RenderWithTokenCount(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 10,
		TokenCount: 50000, // Large token count to test rendering
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-session",
				AnalysisType:   "tool_analysis",
				PromptName:     "test_prompt",
				ModelUsed:      "test-model",
				AnalysisResult: "Test analysis result with some content",
				AnalyzedAt:     time.Now(),
			},
		},
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize with window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Call View to trigger renderContent with token count
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestSessionDetailModel_MaxFunctionBothPaths(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)

	// Test with width that triggers different max() paths
	// Very large width
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	view1 := model.View()
	if view1 == "" {
		t.Error("View should return non-empty string with large width")
	}

	// Small width (triggers max(0, ...) returning 0)
	updatedModel, _ = model.Update(tea.WindowSizeMsg{Width: 5, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	view2 := model.View()
	if view2 == "" {
		t.Error("View should return non-empty string with small width")
	}
}

func TestSessionDetailModel_MultipleAnalyses(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:     "test-session-multi",
		ShortID:       "test-m",
		FirstEvent:    time.Now(),
		LastEvent:     time.Now(),
		EventCount:    10,
		AnalysisCount: 3,
		HasAnalysis:   true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-session-multi",
				AnalysisType:   "analysis_1",
				PromptName:     "prompt_1",
				ModelUsed:      "model_1",
				AnalysisResult: "Result 1 with some content",
				AnalyzedAt:     time.Now().Add(-2 * time.Hour),
			},
			{
				SessionID:      "test-session-multi",
				AnalysisType:   "analysis_2",
				PromptName:     "prompt_2",
				ModelUsed:      "model_2",
				AnalysisResult: "Result 2 with different content",
				AnalyzedAt:     time.Now().Add(-1 * time.Hour),
			},
			{
				SessionID:      "test-session-multi",
				AnalysisType:   "analysis_3",
				PromptName:     "prompt_3",
				ModelUsed:      "model_3",
				AnalysisResult: "Result 3 with more content here",
				AnalyzedAt:     time.Now(),
			},
		},
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// View should render all analyses
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string with multiple analyses")
	}
}

func TestSessionDetailModel_LongAnalysisPreview(t *testing.T) {
	longResult := ""
	for i := 0; i < 500; i++ {
		longResult += "A"
	}

	session := &tui.SessionInfo{
		SessionID:   "test-session-long",
		ShortID:     "test-l",
		FirstEvent:  time.Now(),
		LastEvent:   time.Now(),
		EventCount:  5,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-session-long",
				AnalysisType:   "tool_analysis",
				PromptName:     "test_prompt",
				AnalysisResult: longResult, // > 300 chars to trigger truncation
				AnalyzedAt:     time.Now(),
			},
		},
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// View should truncate long preview
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string with long analysis")
	}
}

func TestSessionDetailModel_Max(t *testing.T) {
	// The max function is internal, but we can test it indirectly through rendering
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize with very small width to trigger max() edge cases
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 10, Height: 10})
	model = updatedModel.(tui.SessionDetailModel)

	// Should not panic even with small dimensions
	view := model.View()
	if view == "" {
		t.Error("View should return non-empty string even with small dimensions")
	}
}

func TestSessionDetailModel_View_AfterInit(t *testing.T) {
	session := &tui.SessionInfo{
		SessionID:  "test-session",
		ShortID:    "test123",
		FirstEvent: time.Now(),
		LastEvent:  time.Now(),
		EventCount: 5,
		HasAnalysis: true,
		Analyses: []*domain.SessionAnalysis{
			{
				SessionID:      "test-session",
				AnalysisType:   "tool_analysis",
				PromptName:     "test_prompt",
				ModelUsed:      "test-model",
				AnalysisResult: "Test result",
				AnalyzedAt:     time.Now(),
			},
		},
		TokenCount: 1500,
	}

	model := tui.NewSessionDetailModel(session)

	// Initialize with window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model = updatedModel.(tui.SessionDetailModel)

	// Call View
	view := model.View()

	if view == "" {
		t.Error("View should return non-empty string after initialization")
	}
}
