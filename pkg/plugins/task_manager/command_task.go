package task_manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// TaskCreateCommand creates a new task in a track
// ============================================================================

type TaskCreateCommand struct {
	Plugin      *TaskManagerPlugin
	project string
	trackID     string
	title       string
	description string
	rank        int
}

func (c *TaskCreateCommand) GetName() string {
	return "task create"
}

func (c *TaskCreateCommand) GetDescription() string {
	return "Create a new task in a track"
}

func (c *TaskCreateCommand) GetUsage() string {
	return "dw task-manager task create --track <track-id> --title <title> [--description <desc>] [--rank <rank>]"
}

func (c *TaskCreateCommand) GetHelp() string {
	return `Creates a new task in the specified track.

A task represents a unit of work within a track. All tasks must belong
to an existing track in the active roadmap.

Flags:
  --track <track-id>        Track ID to create task in (required)
  --title <title>           Task title (required)
  --description <desc>      Task description (optional)
  --rank <rank>             Task rank (optional, default: 500)
                           Range: 1-1000 (lower = higher priority)

Examples:
  # Create a basic task
  dw task-manager task create \
    --track track-plugin-system \
    --title "Implement plugin registry"

  # Create with custom rank
  dw task-manager task create \
    --track track-plugin-system \
    --title "Implement plugin registry" \
    --description "Create registry to discover and load plugins" \
    --rank 100

Notes:
  - Track must exist (create with 'dw task-manager track create')
  - Initial status is automatically set to 'todo'
  - Task ID is generated automatically`
}

func (c *TaskCreateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags FIRST to get project name
	c.rank = 500 // default (medium priority)
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
		}
	}

	// Validate required flags
	if c.trackID == "" || c.title == "" {
		return fmt.Errorf("--track and --title are required")
	}

	// Get repository for project AFTER parsing flags
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Verify track exists with helpful error message
	track, err := repo.GetTrack(ctx, c.trackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: Track \"%s\" not found\n", c.trackID)
			fmt.Fprintf(cmdCtx.GetStdout(), "Hint: Use 'dw task-manager track list' to see available tracks\n")
			return fmt.Errorf("track not found: %s", c.trackID)
		}
		return fmt.Errorf("failed to verify track: %w", err)
	}

	// Generate task ID using project code and sequence number
	projectCode := repo.GetProjectCode(ctx)
	nextNum, err := repo.GetNextSequenceNumber(ctx, "task")
	if err != nil {
		return fmt.Errorf("failed to generate task ID: %w", err)
	}
	taskID := fmt.Sprintf("%s-task-%d", projectCode, nextNum)

	// Create task
	task := NewTaskEntity(
		taskID,
		c.trackID,
		c.title,
		c.description,
		"todo",         // initial status
		c.rank,
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
	project string
	track    string
	status   string
}

func (c *TaskListCommand) GetName() string {
	return "task list"
}

func (c *TaskListCommand) GetDescription() string {
	return "List tasks with optional filtering"
}

func (c *TaskListCommand) GetUsage() string {
	return "dw task-manager task list [--track <track-id>] [--status <status>]"
}

func (c *TaskListCommand) GetHelp() string {
	return `Lists all tasks, optionally filtered by track or status.

Tasks are displayed sorted by rank within each track (lower ranks first).

Flags:
  --track <track-id>        Filter by track ID (optional)
  --status <status>         Filter by status (optional, comma-separated)
                           Values: todo, in-progress, done

Examples:
  # List all tasks
  dw task-manager task list

  # List tasks in a specific track
  dw task-manager task list --track track-plugin-system

  # List tasks with specific status
  dw task-manager task list --status todo

  # List in-progress or done tasks
  dw task-manager task list --status in-progress,done

Notes:
  - Status filter accepts comma-separated values
  - All filters are optional and can be combined
  - Tasks are ordered by rank (1=highest priority, 1000=lowest)`
}

func (c *TaskListCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
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
				c.track = args[i+1]
				i++
			}
		case "--status":
			if i+1 < len(args) {
				c.status = args[i+1]
				i++
			}
		}
	}

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Build filters
	filters := TaskFilters{
		TrackID: c.track,
	}

	if c.status != "" {
		filters.Status = strings.Split(c.status, ",")
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
		"ID", "Title", "Status", "Rank", "Track")
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

		fmt.Fprintf(stdout, "%-20s %-40s %-12s %-8d %-20s\n",
			abbrevID, task.Title, task.Status, task.Rank, trackName)
	}

	fmt.Fprintf(stdout, "\nTotal: %d task(s)\n", len(tasks))
	return nil
}

// ============================================================================
// TaskShowCommand shows detailed task information
// ============================================================================

type TaskShowCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	taskID string
}

func (c *TaskShowCommand) GetName() string {
	return "task show"
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

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

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

	// Get iterations that contain this task
	iterations, err := repo.GetIterationsForTask(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve iterations: %w", err)
	}

	// Display task details
	stdout := cmdCtx.GetStdout()
	fmt.Fprintf(stdout, "Task Details\n")
	fmt.Fprintf(stdout, "%s\n\n", strings.Repeat("=", 60))

	fmt.Fprintf(stdout, "ID:          %s\n", task.ID)
	fmt.Fprintf(stdout, "Title:       %s\n", task.Title)
	fmt.Fprintf(stdout, "Status:      %s\n", task.Status)
	fmt.Fprintf(stdout, "Rank:        %d\n", task.Rank)

	if task.Description != "" {
		fmt.Fprintf(stdout, "Description: %s\n", task.Description)
	}

	fmt.Fprintf(stdout, "\nTrack:       %s (%s)\n", track.Title, track.ID)

	if task.Branch != "" {
		fmt.Fprintf(stdout, "Branch:      %s\n", task.Branch)
	}

	// Display iterations
	fmt.Fprintf(stdout, "\nIterations:\n")
	if len(iterations) == 0 {
		fmt.Fprintf(stdout, "  Not assigned to any iteration\n")
	} else {
		for _, iter := range iterations {
			fmt.Fprintf(stdout, "  - Iteration %d: %s (status: %s)\n",
				iter.Number, iter.Name, iter.Status)
		}
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
	project string
	taskID      string
	title       string
	description string
	status      string
	rank        int
	branch      string
	hasUpdates  bool
	hasRank     bool
}

func (c *TaskUpdateCommand) GetName() string {
	return "task update"
}

func (c *TaskUpdateCommand) GetDescription() string {
	return "Update an existing task"
}

func (c *TaskUpdateCommand) GetUsage() string {
	return "dw task-manager task update <task-id> [--title <title>] [--description <desc>] [--status <status>] [--rank <rank>] [--branch <branch>]"
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
  --rank <rank>             New task rank (optional, 1-1000)
  --branch <branch>         Git branch name (optional)

Examples:
  # Mark task as in-progress
  dw task-manager task update task-123 --status in-progress

  # Update multiple fields
  dw task-manager task update task-123 \
    --status done \
    --rank 100

  # Set a git branch
  dw task-manager task update task-123 --branch feat/new-feature

Notes:
  - At least one flag must be provided
  - Use --branch to associate a git branch
  - Use --status to change task progress
  - Rank determines ordering (1=highest priority, 1000=lowest)`
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
		case "--rank":
			if i+1 < len(args) {
				var err error
				c.rank, err = strconv.Atoi(args[i+1])
				if err != nil || c.rank < 1 || c.rank > 1000 {
					return fmt.Errorf("invalid rank: must be between 1 and 1000")
				}
				c.hasUpdates = true
				c.hasRank = true
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

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get existing task
	task, err := repo.GetTask(ctx, c.taskID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("task not found: %s", c.taskID)
		}
		return fmt.Errorf("failed to retrieve task: %w", err)
	}

	// Validate status change to "done"
	if c.status == "done" && task.Status != "done" {
		// Get the parent track
		track, err := repo.GetTrack(ctx, task.TrackID)
		if err != nil {
			return fmt.Errorf("failed to retrieve parent track: %w", err)
		}

		// Check if all acceptance criteria are verified
		acs, err := repo.ListAC(ctx, c.taskID)
		if err != nil {
			return fmt.Errorf("failed to check acceptance criteria: %w", err)
		}

		// Build lists of unverified and failed ACs
		var unverifiedAC []*AcceptanceCriteriaEntity
		var failedAC []*AcceptanceCriteriaEntity
		for _, ac := range acs {
			if ac.IsFailed() {
				failedAC = append(failedAC, ac)
			} else if !ac.IsVerified() {
				unverifiedAC = append(unverifiedAC, ac)
			}
		}

		if len(unverifiedAC) > 0 || len(failedAC) > 0 {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: Cannot mark task as done\n\n")

			if len(unverifiedAC) > 0 {
				fmt.Fprintf(cmdCtx.GetStdout(), "Unverified acceptance criteria (%d):\n", len(unverifiedAC))
				for _, ac := range unverifiedAC {
					statusIcon := ac.StatusIndicator()
					fmt.Fprintf(cmdCtx.GetStdout(), "  %s [%s] %s\n", statusIcon, ac.ID, ac.Description)
				}
				fmt.Fprintf(cmdCtx.GetStdout(), "\n")
			}

			if len(failedAC) > 0 {
				fmt.Fprintf(cmdCtx.GetStdout(), "Failed acceptance criteria (%d):\n", len(failedAC))
				for _, ac := range failedAC {
					statusIcon := ac.StatusIndicator()
					fmt.Fprintf(cmdCtx.GetStdout(), "  %s [%s] %s\n", statusIcon, ac.ID, ac.Description)
					if ac.Notes != "" {
						fmt.Fprintf(cmdCtx.GetStdout(), "     Feedback: %s\n", ac.Notes)
					}
				}
				fmt.Fprintf(cmdCtx.GetStdout(), "\n")
			}

			// Return nil to avoid showing help text (error already printed to stdout)
			return nil
		}

		// Check if ADR requirement is configured and enforce it
		config := c.Plugin.GetConfig()
		if config.ADR.Required && config.ADR.EnforceOnTaskCompletion {
			// Check if track has at least one ADR
			adrs, err := repo.ListADRs(ctx, &task.TrackID)
			if err != nil {
				return fmt.Errorf("failed to check ADR status: %w", err)
			}

			if len(adrs) == 0 {
				fmt.Fprintf(cmdCtx.GetStdout(), "Error: Cannot complete task - track has no ADR\n\n")
				fmt.Fprintf(cmdCtx.GetStdout(), "Track \"%s\" requires an Architecture Decision Record before tasks can be completed.\n\n", track.ID)
				fmt.Fprintf(cmdCtx.GetStdout(), "Create an ADR with:\n")
				fmt.Fprintf(cmdCtx.GetStdout(), "  dw task-manager adr create %s --title \"...\" --context \"...\" --decision \"...\" --consequences \"...\"\n\n", track.ID)
				fmt.Fprintf(cmdCtx.GetStdout(), "Or disable ADR requirement in config:\n")
				fmt.Fprintf(cmdCtx.GetStdout(), "  task_manager.adr.enforce_on_task_completion: false\n")
				return fmt.Errorf("cannot complete task: track has no ADR")
			}
		}
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
	if c.hasRank {
		task.Rank = c.rank
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
	Plugin  *TaskManagerPlugin
	project string
	taskID string
	force  bool
}

func (c *TaskDeleteCommand) GetName() string {
	return "task delete"
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

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

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
	project string
	taskID        string
	newTrackID    string
}

func (c *TaskMoveCommand) GetName() string {
	return "task move"
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

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

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
	return "task migrate"
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

// ============================================================================
// TaskBacklogCommand shows all unassigned tasks (backlog)
// ============================================================================

type TaskBacklogCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *TaskBacklogCommand) GetName() string {
	return "task backlog"
}

func (c *TaskBacklogCommand) GetDescription() string {
	return "Show all unassigned tasks (backlog)"
}

func (c *TaskBacklogCommand) GetUsage() string {
	return "dw task-manager task backlog"
}

func (c *TaskBacklogCommand) GetHelp() string {
	return `Shows all tasks that are not assigned to any iteration and not done.

The backlog represents work that has been planned but not yet scheduled
into an iteration. Tasks are displayed ordered by creation date (oldest first).

Examples:
  # Show backlog tasks
  dw task-manager task backlog

  # Show backlog for a specific project
  dw task-manager task backlog --project production

Notes:
  - Only shows tasks with status 'todo' or 'in-progress'
  - Excludes tasks that are in any iteration
  - Excludes tasks with status 'done'
  - Tasks are ordered by creation date (oldest first)`
}

func (c *TaskBacklogCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
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

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get backlog tasks
	tasks, err := repo.GetBacklogTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get backlog tasks: %w", err)
	}

	// Display results
	stdout := cmdCtx.GetStdout()

	if len(tasks) == 0 {
		fmt.Fprintf(stdout, "No backlog tasks found\n")
		return nil
	}

	// Print header
	fmt.Fprintf(stdout, "%-20s %-40s %-20s %-12s\n",
		"ID", "Title", "Track", "Status")
	fmt.Fprintf(stdout, "%s %s %s %s\n",
		strings.Repeat("-", 20), strings.Repeat("-", 40),
		strings.Repeat("-", 20), strings.Repeat("-", 12))

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

		// Truncate title if too long
		title := task.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		// Truncate track name if too long
		if len(trackName) > 20 {
			trackName = trackName[:17] + "..."
		}

		fmt.Fprintf(stdout, "%-20s %-40s %-20s %-12s\n",
			abbrevID, title, trackName, task.Status)
	}

	fmt.Fprintf(stdout, "\nTotal: %d backlog task(s)\n", len(tasks))
	return nil
}

// ============================================================================
// TaskCheckReadyCommand checks if a task is ready to be marked as done
// ============================================================================

type TaskCheckReadyCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	taskID  string
}

func (c *TaskCheckReadyCommand) GetName() string {
	return "task check-ready"
}

func (c *TaskCheckReadyCommand) GetDescription() string {
	return "Check if a task is ready to be marked as done"
}

func (c *TaskCheckReadyCommand) GetUsage() string {
	return "dw task-manager task check-ready <task-id>"
}

func (c *TaskCheckReadyCommand) GetHelp() string {
	return `Checks if a task is ready to be marked as done by verifying
all acceptance criteria have been verified.

Returns success if all ACs are verified, or an error listing unverified ACs.

Examples:
  # Check if task is ready
  dw task-manager task check-ready DW-task-123

  # Command will show status of each AC
  # Example output:
  #   ✓ DW-ac-1: User can login
  #   ✓ DW-ac-2: Password validation works
  #   ⏸ DW-ac-3: 2FA enabled (Pending human review)

Notes:
  - All acceptance criteria must be verified to mark task as done
  - Failed ACs (✗) will block task completion
  - Pending review ACs (⏸) will block task completion`
}

func (c *TaskCheckReadyCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if c.taskID == "" && !strings.HasPrefix(args[i], "--") {
				c.taskID = args[i]
			}
		}
	}

	if c.taskID == "" {
		return fmt.Errorf("<task-id> is required")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Verify task exists
	task, err := repo.GetTask(ctx, c.taskID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: Task \"%s\" not found\n", c.taskID)
			return fmt.Errorf("task not found: %s", c.taskID)
		}
		return fmt.Errorf("failed to verify task: %w", err)
	}

	// Get ACs for task
	acs, err := repo.ListAC(ctx, c.taskID)
	if err != nil {
		return fmt.Errorf("failed to check acceptance criteria: %w", err)
	}

	// If no ACs, task is ready
	if len(acs) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "Task is ready to be marked as done\n")
		fmt.Fprintf(cmdCtx.GetStdout(), "  %s - No acceptance criteria\n", task.Title)
		return nil
	}

	// Check verification status
	var verifiedCount, unverifiedCount int
	var unverifiedAC []*AcceptanceCriteriaEntity

	for _, ac := range acs {
		if ac.IsVerified() {
			verifiedCount++
		} else if ac.Status == ACStatusFailed {
			unverifiedAC = append(unverifiedAC, ac)
			unverifiedCount++
		} else {
			unverifiedAC = append(unverifiedAC, ac)
			unverifiedCount++
		}
	}

	// Display results
	fmt.Fprintf(cmdCtx.GetStdout(), "Task: %s (%s)\n", task.Title, task.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "Acceptance Criteria: %d/%d verified\n\n", verifiedCount, len(acs))

	for _, ac := range acs {
		statusIcon := ac.StatusIndicator()
		fmt.Fprintf(cmdCtx.GetStdout(), "%s [%s] %s\n", statusIcon, ac.ID, ac.Description)
	}

	// Return error if any unverified
	if len(unverifiedAC) > 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "\nTask is NOT ready - %d acceptance criteria not verified\n", len(unverifiedAC))
		fmt.Fprintf(cmdCtx.GetStdout(), "Hint: Use 'dw task-manager ac verify <ac-id>' to verify criteria\n")
		return fmt.Errorf("task not ready: %d acceptance criteria not verified", len(unverifiedAC))
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "\nTask is ready to be marked as done\n")
	return nil
}
