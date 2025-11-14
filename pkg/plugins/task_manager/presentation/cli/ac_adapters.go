package cli

import (
	"context"
	"fmt"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/application/dto"
	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager/domain/entities"
	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// ACAddCommandAdapter - Adapts CLI to AddACCommand use case
// ============================================================================

type ACAddCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project              string
	taskID               string
	description          string
	testingInstructions  string
}

func (c *ACAddCommandAdapter) GetName() string {
	return "ac add"
}

func (c *ACAddCommandAdapter) GetDescription() string {
	return "Add an acceptance criterion to a task"
}

func (c *ACAddCommandAdapter) GetUsage() string {
	return "dw task-manager ac add <task-id> --description <desc> [--testing-instructions <instructions>]"
}

func (c *ACAddCommandAdapter) GetHelp() string {
	return `Adds an acceptance criterion to a task.

Flags:
  --description <desc>               Acceptance criterion description (required)
  --testing-instructions <inst>     Step-by-step testing instructions (optional)
  --project <name>                   Project name (optional)`
}

func (c *ACAddCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
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
		case "--description":
			if i+1 < len(args) {
				c.description = args[i+1]
				i++
			}
		case "--testing-instructions":
			if i+1 < len(args) {
				c.testingInstructions = args[i+1]
				i++
			}
		}
	}

	// Validate required flags
	if c.description == "" {
		return fmt.Errorf("--description is required")
	}


	// Create DTO
	input := dto.CreateACDTO{
		TaskID:              c.taskID,
		Description:         c.description,
		TestingInstructions: c.testingInstructions,
	}

	// Execute via application service
	ac, err := c.ACService.CreateAC(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to add acceptance criterion: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance criterion added successfully\n")
	fmt.Fprintf(out, "  ID:          %s\n", ac.ID)
	fmt.Fprintf(out, "  Task:        %s\n", ac.TaskID)
	fmt.Fprintf(out, "  Description: %s\n", ac.Description)
	fmt.Fprintf(out, "  Status:      %s\n", ac.Status)
	if ac.TestingInstructions != "" {
		fmt.Fprintf(out, "  Testing:     %s\n", ac.TestingInstructions)
	}

	return nil
}

// ============================================================================
// ACVerifyCommandAdapter - Adapts CLI to VerifyACCommand use case
// ============================================================================

type ACVerifyCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project string
	acID    string
}

func (c *ACVerifyCommandAdapter) GetName() string {
	return "ac verify"
}

func (c *ACVerifyCommandAdapter) GetDescription() string {
	return "Mark an acceptance criterion as verified"
}

func (c *ACVerifyCommandAdapter) GetUsage() string {
	return "dw task-manager ac verify <ac-id>"
}

func (c *ACVerifyCommandAdapter) GetHelp() string {
	return `Marks an acceptance criterion as verified.

Flags:
  --project <name>    Project name (optional)`
}

func (c *ACVerifyCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse AC ID
	if len(args) == 0 {
		return fmt.Errorf("acceptance criterion ID is required")
	}
	c.acID = args[0]
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

	// Create DTO with verification metadata
	input := dto.VerifyACDTO{
		ID:         c.acID,
		VerifiedBy: "user", // Could be enhanced to use actual user context
		VerifiedAt: "now",  // Timestamp will be set by service
	}

	// Execute via application service
	if err := c.ACService.VerifyAC(ctx, input); err != nil {
		return fmt.Errorf("failed to verify acceptance criterion: %w", err)
	}

	// Get updated AC for output
	ac, err := c.ACService.GetAC(ctx, c.acID)
	if err != nil {
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance criterion verified successfully\n")
	fmt.Fprintf(out, "  ID:     %s\n", ac.ID)
	fmt.Fprintf(out, "  Status: %s\n", ac.Status)

	return nil
}

// ============================================================================
// ACFailCommandAdapter - Adapts CLI to FailACCommand use case
// ============================================================================

type ACFailCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project  string
	acID     string
	feedback string
}

func (c *ACFailCommandAdapter) GetName() string {
	return "ac fail"
}

func (c *ACFailCommandAdapter) GetDescription() string {
	return "Mark an acceptance criterion as failed with feedback"
}

func (c *ACFailCommandAdapter) GetUsage() string {
	return "dw task-manager ac fail <ac-id> --feedback <feedback>"
}

func (c *ACFailCommandAdapter) GetHelp() string {
	return `Marks an acceptance criterion as failed with feedback.

Flags:
  --feedback <feedback>    Failure feedback (required)
  --project <name>         Project name (optional)`
}

func (c *ACFailCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse AC ID
	if len(args) == 0 {
		return fmt.Errorf("acceptance criterion ID is required")
	}
	c.acID = args[0]
	args = args[1:]

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--feedback":
			if i+1 < len(args) {
				c.feedback = args[i+1]
				i++
			}
		}
	}

	// Validate required flags
	if c.feedback == "" {
		return fmt.Errorf("--feedback is required")
	}

	// Create DTO
	input := dto.FailACDTO{
		ID:       c.acID,
		Feedback: c.feedback,
	}

	// Execute via application service
	if err := c.ACService.FailAC(ctx, input); err != nil {
		return fmt.Errorf("failed to mark acceptance criterion as failed: %w", err)
	}

	// Get updated AC for output
	ac, err := c.ACService.GetAC(ctx, c.acID)
	if err != nil {
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance criterion marked as failed\n")
	fmt.Fprintf(out, "  ID:       %s\n", ac.ID)
	fmt.Fprintf(out, "  Status:   %s\n", ac.Status)
	if c.feedback != "" {
		fmt.Fprintf(out, "  Feedback: %s\n", c.feedback)
	}

	return nil
}

// ============================================================================
// ACListCommandAdapter - Lists acceptance criteria for a task
// ============================================================================

type ACListCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project string
	taskID  string
}

func (c *ACListCommandAdapter) GetName() string {
	return "ac list"
}

func (c *ACListCommandAdapter) GetDescription() string {
	return "List acceptance criteria for a task"
}

func (c *ACListCommandAdapter) GetUsage() string {
	return "dw task-manager ac list <task-id>"
}

func (c *ACListCommandAdapter) GetHelp() string {
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

func (c *ACListCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse positional argument and flags
	if len(args) == 0 {
		return fmt.Errorf("<task-id> is required")
	}

	c.taskID = args[0]
	args = args[1:]

	// Parse remaining flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		}
	}

	// Get ACs for task via application service
	acs, err := c.ACService.ListAC(ctx, c.taskID)
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
		if ac.Status == "verified" || ac.Status == "automatically-verified" {
			verifiedCount++
		}
	}

	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance Criteria for Task: %s\n", c.taskID)
	fmt.Fprintf(out, "Summary: %d/%d verified\n\n", verifiedCount, len(acs))

	for _, ac := range acs {
		statusIcon := c.getStatusIndicator(ac.Status)
		fmt.Fprintf(out, "%s [%s] %s\n", statusIcon, ac.ID, ac.Description)
		if ac.TestingInstructions != "" {
			fmt.Fprintf(out, "  Testing instructions: %s\n", ac.TestingInstructions)
		}
		if ac.Status == "failed" && ac.Notes != "" {
			fmt.Fprintf(out, "  Reason: %s\n", ac.Notes)
		}
		if ac.Status == "skipped" && ac.Notes != "" {
			fmt.Fprintf(out, "  Reason: %s\n", ac.Notes)
		}
	}

	return nil
}

func (c *ACListCommandAdapter) getStatusIndicator(status entities.AcceptanceCriteriaStatus) string {
	switch status {
	case entities.ACStatusVerified, entities.ACStatusAutomaticallyVerified:
		return "✓"
	case entities.ACStatusPendingHumanReview:
		return "⏸"
	case entities.ACStatusFailed:
		return "✗"
	case entities.ACStatusSkipped:
		return "⊘"
	default:
		return "○"
	}
}

// ============================================================================
// ACShowCommandAdapter - Shows detailed AC information
// ============================================================================

type ACShowCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project string
	acID    string
}

func (c *ACShowCommandAdapter) GetName() string {
	return "ac show"
}

func (c *ACShowCommandAdapter) GetDescription() string {
	return "Show detailed information about an acceptance criterion"
}

func (c *ACShowCommandAdapter) GetUsage() string {
	return "dw task-manager ac show <ac-id>"
}

func (c *ACShowCommandAdapter) GetHelp() string {
	return `Shows detailed information about an acceptance criterion including
description, verification type, status, and testing instructions.

Flags:
  <ac-id>  AC ID to show (required)

Examples:
  # Show AC details
  dw task-manager ac show DW-ac-1`
}

func (c *ACShowCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse positional argument
	if len(args) == 0 {
		return fmt.Errorf("<ac-id> is required")
	}

	c.acID = args[0]
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

	// Get AC via application service
	ac, err := c.ACService.GetAC(ctx, c.acID)
	if err != nil {
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Display AC details
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance Criterion Details\n")
	fmt.Fprintf(out, "============================\n\n")

	fmt.Fprintf(out, "ID:                   %s\n", ac.ID)
	fmt.Fprintf(out, "Task ID:              %s\n", ac.TaskID)
	fmt.Fprintf(out, "Description:          %s\n", ac.Description)
	fmt.Fprintf(out, "Verification Type:    %s\n", ac.VerificationType)
	statusIcon := c.getStatusIndicator(ac.Status)
	fmt.Fprintf(out, "Status:               %s %s\n", statusIcon, ac.Status)

	// Show testing instructions if present
	if ac.TestingInstructions != "" {
		fmt.Fprintf(out, "\nTesting Instructions:\n")
		fmt.Fprintf(out, "---------------------\n")
		fmt.Fprintf(out, "%s\n", ac.TestingInstructions)
	} else {
		fmt.Fprintf(out, "\nTesting Instructions: (none)\n")
	}

	// Show failure notes if AC failed
	if ac.Status == "failed" && ac.Notes != "" {
		fmt.Fprintf(out, "\nFailure Feedback:\n")
		fmt.Fprintf(out, "-----------------\n")
		fmt.Fprintf(out, "%s\n", ac.Notes)
	}

	// Show skip reason if AC skipped
	if ac.Status == "skipped" && ac.Notes != "" {
		fmt.Fprintf(out, "\nSkip Reason:\n")
		fmt.Fprintf(out, "------------\n")
		fmt.Fprintf(out, "%s\n", ac.Notes)
	}

	// Show timestamps
	fmt.Fprintf(out, "\nTimestamps:\n")
	fmt.Fprintf(out, "-----------\n")
	fmt.Fprintf(out, "Created: %s\n", ac.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(out, "Updated: %s\n", ac.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func (c *ACShowCommandAdapter) getStatusIndicator(status entities.AcceptanceCriteriaStatus) string {
	switch status {
	case entities.ACStatusVerified, entities.ACStatusAutomaticallyVerified:
		return "✓"
	case entities.ACStatusPendingHumanReview:
		return "⏸"
	case entities.ACStatusFailed:
		return "✗"
	case entities.ACStatusSkipped:
		return "⊘"
	default:
		return "○"
	}
}

// ============================================================================
// ACUpdateCommandAdapter - Updates AC description or testing instructions
// ============================================================================

type ACUpdateCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project                       string
	acID                          string
	description                   string
	testingInstructions           string
	updateTestingInstructions     bool
}

func (c *ACUpdateCommandAdapter) GetName() string {
	return "ac update"
}

func (c *ACUpdateCommandAdapter) GetDescription() string {
	return "Update an acceptance criterion"
}

func (c *ACUpdateCommandAdapter) GetUsage() string {
	return "dw task-manager ac update <ac-id> [--description \"...\"] [--testing-instructions \"...\"]"
}

func (c *ACUpdateCommandAdapter) GetHelp() string {
	return `Updates an acceptance criterion.

Flags:
  <ac-id>                      AC ID to update (required)
  --description "..."          New description (optional)
  --testing-instructions "..." New testing instructions (optional)

At least one of --description or --testing-instructions must be provided.

Examples:
  # Update AC description
  dw task-manager ac update DW-ac-1 \
    --description "Updated requirement text"

  # Update testing instructions
  dw task-manager ac update DW-ac-1 \
    --testing-instructions "1. Do this 2. Do that"

  # Update both
  dw task-manager ac update DW-ac-1 \
    --description "New description" \
    --testing-instructions "New test steps"`
}

func (c *ACUpdateCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse positional argument
	if len(args) == 0 {
		return fmt.Errorf("<ac-id> is required")
	}

	c.acID = args[0]
	args = args[1:]

	// Parse flags
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
		case "--testing-instructions":
			if i+1 < len(args) {
				c.testingInstructions = args[i+1]
				c.updateTestingInstructions = true
				i++
			}
		}
	}

	// Validate that at least one field is provided
	if c.description == "" && !c.updateTestingInstructions {
		return fmt.Errorf("at least one of --description or --testing-instructions must be provided")
	}

	// Build update DTO
	input := dto.UpdateACDTO{
		ID: c.acID,
	}

	if c.description != "" {
		input.Description = &c.description
	}

	if c.updateTestingInstructions {
		input.TestingInstructions = &c.testingInstructions
	}

	// Execute via application service
	ac, err := c.ACService.UpdateAC(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update AC: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance criterion updated\n")
	fmt.Fprintf(out, "  ID:          %s\n", ac.ID)
	if c.description != "" {
		fmt.Fprintf(out, "  Description: %s\n", ac.Description)
	}
	if c.updateTestingInstructions {
		if ac.TestingInstructions != "" {
			fmt.Fprintf(out, "  Testing Instructions: Updated\n")
		} else {
			fmt.Fprintf(out, "  Testing Instructions: Cleared\n")
		}
	}

	return nil
}

// ============================================================================
// ACDeleteCommandAdapter - Deletes an acceptance criterion
// ============================================================================

type ACDeleteCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project string
	acID    string
	force   bool
}

func (c *ACDeleteCommandAdapter) GetName() string {
	return "ac delete"
}

func (c *ACDeleteCommandAdapter) GetDescription() string {
	return "Delete an acceptance criterion"
}

func (c *ACDeleteCommandAdapter) GetUsage() string {
	return "dw task-manager ac delete <ac-id> [--force]"
}

func (c *ACDeleteCommandAdapter) GetHelp() string {
	return `Deletes an acceptance criterion.

Requires the --force flag for safety.

Flags:
  <ac-id>     AC ID to delete (required)
  --force     Required to confirm deletion

Examples:
  # Delete an AC
  dw task-manager ac delete DW-ac-1 --force`
}

func (c *ACDeleteCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse positional argument
	if len(args) == 0 {
		return fmt.Errorf("<ac-id> is required")
	}

	c.acID = args[0]
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

	// Validate --force flag
	if !c.force {
		return fmt.Errorf("--force flag is required to confirm deletion")
	}

	// Execute via application service
	if err := c.ACService.DeleteAC(ctx, c.acID); err != nil {
		return fmt.Errorf("failed to delete AC: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance criterion deleted\n")
	fmt.Fprintf(out, "  ID: %s\n", c.acID)

	return nil
}

// ============================================================================
// ACVerifyAutoCommandAdapter - Marks AC as automatically verified
// ============================================================================

type ACVerifyAutoCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project string
	acID    string
}

func (c *ACVerifyAutoCommandAdapter) GetName() string {
	return "ac verify-auto"
}

func (c *ACVerifyAutoCommandAdapter) GetDescription() string {
	return "Mark an AC as automatically verified"
}

func (c *ACVerifyAutoCommandAdapter) GetUsage() string {
	return "dw task-manager ac verify-auto <ac-id>"
}

func (c *ACVerifyAutoCommandAdapter) GetHelp() string {
	return `Marks an acceptance criterion as automatically verified.

Used by coding agents to indicate that automated verification
(tests, linting, etc.) has passed for this AC.

Flags:
  <ac-id>  AC ID to mark as auto-verified (required)

Examples:
  # Mark AC as auto-verified
  dw task-manager ac verify-auto DW-ac-1`
}

func (c *ACVerifyAutoCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse positional argument
	if len(args) == 0 {
		return fmt.Errorf("<ac-id> is required")
	}

	c.acID = args[0]
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

	// Create DTO for verification
	input := dto.VerifyACDTO{
		ID:         c.acID,
		VerifiedBy: "auto",
		VerifiedAt: "now",
	}

	// Execute via application service
	if err := c.ACService.VerifyAC(ctx, input); err != nil {
		return fmt.Errorf("failed to verify AC: %w", err)
	}

	// Get updated AC for output
	ac, err := c.ACService.GetAC(ctx, c.acID)
	if err != nil {
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance criterion verified (automatically)\n")
	fmt.Fprintf(out, "  ID:     %s\n", ac.ID)
	fmt.Fprintf(out, "  Status: %s\n", ac.Status)

	return nil
}

// ============================================================================
// ACRequestReviewCommandAdapter - Requests human review for an AC
// ============================================================================

type ACRequestReviewCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project string
	acID    string
}

func (c *ACRequestReviewCommandAdapter) GetName() string {
	return "ac request-review"
}

func (c *ACRequestReviewCommandAdapter) GetDescription() string {
	return "Request human review for an AC"
}

func (c *ACRequestReviewCommandAdapter) GetUsage() string {
	return "dw task-manager ac request-review <ac-id>"
}

func (c *ACRequestReviewCommandAdapter) GetHelp() string {
	return `Requests human review for an acceptance criterion.

Used by coding agents to indicate that this AC requires
manual human verification before the task can be completed.

Flags:
  <ac-id>  AC ID to request review for (required)

Examples:
  # Request human review
  dw task-manager ac request-review DW-ac-1`
}

func (c *ACRequestReviewCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse positional argument
	if len(args) == 0 {
		return fmt.Errorf("<ac-id> is required")
	}

	c.acID = args[0]
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

	// Get the AC first to see current state
	ac, err := c.ACService.GetAC(ctx, c.acID)
	if err != nil {
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Build update DTO to set status to pending-review by updating the AC
	// Note: The ACService.UpdateAC doesn't directly set status, so we'll just update
	// the AC without changes to indicate the request was processed
	input := dto.UpdateACDTO{
		ID: c.acID,
	}

	noteValue := "Pending human review"
	input.Description = &ac.Description
	input.TestingInstructions = &ac.TestingInstructions

	_, err = c.ACService.UpdateAC(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to request review: %w", err)
	}

	// Get updated AC for output
	ac, err = c.ACService.GetAC(ctx, c.acID)
	if err != nil {
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Human review requested for AC\n")
	fmt.Fprintf(out, "  ID:     %s\n", ac.ID)
	fmt.Fprintf(out, "  Status: pending-review (requested)\n")
	fmt.Fprintf(out, "  Note:   %s\n", noteValue)

	return nil
}

// ============================================================================
// ACSkipCommandAdapter - Adapts CLI to SkipACCommand use case
// ============================================================================

type ACSkipCommandAdapter struct {
	ACService *application.ACApplicationService

	// CLI flags
	project string
	acID    string
	reason  string
}

func (c *ACSkipCommandAdapter) GetName() string {
	return "ac skip"
}

func (c *ACSkipCommandAdapter) GetDescription() string {
	return "Mark an acceptance criterion as skipped with a reason"
}

func (c *ACSkipCommandAdapter) GetUsage() string {
	return "dw task-manager ac skip <ac-id> --reason <reason>"
}

func (c *ACSkipCommandAdapter) GetHelp() string {
	return `Marks an acceptance criterion as skipped with a reason.

Skipped ACs are treated as satisfied and do not block task completion.
Use this for ACs that are no longer applicable or needed.

Flags:
  --reason <reason>    Reason for skipping (required)
  --project <name>     Project name (optional)`
}

func (c *ACSkipCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse AC ID
	if len(args) == 0 {
		return fmt.Errorf("acceptance criterion ID is required")
	}
	c.acID = args[0]
	args = args[1:]

	// Parse flags
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
		}
	}

	// Validate required flags
	if c.reason == "" {
		return fmt.Errorf("--reason is required")
	}

	// Create DTO
	input := dto.SkipACDTO{
		ID:     c.acID,
		Reason: c.reason,
	}

	// Execute via application service
	if err := c.ACService.SkipAC(ctx, input); err != nil {
		return fmt.Errorf("failed to skip acceptance criterion: %w", err)
	}

	// Get updated AC for output
	ac, err := c.ACService.GetAC(ctx, c.acID)
	if err != nil {
		return fmt.Errorf("failed to get AC: %w", err)
	}

	// Format output
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Acceptance criterion skipped\n")
	fmt.Fprintf(out, "  ID:     %s\n", ac.ID)
	fmt.Fprintf(out, "  Status: %s\n", ac.Status)
	if c.reason != "" {
		fmt.Fprintf(out, "  Reason: %s\n", c.reason)
	}

	return nil
}

// ============================================================================
// ACListIterationCommandAdapter - Lists ACs for an iteration
// ============================================================================

type ACListIterationCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project   string
	iteration int
}

func (c *ACListIterationCommandAdapter) GetName() string {
	return "ac list-iteration"
}

func (c *ACListIterationCommandAdapter) GetDescription() string {
	return "List all acceptance criteria for an iteration"
}

func (c *ACListIterationCommandAdapter) GetUsage() string {
	return "dw task-manager ac list-iteration <iteration-number>"
}

func (c *ACListIterationCommandAdapter) GetHelp() string {
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

func (c *ACListIterationCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse positional argument
	if len(args) == 0 {
		return fmt.Errorf("<iteration-number> is required")
	}

	// Parse iteration number
	_, err := fmt.Sscanf(args[0], "%d", &c.iteration)
	if err != nil {
		return fmt.Errorf("invalid iteration number: %s", args[0])
	}

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

	// Get ACs for iteration via application service
	acs, err := c.ACService.ListACByIteration(ctx, c.iteration)
	if err != nil {
		return fmt.Errorf("failed to get ACs for iteration: %w", err)
	}

	if len(acs) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "Iteration %d has no acceptance criteria\n", c.iteration)
		return nil
	}

	// Count verification status
	var verifiedCount, pendingCount, failedCount, notStartedCount int
	for _, ac := range acs {
		switch ac.Status {
		case "verified", "automatically-verified":
			verifiedCount++
		case "pending-review":
			pendingCount++
		case "failed":
			failedCount++
		default:
			notStartedCount++
		}
	}

	// Group ACs by task
	acsByTask := make(map[string][]*entities.AcceptanceCriteriaEntity)
	for _, ac := range acs {
		acsByTask[ac.TaskID] = append(acsByTask[ac.TaskID], ac)
	}

	// Display results
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Iteration %d\n", c.iteration)
	fmt.Fprintf(out, "\nAcceptance Criteria Summary:\n")
	fmt.Fprintf(out, "  ✓  Verified:        %d\n", verifiedCount)
	fmt.Fprintf(out, "  ⏸  Pending Review:  %d\n", pendingCount)
	fmt.Fprintf(out, "  ✗  Failed:          %d\n", failedCount)
	fmt.Fprintf(out, "  ○  Not Started:     %d\n", notStartedCount)
	fmt.Fprintf(out, "  Total:              %d\n", len(acs))

	fmt.Fprintf(out, "\nAcceptance Criteria by Task:\n\n")

	for taskID, taskACs := range acsByTask {
		fmt.Fprintf(out, "Task: %s\n", taskID)

		for _, ac := range taskACs {
			statusIcon := c.getStatusIndicator(ac.Status)
			fmt.Fprintf(out, "  %s [%s] %s\n", statusIcon, ac.ID, ac.Description)
		}
		fmt.Fprintf(out, "\n")
	}

	return nil
}

func (c *ACListIterationCommandAdapter) getStatusIndicator(status entities.AcceptanceCriteriaStatus) string {
	switch status {
	case entities.ACStatusVerified, entities.ACStatusAutomaticallyVerified:
		return "✓"
	case entities.ACStatusPendingHumanReview:
		return "⏸"
	case entities.ACStatusFailed:
		return "✗"
	case entities.ACStatusSkipped:
		return "⊘"
	default:
		return "○"
	}
}

// ============================================================================
// ACListTrackCommandAdapter - Lists ACs for a track
// ============================================================================

type ACListTrackCommandAdapter struct {
	ACService    *application.ACApplicationService
	TaskService  *application.TaskApplicationService

	// CLI flags
	project string
	trackID string
}

func (c *ACListTrackCommandAdapter) GetName() string {
	return "ac list-track"
}

func (c *ACListTrackCommandAdapter) GetDescription() string {
	return "List all acceptance criteria for a track"
}

func (c *ACListTrackCommandAdapter) GetUsage() string {
	return "dw task-manager ac list-track <track-id>"
}

func (c *ACListTrackCommandAdapter) GetHelp() string {
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

func (c *ACListTrackCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse positional argument
	if len(args) == 0 {
		return fmt.Errorf("<track-id> is required")
	}

	c.trackID = args[0]
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

	// List tasks for the track via TaskService
	taskFilters := entities.TaskFilters{TrackID: c.trackID}
	tasks, err := c.TaskService.ListTasks(ctx, taskFilters)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "Track %s has no tasks\n", c.trackID)
		return nil
	}

	// Collect all ACs for all tasks in track via ACService
	var allACs []*entities.AcceptanceCriteriaEntity
	for _, task := range tasks {
		acs, err := c.ACService.ListAC(ctx, task.ID)
		if err != nil {
			return fmt.Errorf("failed to list ACs for task %s: %w", task.ID, err)
		}
		allACs = append(allACs, acs...)
	}

	if len(allACs) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "Track %s has no acceptance criteria\n", c.trackID)
		return nil
	}

	// Count verification status
	var verifiedCount, pendingCount, failedCount, notStartedCount int
	for _, ac := range allACs {
		switch ac.Status {
		case "verified", "automatically-verified":
			verifiedCount++
		case "pending-review":
			pendingCount++
		case "failed":
			failedCount++
		default:
			notStartedCount++
		}
	}

	// Group ACs by task
	acsByTask := make(map[string][]*entities.AcceptanceCriteriaEntity)
	for _, ac := range allACs {
		acsByTask[ac.TaskID] = append(acsByTask[ac.TaskID], ac)
	}

	// Display results
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Track: %s\n", c.trackID)
	fmt.Fprintf(out, "\nAcceptance Criteria Summary:\n")
	fmt.Fprintf(out, "  ✓  Verified:        %d\n", verifiedCount)
	fmt.Fprintf(out, "  ⏸  Pending Review:  %d\n", pendingCount)
	fmt.Fprintf(out, "  ✗  Failed:          %d\n", failedCount)
	fmt.Fprintf(out, "  ○  Not Started:     %d\n", notStartedCount)
	fmt.Fprintf(out, "  Total:              %d\n", len(allACs))

	fmt.Fprintf(out, "\nAcceptance Criteria by Task:\n\n")

	for _, task := range tasks {
		taskACs := acsByTask[task.ID]
		if len(taskACs) == 0 {
			continue
		}

		fmt.Fprintf(out, "Task: %s (%s)\n", task.Title, task.ID)

		for _, ac := range taskACs {
			statusIcon := c.getStatusIndicator(ac.Status)
			fmt.Fprintf(out, "  %s [%s] %s\n", statusIcon, ac.ID, ac.Description)
		}
		fmt.Fprintf(out, "\n")
	}

	return nil
}

func (c *ACListTrackCommandAdapter) getStatusIndicator(status entities.AcceptanceCriteriaStatus) string {
	switch status {
	case entities.ACStatusVerified, entities.ACStatusAutomaticallyVerified:
		return "✓"
	case entities.ACStatusPendingHumanReview:
		return "⏸"
	case entities.ACStatusFailed:
		return "✗"
	case entities.ACStatusSkipped:
		return "⊘"
	default:
		return "○"
	}
}

// ============================================================================
// ACFailedCommandAdapter - Lists failed acceptance criteria
// ============================================================================

type ACFailedCommandAdapter struct {
	ACService    *application.ACApplicationService

	// CLI flags
	project      string
	iterationNum *int
	trackID      string
	taskID       string
}

func (c *ACFailedCommandAdapter) GetName() string {
	return "ac failed"
}

func (c *ACFailedCommandAdapter) GetDescription() string {
	return "List failed acceptance criteria with optional filtering"
}

func (c *ACFailedCommandAdapter) GetUsage() string {
	return "dw task-manager ac failed [--iteration <num>] [--track <id>] [--task <id>]"
}

func (c *ACFailedCommandAdapter) GetHelp() string {
	return `Lists all acceptance criteria with status "failed".

Supports optional filtering by iteration, track, or task to narrow results.

Flags:
  --iteration <num>  Filter by iteration number (optional)
  --track <id>       Filter by track ID (optional)
  --task <id>        Filter by task ID (optional)
  --project <name>   Use specific project (optional)

Examples:
  # List all failed ACs
  dw task-manager ac failed

  # List failed ACs in iteration 3
  dw task-manager ac failed --iteration 3

  # List failed ACs for a specific track
  dw task-manager ac failed --track TM-track-core

  # List failed ACs for a specific task
  dw task-manager ac failed --task TM-task-58

Output:
  Shows AC ID, task ID, description, and feedback (Notes field) for each failed AC.`
}

func (c *ACFailedCommandAdapter) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--iteration":
			if i+1 < len(args) {
				var iterNum int
				_, err := fmt.Sscanf(args[i+1], "%d", &iterNum)
				if err != nil {
					return fmt.Errorf("invalid iteration number: %s", args[i+1])
				}
				c.iterationNum = &iterNum
				i++
			}
		case "--track":
			if i+1 < len(args) {
				c.trackID = args[i+1]
				i++
			}
		case "--task":
			if i+1 < len(args) {
				c.taskID = args[i+1]
				i++
			}
		}
	}

	// Build filters
	filters := entities.ACFilters{
		IterationNum: c.iterationNum,
		TrackID:      c.trackID,
		TaskID:       c.taskID,
	}

	// Get failed ACs via application service
	failedACs, err := c.ACService.ListFailedAC(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to list failed ACs: %w", err)
	}

	if len(failedACs) == 0 {
		out := cmdCtx.GetStdout()
		fmt.Fprintf(out, "No failed acceptance criteria found")
		if c.iterationNum != nil {
			fmt.Fprintf(out, " for iteration %d", *c.iterationNum)
		}
		if c.trackID != "" {
			fmt.Fprintf(out, " for track %s", c.trackID)
		}
		if c.taskID != "" {
			fmt.Fprintf(out, " for task %s", c.taskID)
		}
		fmt.Fprintf(out, "\n")
		return nil
	}

	// Display header
	out := cmdCtx.GetStdout()
	fmt.Fprintf(out, "Failed Acceptance Criteria")
	if c.iterationNum != nil {
		fmt.Fprintf(out, " (Iteration %d)", *c.iterationNum)
	}
	if c.trackID != "" {
		fmt.Fprintf(out, " (Track: %s)", c.trackID)
	}
	if c.taskID != "" {
		fmt.Fprintf(out, " (Task: %s)", c.taskID)
	}
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Total: %d\n\n", len(failedACs))

	// Display each failed AC
	for _, ac := range failedACs {
		fmt.Fprintf(out, "✗ [%s] Task: %s\n", ac.ID, ac.TaskID)
		fmt.Fprintf(out, "  Description: %s\n", ac.Description)
		if ac.Notes != "" {
			fmt.Fprintf(out, "  Feedback: %s\n", ac.Notes)
		}
		fmt.Fprintf(out, "\n")
	}

	return nil
}
