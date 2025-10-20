package claude_code

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/domain"
)

// Ensure plugin implements SDK ICommandProvider
var _ domain.ICommandProvider = (*ClaudeCodePlugin)(nil)

// GetCommands returns the CLI commands provided by this plugin (SDK interface)
func (p *ClaudeCodePlugin) GetCommands() []domain.Command {
	return []domain.Command{
		&InitCommand{plugin: p},
		&LogCommand{plugin: p},
		&AutoSummaryCommand{plugin: p},
		&AutoSummaryExecCommand{plugin: p},
		&SessionSummaryCommand{plugin: p},
	}
}

// InitCommand initializes Claude Code logging infrastructure
type InitCommand struct {
	plugin *ClaudeCodePlugin
}

func (c *InitCommand) GetName() string {
	return "init"
}

func (c *InitCommand) GetDescription() string {
	return "Initialize Claude Code logging infrastructure"
}

func (c *InitCommand) GetUsage() string {
	return "init"
}

func (c *InitCommand) Execute(ctx context.Context, cmdCtx domain.CommandContext, args []string) error {
	if c.plugin.handler == nil {
		return fmt.Errorf("handler not initialized")
	}

	// Use the default database path
	return c.plugin.handler.Init(ctx, c.plugin.dbPath)
}

// LogCommand logs a Claude Code event from hook input
type LogCommand struct {
	plugin *ClaudeCodePlugin
}

func (c *LogCommand) GetName() string {
	return "log"
}

func (c *LogCommand) GetDescription() string {
	return "Log a Claude Code event (reads JSON from stdin)"
}

func (c *LogCommand) GetUsage() string {
	return "log <event-type>"
}

func (c *LogCommand) Execute(ctx context.Context, cmdCtx domain.CommandContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("event type required")
	}

	if c.plugin.handler == nil {
		return fmt.Errorf("handler not initialized")
	}

	eventTypeStr := args[0]

	// Get max param length from environment or use default
	maxParamLength := 30
	if envVal := os.Getenv("DW_MAX_PARAM_LENGTH"); envVal != "" {
		if parsed, err := fmt.Sscanf(envVal, "%d", &maxParamLength); err == nil && parsed == 1 {
			// Successfully parsed
		}
	}

	// Read stdin data from command context
	stdinData, err := io.ReadAll(cmdCtx.GetStdin())
	if err != nil {
		// Silently fail - don't disrupt Claude Code
		return nil
	}

	// Execute (silently - errors shouldn't disrupt Claude Code)
	_ = c.plugin.handler.Log(ctx, eventTypeStr, stdinData, maxParamLength)
	return nil
}

// AutoSummaryCommand handles auto-triggered session summaries on SessionEnd
type AutoSummaryCommand struct {
	plugin *ClaudeCodePlugin
}

func (c *AutoSummaryCommand) GetName() string {
	return "auto-summary"
}

func (c *AutoSummaryCommand) GetDescription() string {
	return "Auto-trigger session summary (called by SessionEnd hook)"
}

func (c *AutoSummaryCommand) GetUsage() string {
	return "auto-summary"
}

func (c *AutoSummaryCommand) Execute(ctx context.Context, cmdCtx domain.CommandContext, args []string) error {
	if c.plugin.handler == nil {
		return fmt.Errorf("handler not initialized")
	}

	// Read stdin data from command context
	stdinData, err := io.ReadAll(cmdCtx.GetStdin())
	if err != nil {
		// Silently fail - don't disrupt Claude Code
		return nil
	}

	// Execute (silently - errors shouldn't disrupt Claude Code)
	_ = c.plugin.handler.AutoSummary(ctx, stdinData)
	return nil
}

// AutoSummaryExecCommand executes the actual summary analysis in background
type AutoSummaryExecCommand struct {
	plugin *ClaudeCodePlugin
}

func (c *AutoSummaryExecCommand) GetName() string {
	return "auto-summary-exec"
}

func (c *AutoSummaryExecCommand) GetDescription() string {
	return "Internal: Execute summary in background (do not call directly)"
}

func (c *AutoSummaryExecCommand) GetUsage() string {
	return "auto-summary-exec <session-id>"
}

func (c *AutoSummaryExecCommand) Execute(ctx context.Context, cmdCtx domain.CommandContext, args []string) error {
	if len(args) < 1 {
		// No session ID provided
		return nil
	}

	if c.plugin.handler == nil {
		return fmt.Errorf("handler not initialized")
	}

	sessionID := args[0]

	// Execute (silently - errors shouldn't disrupt background analysis)
	_ = c.plugin.handler.AutoSummaryExec(ctx, sessionID)
	return nil
}

// SessionSummaryCommand provides a quick summary of a session
type SessionSummaryCommand struct {
	plugin *ClaudeCodePlugin
}

func (c *SessionSummaryCommand) GetName() string {
	return "session-summary"
}

func (c *SessionSummaryCommand) GetDescription() string {
	return "Display a summary of a Claude Code session"
}

func (c *SessionSummaryCommand) GetUsage() string {
	return "session-summary --session-id <id> | --last"
}

func (c *SessionSummaryCommand) Execute(ctx context.Context, cmdCtx domain.CommandContext, args []string) error {
	// Parse flags
	sessionID := ""
	last := false

	// Simple flag parsing
	for i := 0; i < len(args); i++ {
		if args[i] == "--session-id" && i+1 < len(args) {
			sessionID = args[i+1]
			i++
		} else if args[i] == "--last" {
			last = true
		}
	}

	// Determine which session to summarize
	var targetSessionID string
	if last {
		lastID, err := c.plugin.analysisService.GetLastSession(ctx)
		if err != nil {
			return fmt.Errorf("failed to get last session: %w", err)
		}
		targetSessionID = lastID
	} else if sessionID != "" {
		targetSessionID = sessionID
	} else {
		return fmt.Errorf("must specify either --session-id or --last")
	}

	// Get session entity
	entity, err := c.plugin.buildSessionEntity(ctx, targetSessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get output writer from command context
	out := cmdCtx.GetStdout()

	// Display summary
	fmt.Fprintln(out, "Session Summary")
	fmt.Fprintln(out, "===============")
	fmt.Fprintf(out, "Session ID: %s\n", entity.GetID())
	fmt.Fprintf(out, "Event Count: %d\n", entity.GetField("event_count"))
	fmt.Fprintf(out, "First Event: %v\n", entity.GetField("first_event"))
	fmt.Fprintf(out, "Last Event: %v\n", entity.GetField("last_event"))
	fmt.Fprintf(out, "Token Count: ~%d\n", entity.GetField("token_count"))
	fmt.Fprintf(out, "Status: %s\n", entity.GetStatus())

	// Display analyses if available
	analyses := entity.GetAnalyses()
	if len(analyses) > 0 {
		fmt.Fprintf(out, "\nAnalyses: %d\n", len(analyses))
		for i, analysis := range analyses {
			fmt.Fprintf(out, "  [%d] %s (%s)\n", i+1, analysis.PromptName, analysis.ModelUsed)
			if analysis.PatternsSummary != "" {
				fmt.Fprintf(out, "      Summary: %s\n", analysis.PatternsSummary)
			}
		}
	} else {
		fmt.Fprintln(out, "\nNo analyses available")
	}

	return nil
}
