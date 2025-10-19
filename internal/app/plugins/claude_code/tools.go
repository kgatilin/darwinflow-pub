package claude_code

import (
	"context"
	"flag"
	"fmt"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// GetTools returns the tools provided by the claude-code plugin
func (p *ClaudeCodePlugin) GetTools() []domain.Tool {
	return []domain.Tool{
		&SessionSummaryTool{plugin: p},
	}
}

// SessionSummaryTool provides a quick summary of a session
type SessionSummaryTool struct {
	plugin *ClaudeCodePlugin
}

// GetName returns the tool's command name
func (t *SessionSummaryTool) GetName() string {
	return "session-summary"
}

// GetDescription returns a brief description
func (t *SessionSummaryTool) GetDescription() string {
	return "Display a summary of a Claude Code session"
}

// GetUsage returns usage instructions
func (t *SessionSummaryTool) GetUsage() string {
	return "session-summary --session-id <id> | --last"
}

// Execute runs the tool
func (t *SessionSummaryTool) Execute(ctx context.Context, args []string, projectCtx *domain.ProjectContext) error {
	fs := flag.NewFlagSet("session-summary", flag.ContinueOnError)
	sessionID := fs.String("session-id", "", "Session ID to summarize")
	last := fs.Bool("last", false, "Summarize the most recent session")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Determine which session to summarize
	var targetSessionID string
	if *last {
		lastID, err := t.plugin.analysisService.GetLastSession(ctx)
		if err != nil {
			return fmt.Errorf("failed to get last session: %w", err)
		}
		targetSessionID = lastID
	} else if *sessionID != "" {
		targetSessionID = *sessionID
	} else {
		return fmt.Errorf("must specify either --session-id or --last")
	}

	// Get session entity
	entity, err := t.plugin.buildSessionEntity(ctx, targetSessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Display summary
	fmt.Println("Session Summary")
	fmt.Println("===============")
	fmt.Printf("Session ID: %s\n", entity.GetID())
	fmt.Printf("Event Count: %d\n", entity.GetField("event_count"))
	fmt.Printf("First Event: %v\n", entity.GetField("first_event"))
	fmt.Printf("Last Event: %v\n", entity.GetField("last_event"))
	fmt.Printf("Token Count: ~%d\n", entity.GetField("token_count"))
	fmt.Printf("Status: %s\n", entity.GetStatus())

	// Display analyses if available
	analyses := entity.GetAnalyses()
	if len(analyses) > 0 {
		fmt.Printf("\nAnalyses: %d\n", len(analyses))
		for i, analysis := range analyses {
			fmt.Printf("  [%d] %s (%s)\n", i+1, analysis.PromptName, analysis.ModelUsed)
			if analysis.PatternsSummary != "" {
				fmt.Printf("      Summary: %s\n", analysis.PatternsSummary)
			}
		}
	} else {
		fmt.Println("\nNo analyses available")
	}

	return nil
}
