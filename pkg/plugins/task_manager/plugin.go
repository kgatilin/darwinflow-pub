package task_manager

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/services"
	infracli "github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/infrastructure/cli"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/infrastructure/persistence"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/cli"
	presentationTui "github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/presentation/tui"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// Ensure plugin implements required SDK interfaces
var (
	_ pluginsdk.Plugin           = (*TaskManagerPlugin)(nil)
	_ pluginsdk.IEntityProvider  = (*TaskManagerPlugin)(nil)
	_ pluginsdk.ICommandProvider = (*TaskManagerPlugin)(nil)
	_ pluginsdk.IEventEmitter    = (*TaskManagerPlugin)(nil)
	_ infracli.PluginProvider    = (*TaskManagerPlugin)(nil) // Infrastructure CLI provider
)

// TaskManagerPlugin provides task management with SQLite database storage.
// It implements Plugin, IEntityProvider, ICommandProvider, and IEventEmitter interfaces.
// Events are emitted by the EventEmittingRepository decorator (not FileWatcher).
type TaskManagerPlugin struct {
	logger     pluginsdk.Logger
	workingDir string
	tasksDir   string
	eventBus   pluginsdk.EventBus
	// Optional: Database repository for hierarchical roadmap model
	repository domain.RoadmapRepository
	// Configuration for plugin behavior
	config *Config
}

// NewTaskManagerPlugin creates a new task manager plugin with file-based storage
// eventBus is passed as interface{} to allow cmd package to avoid importing pluginsdk.
func NewTaskManagerPlugin(logger pluginsdk.Logger, workingDir string, eventBus interface{}) (*TaskManagerPlugin, error) {
	tasksDir := filepath.Join(workingDir, ".darwinflow", "tasks")

	// Type assert eventBus to pluginsdk.EventBus
	var eb pluginsdk.EventBus
	if eventBus != nil {
		if bus, ok := eventBus.(pluginsdk.EventBus); ok {
			eb = bus
		}
	}

	return &TaskManagerPlugin{
		logger:     logger,
		workingDir: workingDir,
		tasksDir:   tasksDir,
		eventBus:   eb,
	}, nil
}

// NewTaskManagerPluginWithDatabase creates a new task manager plugin with SQLite database support
// for hierarchical roadmap storage. The database connection is owned by the caller.
// eventBus is passed as interface{} to allow cmd package to avoid importing pluginsdk.
func NewTaskManagerPluginWithDatabase(
	logger pluginsdk.Logger,
	workingDir string,
	db *sql.DB,
	eventBus interface{},
) (*TaskManagerPlugin, error) {
	tasksDir := filepath.Join(workingDir, ".darwinflow", "tasks")

	// Initialize database schema
	if err := persistence.InitSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Type assert eventBus to pluginsdk.EventBus
	var eb pluginsdk.EventBus
	if eventBus != nil {
		if bus, ok := eventBus.(pluginsdk.EventBus); ok {
			eb = bus
		}
	}

	// Create base repository and wrap with event emission
	baseRepository := persistence.NewSQLiteRoadmapRepository(db, logger)
	var repository domain.RoadmapRepository = baseRepository

	// Wrap with event-emitting decorator if eventBus is available
	if eb != nil {
		repository = persistence.NewEventEmittingRepository(baseRepository, eb, logger)
	}

	// Load configuration
	config, err := LoadConfig(workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	plugin := &TaskManagerPlugin{
		logger:     logger,
		workingDir: workingDir,
		tasksDir:   tasksDir,
		eventBus:   eb,
		repository: repository,
		config:     config,
	}

	// Migrate from old database location to project-based structure
	if err := plugin.migrateToProjects(); err != nil {
		return nil, fmt.Errorf("failed to migrate to projects: %w", err)
	}

	return plugin, nil
}

// GetRepository returns the optional RoadmapRepository for database operations
// Returns nil if the plugin was initialized without database support
func (p *TaskManagerPlugin) GetRepository() domain.RoadmapRepository {
	return p.repository
}

// GetConfig returns the plugin configuration
func (p *TaskManagerPlugin) GetConfig() *Config {
	if p.config == nil {
		return DefaultConfig()
	}
	return p.config
}

// GetInfo returns metadata about this plugin (SDK interface)
func (p *TaskManagerPlugin) GetInfo() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:        "task-manager",
		Version:     "1.0.0",
		Description: "Task tracking with real-time file watching",
		IsCore:      false,
	}
}

// GetCapabilities returns the capability interfaces this plugin implements (SDK interface)
func (p *TaskManagerPlugin) GetCapabilities() []string {
	return []string{"IEntityProvider", "ICommandProvider", "IEventEmitter"}
}

// GetEntityTypes returns the entity types this plugin provides (SDK interface)
func (p *TaskManagerPlugin) GetEntityTypes() []pluginsdk.EntityTypeInfo {
	return []pluginsdk.EntityTypeInfo{
		{
			Type:              "task",
			DisplayName:       "Task",
			DisplayNamePlural: "Tasks",
			Capabilities:      []string{"IExtensible", "ITrackable"},
			Icon:              "✓",
			Description:       "Task with status tracking",
		},
	}
}

// Query returns entities matching the given query (SDK interface)
// Note: This method is deprecated in favor of using CLI commands with the database repository.
// File-based storage has been replaced with SQLite database.
func (p *TaskManagerPlugin) Query(ctx context.Context, query pluginsdk.EntityQuery) ([]pluginsdk.IExtensible, error) {
	return nil, fmt.Errorf("Query method is deprecated; use CLI commands with database repository")
}

// GetEntity retrieves a single entity by ID (SDK interface)
// Note: This method is deprecated in favor of using CLI commands with the database repository.
// File-based storage has been replaced with SQLite database.
func (p *TaskManagerPlugin) GetEntity(ctx context.Context, entityID string) (pluginsdk.IExtensible, error) {
	return nil, fmt.Errorf("GetEntity method is deprecated; use CLI commands with database repository")
}

// UpdateEntity updates an entity's fields (SDK interface)
// Note: This method is deprecated in favor of using CLI commands with the database repository.
// File-based storage has been replaced with SQLite database.
func (p *TaskManagerPlugin) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (pluginsdk.IExtensible, error) {
	return nil, fmt.Errorf("UpdateEntity method is deprecated; use CLI commands with database repository")
}

// GetCommands returns the CLI commands provided by this plugin (SDK interface)
func (p *TaskManagerPlugin) GetCommands() []pluginsdk.Command {
	// Create application services with injected dependencies
	// Note: Services are created per GetCommands() call. Each adapter receives services
	// configured to work with the RepoProvider (plugin) for per-project repository access.

	// Get repository for service initialization (using active project)
	// Services don't hold repository instances directly; adapters fetch repositories
	// per-command execution via RepoProvider.GetRepositoryForProject()
	repo, _, err := p.GetRepositoryForProject("")
	if err != nil {
		p.logger.Warn("failed to get repository for service initialization", "error", err)
		// Return commands without services (will fail if executed, but plugin loads)
		return p.getCommandsWithoutServices()
	}
	// NOTE: We do NOT call cleanup() here because the application services and repositories
	// need to remain open for the lifetime of the application. The database connection
	// will be closed when the application exits.

	// Unwrap repository if it's wrapped in EventEmittingRepository decorator
	var composite *persistence.SQLiteRepositoryComposite
	if eventRepo, ok := repo.(*persistence.EventEmittingRepository); ok {
		// Unwrap to get the underlying composite repository
		composite, ok = eventRepo.Repo.(*persistence.SQLiteRepositoryComposite)
		if !ok {
			p.logger.Warn("wrapped repository is not SQLiteRepositoryComposite")
			return p.getCommandsWithoutServices()
		}
	} else {
		// Not wrapped, try direct cast
		composite, ok = repo.(*persistence.SQLiteRepositoryComposite)
		if !ok {
			p.logger.Warn("repository is not SQLiteRepositoryComposite")
			return p.getCommandsWithoutServices()
		}
	}

	// Initialize domain services (stateless, no dependencies)
	validationSvc := services.NewValidationService()
	iterationSvc := services.NewIterationService()

	// Initialize application services
	trackService := application.NewTrackApplicationService(
		composite.Track,
		composite.Roadmap,
		composite.Aggregate,
		validationSvc,
	)

	taskService := application.NewTaskApplicationService(
		composite.Task,
		composite.Track,
		composite.Aggregate,
		composite.AC,
		validationSvc,
	)

	iterationService := application.NewIterationApplicationService(
		composite.Iteration,
		composite.Task,
		composite.Aggregate,
		iterationSvc,
		validationSvc,
	)

	adrService := application.NewADRApplicationService(
		composite.ADR,
		composite.Track,
		composite.Aggregate,
		validationSvc,
	)

	acService := application.NewACApplicationService(
		composite.AC,
		composite.Task,
		composite.Aggregate,
		validationSvc,
	)

	roadmapService := application.NewRoadmapApplicationService(
		composite.Roadmap,
		composite.Track,
		composite.Task,
		composite.Iteration,
		validationSvc,
	)

	documentService := application.NewDocumentApplicationService(
		composite.Document,
		composite.Track,
		composite.Iteration,
	)

	return []pluginsdk.Command{
		// Project commands (infrastructure layer)
		&infracli.ProjectCreateCommand{Provider: p},
		&infracli.ProjectListCommand{Provider: p},
		&infracli.ProjectSwitchCommand{Provider: p},
		&infracli.ProjectShowCommand{Provider: p},
		&infracli.ProjectDeleteCommand{Provider: p},
		// Roadmap commands (migrated to CLI adapters)
		&cli.RoadmapInitCommandAdapter{RoadmapService: roadmapService},
		&cli.RoadmapShowCommandAdapter{RoadmapService: roadmapService},
		&cli.RoadmapUpdateCommandAdapter{RoadmapService: roadmapService},
		&cli.RoadmapFullCommandAdapter{RoadmapService: roadmapService},
		// ========================================================================
		// MIGRATED TO CLI ADAPTERS (using application layer services)
		// ========================================================================
		// Track commands
		&cli.TrackCreateCommandAdapter{
			TrackService: trackService,
		},
		&cli.TrackUpdateCommandAdapter{
			TrackService: trackService,
		},
		// Task commands
		&cli.TaskCreateCommandAdapter{
			TaskService: taskService,
		},
		&cli.TaskUpdateCommandAdapter{
			TaskService: taskService,
		},
		&cli.TaskDeleteCommandAdapter{
			TaskService: taskService,
		},
		// Iteration commands
		&cli.IterationCreateCommandAdapter{
			IterationService: iterationService,
		},
		&cli.IterationUpdateCommandAdapter{
			IterationService: iterationService,
		},
		&cli.IterationStartCommandAdapter{
			IterationService: iterationService,
		},
		&cli.IterationCompleteCommandAdapter{
			IterationService: iterationService,
		},
		&cli.IterationListCommandAdapter{
			IterationService: iterationService,
		},
		&cli.IterationShowCommandAdapter{
			IterationService:    iterationService,
			DocumentService:     documentService,
		},
		&cli.IterationCurrentCommandAdapter{
			IterationService: iterationService,
			DocumentService:  documentService,
		},
		&cli.IterationDeleteCommandAdapter{
			IterationService: iterationService,
		},
		&cli.IterationAddTaskCommandAdapter{
			IterationService: iterationService,
		},
		&cli.IterationRemoveTaskCommandAdapter{
			IterationService: iterationService,
		},
		&cli.IterationViewCommandAdapter{
			IterationService: iterationService,
		},
		// ADR commands
		&cli.ADRCreateCommandAdapter{
			ADRService: adrService,
		},
		&cli.ADRUpdateCommandAdapter{
			ADRService: adrService,
		},
		// AC commands
		&cli.ACAddCommandAdapter{
			ACService: acService,
		},
		&cli.ACVerifyCommandAdapter{
			ACService: acService,
		},
		&cli.ACFailCommandAdapter{
			ACService: acService,
		},
		&cli.ACSkipCommandAdapter{
			ACService: acService,
		},
		&cli.ACListCommandAdapter{
			ACService: acService,
		},
		&cli.ACShowCommandAdapter{
			ACService: acService,
		},
		&cli.ACUpdateCommandAdapter{
			ACService: acService,
		},
		&cli.ACDeleteCommandAdapter{
			ACService: acService,
		},
		&cli.ACVerifyAutoCommandAdapter{
			ACService: acService,
		},
		&cli.ACRequestReviewCommandAdapter{
			ACService: acService,
		},
		&cli.ACListIterationCommandAdapter{
			ACService: acService,
		},
		&cli.ACListTrackCommandAdapter{
			ACService:   acService,
			TaskService: taskService,
		},
		&cli.ACFailedCommandAdapter{
			ACService: acService,
		},
		// Document commands
		&cli.DocCreateCommandAdapter{
			DocumentService: documentService,
		},
		&cli.DocUpdateCommandAdapter{
			DocumentService: documentService,
		},
		&cli.DocShowCommandAdapter{
			DocumentService: documentService,
		},
		&cli.DocListCommandAdapter{
			DocumentService: documentService,
		},
		&cli.DocAttachCommandAdapter{
			DocumentService: documentService,
		},
		&cli.DocDetachCommandAdapter{
			DocumentService: documentService,
		},
		&cli.DocDeleteCommandAdapter{
			DocumentService: documentService,
		},
		&cli.DocHelpCommandAdapter{},
		// Task commands (query/list operations)
		&cli.TaskListCommandAdapter{
			TaskService: taskService,
		},
		&cli.TaskShowCommandAdapter{
			TaskService: taskService,
		},
		&cli.TaskMoveCommandAdapter{
			TaskService: taskService,
		},
		&cli.TaskBacklogCommandAdapter{
			TaskService: taskService,
		},
		&cli.TaskCheckReadyCommandAdapter{
			TaskService: taskService,
			ACService:   acService,
		},
		&cli.TaskMigrateCommandAdapter{},

		// ========================================================================
		// NOT MIGRATED YET (still using old command pattern)
		// ========================================================================
		// Track commands (query/list operations) - MIGRATED TO CLI ADAPTERS
		&cli.TrackListCommandAdapter{
			TrackService: trackService,
		},
		&cli.TrackShowCommandAdapter{
			TrackService:    trackService,
			DocumentService: documentService,
		},
		&cli.TrackDeleteCommandAdapter{
			TrackService: trackService,
		},
		&cli.TrackAddDependencyCommandAdapter{
			TrackService: trackService,
		},
		&cli.TrackRemoveDependencyCommandAdapter{
			TrackService: trackService,
		},

		// ========================================================================
		// INFRASTRUCTURE COMMANDS (not migrated, appropriately structured)
		// ========================================================================
		// TUI commands (new MVP implementation)
		&presentationTui.TUINewCommand{Plugin: p},
		// Prompt command (presentation layer)
		&cli.PromptCommand{GetPrompt: cli.GetSystemPrompt},
		// Backup commands (infrastructure layer)
		&infracli.BackupCommand{Provider: p},
		&infracli.RestoreCommand{Provider: p},
		&infracli.BackupListCommand{Provider: p},
	}
}

// getCommandsWithoutServices returns commands when service initialization fails
// This allows the plugin to load even if repository access fails temporarily
func (p *TaskManagerPlugin) getCommandsWithoutServices() []pluginsdk.Command {
	return []pluginsdk.Command{
		// Project commands (infrastructure layer)
		&infracli.ProjectCreateCommand{Provider: p},
		&infracli.ProjectListCommand{Provider: p},
		&infracli.ProjectSwitchCommand{Provider: p},
		&infracli.ProjectShowCommand{Provider: p},
		&infracli.ProjectDeleteCommand{Provider: p},

		// Note: CLI adapters that require services are omitted here (including roadmap commands)
		// This function is only called when service initialization fails
		// Commands will fail gracefully if executed without services

		// ========================================================================
		// INFRASTRUCTURE COMMANDS (not migrated, appropriately structured)
		// ========================================================================
		// TUI commands (new MVP implementation)
		&presentationTui.TUINewCommand{Plugin: p},
		// Prompt command (presentation layer)
		&cli.PromptCommand{GetPrompt: cli.GetSystemPrompt},
		// Backup commands (infrastructure layer)
		&infracli.BackupCommand{Provider: p},
		&infracli.RestoreCommand{Provider: p},
		&infracli.BackupListCommand{Provider: p},
	}
}

// StartEventStream begins streaming events to the provided channel (SDK interface)
// This is a no-op because events are already emitted by the EventEmittingRepository decorator
// (see lines 96, 828 in plugin.go where repositories are wrapped with event emission).
func (p *TaskManagerPlugin) StartEventStream(ctx context.Context, eventChan chan<- pluginsdk.Event) error {
	p.logger.Info("task-manager event stream is handled by EventEmittingRepository decorator")
	return nil
}

// StopEventStream stops the event stream (SDK interface)
// This is a no-op because events are handled by repository decorator, not FileWatcher.
func (p *TaskManagerPlugin) StopEventStream() error {
	p.logger.Info("task-manager event stream stop (no-op)")
	return nil
}

// ============================================================================
// Project Management Methods
// ============================================================================

// getProjectDatabase returns a database connection for the specified project.
// It creates the project directory and initializes the schema if needed.
func (p *TaskManagerPlugin) getProjectDatabase(projectName string) (*sql.DB, error) {
	// Get project-specific database path
	projectDir := filepath.Join(p.workingDir, ".darwinflow", "projects", projectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	dbPath := filepath.Join(projectDir, "roadmap.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize schema
	if err := persistence.InitSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// getActiveProject returns the name of the active project.
// Returns "default" if no active project is set.
func (p *TaskManagerPlugin) getActiveProject() (string, error) {
	activeProjectFile := filepath.Join(p.workingDir, ".darwinflow", "active-project.txt")

	// Read active project file
	data, err := os.ReadFile(activeProjectFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Default to "default" project
			return "default", nil
		}
		return "", fmt.Errorf("failed to read active project file: %w", err)
	}

	projectName := strings.TrimSpace(string(data))
	if projectName == "" {
		return "default", nil
	}

	return projectName, nil
}

// setActiveProject sets the active project.
func (p *TaskManagerPlugin) setActiveProject(name string) error {
	activeProjectFile := filepath.Join(p.workingDir, ".darwinflow", "active-project.txt")

	// Ensure .darwinflow directory exists
	if err := os.MkdirAll(filepath.Dir(activeProjectFile), 0755); err != nil {
		return fmt.Errorf("failed to create .darwinflow directory: %w", err)
	}

	// Write active project name
	if err := os.WriteFile(activeProjectFile, []byte(name), 0644); err != nil {
		return fmt.Errorf("failed to write active project file: %w", err)
	}

	return nil
}

// migrateToProjects migrates the old database location to the new project-based structure.
// If .darwinflow/darwinflow.db exists, it moves it to .darwinflow/projects/default/roadmap.db
func (p *TaskManagerPlugin) migrateToProjects() error {
	oldDB := filepath.Join(p.workingDir, ".darwinflow", "darwinflow.db")

	// Check if old database exists
	if _, err := os.Stat(oldDB); err != nil {
		// Old DB doesn't exist, no migration needed
		return nil
	}

	p.logger.Info("migrating database to project-based structure")

	// Create default project directory
	defaultProjectDir := filepath.Join(p.workingDir, ".darwinflow", "projects", "default")
	if err := os.MkdirAll(defaultProjectDir, 0755); err != nil {
		return fmt.Errorf("failed to create default project directory: %w", err)
	}

	// Move old database to default project
	newDB := filepath.Join(defaultProjectDir, "roadmap.db")

	// Check if new database already exists (migration already done)
	if _, err := os.Stat(newDB); err == nil {
		// Migration already done, just remove old database
		p.logger.Info("migration already completed, removing old database")
		if err := os.Remove(oldDB); err != nil {
			p.logger.Warn("failed to remove old database", "error", err)
		}
		return nil
	}

	// Move database file
	if err := os.Rename(oldDB, newDB); err != nil {
		return fmt.Errorf("failed to move database: %w", err)
	}

	p.logger.Info("migrated database to default project", "path", newDB)

	// Set default as active project
	if err := p.setActiveProject("default"); err != nil {
		return fmt.Errorf("failed to set default project: %w", err)
	}

	return nil
}

// GetRepositoryForProject returns a repository for the specified project.
// If projectName is empty, uses the active project.
// Returns the repository and a cleanup function (close DB).
func (p *TaskManagerPlugin) GetRepositoryForProject(projectName string) (domain.RoadmapRepository, func(), error) {
	// Determine which project to use
	if projectName == "" {
		var err error
		projectName, err = p.getActiveProject()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get active project: %w", err)
		}
	}

	// Get project-specific database
	db, err := p.getProjectDatabase(projectName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get project database: %w", err)
	}

	// Create composite repository for this project (provides all focused repositories)
	composite := persistence.NewSQLiteRepositoryComposite(db, p.logger)
	var repo domain.RoadmapRepository = composite

	// Wrap with event-emitting decorator if eventBus is available
	if p.eventBus != nil {
		repo = persistence.NewEventEmittingRepository(composite, p.eventBus, p.logger)
	}

	// Return repository and cleanup function
	cleanup := func() {
		db.Close()
	}

	return repo, cleanup, nil
}

// ============================================================================
// PluginProvider Interface Implementation (for infrastructure CLI commands)
// ============================================================================

// GetWorkingDir returns the working directory
func (p *TaskManagerPlugin) GetWorkingDir() string {
	return p.workingDir
}

// GetLogger returns the plugin logger
func (p *TaskManagerPlugin) GetLogger() pluginsdk.Logger {
	return p.logger
}

// GetActiveProject returns the active project name (public wrapper)
func (p *TaskManagerPlugin) GetActiveProject() (string, error) {
	return p.getActiveProject()
}

// SetActiveProject sets the active project (public wrapper)
func (p *TaskManagerPlugin) SetActiveProject(projectName string) error {
	return p.setActiveProject(projectName)
}

// GetProjectDatabase returns a database connection for the specified project (public wrapper)
func (p *TaskManagerPlugin) GetProjectDatabase(projectName string) (*sql.DB, error) {
	return p.getProjectDatabase(projectName)
}
