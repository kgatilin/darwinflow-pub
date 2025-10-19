package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170")).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(lipgloss.Color("240")).
				PaddingBottom(1).
				MarginBottom(1)

	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("105"))

	detailContentStyle = lipgloss.NewStyle().
				MarginLeft(2)

	actionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	previewStyle = lipgloss.NewStyle().
			MarginTop(1).
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
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
		case "q", "esc":
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
	title := fmt.Sprintf("Session Details: %s", m.session.ShortID)
	return detailTitleStyle.Render(title)
}

func (m SessionDetailModel) footerView() string {
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100),
	)

	actions := []string{
		"[a] Analyze",
		"[r] Re-analyze",
		"[s] Save",
		"[v] View",
		"[q] Back",
	}

	if !m.session.HasAnalysis {
		actions = []string{
			"[a] Analyze",
			"[q] Back",
		}
	}

	actionsStr := actionStyle.Render(strings.Join(actions, " • "))

	line := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", max(0, m.width-lipgloss.Width(info))))

	return fmt.Sprintf("\n%s\n%s %s", line, actionsStr, info)
}

func (m SessionDetailModel) renderContent() string {
	var b strings.Builder

	// Session metadata
	b.WriteString(detailHeaderStyle.Render("Session Information") + "\n")
	b.WriteString(detailContentStyle.Render(fmt.Sprintf("ID: %s\n", m.session.SessionID)))
	b.WriteString(detailContentStyle.Render(fmt.Sprintf("Time Range: %s - %s\n",
		m.session.FirstEvent.Format("2006-01-02 15:04:05"),
		m.session.LastEvent.Format("15:04:05"))))
	b.WriteString(detailContentStyle.Render(fmt.Sprintf("Event Count: %d\n", m.session.EventCount)))
	b.WriteString("\n")

	// Analysis information
	b.WriteString(detailHeaderStyle.Render("Analysis Status") + "\n")
	if m.session.HasAnalysis {
		b.WriteString(detailContentStyle.Render(
			analyzedStyle.Render(fmt.Sprintf("✓ %d analysis/analyses found\n", m.session.AnalysisCount))))

		for i, analysis := range m.session.Analyses {
			b.WriteString(detailContentStyle.Render(
				fmt.Sprintf("\n%d. Type: %s\n", i+1, analysis.AnalysisType)))
			b.WriteString(detailContentStyle.Render(
				fmt.Sprintf("   Prompt: %s\n", analysis.PromptName)))
			b.WriteString(detailContentStyle.Render(
				fmt.Sprintf("   Model: %s\n", analysis.ModelUsed)))
			b.WriteString(detailContentStyle.Render(
				fmt.Sprintf("   Analyzed: %s\n", analysis.AnalyzedAt.Format("2006-01-02 15:04:05"))))

			// Show preview of analysis
			preview := analysis.AnalysisResult
			if len(preview) > 300 {
				preview = preview[:300] + "..."
			}
			b.WriteString(detailContentStyle.Render("\n   Preview:\n"))
			b.WriteString(previewStyle.Render(preview) + "\n")
		}
	} else {
		b.WriteString(detailContentStyle.Render(
			unanalyzedStyle.Render("✗ Not analyzed\n")))
	}

	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Message types
type BackToListMsg struct{}
type AnalyzeSessionMsg struct{ SessionID string }
type ReanalyzeSessionMsg struct{ SessionID string }
type SaveToMarkdownMsg struct{ SessionID string }
type ViewAnalysisMsg struct{ SessionID string }
