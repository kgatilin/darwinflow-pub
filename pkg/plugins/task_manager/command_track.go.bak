package task_manager

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// ============================================================================
// TrackCreateCommand creates a new track
// ============================================================================

type TrackCreateCommand struct {
	Plugin      *TaskManagerPlugin
	id          string
	title       string
	description string
	priority    string
}

func (c *TrackCreateCommand) GetName() string {
	return "track.create"
}

func (c *TrackCreateCommand) GetDescription() string {
	return "Create a new track"
}

func (c *TrackCreateCommand) GetUsage() string {
	return "dw task-manager track create --id <id> --title <title> [--description <desc>] [--priority <priority>]"
}

func (c *TrackCreateCommand) GetHelp() string {
	return `Creates a new track in the active roadmap.

A track represents a major work area with multiple tasks and iterations.
All tracks must belong to an active roadmap - create one first with
'dw task-manager roadmap init'.

Flags:
  --id <id>                Track ID (required, format: track-<slug>)
  --title <title>          Track title (required)
  --description <desc>     Track description (optional)
  --priority <priority>    Track priority (optional, default: medium)
                          Values: critical, high, medium, low

Examples:
  # Create a basic track
  dw task-manager track create --id track-plugin-system --title "Plugin System"

  # Create with full details
  dw task-manager track create \
    --id track-plugin-system \
    --title "Plugin System" \
    --description "Implement extensible plugin architecture" \
    --priority high

Notes:
  - Track ID must follow convention: track-<slug> (alphanumeric with hyphens)
  - An active roadmap must exist (create with 'dw task-manager roadmap init')
  - Initial status is automatically set to 'not-started'
  - No dependencies are added initially (use track add-dependency)`
}

func (c *TrackCreateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	c.priority = "medium" // default
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 < len(args) {
				c.id = args[i+1]
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
	if c.id == "" || c.title == "" {
		return fmt.Errorf("--id and --title are required")
	}

	// Get repository
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("database repository not initialized")
	}

	// Get active roadmap (required)
	roadmap, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			return fmt.Errorf("no active roadmap found - create one first with 'dw task-manager roadmap init'")
		}
		return fmt.Errorf("failed to get active roadmap: %w", err)
	}

	// Create track with initial status
	now := time.Now().UTC()
	track, err := NewTrackEntity(
		c.id,
		roadmap.ID,
		c.title,
		c.description,
		"not-started",
		c.priority,
		[]string{}, // no dependencies initially
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create track: %w", err)
	}

	// Save to repository
	if err := repo.SaveTrack(ctx, track); err != nil {
		return fmt.Errorf("failed to save track: %w", err)
	}

	// Display success message
	fmt.Fprintf(cmdCtx.GetStdout(), "Track created successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "ID:          %s\n", track.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "Title:       %s\n", track.Title)
	if track.Description != "" {
		fmt.Fprintf(cmdCtx.GetStdout(), "Description: %s\n", track.Description)
	}
	fmt.Fprintf(cmdCtx.GetStdout(), "Status:      %s\n", track.Status)
	fmt.Fprintf(cmdCtx.GetStdout(), "Priority:    %s\n", track.Priority)
	fmt.Fprintf(cmdCtx.GetStdout(), "Roadmap:     %s\n", track.RoadmapID)

	return nil
}

// ============================================================================
// TrackListCommand lists tracks with optional filtering
// ============================================================================

type TrackListCommand struct {
	Plugin   *TaskManagerPlugin
	status   string
	priority string
}

func (c *TrackListCommand) GetName() string {
	return "track.list"
}

func (c *TrackListCommand) GetDescription() string {
	return "List tracks"
}

func (c *TrackListCommand) GetUsage() string {
	return "dw task-manager track list [--status <status>] [--priority <priority>]"
}

func (c *TrackListCommand) GetHelp() string {
	return `Lists all tracks in the active roadmap with optional filtering.

Flags:
  --status <status>      Filter by status (can be comma-separated)
                         Values: not-started, in-progress, complete, blocked, waiting
  --priority <priority>  Filter by priority (can be comma-separated)
                         Values: critical, high, medium, low

Examples:
  # List all tracks
  dw task-manager track list

  # List in-progress tracks
  dw task-manager track list --status in-progress

  # List critical and high priority tracks
  dw task-manager track list --priority critical,high

  # Combine filters
  dw task-manager track list --status in-progress,blocked --priority critical

Output:
  A table showing: ID, Title, Status, Priority, Dependencies count`
}

func (c *TrackListCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
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

	// Get repository
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("database repository not initialized")
	}

	// Get active roadmap
	roadmap, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "No active roadmap found.\n")
			return nil
		}
		return fmt.Errorf("failed to get active roadmap: %w", err)
	}

	// Build filters
	filters := TrackFilters{}
	if c.status != "" {
		filters.Status = strings.Split(strings.TrimSpace(c.status), ",")
		for i, s := range filters.Status {
			filters.Status[i] = strings.TrimSpace(s)
		}
	}
	if c.priority != "" {
		filters.Priority = strings.Split(strings.TrimSpace(c.priority), ",")
		for i, p := range filters.Priority {
			filters.Priority[i] = strings.TrimSpace(p)
		}
	}

	// List tracks
	tracks, err := repo.ListTracks(ctx, roadmap.ID, filters)
	if err != nil {
		return fmt.Errorf("failed to list tracks: %w", err)
	}

	// Display tracks
	if len(tracks) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "No tracks found.\n")
		return nil
	}

	// Print header
	fmt.Fprintf(cmdCtx.GetStdout(), "%-25s %-30s %-12s %-10s %s\n",
		"ID", "Title", "Status", "Priority", "Dependencies")
	fmt.Fprintf(cmdCtx.GetStdout(), "%s\n",
		strings.Repeat("-", 90))

	// Print each track
	for _, track := range tracks {
		depCount := len(track.Dependencies)
		depStr := fmt.Sprintf("%d", depCount)
		fmt.Fprintf(cmdCtx.GetStdout(), "%-25s %-30s %-12s %-10s %s\n",
			track.ID, truncateString(track.Title, 29), track.Status, track.Priority, depStr)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "\nTotal: %d track(s)\n", len(tracks))
	return nil
}

// ============================================================================
// TrackShowCommand displays a specific track
// ============================================================================

type TrackShowCommand struct {
	Plugin  *TaskManagerPlugin
	trackID string
}

func (c *TrackShowCommand) GetName() string {
	return "track.show"
}

func (c *TrackShowCommand) GetDescription() string {
	return "Show track details"
}

func (c *TrackShowCommand) GetUsage() string {
	return "dw task-manager track show <track-id>"
}

func (c *TrackShowCommand) GetHelp() string {
	return `Displays detailed information about a specific track.

Arguments:
  <track-id>  The ID of the track to display (required)

Examples:
  dw task-manager track show track-plugin-system

Output:
  Track details including:
  - Basic info (ID, title, description)
  - Status and priority
  - Dependencies (tracks this depends on)
  - Created/Updated timestamps`
}

func (c *TrackShowCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get track ID from arguments
	if len(args) == 0 {
		return fmt.Errorf("track ID is required")
	}
	c.trackID = args[0]

	// Get repository
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("database repository not initialized")
	}

	// Get track
	track, err := repo.GetTrack(ctx, c.trackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Track not found: %s\n", c.trackID)
			return nil
		}
		return fmt.Errorf("failed to get track: %w", err)
	}

	// Display track details
	fmt.Fprintf(cmdCtx.GetStdout(), "Track: %s\n", track.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Title:       %s\n", track.Title)
	if track.Description != "" {
		fmt.Fprintf(cmdCtx.GetStdout(), "  Description: %s\n", track.Description)
	}
	fmt.Fprintf(cmdCtx.GetStdout(), "  Status:      %s\n", track.Status)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Priority:    %s\n", track.Priority)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Roadmap:     %s\n", track.RoadmapID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Created:     %s\n", track.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(cmdCtx.GetStdout(), "  Updated:     %s\n", track.UpdatedAt.Format(time.RFC3339))

	// Display dependencies
	if len(track.Dependencies) > 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "  Dependencies:\n")
		for _, dep := range track.Dependencies {
			fmt.Fprintf(cmdCtx.GetStdout(), "    - %s\n", dep)
		}
	} else {
		fmt.Fprintf(cmdCtx.GetStdout(), "  Dependencies: none\n")
	}

	return nil
}

// ============================================================================
// TrackUpdateCommand updates track properties
// ============================================================================

type TrackUpdateCommand struct {
	Plugin      *TaskManagerPlugin
	trackID     string
	title       *string
	description *string
	status      *string
	priority    *string
}

func (c *TrackUpdateCommand) GetName() string {
	return "track.update"
}

func (c *TrackUpdateCommand) GetDescription() string {
	return "Update track properties"
}

func (c *TrackUpdateCommand) GetUsage() string {
	return "dw task-manager track update <track-id> [--title <title>] [--description <desc>] [--status <status>] [--priority <priority>]"
}

func (c *TrackUpdateCommand) GetHelp() string {
	return `Updates properties of an existing track.

Arguments:
  <track-id>  The ID of the track to update (required)

Flags:
  --title <title>        New track title
  --description <desc>   New track description
  --status <status>      New track status
                         Values: not-started, in-progress, complete, blocked, waiting
  --priority <priority>  New track priority
                         Values: critical, high, medium, low

Examples:
  # Update title
  dw task-manager track update track-plugin-system --title "Core Plugin System"

  # Update status
  dw task-manager track update track-plugin-system --status in-progress

  # Update multiple fields
  dw task-manager track update track-plugin-system \
    --status in-progress \
    --priority critical

Notes:
  - At least one flag must be provided
  - Updated_at timestamp is automatically updated`
}

func (c *TrackUpdateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("track ID is required")
	}
	c.trackID = args[0]

	// Parse flags (skip first arg which is track ID)
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--title":
			if i+1 < len(args) {
				c.title = &args[i+1]
				i++
			}
		case "--description":
			if i+1 < len(args) {
				c.description = &args[i+1]
				i++
			}
		case "--status":
			if i+1 < len(args) {
				c.status = &args[i+1]
				i++
			}
		case "--priority":
			if i+1 < len(args) {
				c.priority = &args[i+1]
				i++
			}
		}
	}

	// At least one flag must be provided
	if c.title == nil && c.description == nil && c.status == nil && c.priority == nil {
		return fmt.Errorf("at least one flag must be provided (--title, --description, --status, or --priority)")
	}

	// Get repository
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("database repository not initialized")
	}

	// Get existing track
	track, err := repo.GetTrack(ctx, c.trackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Track not found: %s\n", c.trackID)
			return nil
		}
		return fmt.Errorf("failed to get track: %w", err)
	}

	// Update provided fields
	if c.title != nil {
		track.Title = *c.title
	}
	if c.description != nil {
		track.Description = *c.description
	}
	if c.status != nil {
		track.Status = *c.status
	}
	if c.priority != nil {
		track.Priority = *c.priority
	}
	track.UpdatedAt = time.Now().UTC()

	// Save to repository
	if err := repo.UpdateTrack(ctx, track); err != nil {
		return fmt.Errorf("failed to update track: %w", err)
	}

	// Display success message
	fmt.Fprintf(cmdCtx.GetStdout(), "Track updated successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "ID:          %s\n", track.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "Title:       %s\n", track.Title)
	if track.Description != "" {
		fmt.Fprintf(cmdCtx.GetStdout(), "Description: %s\n", track.Description)
	}
	fmt.Fprintf(cmdCtx.GetStdout(), "Status:      %s\n", track.Status)
	fmt.Fprintf(cmdCtx.GetStdout(), "Priority:    %s\n", track.Priority)
	fmt.Fprintf(cmdCtx.GetStdout(), "Updated:     %s\n", track.UpdatedAt.Format(time.RFC3339))

	return nil
}

// ============================================================================
// TrackDeleteCommand deletes a track
// ============================================================================

type TrackDeleteCommand struct {
	Plugin  *TaskManagerPlugin
	trackID string
	force   bool
}

func (c *TrackDeleteCommand) GetName() string {
	return "track.delete"
}

func (c *TrackDeleteCommand) GetDescription() string {
	return "Delete a track"
}

func (c *TrackDeleteCommand) GetUsage() string {
	return "dw task-manager track delete <track-id> [--force]"
}

func (c *TrackDeleteCommand) GetHelp() string {
	return `Deletes a track and all its associated tasks and iterations.

Arguments:
  <track-id>  The ID of the track to delete (required)

Flags:
  --force     Skip confirmation prompt (optional)

Examples:
  # Delete with confirmation prompt
  dw task-manager track delete track-plugin-system

  # Delete without confirmation
  dw task-manager track delete track-plugin-system --force

Notes:
  - This operation is permanent
  - All tasks and iterations in the track will be deleted
  - Any tracks depending on this one will have the dependency removed`
}

func (c *TrackDeleteCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("track ID is required")
	}
	c.trackID = args[0]

	// Check for --force flag
	for i := 1; i < len(args); i++ {
		if args[i] == "--force" {
			c.force = true
		}
	}

	// Get repository
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("database repository not initialized")
	}

	// Get track to verify it exists
	track, err := repo.GetTrack(ctx, c.trackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Track not found: %s\n", c.trackID)
			return nil
		}
		return fmt.Errorf("failed to get track: %w", err)
	}

	// Prompt for confirmation unless --force
	if !c.force {
		fmt.Fprintf(cmdCtx.GetStdout(), "Are you sure you want to delete track '%s'? (y/N): ", track.Title)
		scanner := bufio.NewScanner(cmdCtx.GetStdin())
		if !scanner.Scan() {
			fmt.Fprintf(cmdCtx.GetStdout(), "Deletion cancelled.\n")
			return nil
		}
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if response != "y" && response != "yes" {
			fmt.Fprintf(cmdCtx.GetStdout(), "Deletion cancelled.\n")
			return nil
		}
	}

	// Delete track (cascades to tasks)
	if err := repo.DeleteTrack(ctx, c.trackID); err != nil {
		return fmt.Errorf("failed to delete track: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Track deleted successfully: %s\n", c.trackID)
	return nil
}

// ============================================================================
// TrackAddDependencyCommand adds a dependency between tracks
// ============================================================================

type TrackAddDependencyCommand struct {
	Plugin      *TaskManagerPlugin
	trackID     string
	dependsOnID string
}

func (c *TrackAddDependencyCommand) GetName() string {
	return "track.add-dependency"
}

func (c *TrackAddDependencyCommand) GetDescription() string {
	return "Add a dependency between tracks"
}

func (c *TrackAddDependencyCommand) GetUsage() string {
	return "dw task-manager track add-dependency <track-id> <depends-on-id>"
}

func (c *TrackAddDependencyCommand) GetHelp() string {
	return `Adds a dependency from one track to another.

This indicates that <track-id> depends on <depends-on-id>.
The system automatically detects and prevents circular dependencies.

Arguments:
  <track-id>      The track that will depend on another (required)
  <depends-on-id> The track to depend on (required)

Examples:
  # Make plugin-system depend on framework-core
  dw task-manager track add-dependency track-plugin-system track-framework-core

  # Track A depends on both B and C
  dw task-manager track add-dependency track-a track-b
  dw task-manager track add-dependency track-a track-c

Notes:
  - Circular dependencies are automatically detected and prevented
  - Both tracks must exist
  - A track cannot depend on itself
  - If dependency already exists, no action is taken`
}

func (c *TrackAddDependencyCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("track ID and depends-on ID are required")
	}
	c.trackID = args[0]
	c.dependsOnID = args[1]

	// Get repository
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("database repository not initialized")
	}

	// Validate both tracks exist
	track, err := repo.GetTrack(ctx, c.trackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Track not found: %s\n", c.trackID)
			return nil
		}
		return fmt.Errorf("failed to get track: %w", err)
	}

	depTrack, err := repo.GetTrack(ctx, c.dependsOnID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Track not found: %s\n", c.dependsOnID)
			return nil
		}
		return fmt.Errorf("failed to get dependency track: %w", err)
	}

	// Add dependency
	if err := repo.AddTrackDependency(ctx, c.trackID, c.dependsOnID); err != nil {
		if errors.Is(err, pluginsdk.ErrAlreadyExists) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Dependency already exists: %s -> %s\n", c.trackID, c.dependsOnID)
			return nil
		}
		return fmt.Errorf("failed to add dependency: %w", err)
	}

	// Validate no cycles
	if err := repo.ValidateNoCycles(ctx, c.trackID); err != nil {
		// Rollback the dependency addition
		_ = repo.RemoveTrackDependency(ctx, c.trackID, c.dependsOnID)
		return fmt.Errorf("circular dependency detected: %w", err)
	}

	// Display success message
	fmt.Fprintf(cmdCtx.GetStdout(), "Dependency added successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  %s depends on %s\n", track.Title, depTrack.Title)
	fmt.Fprintf(cmdCtx.GetStdout(), "  (%s -> %s)\n", c.trackID, c.dependsOnID)

	return nil
}

// ============================================================================
// TrackRemoveDependencyCommand removes a dependency between tracks
// ============================================================================

type TrackRemoveDependencyCommand struct {
	Plugin      *TaskManagerPlugin
	trackID     string
	dependsOnID string
}

func (c *TrackRemoveDependencyCommand) GetName() string {
	return "track.remove-dependency"
}

func (c *TrackRemoveDependencyCommand) GetDescription() string {
	return "Remove a dependency between tracks"
}

func (c *TrackRemoveDependencyCommand) GetUsage() string {
	return "dw task-manager track remove-dependency <track-id> <depends-on-id>"
}

func (c *TrackRemoveDependencyCommand) GetHelp() string {
	return `Removes a dependency from one track to another.

Arguments:
  <track-id>      The track to remove dependency from (required)
  <depends-on-id> The track to stop depending on (required)

Examples:
  # Remove dependency: plugin-system no longer depends on framework-core
  dw task-manager track remove-dependency track-plugin-system track-framework-core

Notes:
  - If the dependency doesn't exist, no error is returned
  - Both tracks must exist`
}

func (c *TrackRemoveDependencyCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("track ID and depends-on ID are required")
	}
	c.trackID = args[0]
	c.dependsOnID = args[1]

	// Get repository
	repo := c.Plugin.GetRepository()
	if repo == nil {
		return fmt.Errorf("database repository not initialized")
	}

	// Validate tracks exist
	_, err := repo.GetTrack(ctx, c.trackID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Track not found: %s\n", c.trackID)
			return nil
		}
		return fmt.Errorf("failed to get track: %w", err)
	}

	_, err = repo.GetTrack(ctx, c.dependsOnID)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Track not found: %s\n", c.dependsOnID)
			return nil
		}
		return fmt.Errorf("failed to get dependency track: %w", err)
	}

	// Remove dependency
	if err := repo.RemoveTrackDependency(ctx, c.trackID, c.dependsOnID); err != nil {
		// If dependency doesn't exist, don't treat as error
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "Dependency does not exist: %s -> %s\n", c.trackID, c.dependsOnID)
			return nil
		}
		return fmt.Errorf("failed to remove dependency: %w", err)
	}

	// Display success message
	fmt.Fprintf(cmdCtx.GetStdout(), "Dependency removed successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  %s no longer depends on %s\n", c.trackID, c.dependsOnID)

	return nil
}

