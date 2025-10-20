package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the session list
var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	analyzedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // Green for analyzed
	unanalyzedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange for unanalyzed
	multiAnalysisStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("51")) // Cyan for multiple analyses
)

// SessionItem implements list.Item for the Bubble Tea list component
type SessionItem struct {
	session *SessionInfo
}

func (i SessionItem) FilterValue() string { return i.session.SessionID }

func (i SessionItem) Title() string {
	statusIcon := "✗"
	statusStyle := unanalyzedStyle

	if i.session.HasAnalysis {
		if i.session.AnalysisCount > 1 {
			statusIcon = fmt.Sprintf("⟳%d", i.session.AnalysisCount)
			statusStyle = multiAnalysisStyle
		} else {
			statusIcon = "✓"
			statusStyle = analyzedStyle
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
	list     list.Model
	sessions []*SessionInfo
	width    int
	height   int
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
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

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
		m.list.SetHeight(msg.Height - 2)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
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
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the view
func (m SessionListModel) View() string {
	return "\n" + m.list.View()
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
