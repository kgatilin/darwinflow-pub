package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

var (
	viewerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170")).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(lipgloss.Color("240")).
				PaddingBottom(1).
				MarginBottom(1).
				Align(lipgloss.Left)
)

// AnalysisViewerModel displays the full analysis in a scrollable view
type AnalysisViewerModel struct {
	analysis *domain.SessionAnalysis
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewAnalysisViewerModel creates a new analysis viewer
func NewAnalysisViewerModel(analysis *domain.SessionAnalysis) AnalysisViewerModel {
	return AnalysisViewerModel{
		analysis: analysis,
	}
}

// Init initializes the model
func (m AnalysisViewerModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m AnalysisViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m AnalysisViewerModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m AnalysisViewerModel) headerView() string {
	title := fmt.Sprintf("Analysis: %s (%s)", m.analysis.SessionID[:8], m.analysis.PromptName)
	return viewerTitleStyle.Render(title)
}

func (m AnalysisViewerModel) footerView() string {
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100),
	)

	actions := actionStyle.Render("[↑/↓] Scroll • [Esc] Back")

	line := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", max(0, m.width-lipgloss.Width(info))))

	return fmt.Sprintf("\n%s\n%s %s", line, actions, info)
}

func (m AnalysisViewerModel) renderContent() string {
	var b strings.Builder

	// Metadata header
	b.WriteString(detailHeaderStyle.Render("Metadata") + "\n\n")
	b.WriteString(fmt.Sprintf("Session ID:  %s\n", m.analysis.SessionID))
	b.WriteString(fmt.Sprintf("Type:        %s\n", m.analysis.AnalysisType))
	b.WriteString(fmt.Sprintf("Prompt:      %s\n", m.analysis.PromptName))
	b.WriteString(fmt.Sprintf("Model:       %s\n", m.analysis.ModelUsed))
	b.WriteString(fmt.Sprintf("Analyzed At: %s\n\n", m.analysis.AnalyzedAt.Format("2006-01-02 15:04:05")))

	// Render analysis content as markdown
	b.WriteString(detailHeaderStyle.Render("Analysis Result") + "\n\n")

	// Use glamour to render the markdown with dark style for better visibility
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(m.width-4), // Account for padding
	)

	if err == nil {
		renderedMarkdown, err := renderer.Render(m.analysis.AnalysisResult)
		if err == nil {
			b.WriteString(renderedMarkdown)
		} else {
			// Fallback to raw text if rendering fails
			b.WriteString(fmt.Sprintf("[Render error: %v]\n%s", err, m.analysis.AnalysisResult))
		}
	} else {
		// Fallback to raw text if renderer creation fails
		b.WriteString(fmt.Sprintf("[Renderer creation error: %v]\n%s", err, m.analysis.AnalysisResult))
	}

	b.WriteString("\n")

	return b.String()
}

// Message types
type BackToDetailMsg struct{}
