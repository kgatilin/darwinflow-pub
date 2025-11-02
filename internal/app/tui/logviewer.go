package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kgatilin/darwinflow-pub/internal/app"
)

// LogViewerModel displays the session log in markdown format
type LogViewerModel struct {
	sessionID        string
	logMarkdown      string
	logRecords       []*app.LogRecord
	viewport         viewport.Model
	width            int
	height           int
	ready            bool
	searchMode       bool
	searchInput      textinput.Model
	searchQuery      string
	matchLines       []int // Line numbers where matches are found
	currentMatch     int   // Current match index
	renderedContent  string
	highlightedContent string
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

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100

	return LogViewerModel{
		sessionID:   sessionID,
		logMarkdown: markdown,
		logRecords:  logs,
		searchInput: ti,
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
		searchPanelHeight := 0
		if m.searchMode {
			searchPanelHeight = 3 // Search panel takes 3 lines
		}
		verticalMarginHeight := headerHeight + footerHeight + searchPanelHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
			m.renderAndSetContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		return m, nil

	case tea.KeyMsg:
		// Handle search mode
		if m.searchMode {
			switch msg.String() {
			case "esc":
				// Exit search mode and resize viewport
				m.searchMode = false
				m.resizeViewport()
				return m, nil
			case "enter":
				// Exit search mode and keep current search active
				m.searchMode = false
				m.resizeViewport()
				return m, nil
			default:
				// Update search input and search incrementally
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)

				// Update search as they type
				m.searchQuery = m.searchInput.Value()
				m.findMatches()
				m.updateHighlighting()
				if len(m.matchLines) > 0 {
					m.scrollToCurrentMatch()
				}

				return m, tea.Batch(cmds...)
			}
		}

		// Normal mode key handling
		switch msg.String() {
		case "esc":
			if m.searchQuery != "" {
				// Clear search
				m.searchQuery = ""
				m.matchLines = nil
				m.currentMatch = 0
				m.highlightedContent = ""
				m.viewport.SetContent(m.renderedContent)
				return m, nil
			}
			// Return to detail view
			return m, func() tea.Msg {
				return BackToDetailMsg{}
			}
		case "/":
			// Enter search mode and resize viewport
			m.searchMode = true
			m.searchInput.Focus()
			m.searchInput.SetValue(m.searchQuery) // Keep previous search
			m.resizeViewport()
			return m, textinput.Blink
		case "n":
			// Next match
			if len(m.matchLines) > 0 {
				m.currentMatch = (m.currentMatch + 1) % len(m.matchLines)
				m.scrollToCurrentMatch()
			}
			return m, nil
		case "N":
			// Previous match
			if len(m.matchLines) > 0 {
				m.currentMatch--
				if m.currentMatch < 0 {
					m.currentMatch = len(m.matchLines) - 1
				}
				m.scrollToCurrentMatch()
			}
			return m, nil

		// Vim-style navigation
		case "j":
			// Move down (let viewport handle it)
		case "k":
			// Move up (let viewport handle it)
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

	if !m.searchMode {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// resizeViewport adjusts viewport height based on search mode
func (m *LogViewerModel) resizeViewport() {
	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	searchPanelHeight := 0
	if m.searchMode {
		searchPanelHeight = 3
	}
	verticalMarginHeight := headerHeight + footerHeight + searchPanelHeight
	m.viewport.Height = m.height - verticalMarginHeight
}

// findMatches finds all line numbers containing the search query
func (m *LogViewerModel) findMatches() {
	m.matchLines = nil
	m.currentMatch = 0

	if m.searchQuery == "" {
		return
	}

	// Search in the rendered content (case-insensitive)
	query := strings.ToLower(m.searchQuery)
	lines := strings.Split(m.renderedContent, "\n")

	for lineNum, line := range lines {
		if strings.Contains(strings.ToLower(line), query) {
			m.matchLines = append(m.matchLines, lineNum)
		}
	}
}

// updateHighlighting highlights search matches in the rendered content
func (m *LogViewerModel) updateHighlighting() {
	// Trim whitespace and check for empty query
	query := strings.TrimSpace(m.searchQuery)

	if query == "" || m.renderedContent == "" || len(query) == 0 {
		m.highlightedContent = ""
		m.viewport.SetContent(m.renderedContent)
		return
	}

	// Highlight style: yellow background with black text
	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("220")).
		Foreground(lipgloss.Color("0")).
		Bold(true)

	// Case-insensitive search and replace
	content := m.renderedContent
	result := strings.Builder{}
	result.Grow(len(content) + 1000) // Pre-allocate for performance

	// Convert to lowercase for case-insensitive search
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)

	lastPos := 0
	iterations := 0
	maxIterations := 10000 // Safety limit to prevent infinite loops

	for iterations < maxIterations {
		pos := strings.Index(lowerContent[lastPos:], lowerQuery)
		if pos == -1 {
			// No more matches, add the rest
			result.WriteString(content[lastPos:])
			break
		}

		actualPos := lastPos + pos
		// Add text before match
		result.WriteString(content[lastPos:actualPos])
		// Add highlighted match
		matchText := content[actualPos : actualPos+len(query)]
		result.WriteString(highlightStyle.Render(matchText))

		lastPos = actualPos + len(query)

		// Safety check: if lastPos isn't advancing, break to prevent infinite loop
		if lastPos >= len(content) || len(query) == 0 {
			break
		}

		iterations++
	}

	m.highlightedContent = result.String()
	m.viewport.SetContent(m.highlightedContent)
}

// scrollToCurrentMatch scrolls the viewport to show the current match
func (m *LogViewerModel) scrollToCurrentMatch() {
	if len(m.matchLines) == 0 || m.currentMatch >= len(m.matchLines) {
		return
	}

	targetLine := m.matchLines[m.currentMatch]

	// Calculate the total number of lines in the content
	totalLines := strings.Count(m.logMarkdown, "\n") + 1

	// Get viewport height
	viewportHeight := m.viewport.Height

	// Calculate offset to center the match in the viewport
	offset := targetLine - viewportHeight/2
	if offset < 0 {
		offset = 0
	}
	if offset > totalLines-viewportHeight {
		offset = totalLines - viewportHeight
	}
	if offset < 0 {
		offset = 0
	}

	// Set the viewport position
	m.viewport.SetYOffset(offset)
}

// View renders the view
func (m LogViewerModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	if m.searchMode {
		// Show viewport with search panel at the bottom
		return fmt.Sprintf("%s\n%s\n%s\n%s",
			m.headerView(),
			m.viewport.View(),
			m.searchPanelView(),
			m.footerView())
	}

	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m LogViewerModel) searchPanelView() string {
	// Search panel with border
	searchStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	matchInfo := ""
	if len(m.matchLines) > 0 {
		matchInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render(fmt.Sprintf(" (%d matches)", len(m.matchLines)))
	}

	content := fmt.Sprintf("Search: %s%s", m.searchInput.View(), matchInfo)
	return searchStyle.Render(content)
}

func (m LogViewerModel) headerView() string {
	// Breadcrumb
	breadcrumb := RenderBreadcrumb([]string{"Sessions", m.sessionID[:8], "Log"})

	// Title with search info
	title := fmt.Sprintf("Session Log: %s", m.sessionID[:8])
	if m.searchQuery != "" && len(m.matchLines) > 0 {
		title += fmt.Sprintf(" (match %d/%d)", m.currentMatch+1, len(m.matchLines))
	} else if m.searchQuery != "" {
		title += " (no matches)"
	}
	titleRendered := PageTitleStyle.Render(title)

	return breadcrumb + "\n\n" + titleRendered
}

func (m LogViewerModel) footerView() string {
	// Scroll percentage
	scrollInfo := SubtleTextStyle.Render(
		fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100),
	)

	// Build help line based on current mode
	var helpLine string
	if m.searchMode {
		matchInfo := ""
		if len(m.matchLines) > 0 {
			matchInfo = fmt.Sprintf(" (%d matches)", len(m.matchLines))
		}
		helpLine = RenderHelpLine(
			RenderKeyHelp("Enter", "done"),
			RenderKeyHelp("Esc", "cancel"),
		) + HelpTextStyle.Render(matchInfo)
	} else if m.searchQuery != "" {
		helpLine = RenderHelpLine(
			RenderKeyHelp("j/k", "scroll"),
			RenderKeyHelp("n/N", "next/prev match"),
			RenderKeyHelp("/", "edit search"),
			RenderKeyHelp("Esc", "clear"),
		)
	} else {
		helpLine = RenderHelpLine(
			RenderKeyHelp("j/k or ↑/↓", "scroll"),
			RenderKeyHelp("/", "search"),
			RenderKeyHelp("?", "help"),
			RenderKeyHelp("Esc", "back"),
		)
	}

	// Divider
	dividerWidth := max(0, m.width-lipgloss.Width(scrollInfo)-2)
	divider := RenderDivider(dividerWidth)

	return fmt.Sprintf("\n%s %s\n%s", divider, scrollInfo, helpLine)
}

// renderAndSetContent renders the markdown and sets it in the viewport
func (m *LogViewerModel) renderAndSetContent() {
	// Use glamour to render the markdown with dark style for better visibility
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(m.width-4), // Account for padding
	)

	if err != nil {
		// Fallback to raw text if renderer creation fails
		m.renderedContent = fmt.Sprintf("[Renderer creation error: %v]\n\n%s", err, m.logMarkdown)
	} else {
		renderedMarkdown, err := renderer.Render(m.logMarkdown)
		if err != nil {
			// Fallback to raw text if rendering fails
			m.renderedContent = fmt.Sprintf("[Render error: %v]\n\n%s", err, m.logMarkdown)
		} else {
			m.renderedContent = renderedMarkdown
		}
	}

	m.viewport.SetContent(m.renderedContent)
}
