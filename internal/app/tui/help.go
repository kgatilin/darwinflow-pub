package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpOverlay provides contextual help for each view

// HelpSection represents a section of help content
type HelpSection struct {
	Title string
	Items []HelpItem
}

// HelpItem represents a single help item
type HelpItem struct {
	Key         string
	Description string
}

// GetHelpForView returns help sections for the specified view
func GetHelpForView(view ViewState) []HelpSection {
	switch view {
	case ViewSessionList:
		return getSessionListHelp()
	case ViewSessionDetail:
		return getSessionDetailHelp()
	case ViewAnalysisViewer:
		return getAnalysisViewerHelp()
	case ViewLogViewer:
		return getLogViewerHelp()
	default:
		return getDefaultHelp()
	}
}

func getSessionListHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"↑/↓ or j/k", "Move selection up/down"},
				{"Enter", "View session details"},
				{"/", "Filter sessions"},
				{"Esc", "Clear filter / Quit"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"r", "Refresh session list"},
				{"?", "Toggle this help"},
				{"Ctrl+C or q", "Quit application"},
			},
		},
		{
			Title: "Status Indicators",
			Items: []HelpItem{
				{"✓", "Session has analysis"},
				{"⟳N", "Session has N analyses"},
				{"✗", "Session not analyzed"},
			},
		},
	}
}

func getSessionDetailHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"↑/↓ or j/k", "Scroll content"},
				{"Esc", "Back to session list"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"a", "Analyze session"},
				{"r", "Re-analyze session"},
				{"v", "View analysis"},
				{"l", "View session log"},
				{"s", "Save analysis to markdown"},
				{"?", "Toggle this help"},
			},
		},
		{
			Title: "Notes",
			Items: []HelpItem{
				{"", "Some actions only available if analysis exists"},
			},
		},
	}
}

func getAnalysisViewerHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"↑/↓ or j/k", "Scroll content"},
				{"PgUp/PgDn", "Page up/down"},
				{"Home/End", "Jump to top/bottom"},
				{"Esc", "Back to session detail"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"s", "Save to markdown file"},
				{"?", "Toggle this help"},
			},
		},
	}
}

func getLogViewerHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"↑/↓ or j/k", "Scroll content"},
				{"PgUp/PgDn", "Page up/down"},
				{"Esc", "Back to session detail"},
			},
		},
		{
			Title: "Search",
			Items: []HelpItem{
				{"/", "Start search"},
				{"n", "Next match"},
				{"N", "Previous match"},
				{"Esc", "Clear search / Exit search mode"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"?", "Toggle this help"},
			},
		},
	}
}

func getDefaultHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "General",
			Items: []HelpItem{
				{"?", "Toggle help"},
				{"Ctrl+C", "Quit"},
			},
		},
	}
}

// RenderHelpOverlay renders the help content as a centered overlay
func RenderHelpOverlay(view ViewState, width, height int) string {
	sections := GetHelpForView(view)

	// Calculate max width for help content
	maxWidth := width - 12
	if maxWidth < 40 {
		maxWidth = 40
	}
	if maxWidth > 80 {
		maxWidth = 80
	}

	// Build help content
	var content strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Render("Help")
	content.WriteString(header + "\n\n")

	// Sections
	for i, section := range sections {
		if i > 0 {
			content.WriteString("\n")
		}

		// Section title
		sectionTitle := SectionTitleStyle.Render(section.Title)
		content.WriteString(sectionTitle + "\n")

		// Section items
		for _, item := range section.Items {
			if item.Key == "" {
				// Note without key binding
				content.WriteString(fmt.Sprintf("  %s\n", item.Description))
			} else {
				// Key binding with description
				keyStyle := lipgloss.NewStyle().
					Foreground(ColorHighlight).
					Bold(true).
					Width(15).
					Align(lipgloss.Left)
				key := keyStyle.Render(item.Key)
				content.WriteString(fmt.Sprintf("  %s  %s\n", key, item.Description))
			}
		}
	}

	content.WriteString("\n")

	// Footer
	footer := HelpTextStyle.Render("Press ? or Esc to close help")
	content.WriteString(footer)

	// Apply box style
	helpBox := InfoBoxStyle.
		Width(maxWidth).
		Render(content.String())

	// Center on screen
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		helpBox,
	)
}
