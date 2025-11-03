package task_manager

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// IterationCreateCommand creates a new iteration
// ============================================================================

type IterationCreateCommand struct {
	Plugin      *TaskManagerPlugin
	project string
	name        string
	goal        string
	deliverable string
}

func (c *IterationCreateCommand) GetName() string {
	return "iteration create"
}

func (c *IterationCreateCommand) GetDescription() string {
	return "Create a new iteration"
}

func (c *IterationCreateCommand) GetUsage() string {
	return "dw task-manager iteration create --name <name> --goal <goal> [--deliverable <deliverable>]"
}

func (c *IterationCreateCommand) GetHelp() string {
	return `Creates a new iteration with auto-incrementing number.

An iteration is a time-boxed grouping of tasks for sprint planning.
Each iteration must have a name and goal. Deliverable is optional.

Flags:
  --name <name>              Iteration name (required)
  --goal <goal>              Iteration goal (required)
  --deliverable <output>     Expected deliverable (optional)

Examples:
  # Create a basic iteration
  dw task-manager iteration create \
    --name "Foundation Sprint" \
    --goal "Complete view-based analysis"

  # With deliverable
  dw task-manager iteration create \
    --name "Foundation Sprint" \
    --goal "Complete view-based analysis" \
    --deliverable "Plugin-agnostic analysis framework"

Notes:
  - Iteration number is auto-incremented starting from 1
  - Initial status is set to 'planned'
  - No tasks are added initially (use iteration add-task)
  - Only one iteration can have status 'current' at a time
  - Deadline support is planned for a future iteration (will require schema update)`
}

func (c *IterationCreateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--name":
			if i+1 < len(args) {
				c.name = args[i+1]
				i++
			}
		case "--goal":
			if i+1 < len(args) {
				c.goal = args[i+1]
				i++
			}
		case "--deliverable":
			if i+1 < len(args) {
				c.deliverable = args[i+1]
				i++
			}
		}
	}

	// Validate required flags
	if c.name == "" || c.goal == "" {
		return fmt.Errorf("--name and --goal are required")
	}

	// Auto-generate iteration number
	iterations, err := repo.ListIterations(ctx)
	if err != nil {
		return fmt.Errorf("failed to list iterations: %w", err)
	}

	number := 1
	if len(iterations) > 0 {
		for _, it := range iterations {
			if it.Number >= number {
				number = it.Number + 1
			}
		}
	}

	// Create iteration entity
	now := time.Now().UTC()
	iteration, err := NewIterationEntity(
		number,
		c.name,
		c.goal,
		c.deliverable,
		[]string{},
		"planned",
		500, // default rank (medium priority)
		time.Time{},
		time.Time{},
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create iteration entity: %w", err)
	}

	// Save to repository
	if err := repo.SaveIteration(ctx, iteration); err != nil {
		return fmt.Errorf("failed to save iteration: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Iteration created successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "Number:       %d\n", iteration.Number)
	fmt.Fprintf(cmdCtx.GetStdout(), "Name:         %s\n", iteration.Name)
	fmt.Fprintf(cmdCtx.GetStdout(), "Goal:         %s\n", iteration.Goal)
	if c.deliverable != "" {
		fmt.Fprintf(cmdCtx.GetStdout(), "Deliverable:  %s\n", iteration.Deliverable)
	}
	fmt.Fprintf(cmdCtx.GetStdout(), "Status:       %s\n", iteration.Status)

	return nil
}

// ============================================================================
// IterationListCommand lists all iterations
// ============================================================================

type IterationListCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *IterationListCommand) GetName() string {
	return "iteration list"
}

func (c *IterationListCommand) GetDescription() string {
	return "List all iterations"
}

func (c *IterationListCommand) GetUsage() string {
	return "dw task-manager iteration list"
}

func (c *IterationListCommand) GetHelp() string {
	return `Lists all iterations in the order they were created.

Each iteration shows its number, name, goal, status, task count, and timestamps.

Examples:
  dw task-manager iteration list

Notes:
  - Iterations are displayed in order by number
  - Current iteration is highlighted
  - Status values: planned, current, complete`
}

func (c *IterationListCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get all iterations
	iterations, err := repo.ListIterations(ctx)
	if err != nil {
		return fmt.Errorf("failed to list iterations: %w", err)
	}

	if len(iterations) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "No iterations found.\n")
		fmt.Fprintf(cmdCtx.GetStdout(), "Create one with 'dw task-manager iteration create --name <name> --goal <goal>'\n")
		return nil
	}

	// Display header
	fmt.Fprintf(cmdCtx.GetStdout(), "%-3s %-30s %-20s %-10s %-5s %-19s %-19s\n",
		"#", "Name", "Goal", "Status", "Tasks", "Started", "Completed")
	fmt.Fprintf(cmdCtx.GetStdout(), "%-3s %-30s %-20s %-10s %-5s %-19s %-19s\n",
		strings.Repeat("-", 3),
		strings.Repeat("-", 30),
		strings.Repeat("-", 20),
		strings.Repeat("-", 10),
		strings.Repeat("-", 5),
		strings.Repeat("-", 19),
		strings.Repeat("-", 19),
	)

	// Display iterations
	for _, iter := range iterations {
		startedStr := "-"
		if iter.StartedAt != nil {
			startedStr = iter.StartedAt.Format("2006-01-02 15:04")
		}

		completedStr := "-"
		if iter.CompletedAt != nil {
			completedStr = iter.CompletedAt.Format("2006-01-02 15:04")
		}

		// Truncate long strings for display
		name := iter.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		goal := iter.Goal
		if len(goal) > 20 {
			goal = goal[:17] + "..."
		}

		fmt.Fprintf(cmdCtx.GetStdout(), "%-3d %-30s %-20s %-10s %-5d %-19s %-19s\n",
			iter.Number,
			name,
			goal,
			iter.Status,
			len(iter.TaskIDs),
			startedStr,
			completedStr,
		)
	}

	return nil
}

// ============================================================================
// IterationShowCommand displays a specific iteration
// ============================================================================

type IterationShowCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	full    bool
}

func (c *IterationShowCommand) GetName() string {
	return "iteration show"
}

func (c *IterationShowCommand) GetDescription() string {
	return "Display a specific iteration"
}

func (c *IterationShowCommand) GetUsage() string {
	return "dw task-manager iteration show <number> [--full]"
}

func (c *IterationShowCommand) GetHelp() string {
	return `Displays detailed information about a specific iteration.

Shows the iteration's properties, timestamps, and all associated tasks
with their status breakdown.

Arguments:
  <number>  Iteration number (required)

Flags:
  --full    Show full task titles and descriptions (default: truncated)

Examples:
  dw task-manager iteration show 1
  dw task-manager iteration show 2 --full

Notes:
  - Run 'dw task-manager iteration list' to see all iteration numbers
  - Task counts show completed/total breakdown
  - Use --full to see complete task titles and descriptions`
}

func (c *IterationShowCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	c.full = false
	iterationNum := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--full":
			c.full = true
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "--") && iterationNum == "" {
				iterationNum = args[i]
			}
		}
	}

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Parse iteration number
	if iterationNum == "" {
		return fmt.Errorf("iteration number is required")
	}

	number, err := strconv.Atoi(iterationNum)
	if err != nil {
		return fmt.Errorf("invalid iteration number: %v", err)
	}

	// Get iteration
	iteration, err := repo.GetIteration(ctx, number)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d not found.\n", number)
			return nil
		}
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Get iteration tasks
	tasks, err := repo.GetIterationTasks(ctx, number)
	if err != nil {
		return fmt.Errorf("failed to get iteration tasks: %w", err)
	}

	// Display iteration details
	fmt.Fprintf(cmdCtx.GetStdout(), "Iteration #%d: %s\n", iteration.Number, iteration.Name)
	fmt.Fprintf(cmdCtx.GetStdout(), "===============================\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "Goal:          %s\n", iteration.Goal)
	if iteration.Deliverable != "" {
		fmt.Fprintf(cmdCtx.GetStdout(), "Deliverable:   %s\n", iteration.Deliverable)
	}
	fmt.Fprintf(cmdCtx.GetStdout(), "Status:        %s\n", iteration.Status)
	fmt.Fprintf(cmdCtx.GetStdout(), "Created:       %s\n", iteration.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(cmdCtx.GetStdout(), "Updated:       %s\n", iteration.UpdatedAt.Format(time.RFC3339))

	if iteration.StartedAt != nil {
		fmt.Fprintf(cmdCtx.GetStdout(), "Started:       %s\n", iteration.StartedAt.Format(time.RFC3339))
	}
	if iteration.CompletedAt != nil {
		fmt.Fprintf(cmdCtx.GetStdout(), "Completed:     %s\n", iteration.CompletedAt.Format(time.RFC3339))
	}

	// Display task information as list blocks
	fmt.Fprintf(cmdCtx.GetStdout(), "\nTasks: %d total\n", len(iteration.TaskIDs))
	if len(tasks) > 0 {
		completedCount := 0
		for _, task := range tasks {
			if task.Status == "done" {
				completedCount++
			}

			// Display as list block
			fmt.Fprintf(cmdCtx.GetStdout(), "\n- %s: %s\n", task.ID, task.Title)
			fmt.Fprintf(cmdCtx.GetStdout(), "  Status: %s\n", task.Status)

			// Show description if --full flag is set
			if c.full && task.Description != "" {
				fmt.Fprintf(cmdCtx.GetStdout(), "  Description: %s\n", task.Description)
			}
		}

		fmt.Fprintf(cmdCtx.GetStdout(), "\nProgress: %d/%d tasks completed (%.0f%%)\n",
			completedCount,
			len(tasks),
			float64(completedCount)/float64(len(tasks))*100,
		)
	} else {
		fmt.Fprintf(cmdCtx.GetStdout(), "No tasks in this iteration.\n")
	}

	return nil
}

// ============================================================================
// IterationCurrentCommand displays the current iteration
// ============================================================================

type IterationCurrentCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	full    bool
}

func (c *IterationCurrentCommand) GetName() string {
	return "iteration current"
}

func (c *IterationCurrentCommand) GetDescription() string {
	return "Display the current iteration"
}

func (c *IterationCurrentCommand) GetUsage() string {
	return "dw task-manager iteration current [--full]"
}

func (c *IterationCurrentCommand) GetHelp() string {
	return `Displays the current active iteration (status: current).

If no iteration is currently active, provides guidance on how to start one.
By default, only shows non-completed tasks.

Flags:
  --full    Show task descriptions (default: only titles)

Examples:
  dw task-manager iteration current
  dw task-manager iteration current --full

Notes:
  - Only one iteration can be current at a time
  - Only non-completed tasks (todo, in-progress) are shown by default
  - Start an iteration with 'dw task-manager iteration start <number>'
  - Complete current iteration with 'dw task-manager iteration complete'`
}

func (c *IterationCurrentCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	c.full = false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--full":
			c.full = true
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

	// Get current iteration
	iteration, err := repo.GetCurrentIteration(ctx)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			// No current iteration - show next planned iteration (ordered by rank)
			iterations, err := repo.ListIterations(ctx)
			if err != nil {
				return fmt.Errorf("failed to list iterations: %w", err)
			}

			// Filter for planned iterations
			var plannedIterations []*IterationEntity
			for _, iter := range iterations {
				if iter.Status == "planned" {
					plannedIterations = append(plannedIterations, iter)
				}
			}

			// Sort by rank (lowest rank = highest priority)
			sort.Slice(plannedIterations, func(i, j int) bool {
				return plannedIterations[i].Rank < plannedIterations[j].Rank
			})

			fmt.Fprintf(cmdCtx.GetStdout(), "No current iteration.\n\n")

			if len(plannedIterations) > 0 {
				// Show only the next iteration (highest priority)
				iter := plannedIterations[0]

				// Get tasks for progress calculation
				tasks, err := repo.GetIterationTasks(ctx, iter.Number)
				if err != nil {
					return fmt.Errorf("failed to get iteration tasks: %w", err)
				}

				completedCount := 0
				for _, task := range tasks {
					if task.Status == "done" {
						completedCount++
					}
				}

				completePct := 0
				if len(tasks) > 0 {
					completePct = (completedCount * 100) / len(tasks)
				}

				fmt.Fprintf(cmdCtx.GetStdout(), "Next planned iteration:\n")
				fmt.Fprintf(cmdCtx.GetStdout(), "#%d: %s - %s (%d tasks, %d%% complete)\n",
					iter.Number, iter.Name, iter.Goal, len(tasks), completePct)

				fmt.Fprintf(cmdCtx.GetStdout(), "\nTo review this iteration:\n")
				fmt.Fprintf(cmdCtx.GetStdout(), "  dw task-manager iteration show %d           # View tasks\n", iter.Number)
				fmt.Fprintf(cmdCtx.GetStdout(), "  dw task-manager ac list-iteration %d        # View all acceptance criteria\n", iter.Number)
				fmt.Fprintf(cmdCtx.GetStdout(), "\nTo start working:\n")
				fmt.Fprintf(cmdCtx.GetStdout(), "  dw task-manager iteration start %d          # Start this iteration\n", iter.Number)
			} else {
				fmt.Fprintf(cmdCtx.GetStdout(), "No planned iterations available.\n")
				fmt.Fprintf(cmdCtx.GetStdout(), "Hint: Use 'dw task-manager iteration create' to create an iteration\n")
			}

			return nil
		}
		return fmt.Errorf("failed to get current iteration: %w", err)
	}

	// Get iteration tasks
	tasks, err := repo.GetIterationTasks(ctx, iteration.Number)
	if err != nil {
		return fmt.Errorf("failed to get iteration tasks: %w", err)
	}

	// Filter non-completed tasks (todo, in-progress only)
	var activeTasks []*TaskEntity
	completedCount := 0
	for _, task := range tasks {
		if task.Status == "done" {
			completedCount++
		} else {
			activeTasks = append(activeTasks, task)
		}
	}

	// Display current iteration details
	fmt.Fprintf(cmdCtx.GetStdout(), "Current Iteration: #%d: %s\n", iteration.Number, iteration.Name)
	fmt.Fprintf(cmdCtx.GetStdout(), "===============================\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "Goal:          %s\n", iteration.Goal)
	if iteration.Deliverable != "" {
		fmt.Fprintf(cmdCtx.GetStdout(), "Deliverable:   %s\n", iteration.Deliverable)
	}
	fmt.Fprintf(cmdCtx.GetStdout(), "Status:        %s\n", iteration.Status)
	fmt.Fprintf(cmdCtx.GetStdout(), "Created:       %s\n", iteration.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(cmdCtx.GetStdout(), "Updated:       %s\n", iteration.UpdatedAt.Format(time.RFC3339))

	if iteration.StartedAt != nil {
		fmt.Fprintf(cmdCtx.GetStdout(), "Started:       %s\n", iteration.StartedAt.Format(time.RFC3339))
	}

	// Display active (non-completed) tasks
	fmt.Fprintf(cmdCtx.GetStdout(), "\nActive Tasks: %d (of %d total)\n", len(activeTasks), len(tasks))
	if len(activeTasks) > 0 {
		for _, task := range activeTasks {
			// Display as list block
			fmt.Fprintf(cmdCtx.GetStdout(), "\n- %s: %s\n", task.ID, task.Title)
			fmt.Fprintf(cmdCtx.GetStdout(), "  Status: %s\n", task.Status)

			// Show description if --full flag is set
			if c.full && task.Description != "" {
				fmt.Fprintf(cmdCtx.GetStdout(), "  Description: %s\n", task.Description)
			}
		}

		fmt.Fprintf(cmdCtx.GetStdout(), "\nProgress: %d/%d tasks completed (%.0f%%)\n",
			completedCount,
			len(tasks),
			float64(completedCount)/float64(len(tasks))*100,
		)
	} else if len(tasks) > 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "All tasks completed! Use 'dw task-manager iteration complete %d' to finish this iteration.\n", iteration.Number)
	} else {
		fmt.Fprintf(cmdCtx.GetStdout(), "No tasks in current iteration.\n")
	}

	return nil
}

// ============================================================================
// IterationUpdateCommand updates an iteration
// ============================================================================

type IterationUpdateCommand struct {
	Plugin      *TaskManagerPlugin
	project string
	name        *string
	goal        *string
	deliverable *string
	status      *string
}

func (c *IterationUpdateCommand) GetName() string {
	return "iteration update"
}

func (c *IterationUpdateCommand) GetDescription() string {
	return "Update an iteration"
}

func (c *IterationUpdateCommand) GetUsage() string {
	return "dw task-manager iteration update <number> [--name <name>] [--goal <goal>] [--deliverable <deliverable>] [--status <status>]"
}

func (c *IterationUpdateCommand) GetHelp() string {
	return `Updates properties of a specific iteration.

At least one flag must be provided to update.

Arguments:
  <number>  Iteration number (required)

Flags:
  --name <name>              New iteration name
  --goal <goal>              New iteration goal
  --deliverable <output>     New expected deliverable
  --status <status>          New status (planned, current, complete)

Examples:
  # Update name
  dw task-manager iteration update 1 --name "Sprint 1"

  # Update multiple fields
  dw task-manager iteration update 1 \
    --name "Sprint 1" \
    --goal "Complete framework"

  # Reset status to planned
  dw task-manager iteration update 1 --status planned

Notes:
  - At least one flag is required
  - Updated_at timestamp is automatically updated
  - Valid status values: planned, current, complete
  - Use start/complete commands for typical workflow (recommended)
  - Direct status updates are for manual corrections/resets`
}

func (c *IterationUpdateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Parse iteration number
	if len(args) == 0 {
		return fmt.Errorf("iteration number is required")
	}

	number, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid iteration number: %v", err)
	}

	// Parse flags
	c.name = nil
	c.goal = nil
	c.deliverable = nil
	c.status = nil

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 < len(args) {
				v := args[i+1]
				c.name = &v
				i++
			}
		case "--goal":
			if i+1 < len(args) {
				v := args[i+1]
				c.goal = &v
				i++
			}
		case "--deliverable":
			if i+1 < len(args) {
				v := args[i+1]
				c.deliverable = &v
				i++
			}
		case "--status":
			if i+1 < len(args) {
				v := args[i+1]
				c.status = &v
				i++
			}
		}
	}

	// At least one flag must be provided
	if c.name == nil && c.goal == nil && c.deliverable == nil && c.status == nil {
		return fmt.Errorf("at least one flag is required (--name, --goal, --deliverable, or --status)")
	}

	// Validate status if provided
	if c.status != nil {
		validStatuses := map[string]bool{
			"planned":  true,
			"current":  true,
			"complete": true,
		}
		if !validStatuses[*c.status] {
			return fmt.Errorf("invalid status '%s'; must be one of: planned, current, complete", *c.status)
		}
	}

	// Get iteration
	iteration, err := repo.GetIteration(ctx, number)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d not found.\n", number)
			return nil
		}
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Update provided fields
	if c.name != nil {
		iteration.Name = *c.name
	}
	if c.goal != nil {
		iteration.Goal = *c.goal
	}
	if c.deliverable != nil {
		iteration.Deliverable = *c.deliverable
	}
	if c.status != nil {
		iteration.Status = *c.status
		// Update timestamps based on status changes
		now := time.Now().UTC()
		if *c.status == "current" && iteration.StartedAt == nil {
			iteration.StartedAt = &now
		}
		if *c.status == "complete" && iteration.CompletedAt == nil {
			iteration.CompletedAt = &now
		}
		// Reset timestamps if moving back to planned
		if *c.status == "planned" {
			iteration.StartedAt = nil
			iteration.CompletedAt = nil
		}
	}
	iteration.UpdatedAt = time.Now().UTC()

	// Save to repository
	if err := repo.UpdateIteration(ctx, iteration); err != nil {
		return fmt.Errorf("failed to update iteration: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Iteration updated successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "Number:       %d\n", iteration.Number)
	fmt.Fprintf(cmdCtx.GetStdout(), "Name:         %s\n", iteration.Name)
	fmt.Fprintf(cmdCtx.GetStdout(), "Goal:         %s\n", iteration.Goal)
	if iteration.Deliverable != "" {
		fmt.Fprintf(cmdCtx.GetStdout(), "Deliverable:  %s\n", iteration.Deliverable)
	}
	fmt.Fprintf(cmdCtx.GetStdout(), "Status:       %s\n", iteration.Status)
	fmt.Fprintf(cmdCtx.GetStdout(), "Updated:      %s\n", iteration.UpdatedAt.Format(time.RFC3339))

	return nil
}

// ============================================================================
// IterationDeleteCommand deletes an iteration
// ============================================================================

type IterationDeleteCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	force  bool
}

func (c *IterationDeleteCommand) GetName() string {
	return "iteration delete"
}

func (c *IterationDeleteCommand) GetDescription() string {
	return "Delete an iteration"
}

func (c *IterationDeleteCommand) GetUsage() string {
	return "dw task-manager iteration delete <number> [--force]"
}

func (c *IterationDeleteCommand) GetHelp() string {
	return `Deletes a specific iteration.

By default, will prompt for confirmation. Use --force to skip confirmation.

Arguments:
  <number>  Iteration number (required)

Flags:
  --force   Skip confirmation prompt

Examples:
  # Delete with confirmation
  dw task-manager iteration delete 1

  # Delete without confirmation
  dw task-manager iteration delete 1 --force

Notes:
  - Deletion is permanent
  - Only the iteration is deleted; tasks remain and lose iteration association
  - Cannot delete the current iteration`
}

func (c *IterationDeleteCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Parse iteration number and flags
	if len(args) == 0 {
		return fmt.Errorf("iteration number is required")
	}

	number, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid iteration number: %v", err)
	}

	c.force = false
	for i := 1; i < len(args); i++ {
		if args[i] == "--force" {
			c.force = true
		}
	}

	// Get iteration to verify it exists
	iteration, err := repo.GetIteration(ctx, number)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d not found.\n", number)
			return nil
		}
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Prompt for confirmation unless --force
	if !c.force {
		fmt.Fprintf(cmdCtx.GetStdout(), "Delete iteration #%d '%s'? (y/n) ", number, iteration.Name)
		scanner := bufio.NewScanner(cmdCtx.GetStdin())
		if !scanner.Scan() {
			fmt.Fprintf(cmdCtx.GetStdout(), "Cancelled.\n")
			return nil
		}
		response := strings.TrimSpace(scanner.Text())
		if response != "y" && response != "yes" {
			fmt.Fprintf(cmdCtx.GetStdout(), "Cancelled.\n")
			return nil
		}
	}

	// Delete iteration
	if err := repo.DeleteIteration(ctx, number); err != nil {
		return fmt.Errorf("failed to delete iteration: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Iteration #%d deleted successfully\n", number)

	return nil
}

// ============================================================================
// IterationAddTaskCommand adds tasks to an iteration
// ============================================================================

type IterationAddTaskCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *IterationAddTaskCommand) GetName() string {
	return "iteration add-task"
}

func (c *IterationAddTaskCommand) GetDescription() string {
	return "Add tasks to an iteration"
}

func (c *IterationAddTaskCommand) GetUsage() string {
	return "dw task-manager iteration add-task <number> <task-id> [<task-id> ...]"
}

func (c *IterationAddTaskCommand) GetHelp() string {
	return `Adds one or more tasks to a specific iteration.

Arguments:
  <number>      Iteration number (required)
  <task-id>     Task ID (required, can specify multiple)

Examples:
  # Add single task
  dw task-manager iteration add-task 1 DW-task-1

  # Add multiple tasks in one command
  dw task-manager iteration add-task 1 DW-task-1 DW-task-2 DW-task-3

Notes:
  - Multiple tasks can be added in a single command
  - All task IDs must exist
  - Tasks already in the iteration will be skipped with a warning
  - Run 'dw task-manager task list' to see available tasks`
}

func (c *IterationAddTaskCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Parse iteration number and task IDs
	if len(args) < 2 {
		return fmt.Errorf("iteration number and at least one task ID are required")
	}

	number, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid iteration number: %v", err)
	}

	taskIDs := args[1:]

	// Verify iteration exists
	_, err = repo.GetIteration(ctx, number)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d not found.\n", number)
			return nil
		}
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Verify all tasks exist
	for _, taskID := range taskIDs {
		_, err = repo.GetTask(ctx, taskID)
		if err != nil {
			if errors.Is(err, pluginsdk.ErrNotFound) {
				fmt.Fprintf(cmdCtx.GetStdout(), "Task %s not found.\n", taskID)
				return nil
			}
			return fmt.Errorf("failed to get task %s: %w", taskID, err)
		}
	}

	// Add tasks to iteration
	addedCount := 0
	for _, taskID := range taskIDs {
		if err := repo.AddTaskToIteration(ctx, number, taskID); err != nil {
			if errors.Is(err, pluginsdk.ErrAlreadyExists) {
				fmt.Fprintf(cmdCtx.GetStdout(), "Task %s already in iteration #%d (skipped)\n", taskID, number)
				continue
			}
			return fmt.Errorf("failed to add task %s: %w", taskID, err)
		}
		addedCount++
	}

	if addedCount == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "No tasks added (all already in iteration).\n")
	} else {
		fmt.Fprintf(cmdCtx.GetStdout(), "%d task(s) added to iteration #%d\n", addedCount, number)
	}

	return nil
}

// ============================================================================
// IterationRemoveTaskCommand removes a task from an iteration
// ============================================================================

type IterationRemoveTaskCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *IterationRemoveTaskCommand) GetName() string {
	return "iteration remove-task"
}

func (c *IterationRemoveTaskCommand) GetDescription() string {
	return "Remove a task from an iteration"
}

func (c *IterationRemoveTaskCommand) GetUsage() string {
	return "dw task-manager iteration remove-task <number> <task-id>"
}

func (c *IterationRemoveTaskCommand) GetHelp() string {
	return `Removes a specific task from an iteration.

Arguments:
  <number>    Iteration number (required)
  <task-id>   Task ID (required)

Examples:
  dw task-manager iteration remove-task 1 task-fc-001

Notes:
  - Only removes the task from the iteration; task itself is not deleted
  - Run 'dw task-manager iteration show <number>' to see tasks in iteration`
}

func (c *IterationRemoveTaskCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Parse iteration number and task ID
	if len(args) < 2 {
		return fmt.Errorf("iteration number and task ID are required")
	}

	number, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid iteration number: %v", err)
	}

	taskID := args[1]

	// Remove task from iteration
	if err := repo.RemoveTaskFromIteration(ctx, number, taskID); err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Task %s not found in iteration %d.\n", taskID, number)
			return nil
		}
		return fmt.Errorf("failed to remove task: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Task %s removed from iteration #%d\n", taskID, number)

	return nil
}

// ============================================================================
// IterationStartCommand starts an iteration
// ============================================================================

type IterationStartCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *IterationStartCommand) GetName() string {
	return "iteration start"
}

func (c *IterationStartCommand) GetDescription() string {
	return "Start an iteration"
}

func (c *IterationStartCommand) GetUsage() string {
	return "dw task-manager iteration start <number>"
}

func (c *IterationStartCommand) GetHelp() string {
	return `Marks an iteration as current and sets the started timestamp.

Arguments:
  <number>  Iteration number (required)

Examples:
  dw task-manager iteration start 1

Notes:
  - Only one iteration can be current at a time
  - If another iteration is current, it will be marked as planned
  - Started timestamp is automatically set
  - Use 'dw task-manager iteration complete' to mark as finished`
}

func (c *IterationStartCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Parse iteration number
	if len(args) == 0 {
		return fmt.Errorf("iteration number is required")
	}

	number, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid iteration number: %v", err)
	}

	// Start iteration
	if err := repo.StartIteration(ctx, number); err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d not found.\n", number)
			return nil
		}
		if errors.Is(err, pluginsdk.ErrInvalidArgument) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Cannot start iteration: %v\n", err)
			return nil
		}
		return fmt.Errorf("failed to start iteration: %w", err)
	}

	// Get updated iteration to display
	iteration, err := repo.GetIteration(ctx, number)
	if err != nil {
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Iteration #%d started\n", number)
	fmt.Fprintf(cmdCtx.GetStdout(), "Name:    %s\n", iteration.Name)
	fmt.Fprintf(cmdCtx.GetStdout(), "Status:  %s\n", iteration.Status)
	if iteration.StartedAt != nil {
		fmt.Fprintf(cmdCtx.GetStdout(), "Started: %s\n", iteration.StartedAt.Format(time.RFC3339))
	}

	return nil
}

// ============================================================================
// IterationCompleteCommand completes an iteration
// ============================================================================

type IterationCompleteCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *IterationCompleteCommand) GetName() string {
	return "iteration complete"
}

func (c *IterationCompleteCommand) GetDescription() string {
	return "Complete an iteration"
}

func (c *IterationCompleteCommand) GetUsage() string {
	return "dw task-manager iteration complete <number>"
}

func (c *IterationCompleteCommand) GetHelp() string {
	return `Marks an iteration as complete and sets the completed timestamp.

Arguments:
  <number>  Iteration number (required)

Examples:
  dw task-manager iteration complete 1

Notes:
  - Iteration must be in current status to complete
  - Completed timestamp is automatically set
  - Use 'dw task-manager iteration start' to begin a new iteration
  - Completed iteration can still be viewed but not modified`
}

func (c *IterationCompleteCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Parse iteration number
	if len(args) == 0 {
		return fmt.Errorf("iteration number is required")
	}

	number, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid iteration number: %v", err)
	}

	// Complete iteration
	if err := repo.CompleteIteration(ctx, number); err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d not found.\n", number)
			return nil
		}
		if errors.Is(err, pluginsdk.ErrInvalidArgument) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Cannot complete iteration: %v\n", err)
			return nil
		}
		return fmt.Errorf("failed to complete iteration: %w", err)
	}

	// Get updated iteration
	iteration, err := repo.GetIteration(ctx, number)
	if err != nil {
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Get tasks for summary
	tasks, err := repo.GetIterationTasks(ctx, number)
	if err != nil {
		return fmt.Errorf("failed to get iteration tasks: %w", err)
	}

	// Calculate task summary
	completedCount := 0
	for _, task := range tasks {
		if task.Status == "done" {
			completedCount++
		}
	}

	// Calculate duration
	durationStr := "-"
	if iteration.StartedAt != nil && iteration.CompletedAt != nil {
		duration := iteration.CompletedAt.Sub(*iteration.StartedAt)
		durationStr = fmt.Sprintf("%.0f hours", duration.Hours())
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Iteration #%d completed\n", number)
	fmt.Fprintf(cmdCtx.GetStdout(), "Name:       %s\n", iteration.Name)
	fmt.Fprintf(cmdCtx.GetStdout(), "Status:     %s\n", iteration.Status)
	fmt.Fprintf(cmdCtx.GetStdout(), "Tasks:      %d/%d completed\n", completedCount, len(tasks))
	if iteration.CompletedAt != nil {
		fmt.Fprintf(cmdCtx.GetStdout(), "Completed:  %s\n", iteration.CompletedAt.Format(time.RFC3339))
	}
	if iteration.StartedAt != nil {
		fmt.Fprintf(cmdCtx.GetStdout(), "Duration:   %s\n", durationStr)
	}

	return nil
}
