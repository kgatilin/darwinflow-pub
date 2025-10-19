package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func handleClaudeCommand(args []string) {
	if len(args) < 1 {
		printClaudeUsage()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "init":
		handleInit(args[1:])
	case "log":
		handleLog(args[1:])
	case "auto-summary":
		handleAutoSummary(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown claude subcommand: %s\n\n", subcommand)
		printClaudeUsage()
		os.Exit(1)
	}
}

func printClaudeUsage() {
	fmt.Println("Usage: dw claude <subcommand>")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  init              Initialize Claude Code logging infrastructure")
	fmt.Println("  log <event-type>  Log a Claude Code event (reads JSON from stdin)")
	fmt.Println("  auto-summary      Auto-trigger session summary (called by SessionEnd hook)")
	fmt.Println()
}

// handleAutoSummary handles auto-triggered session summaries on SessionEnd
// This is called by the SessionEnd hook and only runs if auto_summary_enabled is true in config
func handleAutoSummary(args []string) {
	// Silently execute (errors shouldn't disrupt Claude Code)
	if err := autoSummaryFromStdin(); err != nil {
		// Fail silently - don't disrupt Claude Code
		return
	}
}

func autoSummaryFromStdin() error {
	// Read hook input from stdin
	stdinData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	// Try to parse as hook input
	hookInput, err := infra.ParseHookInput(io.NopCloser(bytes.NewReader(stdinData)))
	if err != nil {
		// Not valid hook input, fail silently
		return nil
	}

	// Load config to check if auto-summary is enabled
	logger := infra.NewDefaultLogger()
	configLoader := infra.NewConfigLoader(logger)
	config, err := configLoader.LoadConfig("")
	if err != nil {
		// Config load failed, fail silently
		return nil
	}

	// Check if auto-summary is enabled
	if !config.Analysis.AutoSummaryEnabled {
		// Auto-summary disabled, silently exit
		return nil
	}

	// Get the session ID from hook input
	sessionID := hookInput.SessionID
	if sessionID == "" {
		// No session ID, can't analyze
		return nil
	}

	// Get the prompt name from config
	promptName := config.Analysis.AutoSummaryPrompt
	if promptName == "" {
		promptName = "session_summary"
	}

	// Create repository and services
	repo, err := infra.NewSQLiteEventRepository(app.DefaultDBPath)
	if err != nil {
		return err
	}
	defer repo.Close()

	logsService := app.NewLogsService(repo, repo)
	llmExecutor := app.NewClaudeCLIExecutorWithConfig(logger, config)
	analysisService := app.NewAnalysisService(repo, repo, logsService, llmExecutor, logger, config)

	// Trigger analysis in background (don't block)
	// Use a goroutine so the hook returns quickly
	go func() {
		_, _ = analysisService.AnalyzeSessionWithPrompt(context.Background(), sessionID, promptName)
		// Ignore errors - this is best-effort background analysis
	}()

	return nil
}

func handleInit(args []string) {
	dbPath := app.DefaultDBPath

	fmt.Println("Initializing Claude Code logging for DarwinFlow...")
	fmt.Println()

	// Create infrastructure dependencies
	repository, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating repository: %v\n", err)
		os.Exit(1)
	}
	defer repository.Close()

	hookConfigManager, err := infra.NewHookConfigManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating hook config manager: %v\n", err)
		os.Exit(1)
	}

	// Create application service
	setupService := app.NewSetupService(repository, hookConfigManager)

	// Initialize logging infrastructure
	ctx := context.Background()
	if err := setupService.Initialize(ctx, dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Created logging database:", dbPath)
	fmt.Println("✓ Added hooks to Claude Code settings:", setupService.GetSettingsPath())
	fmt.Println()
	fmt.Println("DarwinFlow logging is now active for all Claude Code sessions.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Restart Claude Code to activate the hooks")
	fmt.Println("  2. Events will be automatically logged to", dbPath)
}

func handleLog(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: event type required")
		fmt.Fprintln(os.Stderr, "Usage: dw claude log <event-type>")
		os.Exit(1)
	}

	eventTypeStr := args[0]

	// Get max param length from environment or use default
	maxParamLength := 30
	if envVal := os.Getenv("DW_MAX_PARAM_LENGTH"); envVal != "" {
		if parsed, err := fmt.Sscanf(envVal, "%d", &maxParamLength); err == nil && parsed == 1 {
			// Successfully parsed
		}
	}

	// Silently execute (errors shouldn't disrupt Claude Code)
	if err := logFromStdin(eventTypeStr, maxParamLength); err != nil {
		// Silently fail - don't disrupt Claude Code
		return
	}
}

func logFromStdin(eventTypeStr string, maxParamLength int) error {
	// Read hook input from stdin
	stdinData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	// Try to parse as hook input
	hookInput, err := infra.ParseHookInput(io.NopCloser(bytes.NewReader(stdinData)))
	if err != nil {
		// Not valid hook input, fail silently
		return nil
	}

	// Map event type string to domain event type
	eventMapper := &app.EventMapper{}
	eventType := eventMapper.MapEventType(eventTypeStr)

	// Create infrastructure dependencies
	repository, err := infra.NewSQLiteEventRepository(app.DefaultDBPath)
	if err != nil {
		return err
	}
	defer repository.Close()

	transcriptParser := infra.NewTranscriptParser()
	contextDetector := infra.NewContextDetector()

	// Create application service
	loggerService := app.NewLoggerService(
		repository,
		transcriptParser,
		contextDetector,
		infra.NormalizeContent,
	)
	defer loggerService.Close()

	// Convert infra.HookInput to app.HookInputData
	hookInputData := app.HookInputData{
		SessionID:      hookInput.SessionID,
		TranscriptPath: hookInput.TranscriptPath,
		CWD:            hookInput.CWD,
		PermissionMode: hookInput.PermissionMode,
		HookEventName:  hookInput.HookEventName,
		ToolName:       hookInput.ToolName,
		ToolInput:      hookInput.ToolInput,
		ToolOutput:     hookInput.ToolOutput,
		Error:          hookInput.Error,
		UserMessage:    hookInput.UserMessage,
		Prompt:         hookInput.Prompt,
	}

	// Log event
	ctx := context.Background()
	return loggerService.LogFromHookInput(ctx, hookInputData, eventType, maxParamLength)
}
