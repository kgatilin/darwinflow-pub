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
	sessionList     SessionListModel
	sessionDetail   SessionDetailModel
	analysisViewer  AnalysisViewerModel
	logViewer       LogViewerModel
	spinner         spinner.Model

	// Selected session for operations
	selectedSession *SessionInfo

	// Flag to track if we should show detail view after refresh
	showDetailAfterRefresh bool

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
		if !m.loading && m.currentView == ViewAnalysisViewer {
			var model tea.Model
			var cmd tea.Cmd
			model, cmd = m.analysisViewer.Update(msg)
			m.analysisViewer = model.(AnalysisViewerModel)
			return m, cmd
		}
		if !m.loading && m.currentView == ViewLogViewer {
			var model tea.Model
			var cmd tea.Cmd
			model, cmd = m.logViewer.Update(msg)
			m.logViewer = model.(LogViewerModel)
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

		// If we should show detail view after refresh (e.g., after analysis)
		if m.showDetailAfterRefresh && m.selectedSession != nil {
			m.showDetailAfterRefresh = false
			// Find the updated session info
			for _, session := range msg.Sessions {
				if session.SessionID == m.selectedSession.SessionID {
					m.selectedSession = session
					m.sessionDetail = NewSessionDetailModel(session)
					m.currentView = ViewSessionDetail
					// Send initial window size to the detail view
					if m.width > 0 && m.height > 0 {
						return m, func() tea.Msg {
							return tea.WindowSizeMsg{Width: m.width, Height: m.height}
						}
					}
					return m, nil
				}
			}
		}

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

	case ViewAnalysisMsg:
		// Get the analysis for this session
		analyses, err := m.analysisService.GetAnalysesBySessionID(m.ctx, msg.SessionID)
		if err != nil || len(analyses) == 0 {
			m.err = fmt.Errorf("no analysis found for session")
			return m, nil
		}
		// Use the most recent analysis
		m.analysisViewer = NewAnalysisViewerModel(analyses[0])
		m.currentView = ViewAnalysisViewer
		// Send initial window size to the viewer
		if m.width > 0 && m.height > 0 {
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}
		return m, nil

	case ViewLogMsg:
		// Get the logs for this session
		logs, err := m.logsService.ListRecentLogs(m.ctx, 0, 0, msg.SessionID, true)
		if err != nil || len(logs) == 0 {
			m.err = fmt.Errorf("no logs found for session")
			return m, nil
		}
		m.logViewer = NewLogViewerModel(msg.SessionID, logs)
		m.currentView = ViewLogViewer
		// Send initial window size to the viewer
		if m.width > 0 && m.height > 0 {
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}
		return m, nil

	case BackToDetailMsg:
		m.currentView = ViewSessionDetail
		// Send window size when returning to detail view
		if m.width > 0 && m.height > 0 {
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}
		return m, nil

	case AnalysisCompleteMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
		} else {
			// Set flag to show detail view after refresh
			m.showDetailAfterRefresh = true
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

	case ViewAnalysisViewer:
		var model tea.Model
		model, cmd = m.analysisViewer.Update(msg)
		m.analysisViewer = model.(AnalysisViewerModel)

	case ViewLogViewer:
		var model tea.Model
		model, cmd = m.logViewer.Update(msg)
		m.logViewer = model.(LogViewerModel)
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
	case ViewAnalysisViewer:
		return m.analysisViewer.View()
	case ViewLogViewer:
		return m.logViewer.View()
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

		// Estimate token count for the session
		tokenCount, err := m.analysisService.EstimateTokenCount(m.ctx, sessionID)
		if err != nil {
			tokenCount = 0 // Default to 0 if estimation fails
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
			TokenCount:    tokenCount,
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
