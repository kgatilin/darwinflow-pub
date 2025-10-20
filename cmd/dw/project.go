package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/app/plugins/claude_code"
	"github.com/kgatilin/darwinflow-pub/internal/domain"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

func projectCommand(args []string) {
	// Parse global flags
	fs := flag.NewFlagSet("project", flag.ContinueOnError)
	dbPath := fs.String("db", app.DefaultDBPath, "Path to SQLite database")
	configPath := fs.String("config", "", "Path to config file (default: .darwinflow.yaml in current dir)")
	debugMode := fs.Bool("debug", false, "Enable debug logging")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			printProjectUsage()
			return
		}
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	// Get remaining args (tool name and tool args)
	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		printProjectUsage()
		return
	}

	toolName := remainingArgs[0]
	toolArgs := remainingArgs[1:]

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

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Create services
	logsService := app.NewLogsService(repo, repo)
	llmExecutor := app.NewClaudeCLIExecutorWithConfig(logger, config)
	analysisService := app.NewAnalysisService(repo, repo, logsService, llmExecutor, logger, config)

	// Create plugin registry
	pluginRegistry := app.NewPluginRegistry(logger)

	// Register claude-code core plugin
	claudeCodePlugin := claude_code.NewClaudeCodePlugin(analysisService, logsService, logger)
	if err := pluginRegistry.RegisterPlugin(claudeCodePlugin); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering claude-code plugin: %v\n", err)
		os.Exit(1)
	}

	// Create tool registry
	toolRegistry := app.NewToolRegistry(pluginRegistry, logger)

	// Create handler
	handler := app.NewProjectCommandHandler(toolRegistry, os.Stdout)

	// Handle special commands
	ctx := context.Background()
	if toolName == "list" || toolName == "--help" || toolName == "-h" {
		if err := handler.ListTools(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing tools: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create project context
	projectCtx := &domain.ProjectContext{
		EventRepo:    repo,
		AnalysisRepo: repo,
		Config:       config,
		CWD:          cwd,
		DBPath:       *dbPath,
	}

	// Execute tool
	if err := handler.ExecuteTool(ctx, toolName, toolArgs, projectCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing tool '%s': %v\n", toolName, err)
		os.Exit(1)
	}
}

func printProjectUsage() {
	fmt.Println("Usage: dw project [flags] <toolname> [tool-args...]")
	fmt.Println()
	fmt.Println("Run project-specific tools provided by plugins.")
	fmt.Println()
	fmt.Println("Special commands:")
	fmt.Println("  dw project list       List all available tools")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --db string          Path to SQLite database (default: ~/.darwinflow/darwinflow.db)")
	fmt.Println("  --config string      Path to config file (default: .darwinflow.yaml)")
	fmt.Println("  --debug              Enable debug logging")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dw project list")
	fmt.Println("  dw project session-summary --session-id abc123")
}
