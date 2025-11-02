package task_manager

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// ADRCreateCommand creates a new ADR
// ============================================================================

type ADRCreateCommand struct {
	Plugin       *TaskManagerPlugin
	project      string
	trackID      string
	title        string
	context      string
	decision     string
	consequences string
	alternatives string
}

func (c *ADRCreateCommand) GetName() string {
	return "adr create"
}

func (c *ADRCreateCommand) GetDescription() string {
	return "Create a new Architecture Decision Record (ADR)"
}

func (c *ADRCreateCommand) GetUsage() string {
	return "dw task-manager adr create <track-id> --title \"...\" --context \"...\" --decision \"...\" --consequences \"...\" [--alternatives \"...\"]"
}

func (c *ADRCreateCommand) GetHelp() string {
	return `Creates a new Architecture Decision Record for a track.

An ADR documents important architectural decisions made for a track.

Flags:
  --title <title>                ADR title (required)
  --context <context>            Context and background (required)
  --decision <decision>          The decision made (required)
  --consequences <consequences>  Consequences of the decision (required)
  --alternatives <alternatives>  Alternative options considered (optional)

Examples:
  # Create an ADR
  dw task-manager adr create track-1 \
    --title "Use gRPC for inter-service communication" \
    --context "Need efficient RPC mechanism with strict typing" \
    --decision "Adopt gRPC with Protocol Buffers" \
    --consequences "Adds compile-time code generation step, requires Go 1.13+"

Notes:
  - ADR ID is auto-generated as DW-adr-<number>
  - Initial status is 'proposed'
  - Use 'adr update' to change status to 'accepted'
  - Use 'adr supersede' to mark as superseded by another ADR`
}

func (c *ADRCreateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
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
				c.title = args[i+1]
				i++
			}
		case "--context":
			if i+1 < len(args) {
				c.context = args[i+1]
				i++
			}
		case "--decision":
			if i+1 < len(args) {
				c.decision = args[i+1]
				i++
			}
		case "--consequences":
			if i+1 < len(args) {
				c.consequences = args[i+1]
				i++
			}
		case "--alternatives":
			if i+1 < len(args) {
				c.alternatives = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "--") && c.trackID == "" {
				c.trackID = args[i]
			}
		}
	}

	// Validate required fields
	if c.trackID == "" {
		return fmt.Errorf("track ID is required")
	}
	if c.title == "" {
		return fmt.Errorf("--title is required")
	}
	if c.context == "" {
		return fmt.Errorf("--context is required")
	}
	if c.decision == "" {
		return fmt.Errorf("--decision is required")
	}
	if c.consequences == "" {
		return fmt.Errorf("--consequences is required")
	}

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Generate ADR ID
	adrNum, err := repo.GetNextSequenceNumber(ctx, "adr")
	if err != nil {
		return fmt.Errorf("failed to generate ADR ID: %w", err)
	}

	projectCode := repo.GetProjectCode(ctx)
	adrID := fmt.Sprintf("%s-adr-%d", projectCode, adrNum)

	// Create ADR entity
	now := GetCurrentTime()
	adr, err := NewADREntity(
		adrID,
		c.trackID,
		c.title,
		string(ADRStatusProposed),
		c.context,
		c.decision,
		c.consequences,
		c.alternatives,
		now,
		now,
		nil,
	)
	if err != nil {
		return err
	}

	// Save to repository
	if err := repo.SaveADR(ctx, adr); err != nil {
		return err
	}

	// Event is emitted by the repository

	fmt.Printf("Created ADR %s in track %s\n", adrID, c.trackID)
	fmt.Printf("Title: %s\n", c.title)
	fmt.Printf("Status: %s\n", string(ADRStatusProposed))

	return nil
}

// ============================================================================
// ADRListCommand lists all ADRs
// ============================================================================

type ADRListCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	trackID string
}

func (c *ADRListCommand) GetName() string {
	return "adr list"
}

func (c *ADRListCommand) GetDescription() string {
	return "List all ADRs"
}

func (c *ADRListCommand) GetUsage() string {
	return "dw task-manager adr list [track-id]"
}

func (c *ADRListCommand) GetHelp() string {
	return `Lists all Architecture Decision Records.

Can optionally filter by track.

Flags:
  track-id    Optional: filter to specific track

Examples:
  # List all ADRs
  dw task-manager adr list

  # List ADRs for a track
  dw task-manager adr list track-1`
}

func (c *ADRListCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "--") && c.trackID == "" {
				c.trackID = args[i]
			}
		}
	}

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// List ADRs
	var adrs []*ADREntity
	if c.trackID != "" {
		adrs, err = repo.GetADRsByTrack(ctx, c.trackID)
	} else {
		adrs, err = repo.ListADRs(ctx, nil)
	}
	if err != nil {
		return err
	}

	if len(adrs) == 0 {
		fmt.Println("No ADRs found")
		return nil
	}

	fmt.Printf("Found %d ADR(s):\n\n", len(adrs))
	for _, adr := range adrs {
		statusIcon := "○"
		switch adr.Status {
		case string(ADRStatusAccepted):
			statusIcon = "✓"
		case string(ADRStatusDeprecated):
			statusIcon = "✗"
		case string(ADRStatusSuperseded):
			statusIcon = "⤳"
		}

		fmt.Printf("%s [%s] %s: %s\n", statusIcon, adr.ID, adr.TrackID, adr.Title)
		fmt.Printf("   Status: %s\n", adr.Status)
		fmt.Printf("   Created: %s\n", adr.CreatedAt.Format(time.RFC3339))
		if adr.SupersededBy != nil {
			fmt.Printf("   Superseded by: %s\n", *adr.SupersededBy)
		}
		fmt.Println()
	}

	return nil
}

// ============================================================================
// ADRShowCommand shows an ADR's details
// ============================================================================

type ADRShowCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	adrID   string
}

func (c *ADRShowCommand) GetName() string {
	return "adr show"
}

func (c *ADRShowCommand) GetDescription() string {
	return "Show ADR details"
}

func (c *ADRShowCommand) GetUsage() string {
	return "dw task-manager adr show <adr-id>"
}

func (c *ADRShowCommand) GetHelp() string {
	return `Shows details of an Architecture Decision Record in markdown format.

Examples:
  # Show an ADR
  dw task-manager adr show DW-adr-1`
}

func (c *ADRShowCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "--") && c.adrID == "" {
				c.adrID = args[i]
			}
		}
	}

	if c.adrID == "" {
		return fmt.Errorf("ADR ID is required")
	}

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get ADR
	adr, err := repo.GetADR(ctx, c.adrID)
	if err != nil {
		return err
	}

	// Output as markdown
	fmt.Println(adr.ToMarkdown())

	return nil
}

// ============================================================================
// ADRUpdateCommand updates an ADR
// ============================================================================

type ADRUpdateCommand struct {
	Plugin       *TaskManagerPlugin
	project      string
	adrID        string
	title        string
	context      string
	decision     string
	consequences string
	alternatives string
}

func (c *ADRUpdateCommand) GetName() string {
	return "adr update"
}

func (c *ADRUpdateCommand) GetDescription() string {
	return "Update an ADR"
}

func (c *ADRUpdateCommand) GetUsage() string {
	return "dw task-manager adr update <adr-id> [--title|--context|--decision|--consequences|--alternatives \"...\"]"
}

func (c *ADRUpdateCommand) GetHelp() string {
	return `Updates an existing Architecture Decision Record.

At least one field must be provided.

Examples:
  # Update title
  dw task-manager adr update DW-adr-1 --title "New title"`
}

func (c *ADRUpdateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--title":
			if i+1 < len(args) {
				c.title = args[i+1]
				i++
			}
		case "--context":
			if i+1 < len(args) {
				c.context = args[i+1]
				i++
			}
		case "--decision":
			if i+1 < len(args) {
				c.decision = args[i+1]
				i++
			}
		case "--consequences":
			if i+1 < len(args) {
				c.consequences = args[i+1]
				i++
			}
		case "--alternatives":
			if i+1 < len(args) {
				c.alternatives = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "--") && c.adrID == "" {
				c.adrID = args[i]
			}
		}
	}

	if c.adrID == "" {
		return fmt.Errorf("ADR ID is required")
	}

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Get existing ADR
	adr, err := repo.GetADR(ctx, c.adrID)
	if err != nil {
		return err
	}

	// Update fields if provided
	if c.title != "" {
		adr.Title = c.title
	}
	if c.context != "" {
		adr.Context = c.context
	}
	if c.decision != "" {
		adr.Decision = c.decision
	}
	if c.consequences != "" {
		adr.Consequences = c.consequences
	}
	if c.alternatives != "" {
		adr.Alternatives = c.alternatives
	}

	adr.UpdatedAt = GetCurrentTime()

	// Update in repository
	if err := repo.UpdateADR(ctx, adr); err != nil {
		return err
	}

	// Event is emitted by the repository

	fmt.Printf("Updated ADR %s\n", c.adrID)

	return nil
}

// ============================================================================
// ADRSupersdeCommand marks an ADR as superseded
// ============================================================================

type ADRSupersdeCommand struct {
	Plugin      *TaskManagerPlugin
	project     string
	adrID       string
	supersdeBy  string
}

func (c *ADRSupersdeCommand) GetName() string {
	return "adr supersede"
}

func (c *ADRSupersdeCommand) GetDescription() string {
	return "Mark an ADR as superseded by another ADR"
}

func (c *ADRSupersdeCommand) GetUsage() string {
	return "dw task-manager adr supersede <adr-id> --by <new-adr-id>"
}

func (c *ADRSupersdeCommand) GetHelp() string {
	return `Marks an ADR as superseded by another ADR.

Examples:
  # Supersede an ADR
  dw task-manager adr supersede DW-adr-1 --by DW-adr-2`
}

func (c *ADRSupersdeCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--by":
			if i+1 < len(args) {
				c.supersdeBy = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "--") && c.adrID == "" {
				c.adrID = args[i]
			}
		}
	}

	if c.adrID == "" {
		return fmt.Errorf("ADR ID is required")
	}
	if c.supersdeBy == "" {
		return fmt.Errorf("--by is required")
	}

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Supersede ADR
	if err := repo.SupersedeADR(ctx, c.adrID, c.supersdeBy); err != nil {
		return err
	}

	// Event is emitted by the repository

	fmt.Printf("Superseded ADR %s with %s\n", c.adrID, c.supersdeBy)

	return nil
}

// ============================================================================
// ADRDeprecateCommand marks an ADR as deprecated
// ============================================================================

type ADRDeprecateCommand struct {
	Plugin  *TaskManagerPlugin
	project string
	adrID   string
}

func (c *ADRDeprecateCommand) GetName() string {
	return "adr deprecate"
}

func (c *ADRDeprecateCommand) GetDescription() string {
	return "Mark an ADR as deprecated"
}

func (c *ADRDeprecateCommand) GetUsage() string {
	return "dw task-manager adr deprecate <adr-id>"
}

func (c *ADRDeprecateCommand) GetHelp() string {
	return `Marks an ADR as deprecated.

Examples:
  # Deprecate an ADR
  dw task-manager adr deprecate DW-adr-1`
}

func (c *ADRDeprecateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "--") && c.adrID == "" {
				c.adrID = args[i]
			}
		}
	}

	if c.adrID == "" {
		return fmt.Errorf("ADR ID is required")
	}

	// Get repository for project
	repo, cleanup, err := c.Plugin.getRepositoryForProject(c.project)
	if err != nil {
		return err
	}
	defer cleanup()

	// Deprecate ADR
	if err := repo.DeprecateADR(ctx, c.adrID); err != nil {
		return err
	}

	// Event is emitted by the repository

	fmt.Printf("Deprecated ADR %s\n", c.adrID)

	return nil
}
