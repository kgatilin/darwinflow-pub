package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/app/tui"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
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

	// Create services
	logsService := app.NewLogsService(repo, repo)
	llmExecutor := app.NewClaudeCLIExecutorWithConfig(logger, config)
	analysisService := app.NewAnalysisService(repo, repo, logsService, llmExecutor, logger, config)

	// Create plugin registry
	registry := app.NewPluginRegistry(logger)

	// Register built-in plugins
	// Note: setupService and handler are nil because UI doesn't use command execution
	if err := RegisterBuiltInPlugins(registry, analysisService, logsService, logger, nil, nil, *dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering built-in plugins: %v\n", err)
		os.Exit(1)
	}

	// Run TUI
	ctx := context.Background()
	if err := tui.Run(ctx, registry, analysisService, logsService, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}
}
