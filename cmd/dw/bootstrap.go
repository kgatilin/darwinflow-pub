package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kgatilin/darwinflow-pub/internal/app"
	"github.com/kgatilin/darwinflow-pub/internal/infra"
)

// AppServices contains all app-layer services needed by commands.
// Note: This struct only uses app-layer types, no domain or plugin imports.
type AppServices struct {
	PluginRegistry  *app.PluginRegistry
	CommandRegistry *app.CommandRegistry
	LogsService     *app.LogsService
	AnalysisService *app.AnalysisService
	SetupService    *app.SetupService
	ConfigLoader    app.ConfigLoader
	Logger          app.Logger
	EventRepo       interface{} // EventRepository for plugin contexts (type from internal/domain)
	DBPath          string
	WorkingDir      string
}

// InitializeApp creates all infrastructure and app services
func InitializeApp(dbPath, configPath string, debugMode bool) (*AppServices, error) {
	// 1. Create logger
	var logger *infra.Logger
	if debugMode {
		logger = infra.NewDebugLogger()
	} else {
		logger = infra.NewDefaultLogger()
	}

	// 2. Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// 3. Create repository
	repo, err := infra.NewSQLiteEventRepository(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// 4. Load config (keep internally, cmd doesn't need it)
	configLoader := infra.NewConfigLoader(logger)
	config, err := configLoader.LoadConfig(configPath)
	if err != nil {
		// Non-fatal - load default config via config loader
		logger.Warn("Failed to load config, using defaults: %v", err)
		config, _ = configLoader.LoadConfig("") // Will return default
	}

	// 4b. Apply console log level from config (unless debugMode overrides)
	if !debugMode {
		logLevel, _ := infra.ParseLogLevel(config.Logging.ConsoleLogLevel)
		logger.SetLevel(logLevel)
	}

	// 5. Create app services
	logsService := app.NewLogsService(repo, repo)
	llmExecutor := app.NewClaudeCLIExecutorWithConfig(logger, config)
	analysisService := app.NewAnalysisService(repo, repo, logsService, llmExecutor, logger, config)

	// 6. Create setup service (for framework-level initialization)
	// SetupService handles framework infrastructure only (database, schema, etc.)
	// Plugin-specific setup (hooks, etc.) is handled by plugin init commands
	setupService := app.NewSetupService(repo, logger)

	// 7. Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	// 8. Create plugin registry
	pluginRegistry := app.NewPluginRegistry(logger)

	// 9. Register built-in plugins (cmd layer handles plugin imports)
	if err := RegisterBuiltInPlugins(
		pluginRegistry,
		analysisService,
		logsService,
		logger,
		setupService,
		configLoader,
		dbPath,
		workingDir,
	); err != nil {
		return nil, fmt.Errorf("failed to register built-in plugins: %w", err)
	}

	// 10. Load external plugins from .darwinflow/plugins.yaml
	// dbPath is .darwinflow/logs/events.db, so we need to go up two levels to .darwinflow
	pluginsConfigPath := filepath.Join(filepath.Dir(filepath.Dir(dbPath)), "plugins.yaml")
	if _, err := os.Stat(pluginsConfigPath); err == nil {
		loader := infra.NewPluginLoader(logger)
		externalPlugins, err := loader.LoadFromConfig(pluginsConfigPath)
		if err != nil {
			logger.Warn("Failed to load plugins from config: %v", err)
		} else {
			// Initialize and register each plugin
			ctx := context.Background()
			successCount := 0
			for _, plugin := range externalPlugins {
				// Initialize the plugin (required for SubprocessPlugin to get info)
				if initializer, ok := plugin.(interface {
					Initialize(context.Context, string, map[string]interface{}) error
				}); ok {
					if err := initializer.Initialize(ctx, workingDir, nil); err != nil {
						logger.Warn("Failed to initialize external plugin: %v", err)
						continue
					}
				}

				// Register the plugin
				if err := pluginRegistry.RegisterPlugin(plugin); err != nil {
					logger.Warn("Failed to register external plugin: %v", err)
				} else {
					successCount++
				}
			}
			if successCount > 0 {
				logger.Info("Loaded %d external plugin(s) from config", successCount)
			}
		}
	}

	// 11. Create command registry
	commandRegistry := app.NewCommandRegistry(pluginRegistry, logger)

	return &AppServices{
		PluginRegistry:  pluginRegistry,
		CommandRegistry: commandRegistry,
		LogsService:     logsService,
		AnalysisService: analysisService,
		SetupService:    setupService,
		ConfigLoader:    configLoader,
		Logger:          logger,
		EventRepo:       repo,
		DBPath:          dbPath,
		WorkingDir:      workingDir,
	}, nil
}
