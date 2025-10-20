package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// ClaudeCommandHandler handles all claude subcommands
type ClaudeCommandHandler struct {
	setupService    *SetupService
	loggerService   *LoggerService
	analysisService *AnalysisService
	hookInputParser HookInputParser
	eventMapper     *EventMapper
	configLoader    ConfigLoader
	logger          Logger
	output          io.Writer
}

// HookInputParser defines the interface for parsing hook input from stdin
type HookInputParser interface {
	Parse(data []byte) (*HookInputData, error)
}

// NewClaudeCommandHandler creates a new ClaudeCommandHandler
func NewClaudeCommandHandler(
	setupService *SetupService,
	loggerService *LoggerService,
	analysisService *AnalysisService,
	hookInputParser HookInputParser,
	eventMapper *EventMapper,
	configLoader ConfigLoader,
	logger Logger,
	output io.Writer,
) *ClaudeCommandHandler {
	return &ClaudeCommandHandler{
		setupService:    setupService,
		loggerService:   loggerService,
		analysisService: analysisService,
		hookInputParser: hookInputParser,
		eventMapper:     eventMapper,
		configLoader:    configLoader,
		logger:          logger,
		output:          output,
	}
}

// Init initializes Claude Code logging infrastructure
func (h *ClaudeCommandHandler) Init(ctx context.Context, dbPath string) error {
	fmt.Fprintln(h.output, "Initializing Claude Code logging for DarwinFlow...")
	fmt.Fprintln(h.output)

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("error creating database directory: %w", err)
	}

	// Initialize logging infrastructure
	if err := h.setupService.Initialize(ctx, dbPath); err != nil {
		return err
	}

	fmt.Fprintln(h.output, "✓ Created logging database:", dbPath)
	fmt.Fprintln(h.output, "✓ Added hooks to Claude Code settings:", h.setupService.GetSettingsPath())
	fmt.Fprintln(h.output)
	fmt.Fprintln(h.output, "DarwinFlow logging is now active for all Claude Code sessions.")
	fmt.Fprintln(h.output)
	fmt.Fprintln(h.output, "Next steps:")
	fmt.Fprintln(h.output, "  1. Restart Claude Code to activate the hooks")
	fmt.Fprintln(h.output, "  2. Events will be automatically logged to", dbPath)

	return nil
}

// Log logs a Claude Code event from hook input
func (h *ClaudeCommandHandler) Log(ctx context.Context, eventTypeStr string, stdinData []byte, maxParamLength int) error {
	// Parse hook input
	hookInputData, err := h.hookInputParser.Parse(stdinData)
	if err != nil {
		// Not valid hook input, fail silently
		return nil
	}

	// Map event type string to domain event type
	eventType := h.eventMapper.MapEventType(eventTypeStr)

	// Log event using logger service
	return h.loggerService.LogFromHookInput(ctx, *hookInputData, eventType, maxParamLength)
}

// AutoSummary handles auto-triggered session summaries on SessionEnd
// Returns immediately after spawning background process
func (h *ClaudeCommandHandler) AutoSummary(ctx context.Context, stdinData []byte) error {
	// Parse hook input to extract session ID
	hookInputData, err := h.hookInputParser.Parse(stdinData)
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
	config, err := h.configLoader.LoadConfig("")
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
	if err := h.spawnBackgroundSummary(sessionID); err != nil {
		// Fail silently - don't disrupt Claude Code
		return nil
	}

	return nil
}

// spawnBackgroundSummary spawns a detached background process to execute the summary
func (h *ClaudeCommandHandler) spawnBackgroundSummary(sessionID string) error {
	// Get the path to the current executable
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// Create command: dw claude auto-summary-exec <session-id>
	cmd := exec.Command(executable, "claude", "auto-summary-exec", sessionID)

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

// AutoSummaryExec executes the actual summary analysis in background
func (h *ClaudeCommandHandler) AutoSummaryExec(ctx context.Context, sessionID string) error {
	// Load config
	config, err := h.configLoader.LoadConfig("")
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
	_, _ = h.analysisService.AnalyzeSessionWithPrompt(ctx, sessionID, promptName)
	// Ignore errors - this is best-effort background analysis
	return nil
}
