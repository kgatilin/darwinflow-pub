package task_manager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// ACListIterationCommand lists all ACs for an iteration grouped by task
// ============================================================================

type ACListIterationCommand struct {
	Plugin    *TaskManagerPlugin
	project   string
	iteration int
}

func (c *ACListIterationCommand) GetName() string {
	return "ac list-iteration"
}

func (c *ACListIterationCommand) GetDescription() string {
	return "List all acceptance criteria for an iteration"
}

func (c *ACListIterationCommand) GetUsage() string {
	return "dw task-manager ac list-iteration <iteration-number>"
}

func (c *ACListIterationCommand) GetHelp() string {
	return `Lists all acceptance criteria for all tasks in an iteration,
grouped by task with status indicators.

Status indicators:
  ✓   Verified (manually or automatically)
  ⏸   Pending human review
  ○   Not started
  ✗   Failed

Examples:
  # List ACs for iteration 1
  dw task-manager ac list-iteration 1

Notes:
  - Shows each task and its ACs
  - Summary shows overall verification progress`
}

func (c *ACListIterationCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if c.iteration == 0 && !strings.HasPrefix(args[i], "--") {
				// Parse iteration number
				_, err := fmt.Sscanf(args[i], "%d", &c.iteration)
				if err != nil {
					return fmt.Errorf("invalid iteration number: %s", args[i])
				}
			}
		}
	}

	if c.iteration == 0 {
		return fmt.Errorf("<iteration-number> is required")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Verify iteration exists
	iter, err := repo.GetIteration(ctx, c.iteration)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: Iteration %d not found\n", c.iteration)
			return fmt.Errorf("iteration not found: %d", c.iteration)
		}
		return fmt.Errorf("failed to get iteration: %w", err)
	}

	// Get tasks in iteration
	tasks, err := repo.GetIterationTasks(ctx, c.iteration)
	if err != nil {
		return fmt.Errorf("failed to get iteration tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d has no tasks\n", c.iteration)
		return nil
	}

	// Get ACs for iteration
	acs, err := repo.ListACByIteration(ctx, c.iteration)
	if err != nil {
		return fmt.Errorf("failed to get ACs: %w", err)
	}

	// Count verification status
	var verifiedCount, pendingCount, failedCount, notStartedCount int
	for _, ac := range acs {
		switch ac.Status {
		case ACStatusVerified, ACStatusAutomaticallyVerified:
			verifiedCount++
		case ACStatusPendingHumanReview:
			pendingCount++
		case ACStatusFailed:
			failedCount++
		default:
			notStartedCount++
		}
	}

	// Group ACs by task
	acsByTask := make(map[string][]*AcceptanceCriteriaEntity)
	for _, ac := range acs {
		acsByTask[ac.TaskID] = append(acsByTask[ac.TaskID], ac)
	}

	// Display results
	fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d: %s\n", iter.Number, iter.Name)
	fmt.Fprintf(cmdCtx.GetStdout(), "Status: %s\n", iter.Status)
	fmt.Fprintf(cmdCtx.GetStdout(), "\nAcceptance Criteria Summary:\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ✓  Verified:        %d\n", verifiedCount)
	fmt.Fprintf(cmdCtx.GetStdout(), "  ⏸  Pending Review:  %d\n", pendingCount)
	fmt.Fprintf(cmdCtx.GetStdout(), "  ✗  Failed:          %d\n", failedCount)
	fmt.Fprintf(cmdCtx.GetStdout(), "  ○  Not Started:     %d\n", notStartedCount)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Total:              %d\n", len(acs))

	fmt.Fprintf(cmdCtx.GetStdout(), "\nAcceptance Criteria by Task:\n\n")

	for _, task := range tasks {
		taskACs := acsByTask[task.ID]
		if len(taskACs) == 0 {
			continue
		}

		fmt.Fprintf(cmdCtx.GetStdout(), "Task: %s (%s)\n", task.Title, task.ID)

		for _, ac := range taskACs {
			statusIcon := ac.StatusIndicator()
			fmt.Fprintf(cmdCtx.GetStdout(), "  %s [%s] %s\n", statusIcon, ac.ID, ac.Description)
		}
		fmt.Fprintf(cmdCtx.GetStdout(), "\n")
	}

	return nil
}

// ============================================================================
// ACListTrackCommand lists all ACs for a track grouped by task
// ============================================================================

type ACListTrackCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	trackID string
}

func (c *ACListTrackCommand) GetName() string {
	return "ac list-track"
}

func (c *ACListTrackCommand) GetDescription() string {
	return "List all acceptance criteria for a track"
}

func (c *ACListTrackCommand) GetUsage() string {
	return "dw task-manager ac list-track <track-id>"
}

func (c *ACListTrackCommand) GetHelp() string {
	return `Lists all acceptance criteria for all tasks in a track,
grouped by task with status indicators.

Status indicators:
  ✓   Verified (manually or automatically)
  ⏸   Pending human review
  ○   Not started
  ✗   Failed

Examples:
  # List ACs for a track
  dw task-manager ac list-track track-core-framework

Notes:
  - Shows each task and its ACs
  - Summary shows overall verification progress`
}

func (c *ACListTrackCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if c.trackID == "" && !strings.HasPrefix(args[i], "--") {
				c.trackID = args[i]
			}
		}
	}

	if c.trackID == "" {
		return fmt.Errorf("<track-id> is required")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Verify track exists
	track, err := repo.GetTrack(ctx, c.trackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: Track \"%s\" not found\n", c.trackID)
			return fmt.Errorf("track not found: %s", c.trackID)
		}
		return fmt.Errorf("failed to get track: %w", err)
	}

	// Get tasks in track
	tasks, err := repo.ListTasks(ctx, TaskFilters{TrackID: c.trackID})
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "Track %s has no tasks\n", c.trackID)
		return nil
	}

	// Get ACs for track
	acs, err := repo.ListACByTrack(ctx, c.trackID)
	if err != nil {
		return fmt.Errorf("failed to get ACs: %w", err)
	}

	// Count verification status
	var verifiedCount, pendingCount, failedCount, notStartedCount int
	for _, ac := range acs {
		switch ac.Status {
		case ACStatusVerified, ACStatusAutomaticallyVerified:
			verifiedCount++
		case ACStatusPendingHumanReview:
			pendingCount++
		case ACStatusFailed:
			failedCount++
		default:
			notStartedCount++
		}
	}

	// Group ACs by task
	acsByTask := make(map[string][]*AcceptanceCriteriaEntity)
	for _, ac := range acs {
		acsByTask[ac.TaskID] = append(acsByTask[ac.TaskID], ac)
	}

	// Display results
	fmt.Fprintf(cmdCtx.GetStdout(), "Track: %s\n", track.Title)
	fmt.Fprintf(cmdCtx.GetStdout(), "ID: %s\n", track.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "\nAcceptance Criteria Summary:\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ✓  Verified:        %d\n", verifiedCount)
	fmt.Fprintf(cmdCtx.GetStdout(), "  ⏸  Pending Review:  %d\n", pendingCount)
	fmt.Fprintf(cmdCtx.GetStdout(), "  ✗  Failed:          %d\n", failedCount)
	fmt.Fprintf(cmdCtx.GetStdout(), "  ○  Not Started:     %d\n", notStartedCount)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Total:              %d\n", len(acs))

	fmt.Fprintf(cmdCtx.GetStdout(), "\nAcceptance Criteria by Task:\n\n")

	for _, task := range tasks {
		taskACs := acsByTask[task.ID]
		if len(taskACs) == 0 {
			fmt.Fprintf(cmdCtx.GetStdout(), "Task: %s (%s) - No acceptance criteria\n\n", task.Title, task.ID)
			continue
		}

		fmt.Fprintf(cmdCtx.GetStdout(), "Task: %s (%s)\n", task.Title, task.ID)

		for _, ac := range taskACs {
			statusIcon := ac.StatusIndicator()
			fmt.Fprintf(cmdCtx.GetStdout(), "  %s [%s] %s\n", statusIcon, ac.ID, ac.Description)
		}
		fmt.Fprintf(cmdCtx.GetStdout(), "\n")
	}

	return nil
}
