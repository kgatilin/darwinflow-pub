package task_manager

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// ACAddCommand adds a new acceptance criterion to a task
// ============================================================================

type ACAddCommand struct {
	Plugin           *TaskManagerPlugin
	project          string
	taskID           string
	description      string
	verificationType string
}

func (c *ACAddCommand) GetName() string {
	return "ac add"
}

func (c *ACAddCommand) GetDescription() string {
	return "Add a new acceptance criterion to a task"
}

func (c *ACAddCommand) GetUsage() string {
	return "dw task-manager ac add <task-id> --description \"...\" [--type manual|automated]"
}

func (c *ACAddCommand) GetHelp() string {
	return `Adds a new acceptance criterion to a task.

An acceptance criterion defines what must be verified before a task can be
considered complete. It can be manually verified by a human or automatically
verified by a coding agent.

Flags:
  <task-id>                 Task ID to add AC to (required)
  --description "..."       Description of what must be verified (required)
  --type <type>            Verification type (optional, default: manual)
                           Values: manual, automated

Examples:
  # Add a manual verification criterion
  dw task-manager ac add DW-task-123 \
    --description "User can create tasks"

  # Add an automated verification criterion
  dw task-manager ac add DW-task-123 \
    --description "Tests pass with 80% coverage" \
    --type automated

Notes:
  - Initial status is automatically set to "not_started"
  - AC ID is generated automatically`
}

func (c *ACAddCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	c.verificationType = "manual" // default
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				c.description = args[i+1]
				i++
			}
		case "--type":
			if i+1 < len(args) {
				c.verificationType = args[i+1]
				i++
			}
		default:
			if c.taskID == "" && !strings.HasPrefix(args[i], "--") {
				c.taskID = args[i]
			}
		}
	}

	if c.taskID == "" || c.description == "" {
		return fmt.Errorf("<task-id> and --description are required")
	}

	// Validate verification type
	if c.verificationType != "manual" && c.verificationType != "automated" {
		return fmt.Errorf("invalid verification type: %s (must be manual or automated)", c.verificationType)
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Verify task exists
	_, err = repo.GetTask(ctx, c.taskID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: Task \"%s\" not found\n", c.taskID)
			return fmt.Errorf("task not found: %s", c.taskID)
		}
		return fmt.Errorf("failed to verify task: %w", err)
	}

	// Generate AC ID
	projectCode := repo.GetProjectCode(ctx)
	nextNum, err := repo.GetNextSequenceNumber(ctx, "ac")
	if err != nil {
		return fmt.Errorf("failed to generate AC ID: %w", err)
	}
	acID := fmt.Sprintf("%s-ac-%d", projectCode, nextNum)

	// Create AC
	ac := NewAcceptanceCriteriaEntity(
		acID,
		c.taskID,
		c.description,
		AcceptanceCriteriaVerificationType(c.verificationType),
		time.Now().UTC(),
		time.Now().UTC(),
	)

	// Save AC
	if err := repo.SaveAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to save AC: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Acceptance criterion created successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID:               %s\n", ac.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Task:             %s\n", ac.TaskID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Description:      %s\n", ac.Description)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Verification:     %s\n", ac.VerificationType)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Status:           %s\n", ac.Status)

	return nil
}

// ============================================================================
// ACListCommand lists acceptance criteria for a task
// ============================================================================

type ACListCommand struct {
	Plugin *TaskManagerPlugin
	project string
	taskID string
}

func (c *ACListCommand) GetName() string {
	return "ac list"
}

func (c *ACListCommand) GetDescription() string {
	return "List acceptance criteria for a task"
}

func (c *ACListCommand) GetUsage() string {
	return "dw task-manager ac list <task-id>"
}

func (c *ACListCommand) GetHelp() string {
	return `Lists all acceptance criteria for a task with their verification status.

Status indicators:
  ✓   Verified (manually or automatically)
  ⏸   Pending human review
  ○   Not started
  ✗   Failed

Examples:
  # List ACs for a task
  dw task-manager ac list DW-task-123

Notes:
  - Shows verification type and current status for each AC
  - Summary shows total and verified counts`
}

func (c *ACListCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
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
		return fmt.Errorf("failed to list ACs: %w", err)
	}

	if len(acs) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "No acceptance criteria found for task %s\n", c.taskID)
		return nil
	}

	// Count verified
	verifiedCount := 0
	for _, ac := range acs {
		if ac.IsVerified() {
			verifiedCount++
		}
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Acceptance Criteria for Task: %s\n", task.Title)
	fmt.Fprintf(cmdCtx.GetStdout(), "Summary: %d/%d verified\n\n", verifiedCount, len(acs))

	for _, ac := range acs {
		statusIcon := ac.StatusIndicator()
		fmt.Fprintf(cmdCtx.GetStdout(), "%s [%s] %s (%s)\n", statusIcon, ac.ID, ac.Description, ac.VerificationType)
		if ac.Status == ACStatusFailed && ac.Notes != "" {
			fmt.Fprintf(cmdCtx.GetStdout(), "  Reason: %s\n", ac.Notes)
		}
	}

	return nil
}

// ============================================================================
// ACVerifyCommand marks an AC as verified by human
// ============================================================================

type ACVerifyCommand struct {
	Plugin *TaskManagerPlugin
	project string
	acID   string
	notes  string
}

func (c *ACVerifyCommand) GetName() string {
	return "ac verify"
}

func (c *ACVerifyCommand) GetDescription() string {
	return "Mark an acceptance criterion as verified"
}

func (c *ACVerifyCommand) GetUsage() string {
	return "dw task-manager ac verify <ac-id> [--notes \"...\"]"
}

func (c *ACVerifyCommand) GetHelp() string {
	return `Marks an acceptance criterion as verified by human review.

This command is used when a human has manually verified that the AC
requirements have been met. It updates the AC status to "verified".

Flags:
  <ac-id>           AC ID to verify (required)
  --notes "..."     Verification notes (optional)

Examples:
  # Verify an AC
  dw task-manager ac verify DW-ac-1

  # Verify with notes
  dw task-manager ac verify DW-ac-1 \
    --notes "Tested manually on Chrome, Firefox, Safari"`
}

func (c *ACVerifyCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--notes":
			if i+1 < len(args) {
				c.notes = args[i+1]
				i++
			}
		default:
			if c.acID == "" && !strings.HasPrefix(args[i], "--") {
				c.acID = args[i]
			}
		}
	}

	if c.acID == "" {
		return fmt.Errorf("<ac-id> is required")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get AC
	ac, err := repo.GetAC(ctx, c.acID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: AC \"%s\" not found\n", c.acID)
			return fmt.Errorf("AC not found: %s", c.acID)
		}
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Update AC status
	ac.Status = ACStatusVerified
	ac.Notes = c.notes
	ac.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to update AC: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Acceptance criterion verified\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID:     %s\n", ac.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Status: %s\n", ac.Status)

	return nil
}

// ============================================================================
// ACFailCommand marks an AC as failed
// ============================================================================

type ACFailCommand struct {
	Plugin *TaskManagerPlugin
	project string
	acID   string
	reason string
}

func (c *ACFailCommand) GetName() string {
	return "ac fail"
}

func (c *ACFailCommand) GetDescription() string {
	return "Mark an acceptance criterion as failed"
}

func (c *ACFailCommand) GetUsage() string {
	return "dw task-manager ac fail <ac-id> [--reason \"...\"]"
}

func (c *ACFailCommand) GetHelp() string {
	return `Marks an acceptance criterion as failed.

This command is used when verification shows that the AC requirements
have not been met. It updates the AC status to "failed".

Flags:
  <ac-id>            AC ID to mark as failed (required)
  --reason "..."     Failure reason (optional but recommended)

Examples:
  # Mark AC as failed
  dw task-manager ac fail DW-ac-1

  # Mark with reason
  dw task-manager ac fail DW-ac-1 \
    --reason "Tests fail on Safari browser"`
}

func (c *ACFailCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--reason":
			if i+1 < len(args) {
				c.reason = args[i+1]
				i++
			}
		default:
			if c.acID == "" && !strings.HasPrefix(args[i], "--") {
				c.acID = args[i]
			}
		}
	}

	if c.acID == "" {
		return fmt.Errorf("<ac-id> is required")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get AC
	ac, err := repo.GetAC(ctx, c.acID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: AC \"%s\" not found\n", c.acID)
			return fmt.Errorf("AC not found: %s", c.acID)
		}
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Update AC status
	ac.Status = ACStatusFailed
	ac.Notes = c.reason
	ac.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to update AC: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Acceptance criterion marked as failed\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID:     %s\n", ac.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Status: %s\n", ac.Status)
	if c.reason != "" {
		fmt.Fprintf(cmdCtx.GetStdout(), "  Reason: %s\n", c.reason)
	}

	return nil
}

// ============================================================================
// ACUpdateCommand updates an AC description
// ============================================================================

type ACUpdateCommand struct {
	Plugin      *TaskManagerPlugin
	project     string
	acID        string
	description string
}

func (c *ACUpdateCommand) GetName() string {
	return "ac update"
}

func (c *ACUpdateCommand) GetDescription() string {
	return "Update an acceptance criterion"
}

func (c *ACUpdateCommand) GetUsage() string {
	return "dw task-manager ac update <ac-id> --description \"...\""
}

func (c *ACUpdateCommand) GetHelp() string {
	return `Updates an acceptance criterion description.

Flags:
  <ac-id>            AC ID to update (required)
  --description "..."  New description (required)

Examples:
  # Update AC description
  dw task-manager ac update DW-ac-1 \
    --description "Updated requirement text"`
}

func (c *ACUpdateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				c.description = args[i+1]
				i++
			}
		default:
			if c.acID == "" && !strings.HasPrefix(args[i], "--") {
				c.acID = args[i]
			}
		}
	}

	if c.acID == "" || c.description == "" {
		return fmt.Errorf("<ac-id> and --description are required")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get AC
	ac, err := repo.GetAC(ctx, c.acID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: AC \"%s\" not found\n", c.acID)
			return fmt.Errorf("AC not found: %s", c.acID)
		}
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Update AC
	ac.Description = c.description
	ac.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to update AC: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Acceptance criterion updated\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID:          %s\n", ac.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Description: %s\n", ac.Description)

	return nil
}

// ============================================================================
// ACDeleteCommand deletes an AC
// ============================================================================

type ACDeleteCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	acID    string
	force   bool
}

func (c *ACDeleteCommand) GetName() string {
	return "ac delete"
}

func (c *ACDeleteCommand) GetDescription() string {
	return "Delete an acceptance criterion"
}

func (c *ACDeleteCommand) GetUsage() string {
	return "dw task-manager ac delete <ac-id> [--force]"
}

func (c *ACDeleteCommand) GetHelp() string {
	return `Deletes an acceptance criterion.

Requires the --force flag for safety.

Flags:
  <ac-id>     AC ID to delete (required)
  --force     Required to confirm deletion

Examples:
  # Delete an AC
  dw task-manager ac delete DW-ac-1 --force`
}

func (c *ACDeleteCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--force":
			c.force = true
		default:
			if c.acID == "" && !strings.HasPrefix(args[i], "--") {
				c.acID = args[i]
			}
		}
	}

	if c.acID == "" {
		return fmt.Errorf("<ac-id> is required")
	}

	if !c.force {
		return fmt.Errorf("--force flag is required to confirm deletion")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Verify AC exists
	_, err = repo.GetAC(ctx, c.acID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: AC \"%s\" not found\n", c.acID)
			return fmt.Errorf("AC not found: %s", c.acID)
		}
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Delete AC
	if err := repo.DeleteAC(ctx, c.acID); err != nil {
		return fmt.Errorf("failed to delete AC: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Acceptance criterion deleted\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID: %s\n", c.acID)

	return nil
}

// ============================================================================
// ACVerifyAutoCommand marks an AC as automatically verified
// ============================================================================

type ACVerifyAutoCommand struct {
	Plugin *TaskManagerPlugin
	project string
	acID   string
}

func (c *ACVerifyAutoCommand) GetName() string {
	return "ac verify-auto"
}

func (c *ACVerifyAutoCommand) GetDescription() string {
	return "Mark an AC as automatically verified"
}

func (c *ACVerifyAutoCommand) GetUsage() string {
	return "dw task-manager ac verify-auto <ac-id>"
}

func (c *ACVerifyAutoCommand) GetHelp() string {
	return `Marks an acceptance criterion as automatically verified.

Used by coding agents to indicate that automated verification
(tests, linting, etc.) has passed for this AC.

Flags:
  <ac-id>  AC ID to mark as auto-verified (required)

Examples:
  # Mark AC as auto-verified
  dw task-manager ac verify-auto DW-ac-1`
}

func (c *ACVerifyAutoCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if c.acID == "" && !strings.HasPrefix(args[i], "--") {
				c.acID = args[i]
			}
		}
	}

	if c.acID == "" {
		return fmt.Errorf("<ac-id> is required")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get AC
	ac, err := repo.GetAC(ctx, c.acID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: AC \"%s\" not found\n", c.acID)
			return fmt.Errorf("AC not found: %s", c.acID)
		}
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Update AC status
	ac.Status = ACStatusAutomaticallyVerified
	ac.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to update AC: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Acceptance criterion marked as automatically verified\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID:     %s\n", ac.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Status: %s\n", ac.Status)

	return nil
}

// ============================================================================
// ACRequestReviewCommand requests human review for an AC
// ============================================================================

type ACRequestReviewCommand struct {
	Plugin *TaskManagerPlugin
	project string
	acID   string
}

func (c *ACRequestReviewCommand) GetName() string {
	return "ac request-review"
}

func (c *ACRequestReviewCommand) GetDescription() string {
	return "Request human review for an AC"
}

func (c *ACRequestReviewCommand) GetUsage() string {
	return "dw task-manager ac request-review <ac-id>"
}

func (c *ACRequestReviewCommand) GetHelp() string {
	return `Requests human review for an acceptance criterion.

Used by coding agents to indicate that this AC requires
manual human verification before the task can be completed.

Flags:
  <ac-id>  AC ID to request review for (required)

Examples:
  # Request human review
  dw task-manager ac request-review DW-ac-1`
}

func (c *ACRequestReviewCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if c.acID == "" && !strings.HasPrefix(args[i], "--") {
				c.acID = args[i]
			}
		}
	}

	if c.acID == "" {
		return fmt.Errorf("<ac-id> is required")
	}

	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get AC
	ac, err := repo.GetAC(ctx, c.acID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Error: AC \"%s\" not found\n", c.acID)
			return fmt.Errorf("AC not found: %s", c.acID)
		}
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Update AC status
	ac.Status = ACStatusPendingHumanReview
	ac.UpdatedAt = time.Now().UTC()

	if err := repo.UpdateAC(ctx, ac); err != nil {
		return fmt.Errorf("failed to update AC: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Human review requested for acceptance criterion\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID:     %s\n", ac.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Status: %s\n", ac.Status)

	return nil
}
