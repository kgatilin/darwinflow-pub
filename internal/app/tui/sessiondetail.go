package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SessionDetailModel shows detailed information about a session
type SessionDetailModel struct {
	session  *SessionInfo
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewSessionDetailModel creates a new session detail model
func NewSessionDetailModel(session *SessionInfo) SessionDetailModel {
	return SessionDetailModel{
		session: session,
	}
}

// Init initializes the model
func (m SessionDetailModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SessionDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// Return to list
			return m, func() tea.Msg {
				return BackToListMsg{}
			}

		case "a":
			// Analyze session
			return m, func() tea.Msg {
				return AnalyzeSessionMsg{SessionID: m.session.SessionID}
			}

		case "r":
			// Re-analyze session
			return m, func() tea.Msg {
				return ReanalyzeSessionMsg{SessionID: m.session.SessionID}
			}

		case "s":
			// Save to markdown
			return m, func() tea.Msg {
				return SaveToMarkdownMsg{SessionID: m.session.SessionID}
			}

		case "v":
			// View full analysis
			return m, func() tea.Msg {
				return ViewAnalysisMsg{SessionID: m.session.SessionID}
			}

		case "l":
			// View session log
			return m, func() tea.Msg {
				return ViewLogMsg{SessionID: m.session.SessionID}
			}

		// Vim-style navigation
		case "j", "down":
			// Let viewport handle scrolling
		case "k", "up":
			// Let viewport handle scrolling
		case "g":
			// Go to top
			m.viewport.GotoTop()
			return m, nil
		case "G":
			// Go to bottom
			m.viewport.GotoBottom()
			return m, nil
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the view
func (m SessionDetailModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m SessionDetailModel) headerView() string {
	// Breadcrumb
	breadcrumb := RenderBreadcrumb([]string{"Sessions", m.session.ShortID})

	// Title
	title := PageTitleStyle.Render(fmt.Sprintf("Session Details: %s", m.session.ShortID))

	return breadcrumb + "\n\n" + title
}

func (m SessionDetailModel) footerView() string {
	// Scroll percentage
	scrollInfo := SubtleTextStyle.Render(
		fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100),
	)

	// Build action list based on session state
	var actions []string
	if m.session.HasAnalysis {
		actions = []string{
			RenderKeyHelp("v", "view analysis"),
			RenderKeyHelp("r", "re-analyze"),
			RenderKeyHelp("l", "logs"),
			RenderKeyHelp("s", "save"),
			RenderKeyHelp("Esc", "back"),
		}
	} else {
		actions = []string{
			RenderKeyHelp("a", "analyze"),
			RenderKeyHelp("l", "logs"),
			RenderKeyHelp("Esc", "back"),
		}
	}

	helpLine := RenderHelpLine(actions...)

	// Divider
	dividerWidth := max(0, m.width-lipgloss.Width(scrollInfo)-2)
	divider := RenderDivider(dividerWidth)

	return fmt.Sprintf("\n%s %s\n%s", divider, scrollInfo, helpLine)
}

func (m SessionDetailModel) renderContent() string {
	var b strings.Builder

	// Session metadata
	b.WriteString(SectionTitleStyle.Render("Session Information") + "\n")
	b.WriteString(fmt.Sprintf("  ID: %s\n", m.session.SessionID))
	b.WriteString(fmt.Sprintf("  Time Range: %s - %s\n",
		m.session.FirstEvent.Format("2006-01-02 15:04:05"),
		m.session.LastEvent.Format("15:04:05")))
	b.WriteString(fmt.Sprintf("  Event Count: %s\n",
		InfoStyle.Render(fmt.Sprintf("%d", m.session.EventCount))))

	// Display token count with formatting
	if m.session.TokenCount > 0 {
		tokenCountStr := FormatTokenCount(m.session.TokenCount)
		b.WriteString(fmt.Sprintf("  Log Size: %s\n", tokenCountStr))
	}
	b.WriteString("\n")

	// Analysis information
	b.WriteString(SectionTitleStyle.Render("Analysis Status") + "\n")
	if m.session.HasAnalysis {
		statusLine := SuccessStyle.Render(fmt.Sprintf("%s %d analysis/analyses found",
			IconAnalyzed, m.session.AnalysisCount))
		b.WriteString("  " + statusLine + "\n")

		for i, analysis := range m.session.Analyses {
			b.WriteString(fmt.Sprintf("\n  %s Analysis %d\n", IconInfo, i+1))
			b.WriteString(fmt.Sprintf("     View Type: %s\n", analysis.ViewType))
			b.WriteString(fmt.Sprintf("     Prompt: %s\n", analysis.PromptUsed))
			b.WriteString(fmt.Sprintf("     Model: %s\n", analysis.ModelUsed))
			b.WriteString(fmt.Sprintf("     Analyzed: %s\n", analysis.Timestamp.Format("2006-01-02 15:04:05")))

			// Show preview of analysis
			preview := analysis.Result
			if len(preview) > 300 {
				preview = preview[:300] + "..."
			}
			b.WriteString("\n     Preview:\n")
			previewBox := BoxStyle.
				MarginLeft(4).
				Width(min(m.width-8, 80)).
				Render(preview)
			b.WriteString(previewBox + "\n")
		}
	} else {
		statusLine := WarningStyle.Render(fmt.Sprintf("%s Not analyzed", IconUnanalyzed))
		b.WriteString("  " + statusLine + "\n")
		b.WriteString("\n  " + HelpTextStyle.Render("Press 'a' to analyze this session") + "\n")
	}

	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FormatTokenCount formats a token count with thousands separator and unit
func FormatTokenCount(count int) string {
	// Format with thousands separator
	countStr := fmt.Sprintf("%d", count)
	if count >= 1000 {
		// Add comma separator for thousands
		var result []rune
		for i, c := range countStr {
			if i > 0 && (len(countStr)-i)%3 == 0 {
				result = append(result, ',')
			}
			result = append(result, c)
		}
		countStr = string(result)
	}

	return fmt.Sprintf("~%s tokens", countStr)
}

// Message types
type BackToListMsg struct{}
type AnalyzeSessionMsg struct{ SessionID string }
type ReanalyzeSessionMsg struct{ SessionID string }
type SaveToMarkdownMsg struct{ SessionID string }
type ViewAnalysisMsg struct{ SessionID string }
type ViewLogMsg struct{ SessionID string }
