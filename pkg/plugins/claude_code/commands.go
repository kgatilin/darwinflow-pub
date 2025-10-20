package claude_code

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// Ensure plugin implements SDK ICommandProvider
var _ pluginsdk.ICommandProvider = (*ClaudeCodePlugin)(nil)

// GetCommands returns the CLI commands provided by this plugin (SDK interface)
func (p *ClaudeCodePlugin) GetCommands() []pluginsdk.Command {
	return []pluginsdk.Command{
		&InitCommand{plugin: p},
		&LogCommand{plugin: p},
		&AutoSummaryCommand{plugin: p},
		&AutoSummaryExecCommand{plugin: p},
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

func (c *InitCommand) Execute(ctx context.Context, args []string) error {
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

func (c *LogCommand) Execute(ctx context.Context, args []string) error {
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

	// Read stdin data
	stdinData, err := io.ReadAll(os.Stdin)
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

func (c *AutoSummaryCommand) Execute(ctx context.Context, args []string) error {
	if c.plugin.handler == nil {
		return fmt.Errorf("handler not initialized")
	}

	// Read stdin data
	stdinData, err := io.ReadAll(os.Stdin)
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

func (c *AutoSummaryExecCommand) Execute(ctx context.Context, args []string) error {
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
