package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

func handleClaudeCommand(args []string) {
	if len(args) < 1 {
		printClaudeUsage()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "init":
		handleClaudeInit(args[1:])
	case "log":
		handleLog(args[1:])
	case "auto-summary":
		handleAutoSummary(args[1:])
	case "auto-summary-exec":
		handleAutoSummaryExec(args[1:])
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
	fmt.Println("  auto-summary-exec Internal: Execute summary in background (do not call directly)")
	fmt.Println()
}

func handleClaudeInit(args []string) {
	dbPath := app.DefaultDBPath

	// Ensure database directory exists before creating repository
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating database directory: %v\n", err)
		os.Exit(1)
	}

	// Create infrastructure dependencies
	repository, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating repository: %v\n", err)
		os.Exit(1)
	}
	defer repository.Close()

	hookConfigManager, err := claude_code.NewHookConfigManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating hook config manager: %v\n", err)
		os.Exit(1)
	}

	logger := infra.NewDefaultLogger()
	configLoader := infra.NewConfigLoader(logger)

	// Create setup service
	setupService := app.NewSetupService(repository, hookConfigManager)

	// Create handler
	handler := app.NewClaudeCommandHandler(
		setupService,
		nil, // loggerService not needed for init
		nil, // analysisService not needed for init
		nil, // hookInputParser not needed for init
		nil, // eventMapper not needed for init
		configLoader,
		logger,
		os.Stdout,
	)

	// Execute
	ctx := context.Background()
	if err := handler.Init(ctx, dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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

	// Read stdin data
	stdinData, err := io.ReadAll(os.Stdin)
	if err != nil {
		// Silently fail - don't disrupt Claude Code
		return
	}

	// Create infrastructure dependencies
	repository, err := infra.NewSQLiteEventRepository(app.DefaultDBPath)
	if err != nil {
		// Silently fail
		return
	}
	defer repository.Close()

	logger := infra.NewDefaultLogger()
	configLoader := infra.NewConfigLoader(logger)
	transcriptParser := infra.NewTranscriptParser()
	contextDetector := infra.NewContextDetector()
	hookInputParser := infra.NewHookInputParserAdapter()
	eventMapper := &app.EventMapper{}

	// Create logger service
	loggerService := app.NewLoggerService(
		repository,
		transcriptParser,
		contextDetector,
		infra.NormalizeContent,
	)
	defer loggerService.Close()

	// Create handler
	handler := app.NewClaudeCommandHandler(
		nil, // setupService not needed for log
		loggerService,
		nil, // analysisService not needed for log
		hookInputParser,
		eventMapper,
		configLoader,
		logger,
		os.Stdout,
	)

	// Execute (silently - errors shouldn't disrupt Claude Code)
	ctx := context.Background()
	_ = handler.Log(ctx, eventTypeStr, stdinData, maxParamLength)
}

func handleAutoSummary(args []string) {
	// Read stdin data
	stdinData, err := io.ReadAll(os.Stdin)
	if err != nil {
		// Silently fail - don't disrupt Claude Code
		return
	}

	logger := infra.NewDefaultLogger()
	configLoader := infra.NewConfigLoader(logger)
	hookInputParser := infra.NewHookInputParserAdapter()

	// Create handler
	handler := app.NewClaudeCommandHandler(
		nil, // setupService not needed
		nil, // loggerService not needed
		nil, // analysisService not needed
		hookInputParser,
		nil, // eventMapper not needed
		configLoader,
		logger,
		os.Stdout,
	)

	// Execute (silently - errors shouldn't disrupt Claude Code)
	ctx := context.Background()
	_ = handler.AutoSummary(ctx, stdinData)
}

func handleAutoSummaryExec(args []string) {
	if len(args) < 1 {
		// No session ID provided
		return
	}

	sessionID := args[0]

	// Create infrastructure dependencies
	repository, err := infra.NewSQLiteEventRepository(app.DefaultDBPath)
	if err != nil {
		// Exit silently on error
		return
	}
	defer repository.Close()

	logger := infra.NewDefaultLogger()
	configLoader := infra.NewConfigLoader(logger)

	// Create services
	logsService := app.NewLogsService(repository, repository)
	config, _ := configLoader.LoadConfig("")
	if config == nil {
		// Can't proceed without config
		return
	}

	llmExecutor := app.NewClaudeCLIExecutorWithConfig(logger, config)
	analysisService := app.NewAnalysisService(repository, repository, logsService, llmExecutor, logger, config)

	// Create handler
	handler := app.NewClaudeCommandHandler(
		nil, // setupService not needed
		nil, // loggerService not needed
		analysisService,
		nil, // hookInputParser not needed
		nil, // eventMapper not needed
		configLoader,
		logger,
		os.Stdout,
	)

	// Execute (silently - errors shouldn't disrupt background analysis)
	ctx := context.Background()
	_ = handler.AutoSummaryExec(ctx, sessionID)
}
