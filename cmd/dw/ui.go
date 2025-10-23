package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/app/tui"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/claude_code"
)

func uiCommand(args []string) {
	fs := flag.NewFlagSet("ui", flag.ContinueOnError)
	dbPath := fs.String("db", app.DefaultDBPath, "Path to SQLite database")
	configPath := fs.String("config", "", "Path to config file (default: .darwinflow.yaml in current dir)")
	debugMode := fs.Bool("debug", false, "Enable debug logging")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return
		}
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	var logger *infra.Logger
	if *debugMode {
		logger = infra.NewDebugLogger()
	} else {
		logger = infra.NewDefaultLogger()
	}

	// Load config
	configLoader := infra.NewConfigLoader(logger)
	config, err := configLoader.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Setup repository
	repo, err := infra.NewSQLiteEventRepository(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Initialize database schema (including migration from old databases)
	ctx := context.Background()
	if err := repo.Initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	// Create services
	logsService := app.NewLogsService(repo, repo)
	llm := infra.NewClaudeCodeLLMWithConfig(logger, config)
	analysisService := app.NewAnalysisService(repo, repo, logsService, llm, logger, config)

	// Set the session view factory using the claude_code plugin
	analysisService.SetSessionViewFactory(func(sessionID string, events []pluginsdk.Event) pluginsdk.AnalysisView {
		return claude_code.NewSessionView(sessionID, events)
	})

	// Create setup service
	setupService := app.NewSetupService(repo, logger)

	// Create config loader (reuse existing logger)
	configLoaderForPlugin := infra.NewConfigLoader(logger)

	// Create plugin registry
	registry := app.NewPluginRegistry(logger)

	// Create event bus for cross-plugin communication
	busRepo := infra.NewSQLiteEventBusRepositoryFromRepo(repo)
	eventBus := infra.NewInMemoryEventBus(busRepo)

	// Register built-in plugins
	workingDir, _ := os.Getwd()
	if err := RegisterBuiltInPlugins(registry, analysisService, logsService, logger, setupService, configLoaderForPlugin, *dbPath, workingDir, eventBus); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering built-in plugins: %v\n", err)
		os.Exit(1)
	}

	// Create event dispatcher for real-time event streaming
	pluginCtx := app.NewPluginContext(logger, *dbPath, "", repo)
	eventDispatcher := app.NewEventDispatcher(repo, logger, pluginCtx)

	// Run TUI
	if err := tui.Run(ctx, registry, analysisService, logsService, config, eventDispatcher); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}
}
