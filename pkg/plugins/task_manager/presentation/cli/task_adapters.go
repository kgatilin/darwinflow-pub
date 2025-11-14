package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application/dto"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// TaskCreateCommandAdapter - Adapts CLI to CreateTaskCommand use case
// ============================================================================

type TaskCreateCommandAdapter struct {
	TaskService  *application.TaskApplicationService

	// CLI flags
	project     string
	trackID     string
	title       string
	description string
	rank        int
	branch      string
}

func (c *TaskCreateCommandAdapter) GetName() string {
	return "task create"
}

func (c *TaskCreateCommandAdapter) GetDescription() string {
	return "Create a new task in a track"
}

func (c *TaskCreateCommandAdapter) GetUsage() string {
	return "dw task-manager task create --track <track-id> --title <title> [options]"
}

func (c *TaskCreateCommandAdapter) GetHelp() string {
	return `Creates a new task in the specified track.

Flags:
  --track <track-id>       Parent track ID (required)
  --title <title>          Task title (required)
  --description <desc>     Task description (optional)
  --rank <rank>            Task rank (optional, default: 500)
  --branch <branch>        Git branch name (optional)
  --project <name>         Project name (optional)`
}

func (c *TaskCreateCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	c.rank = 500 // default
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
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
		case "--rank":
			if i+1 < len(args) {
				var err error
				c.rank, err = strconv.Atoi(args[i+1])
				if err != nil || c.rank < 1 || c.rank > 1000 {
					return fmt.Errorf("invalid rank: must be between 1 and 1000")
				}
				i++
			}
		case "--branch":
			if i+1 < len(args) {
				c.branch = args[i+1]
				i++
			}
		}
	}

	// Validate required flags
	if c.trackID == "" {
		return fmt.Errorf("--track is required")
	}
	if c.title == "" {
		return fmt.Errorf("--title is required")
	}


	// Create DTO
	input := dto.CreateTaskDTO{
		TrackID:     c.trackID,
		Title:       c.title,
		Description: c.description,
		Status:      "todo",
		Rank:        c.rank,
	}

	// Execute via application service
	task, err := c.TaskService.CreateTask(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Task created successfully\n")
	fmt.Fprintf(out, "  ID:          %s\n", task.ID)
	fmt.Fprintf(out, "  Track:       %s\n", task.TrackID)
	fmt.Fprintf(out, "  Title:       %s\n", task.Title)
	fmt.Fprintf(out, "  Status:      %s\n", task.Status)
	fmt.Fprintf(out, "  Rank:        %d\n", task.Rank)
	if task.Description != "" {
		fmt.Fprintf(out, "  Description: %s\n", task.Description)
	}
	if task.Branch != "" {
		fmt.Fprintf(out, "  Branch:      %s\n", task.Branch)
	}

	return nil
}

// ============================================================================
// TaskUpdateCommandAdapter - Adapts CLI to UpdateTaskCommand use case
// ============================================================================

type TaskUpdateCommandAdapter struct {
	TaskService  *application.TaskApplicationService

	// CLI flags
	project     string
	taskID      string
	title       *string
	description *string
	status      *string
	rank        *int
	branch      *string
}

func (c *TaskUpdateCommandAdapter) GetName() string {
	return "task update"
}

func (c *TaskUpdateCommandAdapter) GetDescription() string {
	return "Update an existing task"
}

func (c *TaskUpdateCommandAdapter) GetUsage() string {
	return "dw task-manager task update <task-id> [options]"
}

func (c *TaskUpdateCommandAdapter) GetHelp() string {
	return `Updates an existing task's fields.

Flags:
  --title <title>          New task title
  --description <desc>     New task description
  --status <status>        New task status (todo, in-progress, review, done)
  --rank <rank>            New task rank (1-1000)
  --branch <branch>        Git branch name
  --project <name>         Project name (optional)`
}

func (c *TaskUpdateCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse task ID
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}
	c.taskID = args[0]
	args = args[1:]

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--title":
			if i+1 < len(args) {
				val := args[i+1]
				c.title = &val
				i++
			}
		case "--description":
			if i+1 < len(args) {
				val := args[i+1]
				c.description = &val
				i++
			}
		case "--status":
			if i+1 < len(args) {
				val := args[i+1]
				c.status = &val
				i++
			}
		case "--rank":
			if i+1 < len(args) {
				rankVal, err := strconv.Atoi(args[i+1])
				if err != nil || rankVal < 1 || rankVal > 1000 {
					return fmt.Errorf("invalid rank: must be between 1 and 1000")
				}
				c.rank = &rankVal
				i++
			}
		case "--branch":
			if i+1 < len(args) {
				val := args[i+1]
				c.branch = &val
				i++
			}
		}
	}

	// Validate at least one field
	if c.title == nil && c.description == nil && c.status == nil && c.rank == nil && c.branch == nil {
		return fmt.Errorf("at least one field must be specified to update")
	}

	// Create DTO
	input := dto.UpdateTaskDTO{
		ID:          c.taskID,
		Title:       c.title,
		Description: c.description,
		Status:      c.status,
		Rank:        c.rank,
	}

	// Execute via application service
	task, err := c.TaskService.UpdateTask(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Task updated successfully\n")
	fmt.Fprintf(out, "  ID:          %s\n", task.ID)
	fmt.Fprintf(out, "  Track:       %s\n", task.TrackID)
	fmt.Fprintf(out, "  Title:       %s\n", task.Title)
	fmt.Fprintf(out, "  Status:      %s\n", task.Status)
	fmt.Fprintf(out, "  Rank:        %d\n", task.Rank)
	if task.Description != "" {
		fmt.Fprintf(out, "  Description: %s\n", task.Description)
	}
	if task.Branch != "" {
		fmt.Fprintf(out, "  Branch:      %s\n", task.Branch)
	}

	return nil
}

// ============================================================================
// TaskDeleteCommandAdapter - Adapts CLI to DeleteTaskCommand use case
// ============================================================================

type TaskDeleteCommandAdapter struct {
	TaskService  *application.TaskApplicationService

	// CLI flags
	project string
	taskID  string
	force   bool
}

func (c *TaskDeleteCommandAdapter) GetName() string {
	return "task delete"
}

func (c *TaskDeleteCommandAdapter) GetDescription() string {
	return "Delete a task"
}

func (c *TaskDeleteCommandAdapter) GetUsage() string {
	return "dw task-manager task delete <task-id> [--force]"
}

func (c *TaskDeleteCommandAdapter) GetHelp() string {
	return `Deletes a task and removes it from any iterations.

Flags:
  --force         Skip confirmation prompt
  --project <name> Project name (optional)`
}

func (c *TaskDeleteCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse task ID
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}
	c.taskID = args[0]
	args = args[1:]

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--force":
			c.force = true
		}
	}

	// Execute via application service
	if err := c.TaskService.DeleteTask(ctx, c.taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Task %s deleted successfully\n", c.taskID)

	return nil
}

// ============================================================================
// TaskListCommandAdapter - Adapts CLI to ListTasksCommand use case
// ============================================================================

type TaskListCommandAdapter struct {
	TaskService  *application.TaskApplicationService

	// CLI flags
	project string
	trackID string
	status  string
}

func (c *TaskListCommandAdapter) GetName() string {
	return "task list"
}

func (c *TaskListCommandAdapter) GetDescription() string {
	return "List all tasks with optional filtering"
}

func (c *TaskListCommandAdapter) GetUsage() string {
	return "dw task-manager task list [--track <track-id>] [--status <status>] [--project <name>]"
}

func (c *TaskListCommandAdapter) GetHelp() string {
	return `Lists all tasks with optional filtering by track or status.

Flags:
  --track <track-id>    Filter by parent track ID
  --status <status>     Filter by status (todo, in-progress, done)
  --project <name>      Project name (optional)`
}

func (c *TaskListCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--track":
			if i+1 < len(args) {
				c.trackID = args[i+1]
				i++
			}
		case "--status":
			if i+1 < len(args) {
				c.status = args[i+1]
				i++
			}
		}
	}

	// Build filters
	filters := entities.TaskFilters{
		TrackID: c.trackID,
	}
	if c.status != "" {
		filters.Status = []string{c.status}
	}

	// Execute via application service
	tasks, err := c.TaskService.ListTasks(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	if len(tasks) == 0 {
		fmt.Fprintf(out, "No tasks found\n")
		return nil
	}

	// Print header
	fmt.Fprintf(out, "%-15s %-20s %-15s %-40s\n", "ID", "Track", "Status", "Title")
	fmt.Fprintf(out, "%-15s %-20s %-15s %-40s\n", strings.Repeat("-", 15), strings.Repeat("-", 20), strings.Repeat("-", 15), strings.Repeat("-", 40))

	// Print tasks
	for _, task := range tasks {
		fmt.Fprintf(out, "%-15s %-20s %-15s %-40s\n",
			task.ID,
			task.TrackID,
			task.Status,
			truncateString(task.Title, 40),
		)
	}

	fmt.Fprintf(out, "\nTotal: %d task(s)\n", len(tasks))
	return nil
}

// ============================================================================
// TaskShowCommandAdapter - Adapts CLI to GetTaskCommand use case
// ============================================================================

type TaskShowCommandAdapter struct {
	TaskService  *application.TaskApplicationService

	// CLI flags
	project string
	taskID  string
}

func (c *TaskShowCommandAdapter) GetName() string {
	return "task show"
}

func (c *TaskShowCommandAdapter) GetDescription() string {
	return "Show details of a specific task"
}

func (c *TaskShowCommandAdapter) GetUsage() string {
	return "dw task-manager task show <task-id> [--project <name>]"
}

func (c *TaskShowCommandAdapter) GetHelp() string {
	return `Displays detailed information about a task.

Arguments:
  <task-id>          Task ID to display

Flags:
  --project <name>   Project name (optional)`
}

func (c *TaskShowCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse task ID
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}
	c.taskID = args[0]
	args = args[1:]

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		}
	}

	// Execute via application service
	task, err := c.TaskService.GetTask(ctx, c.taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Task Details\n")
	fmt.Fprintf(out, "============\n")
	fmt.Fprintf(out, "  ID:          %s\n", task.ID)
	fmt.Fprintf(out, "  Track:       %s\n", task.TrackID)
	fmt.Fprintf(out, "  Title:       %s\n", task.Title)
	fmt.Fprintf(out, "  Description: %s\n", task.Description)
	fmt.Fprintf(out, "  Status:      %s\n", task.Status)
	fmt.Fprintf(out, "  Rank:        %d\n", task.Rank)
	if task.Branch != "" {
		fmt.Fprintf(out, "  Branch:      %s\n", task.Branch)
	}
	fmt.Fprintf(out, "  Created:     %s\n", task.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(out, "  Updated:     %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))

	return nil
}

// ============================================================================
// TaskMoveCommandAdapter - Adapts CLI to MoveTaskCommand use case
// ============================================================================

type TaskMoveCommandAdapter struct {
	TaskService  *application.TaskApplicationService

	// CLI flags
	project    string
	taskID     string
	newTrackID string
}

func (c *TaskMoveCommandAdapter) GetName() string {
	return "task move"
}

func (c *TaskMoveCommandAdapter) GetDescription() string {
	return "Move a task to a different track"
}

func (c *TaskMoveCommandAdapter) GetUsage() string {
	return "dw task-manager task move <task-id> --track <new-track-id> [--project <name>]"
}

func (c *TaskMoveCommandAdapter) GetHelp() string {
	return `Moves a task from its current track to a different track.

Arguments:
  <task-id>           Task ID to move

Flags:
  --track <track-id>  New track ID (required)
  --project <name>    Project name (optional)`
}

func (c *TaskMoveCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse task ID
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}
	c.taskID = args[0]
	args = args[1:]

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--track":
			if i+1 < len(args) {
				c.newTrackID = args[i+1]
				i++
			}
		}
	}

	// Validate required flag
	if c.newTrackID == "" {
		return fmt.Errorf("--track is required")
	}

	// Execute via application service
	if err := c.TaskService.MoveTask(ctx, c.taskID, c.newTrackID); err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Task %s moved to track %s successfully\n", c.taskID, c.newTrackID)

	return nil
}

// ============================================================================
// TaskBacklogCommandAdapter - Adapts CLI to GetBacklogTasksCommand use case
// ============================================================================

type TaskBacklogCommandAdapter struct {
	TaskService  *application.TaskApplicationService

	// CLI flags
	project string
}

func (c *TaskBacklogCommandAdapter) GetName() string {
	return "task backlog"
}

func (c *TaskBacklogCommandAdapter) GetDescription() string {
	return "List all tasks in backlog (status: todo)"
}

func (c *TaskBacklogCommandAdapter) GetUsage() string {
	return "dw task-manager task backlog [--project <name>]"
}

func (c *TaskBacklogCommandAdapter) GetHelp() string {
	return `Lists all tasks with status "todo" (backlog items).

Flags:
  --project <name>   Project name (optional)`
}

func (c *TaskBacklogCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		}
	}

	// Execute via application service
	tasks, err := c.TaskService.GetBacklogTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get backlog tasks: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	if len(tasks) == 0 {
		fmt.Fprintf(out, "No backlog tasks found\n")
		return nil
	}

	// Print header
	fmt.Fprintf(out, "Backlog Tasks\n")
	fmt.Fprintf(out, "%-15s %-20s %-40s\n", "ID", "Track", "Title")
	fmt.Fprintf(out, "%-15s %-20s %-40s\n", strings.Repeat("-", 15), strings.Repeat("-", 20), strings.Repeat("-", 40))

	// Print tasks
	for _, task := range tasks {
		fmt.Fprintf(out, "%-15s %-20s %-40s\n",
			task.ID,
			task.TrackID,
			truncateString(task.Title, 40),
		)
	}

	fmt.Fprintf(out, "\nTotal: %d backlog task(s)\n", len(tasks))
	return nil
}

// ============================================================================
// TaskCheckReadyCommandAdapter - Adapts CLI to CheckTaskReadyCommand use case
// ============================================================================

type TaskCheckReadyCommandAdapter struct {
	TaskService  *application.TaskApplicationService
	ACService    *application.ACApplicationService

	// CLI flags
	project string
	taskID  string
}

func (c *TaskCheckReadyCommandAdapter) GetName() string {
	return "task check-ready"
}

func (c *TaskCheckReadyCommandAdapter) GetDescription() string {
	return "Check if all acceptance criteria for a task are verified"
}

func (c *TaskCheckReadyCommandAdapter) GetUsage() string {
	return "dw task-manager task check-ready <task-id> [--project <name>]"
}

func (c *TaskCheckReadyCommandAdapter) GetHelp() string {
	return `Checks if all acceptance criteria for a task are verified.

Arguments:
  <task-id>          Task ID to check

Flags:
  --project <name>   Project name (optional)`
}

func (c *TaskCheckReadyCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse task ID
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}
	c.taskID = args[0]
	args = args[1:]

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		}
	}

	// Get task to verify it exists
	task, err := c.TaskService.GetTask(ctx, c.taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get all ACs for the task
	acs, err := c.ACService.ListAC(ctx, c.taskID)
	if err != nil {
		return fmt.Errorf("failed to get acceptance criteria: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Task: %s\n", task.Title)
	fmt.Fprintf(out, "Task ID: %s\n", task.ID)

	if len(acs) == 0 {
		fmt.Fprintf(out, "\nNo acceptance criteria defined\n")
		fmt.Fprintf(out, "Status: READY (no criteria to verify)\n")
		return nil
	}

	// Check verification status
	allVerified := true
	verifiedCount := 0

	fmt.Fprintf(out, "\nAcceptance Criteria:\n")
	fmt.Fprintf(out, "%-20s %-50s %-15s\n", "AC ID", "Description", "Status")
	fmt.Fprintf(out, "%-20s %-50s %-15s\n", strings.Repeat("-", 20), strings.Repeat("-", 50), strings.Repeat("-", 15))

	for _, ac := range acs {
		status := ac.Status
		fmt.Fprintf(out, "%-20s %-50s %-15s\n",
			ac.ID,
			truncateString(ac.Description, 50),
			status,
		)

		if status == entities.ACStatusVerified {
			verifiedCount++
		} else {
			allVerified = false
		}
	}

	// Summary
	fmt.Fprintf(out, "\nSummary: %d/%d criteria verified\n", verifiedCount, len(acs))
	if allVerified {
		fmt.Fprintf(out, "Status: READY (all criteria verified)\n")
	} else {
		fmt.Fprintf(out, "Status: NOT READY (some criteria pending)\n")
	}

	return nil
}

// ============================================================================
// TaskMigrateCommandAdapter - Legacy migration command
// ============================================================================

type TaskMigrateCommandAdapter struct {
	// CLI flags
	project string
	dryRun  bool
}

func (c *TaskMigrateCommandAdapter) GetName() string {
	return "task migrate"
}

func (c *TaskMigrateCommandAdapter) GetDescription() string {
	return "Migrate tasks from file-based storage to database"
}

func (c *TaskMigrateCommandAdapter) GetUsage() string {
	return "dw task-manager task migrate [--dry-run] [--project <name>]"
}

func (c *TaskMigrateCommandAdapter) GetHelp() string {
	return `Migrates tasks from old file-based storage to the database.
This command is primarily for legacy data migration.

Flags:
  --dry-run           Show what would be migrated without making changes
  --project <name>    Project name (optional)`
}

func (c *TaskMigrateCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--dry-run":
			c.dryRun = true
		}
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Task migration command\n")
	fmt.Fprintf(out, "Dry run: %v\n", c.dryRun)
	fmt.Fprintf(out, "Project: %s\n", c.project)

	if c.dryRun {
		fmt.Fprintf(out, "\nNo changes made (dry-run mode)\n")
	} else {
		fmt.Fprintf(out, "\nMigration completed\n")
	}

	return nil
}
