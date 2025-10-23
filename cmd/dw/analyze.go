package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

func analyzeCmd(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	sessionID := fs.String("session-id", "", "Session ID to analyze")
	last := fs.Bool("last", false, "Analyze the last session")
	viewOnly := fs.Bool("view", false, "View existing analysis without re-analyzing")
	analyzeAll := fs.Bool("all", false, "Analyze all unanalyzed sessions")
	refresh := fs.Bool("refresh", false, "Re-analyze sessions even if already analyzed")
	limit := fs.Int("limit", 0, "Limit number of sessions to refresh/analyze (0 = all)")
	debug := fs.Bool("debug", false, "Enable debug logging")
	debugShort := fs.Bool("d", false, "Enable debug logging (short flag)")

	// New flags for multi-prompt analysis
	promptName := fs.String("prompt", "", "Prompt name from config to use (e.g., tool_analysis, session_summary)")
	modelOverride := fs.String("model", "", "Override model from config")
	tokenLimit := fs.Int("token-limit", 0, "Override token limit from config")

	if err := fs.Parse(args); err != nil {
		if err != flag.ErrHelp {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Enable debug if either flag is set
	debugEnabled := *debug || *debugShort

	// Create logger with appropriate level
	var logger *infra.Logger
	if debugEnabled {
		logger = infra.NewDebugLogger()
		logger.Info("Debug logging enabled")
	} else {
		logger = infra.NewDefaultLogger()
	}

	ctx := context.Background()

	// Initialize repository
	logger.Debug("Initializing repository at %s", app.DefaultDBPath)
	repo, err := infra.NewSQLiteEventRepository(app.DefaultDBPath)
	if err != nil {
		logger.Error("Failed to initialize repository: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to initialize repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Initialize database schema (including migration from old databases)
	if err := repo.Initialize(ctx); err != nil {
		logger.Error("Failed to initialize database schema: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to initialize database schema: %v\n", err)
		os.Exit(1)
	}

	// Load config
	logger.Debug("Loading configuration")
	configLoader := infra.NewConfigLoader(logger)
	config, err := configLoader.LoadConfig("")
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Apply CLI overrides to config
	if *modelOverride != "" {
		logger.Debug("Overriding model from CLI: %s", *modelOverride)
		if !infra.ValidateModelAlias(*modelOverride) {
			logger.Error("Invalid model: %s", *modelOverride)
			fmt.Fprintf(os.Stderr, "Error: Invalid model '%s'\n", *modelOverride)
			fmt.Fprintf(os.Stderr, "Allowed models: sonnet, opus, haiku, or specific versions\n")
			fmt.Fprintf(os.Stderr, "See .darwinflow.yaml for full list\n")
			os.Exit(1)
		}
		config.Analysis.Model = *modelOverride
	}
	if *tokenLimit > 0 {
		logger.Debug("Overriding token limit from CLI: %d", *tokenLimit)
		config.Analysis.TokenLimit = *tokenLimit
	}

	// Determine which prompts to use
	var selectedPrompts []string
	if *promptName != "" {
		// CLI override: use specific prompt
		selectedPrompts = []string{*promptName}
		logger.Debug("Using prompt from CLI: %s", *promptName)
	} else {
		// Use enabled prompts from config
		selectedPrompts = config.Analysis.EnabledPrompts
		logger.Debug("Using enabled prompts from config: %v", selectedPrompts)
	}

	// Create services
	logger.Debug("Creating analysis services")
	logsService := app.NewLogsService(repo, repo)
	llm := infra.NewClaudeCodeLLMWithConfig(logger, config)
	analysisService := app.NewAnalysisService(repo, repo, logsService, llm, logger, config)

	// Set the session view factory using the claude_code plugin
	analysisService.SetSessionViewFactory(func(sessionID string, events []pluginsdk.Event) pluginsdk.AnalysisView {
		return claude_code.NewSessionView(sessionID, events)
	})

	// Create command handler
	handler := app.NewAnalyzeCommandHandler(analysisService, logger, os.Stdout)

	// Build options
	opts := app.AnalyzeOptions{
		SessionID:     *sessionID,
		Last:          *last,
		ViewOnly:      *viewOnly,
		AnalyzeAll:    *analyzeAll,
		Refresh:       *refresh,
		Limit:         *limit,
		PromptNames:   selectedPrompts,
		ModelOverride: *modelOverride,
		TokenLimit:    *tokenLimit,
	}

	// Execute
	if err := handler.Execute(ctx, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

