package claude_code

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// Ensure plugin implements SDK ICommandProvider
var _ pluginsdk.ICommandProvider = (*ClaudeCodePlugin)(nil)

// logToFile appends a log message to .darwinflow/claude-code.log
// This is used for debugging hook failures. All errors are swallowed to prevent
// disrupting Claude Code hooks.
// configuredLevel should be one of: "debug", "info", "error", "off"
func logToFile(workingDir, level, message, configuredLevel string) {
	// Recover from any panics to prevent disrupting hooks
	defer func() {
		_ = recover()
	}()

	// Normalize level strings to lowercase
	level = strings.ToLower(level)
	configuredLevel = strings.ToLower(configuredLevel)

	// Check if logging is disabled
	if configuredLevel == "off" {
		return
	}

	// Filter based on configured log level
	// "error" = only ERROR
	// "info" = INFO and ERROR
	// "debug" = DEBUG, INFO, and ERROR
	switch configuredLevel {
	case "error":
		if level != "error" {
			return
		}
	case "info":
		if level != "info" && level != "error" {
			return
		}
	case "debug":
		// Log everything
	default:
		// Unknown level, default to error-only
		if level != "error" {
			return
		}
	}

	// Create .darwinflow directory if it doesn't exist
	logDir := filepath.Join(workingDir, ".darwinflow")
	_ = os.MkdirAll(logDir, 0755)

	// Open log file in append mode
	logPath := filepath.Join(logDir, "claude-code.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Silently fail
	}
	defer f.Close()

	// Write log entry with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logEntry := fmt.Sprintf("[%s] %s: %s\n", timestamp, strings.ToUpper(level), message)
	_, _ = f.WriteString(logEntry)
}

// GetCommands returns the CLI commands provided by this plugin (SDK interface)
func (p *ClaudeCodePlugin) GetCommands() []pluginsdk.Command {
	return []pluginsdk.Command{
		&InitCommand{plugin: p},
		&EmitEventCommand{plugin: p},
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

func (c *InitCommand) GetHelp() string {
	return `Initializes Claude Code logging infrastructure for DarwinFlow.

This command sets up the necessary hooks in .claude/settings.json to
automatically capture Claude Code events and log them to DarwinFlow.

Examples:
  # Normal initialization
  dw claude-code init

  # Force reinstall hooks (reinstall even if they already exist)
  dw claude-code init --force

Flags:
  --force    Reinstall hooks even if they already exist

What this does:
  - Creates .darwinflow/logs/ directory
  - Initializes SQLite events database
  - Installs hooks in .claude/settings.json
  - Sets up automatic event logging

Notes:
  - Safe to run multiple times
  - Restarts Claude Code to activate hooks
  - Events are logged to .darwinflow/logs/events.db`
}

func (c *InitCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	out := cmdCtx.GetStdout()

	fmt.Fprintln(out, "Initializing Claude Code logging for DarwinFlow...")
	fmt.Fprintln(out)

	// Initialize framework infrastructure (database, schema)
	if err := c.plugin.setupService.Initialize(ctx, c.plugin.dbPath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	fmt.Fprintln(out, "✓ Created logging database:", c.plugin.dbPath)

	// Install Claude Code hooks (plugin's responsibility, not framework's)
	hookMgr, err := NewHookConfigManager()
	if err != nil {
		// Log warning but don't fail - hooks are optional
		c.plugin.logger.Warn("Failed to create hook config manager: %v", err)
	} else if err := hookMgr.InstallDarwinFlowHooks(); err != nil {
		// Log warning but don't fail - hooks are optional
		c.plugin.logger.Warn("Failed to install hooks: %v", err)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "DarwinFlow logging is now active for all Claude Code sessions.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Next steps:")
	fmt.Fprintln(out, "  1. Restart Claude Code to activate the hooks")
	fmt.Fprintln(out, "  2. Events will be automatically logged to", c.plugin.dbPath)

	return nil
}

// EmitEventCommand emits an event via the plugin SDK context
// This command reads a structured event from stdin and emits it through the plugin context.
// Supports two input formats:
//
// 1. Claude Code Native Hook Format (HookInput):
//
//	{
//	  "session_id": "abc123",
//	  "hook_event_name": "PreToolUse",
//	  "tool_name": "Read",
//	  "tool_input": {...},
//	  "cwd": "/workspace",
//	  ...
//	}
//
// 2. Plugin SDK Event Format:
//
//	{
//	  "type": "tool.invoked",
//	  "source": "claude-code",
//	  "timestamp": "2025-10-20T10:30:00Z",
//	  "payload": { "tool": "Read", "parameters": {...} },
//	  "metadata": { "session_id": "abc123", "cwd": "/workspace" },
//	  "version": "1.0"
//	}
//
// The command auto-detects the format and converts as needed.
// All errors are logged but never propagated - this ensures hook execution is never disrupted.
// Required fields depend on format:
//   - HookInput: session_id, hook_event_name
//   - SDK Event: type, source, metadata.session_id
//
// The command validates input and emits to the framework's event store.
type EmitEventCommand struct {
	plugin *ClaudeCodePlugin
}

func NewEmitEventCommand(plugin *ClaudeCodePlugin) *EmitEventCommand {
	return &EmitEventCommand{
		plugin: plugin,
	}
}

func (c *EmitEventCommand) GetName() string {
	return "emit-event"
}

func (c *EmitEventCommand) GetDescription() string {
	return "Emit an event via plugin context (reads JSON from stdin)"
}

func (c *EmitEventCommand) GetUsage() string {
	return "emit-event"
}

func (c *EmitEventCommand) GetHelp() string {
	return `Emit an event via plugin context (reads JSON from stdin).

This command reads a structured event from stdin and emits it through the
DarwinFlow event system. Supports both Claude Code hook format and SDK event
format. All errors are logged but never propagated - this ensures hook
execution is never disrupted.

Input Formats:

1. Claude Code Hook Format (HookInput):
   {
     "session_id": "abc123",
     "hook_event_name": "PreToolUse",
     "tool_name": "Read",
     "tool_input": {...},
     "cwd": "/workspace"
   }

2. SDK Event Format:
   {
     "type": "tool.invoked",
     "source": "claude-code",
     "timestamp": "2025-10-20T10:30:00Z",
     "payload": {"tool": "Read", "parameters": {...}},
     "metadata": {"session_id": "abc123", "cwd": "/workspace"},
     "version": "1.0"
   }

Examples:
  # Emit via SDK format
  echo '{"type":"tool.invoked","source":"claude-code","metadata":{"session_id":"abc123"}}' | dw claude-code emit-event

  # Emit via hook format
  echo '{"session_id":"abc123","hook_event_name":"PostToolUse"}' | dw claude-code emit-event

Required Fields:
  HookInput:
    - session_id
    - hook_event_name

  SDK Event:
    - type
    - source
    - metadata.session_id

Notes:
  - Auto-detects input format (Hook vs SDK)
  - Safe to call from Claude Code hooks
  - Failures are logged, not propagated
  - Empty stdin is silently ignored`
}

func (c *EmitEventCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get working directory for file logging
	workingDir := cmdCtx.GetWorkingDir()

	// Load config to get log level setting
	config, err := c.plugin.configLoader.LoadConfig("")
	logLevel := "error" // Default to error-only if config fails to load
	if err == nil {
		logLevel = config.Logging.FileLogLevel
		if logLevel == "" {
			logLevel = "error" // Use default if not set
		}
	}

	// Add timeout to prevent infinite hangs
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Safely recover from panics
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("panic recovered: %v", r)
			c.plugin.logger.Error("emit-event: %s", errMsg)
			logToFile(workingDir, "ERROR", errMsg, logLevel)
		}
	}()

	// Read stdin
	stdinData, err := io.ReadAll(cmdCtx.GetStdin())
	if err != nil {
		errMsg := fmt.Sprintf("failed to read stdin: %v", err)
		c.plugin.logger.Debug("emit-event: %s", errMsg)
		logToFile(workingDir, "DEBUG", errMsg, logLevel)
		return nil // Silently fail - don't disrupt Claude Code
	}

	if len(stdinData) == 0 {
		errMsg := "empty stdin"
		c.plugin.logger.Debug("emit-event: %s", errMsg)
		logToFile(workingDir, "DEBUG", errMsg, logLevel)
		return nil // Silently fail - don't disrupt Claude Code
	}

	// Try to parse as SDK Event first
	var event pluginsdk.Event
	if err := json.Unmarshal(stdinData, &event); err != nil {
		errMsg := fmt.Sprintf("invalid JSON: %v", err)
		c.plugin.logger.Debug("emit-event: %s", errMsg)
		logToFile(workingDir, "DEBUG", errMsg, logLevel)
		return nil // Silently fail - don't disrupt Claude Code
	}

	// Check if this looks like a Claude Code hook input (HookInput format)
	// vs a SDK Event format
	if event.Type == "" && event.Source == "" {
		// Try to parse as HookInput (Claude Code native format)
		hookData, err := c.parseAsHookInput(stdinData)
		if err != nil {
			errMsg := fmt.Sprintf("failed to parse as hook input: %v", err)
			c.plugin.logger.Debug("emit-event: %s", errMsg)
			logToFile(workingDir, "DEBUG", errMsg, logLevel)
			return nil // Silently fail - don't disrupt Claude Code
		}

		// Convert HookInput to SDK Event
		event = *HookInputToEvent(hookData)
		if event.Type == "unknown" {
			errMsg := "unknown hook event type"
			c.plugin.logger.Debug("emit-event: %s", errMsg)
			logToFile(workingDir, "DEBUG", errMsg, logLevel)
			return nil
		}
	}

	// Validate required fields
	if event.Type == "" {
		errMsg := "missing required field: type"
		c.plugin.logger.Debug("emit-event: %s", errMsg)
		logToFile(workingDir, "DEBUG", errMsg, logLevel)
		return nil
	}

	if event.Source == "" {
		errMsg := "missing required field: source"
		c.plugin.logger.Debug("emit-event: %s", errMsg)
		logToFile(workingDir, "DEBUG", errMsg, logLevel)
		return nil
	}

	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}

	sessionID, ok := event.Metadata["session_id"]
	if !ok || sessionID == "" {
		errMsg := "missing required field: metadata.session_id"
		c.plugin.logger.Debug("emit-event: %s", errMsg)
		logToFile(workingDir, "DEBUG", errMsg, logLevel)
		return nil
	}

	// Set default timestamp if missing
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set default version if missing
	if event.Version == "" {
		event.Version = "1.0"
	}

	// Initialize empty payload if nil
	if event.Payload == nil {
		event.Payload = make(map[string]interface{})
	}

	// Emit event via plugin context (silently fail if DB error)
	if err := cmdCtx.EmitEvent(ctxWithTimeout, event); err != nil {
		errMsg := fmt.Sprintf("failed to emit event: %v", err)
		c.plugin.logger.Debug("emit-event: %s", errMsg)
		logToFile(workingDir, "ERROR", errMsg, logLevel)
		return nil // Silently fail - don't disrupt Claude Code
	}

	// Log success
	logToFile(workingDir, "INFO", fmt.Sprintf("successfully emitted event: type=%s, source=%s, session_id=%s", event.Type, event.Source, sessionID), logLevel)

	return nil
}

// parseAsHookInput attempts to parse stdin data as HookInput format
func (c *EmitEventCommand) parseAsHookInput(data []byte) (*HookInputData, error) {
	parser := newHookInputParser()
	return parser.Parse(data)
}

// LogCommand logs a Claude Code event from hook input
// DEPRECATED: Use EmitEventCommand instead (will be removed in v2.0)
// This command is kept for backward compatibility with existing hooks
type LogCommand struct {
	plugin *ClaudeCodePlugin
}

func (c *LogCommand) GetName() string {
	return "log"
}

func (c *LogCommand) GetDescription() string {
	return "DEPRECATED: Log a Claude Code event (use emit-event instead)"
}

func (c *LogCommand) GetUsage() string {
	return "log <event-type>"
}

func (c *LogCommand) GetHelp() string {
	return `DEPRECATED: Log a Claude Code event (use emit-event instead).

This command is deprecated and kept only for backward compatibility with
older Claude Code hook configurations. All new integrations should use the
emit-event command instead.

Migration Path:
  Old (deprecated):
    dw claude-code log PreToolUse

  New (recommended):
    echo '{"type":"tool.invoked","source":"claude-code","metadata":{"session_id":"abc123"}}' | \
    dw claude-code emit-event

Why Deprecate:
  - emit-event supports both hook format and SDK event format
  - emit-event is more flexible and standardized
  - emit-event provides better error logging

Removal:
  - Scheduled for removal in v2.0
  - No timeline for v2.0 release

Current Behavior:
  - Silently succeeds (for backward compatibility)
  - Does not actually log events
  - Check your logs for warnings

Notes:
  - This command is deprecated
  - Use 'emit-event' for new integrations
  - Existing hooks will continue to work`
}

func (c *LogCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// DEPRECATED: This command is kept for backward compatibility
	// Silently fail - don't disrupt Claude Code hooks
	c.plugin.logger.Warn("LogCommand is deprecated, use EmitEventCommand instead")
	return nil
}

// AutoSummaryCommand handles auto-triggered session summaries on SessionEnd
// Returns immediately after spawning background process
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

func (c *AutoSummaryCommand) GetHelp() string {
	return `Auto-trigger session summary when a Claude Code session ends.

This command is called automatically by the SessionEnd hook when a Claude Code
session terminates. It spawns a background process to generate a summary
without blocking the session exit.

Purpose:
  - Automatically analyzes completed sessions
  - Generates summaries using configured AI prompt
  - Enables pattern detection and workflow analysis
  - Non-blocking (background execution)

How It Works:
  1. SessionEnd hook triggers this command
  2. Extracts session ID from hook data
  3. Spawns background process: dw claude-code auto-summary-exec <session-id>
  4. Returns immediately (doesn't block session exit)
  5. Background process performs analysis

Configuration:
  Enable in .darwinflow/config.json:
    {
      "analysis": {
        "auto_summary_enabled": true,
        "auto_summary_prompt": "session_summary"
      }
    }

Environment:
  - Auto-summary is enabled by default
  - Can be disabled via config
  - Uses default "session_summary" prompt if not configured

Notes:
  - Typically called from Claude Code SessionEnd hook
  - Do not call manually - use session-summary instead
  - Background process runs asynchronously
  - Session exit is never blocked by analysis`
}

func (c *AutoSummaryCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Add timeout to prevent infinite hangs
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Safely recover from panics
	defer func() {
		if r := recover(); r != nil {
			c.plugin.logger.Error("auto-summary: panic recovered: %v", r)
		}
	}()

	// Read stdin data from command context
	stdinData, err := io.ReadAll(cmdCtx.GetStdin())
	if err != nil {
		// Silently fail - don't disrupt Claude Code
		return nil
	}

	// Parse hook input to extract session ID
	hookInputData, err := c.plugin.hookInputParser.Parse(stdinData)
	if err != nil {
		// Not valid hook input, fail silently
		return nil
	}

	// Get session ID
	sessionID := hookInputData.SessionID
	if sessionID == "" {
		// No session ID, can't analyze
		return nil
	}

	// Load config to check if auto-summary is enabled
	config, err := c.plugin.configLoader.LoadConfig("")
	if err != nil {
		// Config load failed, fail silently
		return nil
	}

	// Check if auto-summary is enabled
	if !config.Analysis.AutoSummaryEnabled {
		// Auto-summary disabled, silently exit
		return nil
	}

	// Spawn detached background process to execute the summary
	if err := c.spawnBackgroundSummary(ctxWithTimeout, sessionID); err != nil {
		// Fail silently - don't disrupt Claude Code
		return nil
	}

	return nil
}

// spawnBackgroundSummary spawns a detached background process to execute the summary
func (c *AutoSummaryCommand) spawnBackgroundSummary(ctx context.Context, sessionID string) error {
	// Get the path to the current executable
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// Create command: dw claude-code auto-summary-exec <session-id>
	cmd := exec.Command(executable, "claude-code", "auto-summary-exec", sessionID)

	// Detach from parent process
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Start the process without waiting for it to complete
	if err := cmd.Start(); err != nil {
		return err
	}

	// Don't wait for the process - let it run in background
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

func (c *AutoSummaryExecCommand) GetHelp() string {
	return `INTERNAL: Execute session summary analysis in background.

This command is internal and not intended for direct use. It is spawned
as a background process by the auto-summary command to perform the actual
session analysis asynchronously.

Usage:
  INTERNAL - Called by auto-summary command only
  Do not invoke directly

Arguments:
  <session-id>    The ID of the session to analyze (required)

How It Works:
  1. Receives session ID from auto-summary command
  2. Loads DarwinFlow configuration
  3. Retrieves configured analysis prompt
  4. Runs analysis service on the session
  5. Stores results in database
  6. Exits silently

Error Handling:
  - All errors are silently ignored
  - Logs errors internally
  - Never propagates failures (background execution)
  - Session data is not affected by analysis failures

Configuration:
  Uses these config settings:
    analysis.auto_summary_prompt    - Analysis prompt name to use

Notes:
  - INTERNAL COMMAND - Not for manual use
  - Run as background process (spawned by auto-summary)
  - Best-effort execution (errors ignored)
  - Safe to call concurrently for different sessions
  - Cannot be interrupted once started`
}

func (c *AutoSummaryExecCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) < 1 {
		// No session ID provided
		return nil
	}

	sessionID := args[0]

	// Load config
	config, err := c.plugin.configLoader.LoadConfig("")
	if err != nil {
		// Config load failed, exit silently
		return nil
	}

	// Get the prompt name from config
	promptName := config.Analysis.AutoSummaryPrompt
	if promptName == "" {
		promptName = "session_summary"
	}

	// Execute the analysis using the analysis service
	_, _ = c.plugin.analysisService.AnalyzeSessionWithPrompt(ctx, sessionID, promptName)
	// Ignore errors - this is best-effort background analysis
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

func (c *SessionSummaryCommand) GetHelp() string {
	return `Display a summary of a Claude Code session with analysis results.

This command retrieves and displays comprehensive information about a specific
Claude Code session, including event count, timing, token usage, and any
available analysis results from AI prompts.

Examples:
  # View last session
  dw claude-code session-summary --last

  # View specific session
  dw claude-code session-summary --session-id abc123def456

  # View last session (shorthand)
  dw claude-code session-summary --last

Flags:
  --session-id <id>    Display summary for specific session ID
  --last               Display summary for most recent session
  (must specify one)

Output Fields:
  Session ID          - Unique session identifier
  Event Count         - Number of events in session
  First Event         - Timestamp of first event
  Last Event          - Timestamp of last event
  Token Count         - Approximate total tokens used
  Status              - Session status (active, completed, etc.)
  Analyses            - List of analysis results with prompts

Analysis Section:
  - Shows all available analyses for the session
  - Lists prompt name and model used
  - Displays summary text if available
  - Multiple analyses can exist per session

Notes:
  - Requires either --session-id or --last
  - Shows most recent analysis results
  - Works with sessions that have no analysis
  - Use dw logs to see raw events
  - Use dw analyze to run new analysis`
}

func (c *SessionSummaryCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
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
