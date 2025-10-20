package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kgatilin/darwinflow-pub/internal/app"
)

// LogViewerModel displays the session log in markdown format
type LogViewerModel struct {
	sessionID   string
	logMarkdown string
	viewport    viewport.Model
	width       int
	height      int
	ready       bool
}

// NewLogViewerModel creates a new log viewer
func NewLogViewerModel(sessionID string, logs []*app.LogRecord) LogViewerModel {
	// Format logs as markdown
	var buf bytes.Buffer
	err := app.FormatLogsAsMarkdown(&buf, logs)

	markdown := buf.String()
	if err != nil {
		markdown = fmt.Sprintf("Error formatting logs: %v", err)
	}

	return LogViewerModel{
		sessionID:   sessionID,
		logMarkdown: markdown,
	}
}

// Init initializes the model
func (m LogViewerModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m LogViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
			m.viewport.SetContent(m.renderContent())
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Return to detail view
			return m, func() tea.Msg {
				return BackToDetailMsg{}
			}
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the view
func (m LogViewerModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m LogViewerModel) headerView() string {
	title := fmt.Sprintf("Session Log: %s", m.sessionID[:8])
	return viewerTitleStyle.Render(title)
}

func (m LogViewerModel) footerView() string {
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100),
	)

	actions := actionStyle.Render("[↑/↓] Scroll • [Esc] Back")

	line := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", max(0, m.width-lipgloss.Width(info))))

	return fmt.Sprintf("\n%s\n%s %s", line, actions, info)
}

func (m LogViewerModel) renderContent() string {
	// Use glamour to render the markdown with dark style for better visibility
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(m.width-4), // Account for padding
	)

	if err != nil {
		// Fallback to raw text if renderer creation fails
		return fmt.Sprintf("[Renderer creation error: %v]\n\n%s", err, m.logMarkdown)
	}

	renderedMarkdown, err := renderer.Render(m.logMarkdown)
	if err != nil {
		// Fallback to raw text if rendering fails
		return fmt.Sprintf("[Render error: %v]\n\n%s", err, m.logMarkdown)
	}

	return renderedMarkdown
}
