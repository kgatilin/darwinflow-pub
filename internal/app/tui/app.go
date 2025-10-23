package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	errorStyle   = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196"))
)

// AppModel is the main orchestrator for the TUI
type AppModel struct {
	ctx             context.Context
	pluginRegistry  *app.PluginRegistry
	analysisService *app.AnalysisService
	logsService     *app.LogsService
	config          *domain.Config

	// State
	currentView  ViewState
	previousView ViewState // Track previous view for error dismissal
	sessions     []*SessionInfo
	loading      bool
	err          error

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

	// Real-time event streaming
	eventDispatcher *app.EventDispatcher
	eventChan       <-chan pluginsdk.Event
	newEventCount   int
	lastEventTime   time.Time

	width  int
	height int
}

// NewAppModel creates a new TUI application model
func NewAppModel(
	ctx context.Context,
	pluginRegistry *app.PluginRegistry,
	analysisService *app.AnalysisService,
	logsService *app.LogsService,
	config *domain.Config,
	eventDispatcher *app.EventDispatcher,
) *AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return &AppModel{
		ctx:             ctx,
		pluginRegistry:  pluginRegistry,
		analysisService: analysisService,
		logsService:     logsService,
		config:          config,
		currentView:     ViewSessionList,
		spinner:         s,
		loading:         true,
		eventDispatcher: eventDispatcher,
	}
}

// Init initializes the application
func (m *AppModel) Init() tea.Cmd {
	subscribeCmd := m.subscribeToEvents()
	return tea.Batch(
		m.spinner.Tick,
		m.loadSessions,
		subscribeCmd,
	)
}

// subscribeToEvents starts listening for real-time events from the dispatcher
func (m *AppModel) subscribeToEvents() tea.Cmd {
	if m.eventDispatcher == nil {
		return nil
	}

	// Get the event channel from the dispatcher
	m.eventChan = m.eventDispatcher.GetEventChannel()

	// Return a command that listens for events in a background goroutine
	return func() tea.Msg {
		return m.listenForEvents()
	}
}

// listenForEvents waits for the next event and returns it as a message
// This is designed to be called repeatedly by the Update loop
func (m *AppModel) listenForEvents() tea.Msg {
	if m.eventChan == nil {
		return nil
	}

	// This blocks until an event arrives
	event := <-m.eventChan

	// Return the event as a message so Update can handle it
	return EventArrivedMsg{
		Event:     event,
		Timestamp: time.Now(),
	}
}

// listenForNextEvent returns a command that listens for the next event
// This is called repeatedly to maintain the event listening loop
func (m *AppModel) listenForNextEvent() tea.Cmd {
	return func() tea.Msg {
		return m.listenForEvents()
	}
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
		// Handle Esc to dismiss error overlay
		if m.err != nil && msg.String() == "esc" {
			m.err = nil
			m.currentView = m.previousView
			return m, nil
		}

	case EventArrivedMsg:
		// Increment unread event counter
		m.newEventCount++
		m.lastEventTime = msg.Timestamp

		// Auto-refresh if user is viewing session list
		if !m.loading && m.currentView == ViewSessionList {
			// Trigger a refresh to show updated event counts
			return m, tea.Batch(
				m.loadSessions,
				m.listenForNextEvent(), // Continue listening for events
			)
		}

		// Continue listening for the next event
		return m, m.listenForNextEvent()

	case SessionsLoadedMsg:
		m.loading = false
		if msg.Error != nil {
			m.previousView = m.currentView
			m.err = msg.Error
			return m, m.listenForNextEvent()
		}
		m.sessions = msg.Sessions
		m.sessionList = NewSessionListModel(msg.Sessions)

		// Sync the new event count to the session list view
		m.sessionList.SetNewEventCount(m.newEventCount)

		// If we're in session list view, clear the new event count when we display it
		if m.currentView == ViewSessionList {
			m.newEventCount = 0
			m.sessionList.SetNewEventCount(0)
		}

		// If we should show detail view after refresh (e.g., after analysis)
		if m.showDetailAfterRefresh && m.selectedSession != nil {
			m.showDetailAfterRefresh = false
			// Find the updated session info
			for _, session := range msg.Sessions {
				if session.SessionID == m.selectedSession.SessionID {
					m.selectedSession = session
					m.sessionDetail = NewSessionDetailModel(session)
					m.currentView = ViewSessionDetail
					// Send initial window size to the detail view and continue listening
					if m.width > 0 && m.height > 0 {
						return m, tea.Batch(
							func() tea.Msg {
								return tea.WindowSizeMsg{Width: m.width, Height: m.height}
							},
							m.listenForNextEvent(),
						)
					}
					return m, m.listenForNextEvent()
				}
			}
		}

		m.currentView = ViewSessionList
		// Send initial window size to the newly created list and continue listening
		if m.width > 0 && m.height > 0 {
			return m, tea.Batch(
				func() tea.Msg {
					return tea.WindowSizeMsg{Width: m.width, Height: m.height}
				},
				m.listenForNextEvent(),
			)
		}
		return m, m.listenForNextEvent()

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
		// Get the analysis for this session (using generic analysis)
		analyses, err := m.analysisService.GetAnalysesByViewID(m.ctx, msg.SessionID)
		if err != nil || len(analyses) == 0 {
			m.previousView = m.currentView
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
			m.previousView = m.currentView
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
			m.previousView = m.currentView
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
			m.previousView = m.currentView
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
	// Show error overlay if error is set
	if m.err != nil {
		return m.renderErrorOverlay()
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

// renderErrorOverlay renders the error message as a centered overlay
func (m *AppModel) renderErrorOverlay() string {
	// Calculate max width for error message (leave padding for border)
	maxWidth := m.width - 12
	if maxWidth < 40 {
		maxWidth = 40
	}
	if maxWidth > 80 {
		maxWidth = 80
	}

	// Build error message with header and instructions
	errorHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Render("Error")

	errorBody := lipgloss.NewStyle().
		Width(maxWidth).
		Render(m.err.Error())

	errorFooter := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		Render("Press Esc to go back")

	// Combine all parts with spacing
	errorContent := lipgloss.JoinVertical(
		lipgloss.Left,
		errorHeader,
		"",
		errorBody,
		"",
		errorFooter,
	)

	// Apply border and padding
	errorBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Render(errorContent)

	// Center the error box on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		errorBox,
	)
}

// Command functions

func (m *AppModel) loadSessions() tea.Msg {
	// Query all sessions from the plugin registry
	entities, err := m.pluginRegistry.Query(m.ctx, pluginsdk.EntityQuery{
		EntityType: "session",
	})
	if err != nil {
		return SessionsLoadedMsg{Error: err}
	}

	sessions := make([]*SessionInfo, 0, len(entities))

	for _, entity := range entities {
		// Extract session info from entity fields
		fields := entity.GetAllFields()

		sessionInfo := &SessionInfo{
			SessionID:     entity.GetID(),
			ShortID:       getStringField(fields, "short_id", ""),
			FirstEvent:    getTimeField(fields, "first_event"),
			LastEvent:     getTimeField(fields, "last_event"),
			EventCount:    getIntField(fields, "event_count", 0),
			AnalysisCount: getIntField(fields, "analysis_count", 0),
			AnalysisTypes: getStringSliceField(fields, "analysis_types"),
			TokenCount:    getIntField(fields, "token_count", 0),
			HasAnalysis:   getBoolField(fields, "has_analysis", false),
		}

		// Get analyses from analysis service (using generic analysis)
		analyses, err := m.analysisService.GetAnalysesByViewID(m.ctx, entity.GetID())
		if err == nil {
			sessionInfo.Analyses = analyses
			if len(analyses) > 0 {
				sessionInfo.LatestAnalysis = analyses[0]
			}
		}

		sessions = append(sessions, sessionInfo)
	}

	return SessionsLoadedMsg{Sessions: sessions}
}

// Helper functions to safely extract typed values from field map

func getStringField(fields map[string]interface{}, key, defaultValue string) string {
	if val, ok := fields[key].(string); ok {
		return val
	}
	return defaultValue
}

func getIntField(fields map[string]interface{}, key string, defaultValue int) int {
	if val, ok := fields[key].(int); ok {
		return val
	}
	return defaultValue
}

func getBoolField(fields map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := fields[key].(bool); ok {
		return val
	}
	return defaultValue
}

func getTimeField(fields map[string]interface{}, key string) time.Time {
	if val, ok := fields[key].(time.Time); ok {
		return val
	}
	return time.Time{}
}

func getStringSliceField(fields map[string]interface{}, key string) []string {
	if val, ok := fields[key].([]string); ok {
		return val
	}
	return []string{}
}

func (m *AppModel) analyzeSession(sessionID, promptName string) tea.Cmd {
	return func() tea.Msg {
		sessionAnalysis, err := m.analysisService.AnalyzeSessionWithPrompt(m.ctx, sessionID, promptName)
		var genericAnalysis *domain.Analysis
		if sessionAnalysis != nil {
			genericAnalysis = sessionAnalysis.ToGenericAnalysis()
		}
		return AnalysisCompleteMsg{
			SessionID: sessionID,
			Analysis:  genericAnalysis,
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
	pluginRegistry *app.PluginRegistry,
	analysisService *app.AnalysisService,
	logsService *app.LogsService,
	config *domain.Config,
	eventDispatcher *app.EventDispatcher,
) error {
	m := NewAppModel(ctx, pluginRegistry, analysisService, logsService, config, eventDispatcher)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
