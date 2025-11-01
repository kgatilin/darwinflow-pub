package task_manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// TaskCreateCommand creates a new task in a track
// ============================================================================

type TaskCreateCommand struct {
	Plugin      *TaskManagerPlugin
	trackID     string
	title       string
	description string
	priority    string
}

func (c *TaskCreateCommand) GetName() string {
	return "task.create"
}

func (c *TaskCreateCommand) GetDescription() string {
	return "Create a new task in a track"
}

func (c *TaskCreateCommand) GetUsage() string {
	return "dw task-manager task create --track <track-id> --title <title> [--description <desc>] [--priority <priority>]"
}

func (c *TaskCreateCommand) GetHelp() string {
	return `Creates a new task in the specified track.

A task represents a unit of work within a track. All tasks must belong
to an existing track in the active roadmap.

Flags:
  --track <track-id>        Track ID to create task in (required)
  --title <title>           Task title (required)
  --description <desc>      Task description (optional)
  --priority <priority>     Task priority (optional, default: medium)
                           Values: critical, high, medium, low

Examples:
  # Create a basic task
  dw task-manager task create \
    --track track-plugin-system \
    --title "Implement plugin registry"

  # Create with full details
  dw task-manager task create \
    --track track-plugin-system \
    --title "Implement plugin registry" \
    --description "Create registry to discover and load plugins" \
    --priority high

Notes:
  - Track must exist (create with 'dw task-manager track create')
  - Initial status is automatically set to 'todo'
  - Task ID is generated automatically`
}

func (c *TaskCreateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	c.priority = "medium" // default
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--track":
			if i+1 < len(args) {
				c.trackID = args[i+1]
				i++
			}
		case "--title":
			if i+1 < len(args) {
				c.title = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				c.description = args[i+1]
				i++
			}
		case "--priority":
			if i+1 < len(args) {
				c.priority = args[i+1]
				i++
			}
		}
	}

	// Validate required flags
	if c.trackID == "" || c.title == "" {
		return fmt.Errorf("--track and --title are required")
	}

	// Ensure repository exists
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("task manager not initialized with database support")
	}

	// Verify track exists
	track, err := repo.GetTrack(ctx, c.trackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("track not found: %s", c.trackID)
		}
		return fmt.Errorf("failed to verify track: %w", err)
	}

	// Generate task ID (timestamp-based)
	taskID := fmt.Sprintf("task-%d", time.Now().UnixNano())

	// Create task
	task := NewTaskEntity(
		taskID,
		c.trackID,
		c.title,
		c.description,
		"todo",         // initial status
		c.priority,
		"",             // no branch initially
		time.Now().UTC(),
		time.Now().UTC(),
	)

	// Save task
	if err := repo.SaveTask(ctx, task); err != nil {
		if errors.Is(err, pluginsdk.ErrAlreadyExists) {
			return fmt.Errorf("task already exists: %s", taskID)
		}
		return fmt.Errorf("failed to save task: %w", err)
	}

	// Output success message
	fmt.Fprintf(cmdCtx.GetStdout(), "Task created successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID:    %s\n", task.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Track: %s (%s)\n", track.Title, track.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Title: %s\n", task.Title)

	return nil
}

// ============================================================================
// TaskListCommand lists tasks with optional filtering
// ============================================================================

type TaskListCommand struct {
	Plugin   *TaskManagerPlugin
	track    string
	status   string
	priority string
}

func (c *TaskListCommand) GetName() string {
	return "task.list"
}

func (c *TaskListCommand) GetDescription() string {
	return "List tasks with optional filtering"
}

func (c *TaskListCommand) GetUsage() string {
	return "dw task-manager task list [--track <track-id>] [--status <status>] [--priority <priority>]"
}

func (c *TaskListCommand) GetHelp() string {
	return `Lists all tasks, optionally filtered by track, status, or priority.

Flags:
  --track <track-id>        Filter by track ID (optional)
  --status <status>         Filter by status (optional, comma-separated)
                           Values: todo, in-progress, done
  --priority <priority>     Filter by priority (optional, comma-separated)
                           Values: critical, high, medium, low

Examples:
  # List all tasks
  dw task-manager task list

  # List tasks in a specific track
  dw task-manager task list --track track-plugin-system

  # List tasks with specific status
  dw task-manager task list --status todo

  # List in-progress or done tasks
  dw task-manager task list --status in-progress,done

  # List critical and high priority tasks
  dw task-manager task list --priority critical,high

Notes:
  - Status and priority filters accept comma-separated values
  - All filters are optional and can be combined`
}

func (c *TaskListCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--track":
			if i+1 < len(args) {
				c.track = args[i+1]
				i++
			}
		case "--status":
			if i+1 < len(args) {
				c.status = args[i+1]
				i++
			}
		case "--priority":
			if i+1 < len(args) {
				c.priority = args[i+1]
				i++
			}
		}
	}

	// Ensure repository exists
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("task manager not initialized with database support")
	}

	// Build filters
	filters := TaskFilters{
		TrackID: c.track,
	}

	if c.status != "" {
		filters.Status = strings.Split(c.status, ",")
	}

	if c.priority != "" {
		filters.Priority = strings.Split(c.priority, ",")
	}

	// List tasks
	tasks, err := repo.ListTasks(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Display results
	stdout := cmdCtx.GetStdout()

	if len(tasks) == 0 {
		fmt.Fprintf(stdout, "No tasks found\n")
		return nil
	}

	// Print header
	fmt.Fprintf(stdout, "%-20s %-40s %-12s %-8s %-20s\n",
		"ID", "Title", "Status", "Priority", "Track")
	fmt.Fprintf(stdout, "%s %s %s %s %s\n",
		strings.Repeat("-", 20), strings.Repeat("-", 40),
		strings.Repeat("-", 12), strings.Repeat("-", 8), strings.Repeat("-", 20))

	// Print tasks
	for _, task := range tasks {
		// Get track info for display
		track, err := repo.GetTrack(ctx, task.TrackID)
		trackName := task.TrackID
		if err == nil {
			trackName = track.Title
		}

		// Abbreviate ID for display
		abbrevID := task.ID
		if len(abbrevID) > 20 {
			abbrevID = abbrevID[:17] + "..."
		}

		fmt.Fprintf(stdout, "%-20s %-40s %-12s %-8s %-20s\n",
			abbrevID, task.Title, task.Status, task.Priority, trackName)
	}

	fmt.Fprintf(stdout, "\nTotal: %d task(s)\n", len(tasks))
	return nil
}

// ============================================================================
// TaskShowCommand shows detailed task information
// ============================================================================

type TaskShowCommand struct {
	Plugin *TaskManagerPlugin
	taskID string
}

func (c *TaskShowCommand) GetName() string {
	return "task.show"
}

func (c *TaskShowCommand) GetDescription() string {
	return "Show detailed task information"
}

func (c *TaskShowCommand) GetUsage() string {
	return "dw task-manager task show <task-id>"
}

func (c *TaskShowCommand) GetHelp() string {
	return `Displays detailed information about a task.

Arguments:
  <task-id>    Task ID to display (required)

Examples:
  # Show a task by full ID
  dw task-manager task show task-1729604400000000000

  # Show a task by abbreviated ID (if unambiguous)
  dw task-manager task show task-16000

Notes:
  - Task ID can be a full ID or unambiguous prefix
  - Displays all task fields including track, status, and timestamps`
}

func (c *TaskShowCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	c.taskID = args[0]

	// Ensure repository exists
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("task manager not initialized with database support")
	}

	// Get task
	task, err := repo.GetTask(ctx, c.taskID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("task not found: %s", c.taskID)
		}
		return fmt.Errorf("failed to retrieve task: %w", err)
	}

	// Get track info
	track, err := repo.GetTrack(ctx, task.TrackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("parent track not found: %s", task.TrackID)
		}
		return fmt.Errorf("failed to retrieve track: %w", err)
	}

	// Display task details
	stdout := cmdCtx.GetStdout()
	fmt.Fprintf(stdout, "Task Details\n")
	fmt.Fprintf(stdout, "%s\n\n", strings.Repeat("=", 60))

	fmt.Fprintf(stdout, "ID:          %s\n", task.ID)
	fmt.Fprintf(stdout, "Title:       %s\n", task.Title)
	fmt.Fprintf(stdout, "Status:      %s\n", task.Status)
	fmt.Fprintf(stdout, "Priority:    %s\n", task.Priority)

	if task.Description != "" {
		fmt.Fprintf(stdout, "Description: %s\n", task.Description)
	}

	fmt.Fprintf(stdout, "\nTrack:       %s (%s)\n", track.Title, track.ID)

	if task.Branch != "" {
		fmt.Fprintf(stdout, "Branch:      %s\n", task.Branch)
	}

	fmt.Fprintf(stdout, "\nCreated:     %s\n", task.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(stdout, "Updated:     %s\n", task.UpdatedAt.Format(time.RFC3339))

	return nil
}

// ============================================================================
// TaskUpdateCommand updates an existing task
// ============================================================================

type TaskUpdateCommand struct {
	Plugin      *TaskManagerPlugin
	taskID      string
	title       string
	description string
	status      string
	priority    string
	branch      string
	hasUpdates  bool
}

func (c *TaskUpdateCommand) GetName() string {
	return "task.update"
}

func (c *TaskUpdateCommand) GetDescription() string {
	return "Update an existing task"
}

func (c *TaskUpdateCommand) GetUsage() string {
	return "dw task-manager task update <task-id> [--title <title>] [--description <desc>] [--status <status>] [--priority <priority>] [--branch <branch>]"
}

func (c *TaskUpdateCommand) GetHelp() string {
	return `Updates fields of an existing task.

Arguments:
  <task-id>    Task ID to update (required)

Flags:
  --title <title>           New task title (optional)
  --description <desc>      New task description (optional)
  --status <status>         New task status (optional)
                           Values: todo, in-progress, done
  --priority <priority>     New task priority (optional)
                           Values: critical, high, medium, low
  --branch <branch>         Git branch name (optional)

Examples:
  # Mark task as in-progress
  dw task-manager task update task-123 --status in-progress

  # Update multiple fields
  dw task-manager task update task-123 \
    --status done \
    --priority critical

  # Set a git branch
  dw task-manager task update task-123 --branch feat/new-feature

Notes:
  - At least one flag must be provided
  - Use --branch to associate a git branch
  - Use --status to change task progress`
}

func (c *TaskUpdateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	c.taskID = args[0]

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				c.title = args[i+1]
				c.hasUpdates = true
				i++
			}
		case "--description":
			if i+1 < len(args) {
				c.description = args[i+1]
				c.hasUpdates = true
				i++
			}
		case "--status":
			if i+1 < len(args) {
				c.status = args[i+1]
				c.hasUpdates = true
				i++
			}
		case "--priority":
			if i+1 < len(args) {
				c.priority = args[i+1]
				c.hasUpdates = true
				i++
			}
		case "--branch":
			if i+1 < len(args) {
				c.branch = args[i+1]
				c.hasUpdates = true
				i++
			}
		}
	}

	// Validate that at least one field was provided
	if !c.hasUpdates {
		return fmt.Errorf("at least one flag must be provided to update")
	}

	// Ensure repository exists
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("task manager not initialized with database support")
	}

	// Get existing task
	task, err := repo.GetTask(ctx, c.taskID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("task not found: %s", c.taskID)
		}
		return fmt.Errorf("failed to retrieve task: %w", err)
	}

	// Update fields
	if c.title != "" {
		task.Title = c.title
	}
	if c.description != "" {
		task.Description = c.description
	}
	if c.status != "" {
		task.Status = c.status
	}
	if c.priority != "" {
		task.Priority = c.priority
	}
	if c.branch != "" {
		task.Branch = c.branch
	}

	task.UpdatedAt = time.Now().UTC()

	// Save updated task
	if err := repo.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Output success message
	stdout := cmdCtx.GetStdout()
	fmt.Fprintf(stdout, "Task updated successfully\n")
	fmt.Fprintf(stdout, "  ID:    %s\n", task.ID)
	fmt.Fprintf(stdout, "  Title: %s\n", task.Title)
	fmt.Fprintf(stdout, "  Status: %s\n", task.Status)

	return nil
}

// ============================================================================
// TaskDeleteCommand deletes a task
// ============================================================================

type TaskDeleteCommand struct {
	Plugin *TaskManagerPlugin
	taskID string
	force  bool
}

func (c *TaskDeleteCommand) GetName() string {
	return "task.delete"
}

func (c *TaskDeleteCommand) GetDescription() string {
	return "Delete a task"
}

func (c *TaskDeleteCommand) GetUsage() string {
	return "dw task-manager task delete <task-id> [--force]"
}

func (c *TaskDeleteCommand) GetHelp() string {
	return `Deletes an existing task.

Arguments:
  <task-id>    Task ID to delete (required)

Flags:
  --force      Skip confirmation prompt (optional)

Examples:
  # Delete a task with confirmation
  dw task-manager task delete task-123

  # Delete without confirmation
  dw task-manager task delete task-123 --force

Notes:
  - Deletion is permanent
  - Without --force, you will be prompted to confirm`
}

func (c *TaskDeleteCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	c.taskID = args[0]

	// Parse flags
	for i := 1; i < len(args); i++ {
		if args[i] == "--force" {
			c.force = true
		}
	}

	// Ensure repository exists
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("task manager not initialized with database support")
	}

	// Get task to verify exists
	task, err := repo.GetTask(ctx, c.taskID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("task not found: %s", c.taskID)
		}
		return fmt.Errorf("failed to retrieve task: %w", err)
	}

	// Prompt for confirmation unless --force
	if !c.force {
		stdout := cmdCtx.GetStdout()
		fmt.Fprintf(stdout, "Delete task '%s'? (y/n): ", task.Title)

		response := make([]byte, 1)
		n, err := cmdCtx.GetStdin().Read(response)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if n == 0 || (response[0] != 'y' && response[0] != 'Y') {
			fmt.Fprintf(stdout, "Cancelled\n")
			return nil
		}
	}

	// Delete task
	if err := repo.DeleteTask(ctx, c.taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Output success message
	stdout := cmdCtx.GetStdout()
	fmt.Fprintf(stdout, "Task deleted successfully\n")
	fmt.Fprintf(stdout, "  ID:    %s\n", task.ID)
	fmt.Fprintf(stdout, "  Title: %s\n", task.Title)

	return nil
}

// ============================================================================
// TaskMoveCommand moves a task to a different track
// ============================================================================

type TaskMoveCommand struct {
	Plugin       *TaskManagerPlugin
	taskID        string
	newTrackID    string
}

func (c *TaskMoveCommand) GetName() string {
	return "task.move"
}

func (c *TaskMoveCommand) GetDescription() string {
	return "Move a task to a different track"
}

func (c *TaskMoveCommand) GetUsage() string {
	return "dw task-manager task move <task-id> --track <new-track-id>"
}

func (c *TaskMoveCommand) GetHelp() string {
	return `Moves a task from its current track to a different track.

Arguments:
  <task-id>     Task ID to move (required)

Flags:
  --track <id>  New track ID (required)

Examples:
  # Move task to a different track
  dw task-manager task move task-123 --track track-plugin-system

Notes:
  - Both old and new tracks must exist
  - Task keeps all other properties when moved
  - Updated timestamp is automatically updated`
}

func (c *TaskMoveCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	c.taskID = args[0]

	// Parse flags
	for i := 1; i < len(args); i++ {
		if args[i] == "--track" && i+1 < len(args) {
			c.newTrackID = args[i+1]
			i++
		}
	}

	// Validate required flags
	if c.newTrackID == "" {
		return fmt.Errorf("--track is required")
	}

	// Ensure repository exists
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("task manager not initialized with database support")
	}

	// Get task
	task, err := repo.GetTask(ctx, c.taskID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("task not found: %s", c.taskID)
		}
		return fmt.Errorf("failed to retrieve task: %w", err)
	}

	oldTrackID := task.TrackID

	// Verify old track exists
	oldTrack, err := repo.GetTrack(ctx, oldTrackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("current track not found: %s", oldTrackID)
		}
		return fmt.Errorf("failed to verify current track: %w", err)
	}

	// Verify new track exists
	newTrack, err := repo.GetTrack(ctx, c.newTrackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("new track not found: %s", c.newTrackID)
		}
		return fmt.Errorf("failed to verify new track: %w", err)
	}

	// Move task
	if err := repo.MoveTaskToTrack(ctx, c.taskID, c.newTrackID); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	// Output success message
	stdout := cmdCtx.GetStdout()
	fmt.Fprintf(stdout, "Task moved successfully\n")
	fmt.Fprintf(stdout, "  ID:         %s\n", task.ID)
	fmt.Fprintf(stdout, "  Title:      %s\n", task.Title)
	fmt.Fprintf(stdout, "  From track: %s (%s)\n", oldTrack.Title, oldTrack.ID)
	fmt.Fprintf(stdout, "  To track:   %s (%s)\n", newTrack.Title, newTrack.ID)

	return nil
}

// ============================================================================
// TaskMigrateCommand migrates file-based tasks to database (placeholder)
// ============================================================================

type TaskMigrateCommand struct {
	Plugin  *TaskManagerPlugin
	taskID  string
	trackID string
}

func (c *TaskMigrateCommand) GetName() string {
	return "task.migrate"
}

func (c *TaskMigrateCommand) GetDescription() string {
	return "Migrate file-based task to database hierarchy"
}

func (c *TaskMigrateCommand) GetUsage() string {
	return "dw task-manager task migrate <task-id> --track <track-id>"
}

func (c *TaskMigrateCommand) GetHelp() string {
	return `Migrates a file-based task to the database-backed hierarchical model.

This command is used to move tasks from the old file-based storage
to the new database-backed model with track ownership.

Arguments:
  <task-id>     File-based task ID to migrate (required)

Flags:
  --track <id>  Target track ID (required)

Examples:
  # Migrate a file-based task
  dw task-manager task migrate task-old-123 --track track-framework-core

Notes:
  - Migration is one-way; source task file is not deleted
  - Target track must exist in the active roadmap`
}

func (c *TaskMigrateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	c.taskID = args[0]

	// Parse flags
	for i := 1; i < len(args); i++ {
		if args[i] == "--track" && i+1 < len(args) {
			c.trackID = args[i+1]
			i++
		}
	}

	// Validate required flags
	if c.trackID == "" {
		return fmt.Errorf("--track is required")
	}

	// Placeholder implementation
	stdout := cmdCtx.GetStdout()
	fmt.Fprintf(stdout, "Task migration is not yet implemented\n")
	fmt.Fprintf(stdout, "Task migration will be implemented in a future phase\n")

	return nil
}
