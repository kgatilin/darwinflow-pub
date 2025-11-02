package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Centralized TUI styles for consistency across all views

// Color palette - consistent semantic colors
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("170") // Purple - main accent
	ColorSecondary = lipgloss.Color("105") // Light purple - headers
	ColorTertiary  = lipgloss.Color("205") // Pink - titles

	// Status colors
	ColorSuccess = lipgloss.Color("42")  // Green - success, analyzed, done
	ColorWarning = lipgloss.Color("208") // Orange - warning, unanalyzed
	ColorError   = lipgloss.Color("196") // Red - error, failed
	ColorInfo    = lipgloss.Color("51")  // Cyan - info, multiple analyses
	ColorMuted   = lipgloss.Color("240") // Gray - borders, secondary text
	ColorSubtle  = lipgloss.Color("241") // Light gray - help text, metadata
	ColorSubtler = lipgloss.Color("244") // Lighter gray - less important info

	// Highlight colors
	ColorHighlight   = lipgloss.Color("229") // Yellow - selected items
	ColorBackground  = lipgloss.Color("240") // Dark gray - selected background
	ColorSearchMatch = lipgloss.Color("220") // Yellow - search highlights
	ColorSearchText  = lipgloss.Color("0")   // Black - text on highlighted bg

	// Special colors
	ColorNewEvent = lipgloss.Color("226") // Yellow - new events indicator
)

// Base styles - reusable building blocks
var (
	// Title styles
	BaseTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTertiary)

	PageTitleStyle = BaseTitleStyle.
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorMuted).
			PaddingBottom(1).
			MarginBottom(1)

	SectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary)

	// Text styles
	NormalTextStyle = lipgloss.NewStyle()

	SubtleTextStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	MutedTextStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	HelpTextStyle = lipgloss.NewStyle().
			Foreground(ColorSubtler).
			Italic(true)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Bold(true)

	// Action styles
	ActionStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorHighlight).
			Bold(true)

	// Border and divider styles
	BorderStyle = lipgloss.NewStyle().
			BorderForeground(ColorMuted)

	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(1, 2)

	InfoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorInfo).
			Padding(1, 2)

	// Selection styles
	SelectedStyle = lipgloss.NewStyle().
			Background(ColorBackground).
			Foreground(ColorHighlight)

	UnselectedStyle = lipgloss.NewStyle()
)

// Status icons - consistent visual indicators
const (
	IconSuccess       = "✓"
	IconError         = "✗"
	IconWarning       = "⚠"
	IconInfo          = "ℹ"
	IconAnalyzed      = "✓"
	IconUnanalyzed    = "✗"
	IconMultiAnalysis = "⟳"
	IconLoading       = "⏳"
	IconNew           = "●"
	IconRefresh       = "↻"
)

// Breadcrumb styles
var (
	BreadcrumbStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	BreadcrumbCurrentStyle = lipgloss.NewStyle().
				Foreground(ColorHighlight).
				Bold(true)

	BreadcrumbSeparator = lipgloss.NewStyle().
				Foreground(ColorMuted).
				Render(" › ")
)

// Helper functions

// RenderBreadcrumb renders a breadcrumb navigation trail
func RenderBreadcrumb(items []string) string {
	if len(items) == 0 {
		return ""
	}

	result := ""
	for i, item := range items {
		if i > 0 {
			result += BreadcrumbSeparator
		}
		if i == len(items)-1 {
			// Last item (current) is highlighted
			result += BreadcrumbCurrentStyle.Render(item)
		} else {
			result += BreadcrumbStyle.Render(item)
		}
	}
	return result
}

// RenderDivider renders a horizontal divider line
func RenderDivider(width int) string {
	if width < 0 {
		width = 0
	}
	return DividerStyle.Render(lipgloss.NewStyle().Width(width).Render("─"))
}

// RenderKeyHelp renders a key-action pair for help text
func RenderKeyHelp(key, action string) string {
	return KeyStyle.Render("["+key+"]") + " " + HelpTextStyle.Render(action)
}

// RenderHelpLine renders a line of help text with multiple key-action pairs
func RenderHelpLine(helps ...string) string {
	result := ""
	for i, help := range helps {
		if i > 0 {
			result += HelpTextStyle.Render(" • ")
		}
		result += help
	}
	return result
}

// FormatStatus returns a styled status indicator
func FormatStatus(status string) string {
	switch status {
	case "success", "done", "analyzed", "complete":
		return SuccessStyle.Render(IconSuccess + " " + status)
	case "error", "failed":
		return ErrorStyle.Render(IconError + " " + status)
	case "warning", "unanalyzed":
		return WarningStyle.Render(IconWarning + " " + status)
	case "loading", "in-progress":
		return InfoStyle.Render(IconLoading + " " + status)
	default:
		return NormalTextStyle.Render(status)
	}
}
