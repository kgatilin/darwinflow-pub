package task_manager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// Ensure plugin implements required SDK interfaces
var (
	_ pluginsdk.Plugin           = (*TaskManagerPlugin)(nil)
	_ pluginsdk.IEntityProvider  = (*TaskManagerPlugin)(nil)
	_ pluginsdk.ICommandProvider = (*TaskManagerPlugin)(nil)
	_ pluginsdk.IEventEmitter    = (*TaskManagerPlugin)(nil)
)

// TaskManagerPlugin provides task management with real-time file watching.
// It implements Plugin, IEntityProvider, ICommandProvider, and IEventEmitter interfaces.
// It can optionally use a RoadmapRepository for hierarchical roadmap storage.
type TaskManagerPlugin struct {
	logger      pluginsdk.Logger
	workingDir  string
	tasksDir    string
	fileWatcher *FileWatcher
	eventBus    pluginsdk.EventBus
	// Optional: Database repository for hierarchical roadmap model
	repository RoadmapRepository
}

// NewTaskManagerPlugin creates a new task manager plugin with file-based storage
// eventBus is passed as interface{} to allow cmd package to avoid importing pluginsdk.
func NewTaskManagerPlugin(logger pluginsdk.Logger, workingDir string, eventBus interface{}) (*TaskManagerPlugin, error) {
	tasksDir := filepath.Join(workingDir, ".darwinflow", "tasks")

	fileWatcher, err := NewFileWatcher(logger, tasksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Type assert eventBus to pluginsdk.EventBus
	var eb pluginsdk.EventBus
	if eventBus != nil {
		if bus, ok := eventBus.(pluginsdk.EventBus); ok {
			eb = bus
		}
	}

	return &TaskManagerPlugin{
		logger:      logger,
		workingDir:  workingDir,
		tasksDir:    tasksDir,
		fileWatcher: fileWatcher,
		eventBus:    eb,
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
	if err := InitSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Migrate existing file-based tasks if any exist
	if err := MigrateFromFileStorage(db, tasksDir); err != nil {
		return nil, fmt.Errorf("failed to migrate tasks from file storage: %w", err)
	}

	fileWatcher, err := NewFileWatcher(logger, tasksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Type assert eventBus to pluginsdk.EventBus
	var eb pluginsdk.EventBus
	if eventBus != nil {
		if bus, ok := eventBus.(pluginsdk.EventBus); ok {
			eb = bus
		}
	}

	// Create base repository and wrap with event emission
	baseRepository := NewSQLiteRoadmapRepository(db, logger)
	var repository RoadmapRepository = baseRepository

	// Wrap with event-emitting decorator if eventBus is available
	if eb != nil {
		repository = NewEventEmittingRepository(baseRepository, eb, logger)
	}

	plugin := &TaskManagerPlugin{
		logger:      logger,
		workingDir:  workingDir,
		tasksDir:    tasksDir,
		fileWatcher: fileWatcher,
		eventBus:    eb,
		repository:  repository,
	}

	// Migrate from old database location to project-based structure
	if err := plugin.migrateToProjects(); err != nil {
		return nil, fmt.Errorf("failed to migrate to projects: %w", err)
	}

	return plugin, nil
}

// GetRepository returns the optional RoadmapRepository for database operations
// Returns nil if the plugin was initialized without database support
func (p *TaskManagerPlugin) GetRepository() RoadmapRepository {
	return p.repository
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
func (p *TaskManagerPlugin) Query(ctx context.Context, query pluginsdk.EntityQuery) ([]pluginsdk.IExtensible, error) {
	// Ensure tasks directory exists
	if err := os.MkdirAll(p.tasksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}

	// Read all task files
	entries, err := os.ReadDir(p.tasksDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	entities := make([]pluginsdk.IExtensible, 0)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(p.tasksDir, entry.Name())
		task, err := p.loadTaskFromFile(filePath)
		if err != nil {
			p.logger.Warn("failed to load task", "path", filePath, "error", err)
			continue
		}

		// Apply filters if specified
		if !p.matchesFilters(task, query.Filters) {
			continue
		}

		entities = append(entities, task)
	}

	// Apply offset and limit
	if query.Offset > 0 {
		if query.Offset >= len(entities) {
			return []pluginsdk.IExtensible{}, nil
		}
		entities = entities[query.Offset:]
	}

	if query.Limit > 0 && len(entities) > query.Limit {
		entities = entities[:query.Limit]
	}

	return entities, nil
}

// GetEntity retrieves a single entity by ID (SDK interface)
func (p *TaskManagerPlugin) GetEntity(ctx context.Context, entityID string) (pluginsdk.IExtensible, error) {
	// Try exact match first
	filePath := filepath.Join(p.tasksDir, entityID+".json")
	task, err := p.loadTaskFromFile(filePath)
	if err != nil {
		// Try prefix match (abbreviated ID)
		matchedFile, matchErr := p.findTaskByPrefix(entityID)
		if matchErr != nil {
			return nil, pluginsdk.ErrNotFound
		}
		task, err = p.loadTaskFromFile(matchedFile)
		if err != nil {
			return nil, pluginsdk.ErrNotFound
		}
	}
	return task, nil
}

// UpdateEntity updates an entity's fields (SDK interface)
func (p *TaskManagerPlugin) UpdateEntity(ctx context.Context, entityID string, fields map[string]interface{}) (pluginsdk.IExtensible, error) {
	// Try exact match first
	filePath := filepath.Join(p.tasksDir, entityID+".json")
	task, err := p.loadTaskFromFile(filePath)
	if err != nil {
		// Try prefix match (abbreviated ID)
		matchedFile, matchErr := p.findTaskByPrefix(entityID)
		if matchErr != nil {
			// Return the specific error from findTaskByPrefix
			return nil, fmt.Errorf("%w: %v", pluginsdk.ErrNotFound, matchErr)
		}
		filePath = matchedFile
		task, err = p.loadTaskFromFile(filePath)
		if err != nil {
			return nil, pluginsdk.ErrNotFound
		}
	}

	// Update fields
	if title, ok := fields["title"]; ok {
		if titleStr, ok := title.(string); ok {
			task.Title = titleStr
		}
	}
	if description, ok := fields["description"]; ok {
		if descStr, ok := description.(string); ok {
			task.Description = descStr
		}
	}
	if status, ok := fields["status"]; ok {
		if statusStr, ok := status.(string); ok {
			task.Status = statusStr
		}
	}
	if priority, ok := fields["priority"]; ok {
		if priorityStr, ok := priority.(string); ok {
			task.Priority = priorityStr
		}
	}

	// Save updated task
	if err := p.saveTaskToFile(filePath, task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}

// GetCommands returns the CLI commands provided by this plugin (SDK interface)
func (p *TaskManagerPlugin) GetCommands() []pluginsdk.Command {
	return []pluginsdk.Command{
		&InitCommand{plugin: p},
		&CreateCommand{plugin: p},
		&ListCommand{plugin: p},
		&UpdateCommand{plugin: p},
		// Project commands
		&ProjectCreateCommand{Plugin: p},
		&ProjectListCommand{Plugin: p},
		&ProjectSwitchCommand{Plugin: p},
		&ProjectShowCommand{Plugin: p},
		&ProjectDeleteCommand{Plugin: p},
		// Roadmap commands
		&RoadmapInitCommand{Plugin: p},
		&RoadmapShowCommand{Plugin: p},
		&RoadmapUpdateCommand{Plugin: p},
		// Track commands
		&TrackCreateCommand{Plugin: p},
		&TrackListCommand{Plugin: p},
		&TrackShowCommand{Plugin: p},
		&TrackUpdateCommand{Plugin: p},
		&TrackDeleteCommand{Plugin: p},
		&TrackAddDependencyCommand{Plugin: p},
		&TrackRemoveDependencyCommand{Plugin: p},
		// Task commands (database-backed hierarchical model)
		&TaskCreateCommand{Plugin: p},
		&TaskListCommand{Plugin: p},
		&TaskShowCommand{Plugin: p},
		&TaskUpdateCommand{Plugin: p},
		&TaskDeleteCommand{Plugin: p},
		&TaskMoveCommand{Plugin: p},
		&TaskMigrateCommand{Plugin: p},
		// Iteration commands
		&IterationCreateCommand{Plugin: p},
		&IterationListCommand{Plugin: p},
		&IterationShowCommand{Plugin: p},
		&IterationCurrentCommand{Plugin: p},
		&IterationUpdateCommand{Plugin: p},
		&IterationDeleteCommand{Plugin: p},
		&IterationAddTaskCommand{Plugin: p},
		&IterationRemoveTaskCommand{Plugin: p},
		&IterationStartCommand{Plugin: p},
		&IterationCompleteCommand{Plugin: p},
		// TUI command
		&TUICommand{Plugin: p},
	}
}

// StartEventStream begins streaming events to the provided channel (SDK interface)
func (p *TaskManagerPlugin) StartEventStream(ctx context.Context, eventChan chan<- pluginsdk.Event) error {
	p.logger.Info("starting event stream for task-manager plugin")
	return p.fileWatcher.Start(ctx, eventChan)
}

// StopEventStream stops the event stream (SDK interface)
func (p *TaskManagerPlugin) StopEventStream() error {
	p.logger.Info("stopping event stream for task-manager plugin")
	return p.fileWatcher.Stop()
}

// Helper methods

// loadTaskFromFile loads a task from a JSON file
func (p *TaskManagerPlugin) loadTaskFromFile(filePath string) (*TaskEntity, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var task TaskEntity
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// saveTaskToFile saves a task to a JSON file
func (p *TaskManagerPlugin) saveTaskToFile(filePath string, task *TaskEntity) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// findTaskByPrefix finds a task file that matches the given ID prefix
// Returns the full file path if exactly one match is found
// Returns error if no matches or multiple matches (ambiguous)
func (p *TaskManagerPlugin) findTaskByPrefix(prefix string) (string, error) {
	entries, err := os.ReadDir(p.tasksDir)
	if err != nil {
		return "", err
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		// Extract ID from filename (remove .json extension)
		if !strings.HasSuffix(filename, ".json") {
			continue
		}
		taskID := strings.TrimSuffix(filename, ".json")

		// Check if ID starts with the prefix
		if strings.HasPrefix(taskID, prefix) {
			matches = append(matches, filepath.Join(p.tasksDir, filename))
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no tasks found matching prefix: %s", prefix)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous task ID prefix: %s matches %d tasks", prefix, len(matches))
	}

	return matches[0], nil
}

// matchesFilters checks if an entity matches the given filters
func (p *TaskManagerPlugin) matchesFilters(entity pluginsdk.IExtensible, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}

	for key, expectedValue := range filters {
		actualValue := entity.GetField(key)
		if actualValue != expectedValue {
			return false
		}
	}

	return true
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
	if err := InitSchema(db); err != nil {
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

// getRepositoryForProject returns a repository for the specified project.
// If projectName is empty, uses the active project.
// Returns the repository and a cleanup function (close DB).
func (p *TaskManagerPlugin) getRepositoryForProject(projectName string) (RoadmapRepository, func(), error) {
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

	// Create repository for this project
	baseRepo := NewSQLiteRoadmapRepository(db, p.logger)
	var repo RoadmapRepository = baseRepo

	// Wrap with event-emitting decorator if eventBus is available
	if p.eventBus != nil {
		repo = NewEventEmittingRepository(baseRepo, p.eventBus, p.logger)
	}

	// Return repository and cleanup function
	cleanup := func() {
		db.Close()
	}

	return repo, cleanup, nil
}
