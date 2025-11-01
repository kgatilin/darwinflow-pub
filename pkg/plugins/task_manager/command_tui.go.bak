package task_manager

import (
	"context"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// TUICommand launches the interactive TUI for browsing roadmaps and tracks
type TUICommand struct {
	Plugin *TaskManagerPlugin
}

// GetName returns the command name
func (c *TUICommand) GetName() string {
	return "tui"
}

// GetDescription returns the command description
func (c *TUICommand) GetDescription() string {
	return "Launch interactive TUI to browse roadmaps and tracks"
}

// GetHelp returns the command help text
func (c *TUICommand) GetHelp() string {
	return `Usage: dw task-manager tui

Launch an interactive terminal user interface to browse and manage roadmaps, tracks, and tasks.

Navigation:
  j/k or ↑/↓     Navigate between items
  Enter          View details / drill down
  esc            Go back to previous view
  r              Refresh data
  q              Quit

Views:
  Roadmap List   Browse all tracks in the active roadmap
  Track Detail   View tasks within a selected track
`
}

// GetUsage returns the command usage
func (c *TUICommand) GetUsage() string {
	return "tui"
}

// Execute runs the TUI command
func (c *TUICommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Repository must be initialized for TUI
	if c.Plugin.repository == nil {
		return fmt.Errorf("database not initialized; run 'dw task-manager roadmap init' first")
	}

	// Create the TUI model
	appModel := NewAppModel(ctx, c.Plugin.repository, c.Plugin.logger)

	// Start the Bubble Tea program
	p := tea.NewProgram(appModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

// TUIScaffold provides utility functions for TUI-related operations
type TUIScaffold struct{}

// RunProgram is a testable wrapper around tea.NewProgram
// This allows us to mock the Bubble Tea program in tests
var RunProgram = func(model tea.Model) error {
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// StartTUI starts the task manager TUI with the given context and output writer
// This function is exposed for external callers (e.g., other plugins, CLI wrappers)
func StartTUI(ctx context.Context, repository RoadmapRepository, logger pluginsdk.Logger, output io.Writer) error {
	if repository == nil {
		return fmt.Errorf("repository is required")
	}

	appModel := NewAppModel(ctx, repository, logger)

	p := tea.NewProgram(appModel, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Fprintf(output, "TUI error: %v\n", err)
		return err
	}

	return nil
}
