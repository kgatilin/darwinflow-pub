package task_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// InitCommand initializes the task directory
type InitCommand struct {
	plugin *TaskManagerPlugin
}

func (c *InitCommand) GetName() string {
	return "init"
}

func (c *InitCommand) GetDescription() string {
	return "Initialize task directory"
}

func (c *InitCommand) GetUsage() string {
	return "dw task-manager init"
}

func (c *InitCommand) GetHelp() string {
	return `Creates the task directory at .darwinflow/tasks/

This command sets up the necessary directory structure for task management.
Tasks are stored as JSON files in this directory, and the file watcher
monitors this location for changes to emit real-time events.

Examples:
  dw task-manager init

Notes:
  - Safe to run multiple times (idempotent)
  - Creates .darwinflow/tasks/ if it doesn't exist
  - Required before creating tasks`
}

func (c *InitCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Create tasks directory
	tasksDir := filepath.Join(cmdCtx.GetWorkingDir(), ".darwinflow", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return fmt.Errorf("failed to create tasks directory: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Task directory initialized at %s\n", tasksDir)
	return nil
}

// CreateCommand creates a new task
type CreateCommand struct {
	plugin *TaskManagerPlugin
}

func (c *CreateCommand) GetName() string {
	return "create"
}

func (c *CreateCommand) GetDescription() string {
	return "Create a new task"
}

func (c *CreateCommand) GetUsage() string {
	return "dw task-manager create <title> [--description <desc>] [--priority <priority>]"
}

func (c *CreateCommand) GetHelp() string {
	return `Creates a new task with the specified title and optional metadata.

The task is saved as a JSON file in .darwinflow/tasks/ with a unique ID
based on the current timestamp. When the file is created, an event is
emitted if the file watcher is running.

Arguments:
  <title>              Task title (required)

Flags:
  --description <desc> Task description (optional)
  --priority <level>   Priority level: low, medium, high (default: medium)

Examples:
  # Simple task
  dw task-manager create "Fix bug in parser"

  # Task with description
  dw task-manager create "Implement feature X" --description "Add support for Y"

  # Task with priority
  dw task-manager create "Critical fix" --priority high

  # Task with all options
  dw task-manager create "Complete refactoring" \
    --description "Refactor authentication module" \
    --priority high

Notes:
  - Task ID is auto-generated as task-<timestamp>
  - Tasks are created with status "todo"
  - Titles with spaces should be quoted`
}

func (c *CreateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	title := args[0]
	description := ""
	priority := "medium"

	// Parse optional flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--description":
			if i+1 < len(args) {
				description = args[i+1]
				i++
			}
		case "--priority":
			if i+1 < len(args) {
				priority = args[i+1]
				i++
			}
		}
	}

	// Generate task ID based on current timestamp
	taskID := fmt.Sprintf("task-%d", time.Now().UnixNano())
	now := time.Now().UTC()

	task := &TaskEntity{
		ID:          taskID,
		Title:       title,
		Description: description,
		Status:      "todo",
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Save task
	tasksDir := filepath.Join(cmdCtx.GetWorkingDir(), ".darwinflow", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return fmt.Errorf("failed to create tasks directory: %w", err)
	}

	filePath := filepath.Join(tasksDir, taskID+".json")
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Task created: %s\n", taskID)
	fmt.Fprintf(cmdCtx.GetStdout(), "Title: %s\n", title)
	fmt.Fprintf(cmdCtx.GetStdout(), "Priority: %s\n", priority)

	return nil
}

// ListCommand lists all tasks
type ListCommand struct {
	plugin *TaskManagerPlugin
}

func (c *ListCommand) GetName() string {
	return "list"
}

func (c *ListCommand) GetDescription() string {
	return "List all tasks"
}

func (c *ListCommand) GetUsage() string {
	return "dw task-manager list [--status <status>]"
}

func (c *ListCommand) GetHelp() string {
	return `Lists all tasks from the .darwinflow/tasks/ directory.

Tasks are displayed in a formatted table showing ID, title, status,
and priority. You can filter tasks by status using the --status flag.

Flags:
  --status <status>    Filter by status: todo, in-progress, done

Examples:
  # List all tasks
  dw task-manager list

  # List only pending tasks
  dw task-manager list --status todo

  # List completed tasks
  dw task-manager list --status done

  # List in-progress tasks
  dw task-manager list --status in-progress

Output format:
  ID                   Title                          Status          Priority
  ---------------------------------------------------------------------------
  task-123             Example task                   todo            high

Notes:
  - IDs are truncated to 20 characters for display
  - Titles are truncated to 30 characters
  - Tasks are read directly from JSON files`
}

func (c *ListCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse optional status filter
	var statusFilter string
	for i := 1; i < len(args); i++ {
		if args[i] == "--status" && i+1 < len(args) {
			statusFilter = args[i+1]
			i++
		}
	}

	// Build query
	query := pluginsdk.EntityQuery{
		EntityType: "task",
	}

	if statusFilter != "" {
		query.Filters = map[string]interface{}{
			"status": statusFilter,
		}
	}

	// Query tasks
	entities, err := c.plugin.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query tasks: %w", err)
	}

	if len(entities) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "No tasks found\n")
		return nil
	}

	// Display tasks
	fmt.Fprintf(cmdCtx.GetStdout(), "Tasks:\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "%-20s %-30s %-15s %-10s\n", "ID", "Title", "Status", "Priority")
	fmt.Fprintf(cmdCtx.GetStdout(), "%s\n", strings.Repeat("-", 75))

	for _, entity := range entities {
		task := entity.(*TaskEntity)
		fmt.Fprintf(
			cmdCtx.GetStdout(),
			"%-20s %-30s %-15s %-10s\n",
			task.ID[:20],
			truncateString(task.Title, 30),
			task.Status,
			task.Priority,
		)
	}

	return nil
}

// UpdateCommand updates a task
type UpdateCommand struct {
	plugin *TaskManagerPlugin
}

func (c *UpdateCommand) GetName() string {
	return "update"
}

func (c *UpdateCommand) GetDescription() string {
	return "Update a task"
}

func (c *UpdateCommand) GetUsage() string {
	return "dw task-manager update <id> [--status <status>] [--title <title>] [--description <desc>] [--priority <priority>]"
}

func (c *UpdateCommand) GetHelp() string {
	return `Updates an existing task's properties.

You can update any combination of task fields: status, title, description,
or priority. At least one field must be specified to update.

Arguments:
  <id>                 Task ID (required, can be abbreviated)

Flags:
  --status <status>    New status: todo, in-progress, done
  --title <title>      New task title
  --description <desc> New task description
  --priority <level>   New priority: low, medium, high

Examples:
  # Update task status
  dw task-manager update task-123 --status done

  # Update multiple fields
  dw task-manager update task-123 --status in-progress --priority high

  # Update title and description
  dw task-manager update task-123 \
    --title "New title" \
    --description "Updated description"

  # Use abbreviated ID
  dw task-manager update task-123 --status done

Notes:
  - Task ID can be the full ID or abbreviated (must be unique)
  - At least one update flag is required
  - Updates are written back to the JSON file
  - Updated_at timestamp is automatically set`
}

func (c *UpdateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: %s", c.GetUsage())
	}

	taskID := args[0]
	updates := make(map[string]interface{})

	// Parse optional flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--status":
			if i+1 < len(args) {
				updates["status"] = args[i+1]
				i++
			}
		case "--title":
			if i+1 < len(args) {
				updates["title"] = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				updates["description"] = args[i+1]
				i++
			}
		case "--priority":
			if i+1 < len(args) {
				updates["priority"] = args[i+1]
				i++
			}
		}
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates specified")
	}

	// Update task
	_, err := c.plugin.UpdateEntity(ctx, taskID, updates)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Task updated: %s\n", taskID)
	for key, value := range updates {
		fmt.Fprintf(cmdCtx.GetStdout(), "  %s: %v\n", key, value)
	}

	return nil
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
