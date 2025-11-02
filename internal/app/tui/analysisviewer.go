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

// AnalysisViewerModel displays the full analysis in a scrollable view
type AnalysisViewerModel struct {
	analysis *domain.Analysis
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewAnalysisViewerModel creates a new analysis viewer
func NewAnalysisViewerModel(analysis *domain.Analysis) AnalysisViewerModel {
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
func (m AnalysisViewerModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m AnalysisViewerModel) headerView() string {
	// Breadcrumb
	var viewInfo string
	if m.analysis.ViewType == "session" {
		viewInfo = m.analysis.ViewID[:8]
	} else {
		viewInfo = fmt.Sprintf("%s: %s", m.analysis.ViewType, m.analysis.ViewID)
	}
	breadcrumb := RenderBreadcrumb([]string{"Sessions", viewInfo, "Analysis"})

	// Title
	title := PageTitleStyle.Render(fmt.Sprintf("Analysis: %s (%s)", viewInfo, m.analysis.PromptUsed))

	return breadcrumb + "\n\n" + title
}

func (m AnalysisViewerModel) footerView() string {
	// Scroll percentage
	scrollInfo := SubtleTextStyle.Render(
		fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100),
	)

	// Help hints
	helpLine := RenderHelpLine(
		RenderKeyHelp("j/k or ↑/↓", "scroll"),
		RenderKeyHelp("g/G", "top/bottom"),
		RenderKeyHelp("?", "help"),
		RenderKeyHelp("Esc", "back"),
	)

	// Divider
	dividerWidth := max(0, m.width-lipgloss.Width(scrollInfo)-2)
	divider := RenderDivider(dividerWidth)

	return fmt.Sprintf("\n%s %s\n%s", divider, scrollInfo, helpLine)
}

func (m AnalysisViewerModel) renderContent() string {
	var b strings.Builder

	// Metadata header
	b.WriteString(SectionTitleStyle.Render("Metadata") + "\n\n")
	b.WriteString(fmt.Sprintf("View ID:     %s\n", m.analysis.ViewID))
	b.WriteString(fmt.Sprintf("View Type:   %s\n", m.analysis.ViewType))
	b.WriteString(fmt.Sprintf("Prompt:      %s\n", m.analysis.PromptUsed))
	b.WriteString(fmt.Sprintf("Model:       %s\n", m.analysis.ModelUsed))
	b.WriteString(fmt.Sprintf("Analyzed At: %s\n", m.analysis.Timestamp.Format("2006-01-02 15:04:05")))

	// Display metadata if present
	if len(m.analysis.Metadata) > 0 {
		b.WriteString("\nAdditional Metadata:\n")
		for key, value := range m.analysis.Metadata {
			b.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}
	b.WriteString("\n")

	// Render analysis content as markdown
	b.WriteString(SectionTitleStyle.Render("Analysis Result") + "\n\n")

	// Use glamour to render the markdown with dark style for better visibility
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(m.width-4), // Account for padding
	)

	if err == nil {
		renderedMarkdown, err := renderer.Render(m.analysis.Result)
		if err == nil {
			b.WriteString(renderedMarkdown)
		} else {
			// Fallback to raw text if rendering fails
			errorMsg := ErrorStyle.Render(fmt.Sprintf("Render error: %v", err))
			b.WriteString(errorMsg + "\n\n" + m.analysis.Result)
		}
	} else {
		// Fallback to raw text if renderer creation fails
		errorMsg := ErrorStyle.Render(fmt.Sprintf("Renderer creation error: %v", err))
		b.WriteString(errorMsg + "\n\n" + m.analysis.Result)
	}

	b.WriteString("\n")

	return b.String()
}

// Message types
type BackToDetailMsg struct{}
