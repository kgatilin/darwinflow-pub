package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	errorStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
)

// AppModel is the main orchestrator for the TUI
type AppModel struct {
	ctx             context.Context
	analysisService *app.AnalysisService
	logsService     *app.LogsService
	config          *domain.Config

	// State
	currentView ViewState
	sessions    []*SessionInfo
	loading     bool
	err         error

	// Sub-models
	sessionList   SessionListModel
	sessionDetail SessionDetailModel
	spinner       spinner.Model

	// Selected session for operations
	selectedSession *SessionInfo

	width  int
	height int
}

// NewAppModel creates a new TUI application model
func NewAppModel(
	ctx context.Context,
	analysisService *app.AnalysisService,
	logsService *app.LogsService,
	config *domain.Config,
) *AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return &AppModel{
		ctx:             ctx,
		analysisService: analysisService,
		logsService:     logsService,
		config:          config,
		currentView:     ViewSessionList,
		spinner:         s,
		loading:         true,
	}
}

// Init initializes the application
func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadSessions,
	)
}

// Update handles all messages
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Only update sub-models if they're initialized
		if !m.loading && m.currentView == ViewSessionList {
			var model tea.Model
			var cmd tea.Cmd
			model, cmd = m.sessionList.Update(msg)
			m.sessionList = model.(SessionListModel)
			return m, cmd
		}
		if !m.loading && m.currentView == ViewSessionDetail {
			var model tea.Model
			var cmd tea.Cmd
			model, cmd = m.sessionDetail.Update(msg)
			m.sessionDetail = model.(SessionDetailModel)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case SessionsLoadedMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
			return m, nil
		}
		m.sessions = msg.Sessions
		m.sessionList = NewSessionListModel(msg.Sessions)
		m.currentView = ViewSessionList
		// Send initial window size to the newly created list
		if m.width > 0 && m.height > 0 {
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}
		return m, nil

	case SelectedSessionMsg:
		m.selectedSession = msg.Session
		m.sessionDetail = NewSessionDetailModel(msg.Session)
		m.currentView = ViewSessionDetail
		// Send initial window size to the newly created detail view
		if m.width > 0 && m.height > 0 {
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}
		return m, nil

	case BackToListMsg:
		m.currentView = ViewSessionList
		// Send window size when returning to list view
		if m.width > 0 && m.height > 0 {
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}
		return m, nil

	case RefreshRequestMsg:
		m.loading = true
		return m, m.loadSessions

	case AnalyzeSessionMsg:
		m.loading = true
		return m, m.analyzeSession(msg.SessionID, "tool_analysis")

	case ReanalyzeSessionMsg:
		m.loading = true
		return m, m.analyzeSession(msg.SessionID, "tool_analysis")

	case SaveToMarkdownMsg:
		m.loading = true
		return m, m.saveToMarkdown(msg.SessionID)

	case AnalysisCompleteMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
		} else {
			// Refresh session data
			return m, m.loadSessions
		}
		return m, nil

	case SaveCompleteMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
		} else {
			// Success message handled in view
			m.err = nil
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Route to appropriate sub-model
	return m.updateCurrentView(msg)
}

func (m *AppModel) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Don't route to sub-models if we're still loading
	if m.loading {
		return m, nil
	}

	var cmd tea.Cmd

	switch m.currentView {
	case ViewSessionList:
		var model tea.Model
		model, cmd = m.sessionList.Update(msg)
		m.sessionList = model.(SessionListModel)

	case ViewSessionDetail:
		var model tea.Model
		model, cmd = m.sessionDetail.Update(msg)
		m.sessionDetail = model.(SessionDetailModel)
	}

	return m, cmd
}

// View renders the current view
func (m *AppModel) View() string {
	if m.err != nil {
		// Wrap error text to terminal width (with some padding for border)
		maxWidth := m.width - 10
		if maxWidth < 40 {
			maxWidth = 40
		}
		errText := fmt.Sprintf("Error: %v\n\nPress ctrl+c to quit", m.err)
		wrappedErr := lipgloss.NewStyle().Width(maxWidth).Render(errText)
		return errorStyle.Render(wrappedErr)
	}

	if m.loading {
		return fmt.Sprintf("\n\n   %s Loading...\n\n", m.spinner.View())
	}

	switch m.currentView {
	case ViewSessionList:
		return m.sessionList.View()
	case ViewSessionDetail:
		return m.sessionDetail.View()
	default:
		return "Unknown view"
	}
}

// Command functions

func (m *AppModel) loadSessions() tea.Msg {
	// Get all session IDs
	sessionIDs, err := m.analysisService.GetAllSessionIDs(m.ctx, 0)
	if err != nil {
		return SessionsLoadedMsg{Error: err}
	}

	sessions := make([]*SessionInfo, 0, len(sessionIDs))

	for _, sessionID := range sessionIDs {
		// Get session logs
		logs, err := m.logsService.ListRecentLogs(m.ctx, 0, 0, sessionID, true)
		if err != nil || len(logs) == 0 {
			continue
		}

		// Get analyses for this session
		analyses, err := m.analysisService.GetAnalysesBySessionID(m.ctx, sessionID)
		if err != nil {
			analyses = []*domain.SessionAnalysis{}
		}

		// Build session info
		sessionInfo := &SessionInfo{
			SessionID:     sessionID,
			ShortID:       sessionID[:8],
			FirstEvent:    logs[0].Timestamp,
			LastEvent:     logs[len(logs)-1].Timestamp,
			EventCount:    len(logs),
			AnalysisCount: len(analyses),
			Analyses:      analyses,
			HasAnalysis:   len(analyses) > 0,
			AnalysisTypes: make([]string, 0, len(analyses)),
		}

		for _, a := range analyses {
			sessionInfo.AnalysisTypes = append(sessionInfo.AnalysisTypes, a.PromptName)
		}

		if len(analyses) > 0 {
			sessionInfo.LatestAnalysis = analyses[0]
		}

		sessions = append(sessions, sessionInfo)
	}

	return SessionsLoadedMsg{Sessions: sessions}
}

func (m *AppModel) analyzeSession(sessionID, promptName string) tea.Cmd {
	return func() tea.Msg {
		analysis, err := m.analysisService.AnalyzeSessionWithPrompt(m.ctx, sessionID, promptName)
		return AnalysisCompleteMsg{
			SessionID: sessionID,
			Analysis:  analysis,
			Error:     err,
		}
	}
}

func (m *AppModel) saveToMarkdown(sessionID string) tea.Cmd {
	return func() tea.Msg {
		// Get the latest analysis
		analyses, err := m.analysisService.GetAnalysesBySessionID(m.ctx, sessionID)
		if err != nil || len(analyses) == 0 {
			return SaveCompleteMsg{Error: fmt.Errorf("no analysis found for session")}
		}

		analysis := analyses[0] // Use most recent

		filePath, err := m.analysisService.SaveToMarkdown(m.ctx, analysis, "", "")
		return SaveCompleteMsg{
			FilePath: filePath,
			Error:    err,
		}
	}
}

// Run starts the TUI application
func Run(
	ctx context.Context,
	analysisService *app.AnalysisService,
	logsService *app.LogsService,
	config *domain.Config,
) error {
	m := NewAppModel(ctx, analysisService, logsService, config)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
