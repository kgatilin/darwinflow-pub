package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// SessionItem implements list.Item for the Bubble Tea list component
type SessionItem struct {
	session *SessionInfo
}

func (i SessionItem) FilterValue() string { return i.session.SessionID }

func (i SessionItem) Title() string {
	statusIcon := IconUnanalyzed
	statusStyle := WarningStyle

	if i.session.HasAnalysis {
		if i.session.AnalysisCount > 1 {
			statusIcon = fmt.Sprintf("%s%d", IconMultiAnalysis, i.session.AnalysisCount)
			statusStyle = InfoStyle
		} else {
			statusIcon = IconAnalyzed
			statusStyle = SuccessStyle
		}
	}

	return fmt.Sprintf("%s %s | %s",
		statusStyle.Render(statusIcon),
		i.session.ShortID,
		i.session.FirstEvent.Format("2006-01-02 15:04"),
	)
}

func (i SessionItem) Description() string {
	desc := fmt.Sprintf("%d events", i.session.EventCount)
	if i.session.HasAnalysis {
		analysisTypes := strings.Join(i.session.AnalysisTypes, ", ")
		desc += fmt.Sprintf(" | Analyzed: %s", analysisTypes)
	}
	return desc
}

// SessionListModel is the Bubble Tea model for the session list view
type SessionListModel struct {
	list          list.Model
	sessions      []*SessionInfo
	width         int
	height        int
	newEventCount int // Number of unread events from dispatcher
}

// NewSessionListModel creates a new session list model
func NewSessionListModel(sessions []*SessionInfo) SessionListModel {
	// Convert sessions to list items
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		items[i] = SessionItem{session: s}
	}

	// Create list with custom delegate
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "DarwinFlow Sessions"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = BaseTitleStyle.MarginLeft(2)
	l.Styles.PaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	l.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)

	return SessionListModel{
		list:     l,
		sessions: sessions,
	}
}

// Init initializes the model
func (m SessionListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SessionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 6) // Account for breadcrumb and footer
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, tea.Quit

		case "enter":
			// Return selected session to parent model
			if item, ok := m.list.SelectedItem().(SessionItem); ok {
				return m, func() tea.Msg {
					return SelectedSessionMsg{Session: item.session}
				}
			}

		case "r":
			// Refresh session list
			return m, func() tea.Msg {
				return RefreshRequestMsg{}
			}

		case "a":
			// Filter to analyzed only
			m.list.SetFilteringEnabled(true)
			return m, nil

		case "u":
			// Filter to unanalyzed only
			m.list.SetFilteringEnabled(true)
			return m, nil

		// Vim-style navigation
		case "j":
			// Move down (let list handle it)
		case "k":
			// Move up (let list handle it)
		case "g":
			// Go to top
			m.list.Select(0)
			return m, nil
		case "G":
			// Go to bottom
			if len(m.sessions) > 0 {
				m.list.Select(len(m.sessions) - 1)
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the view
func (m SessionListModel) View() string {
	// Build breadcrumb
	breadcrumb := RenderBreadcrumb([]string{"Sessions"})

	// Build the title with event counter if there are new events
	title := "DarwinFlow Sessions"
	if m.newEventCount > 0 {
		title = fmt.Sprintf("%s %s",
			title,
			InfoStyle.Render(fmt.Sprintf("(+%d new)", m.newEventCount)))
	}
	m.list.Title = title

	// Footer with help hints
	helpHints := RenderHelpLine(
		RenderKeyHelp("?", "help"),
		RenderKeyHelp("Enter", "view"),
		RenderKeyHelp("r", "refresh"),
		RenderKeyHelp("j/k", "navigate"),
		RenderKeyHelp("Ctrl+C", "quit"),
	)

	return "\n" + breadcrumb + "\n\n" + m.list.View() + "\n\n" + helpHints
}

// SetNewEventCount updates the counter of unread events
func (m *SessionListModel) SetNewEventCount(count int) {
	m.newEventCount = count
}

// GetSelectedSession returns the currently selected session
func (m SessionListModel) GetSelectedSession() *SessionInfo {
	if item, ok := m.list.SelectedItem().(SessionItem); ok {
		return item.session
	}
	return nil
}

// UpdateSessions updates the session list
func (m *SessionListModel) UpdateSessions(sessions []*SessionInfo) {
	m.sessions = sessions
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		items[i] = SessionItem{session: s}
	}
	m.list.SetItems(items)
}

// Message types
type SelectedSessionMsg struct {
	Session *SessionInfo
}

type RefreshRequestMsg struct{}
