package task_manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// RoadmapInitCommand initializes a new roadmap
type RoadmapInitCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *RoadmapInitCommand) GetName() string {
	return "roadmap.init"
}

func (c *RoadmapInitCommand) GetDescription() string {
	return "Initialize a new roadmap"
}

func (c *RoadmapInitCommand) GetUsage() string {
	return "dw task-manager roadmap init --vision <vision> --success-criteria <criteria>"
}

func (c *RoadmapInitCommand) GetHelp() string {
	return `Creates a new roadmap with a vision statement and success criteria.

Only one active roadmap can exist at a time. If you need to replace the current
roadmap, delete it first using 'dw task-manager roadmap delete'.

Flags:
  --vision <vision>              The vision statement for the roadmap (required)
  --success-criteria <criteria>  Success criteria for the roadmap (required)

Examples:
  # Create a simple roadmap
  dw task-manager roadmap init \
    --vision "Build extensible framework" \
    --success-criteria "Support 10 plugins"

  # With multi-line vision
  dw task-manager roadmap init \
    --vision "Create unified productivity platform" \
    --success-criteria "100% test coverage, zero violations"

Notes:
  - Vision must be non-empty
  - Success criteria must be non-empty
  - Only one roadmap can be active at a time`
}

func (c *RoadmapInitCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	var vision, successCriteria string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--vision":
			if i+1 < len(args) {
				vision = args[i+1]
				i++
			}
		case "--success-criteria":
			if i+1 < len(args) {
				successCriteria = args[i+1]
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

	// Validate required flags
	if vision == "" {
		return fmt.Errorf("--vision is required")
	}
	if successCriteria == "" {
		return fmt.Errorf("--success-criteria is required")
	}

	// Check if roadmap already exists
	existing, err := repo.GetActiveRoadmap(ctx)
	if err != nil && !errors.Is(err, pluginsdk.ErrNotFound) {
		return fmt.Errorf("failed to check existing roadmap: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("roadmap already exists: %s - delete it first to create a new one", existing.ID)
	}

	// Generate roadmap ID
	roadmapID := fmt.Sprintf("roadmap-%d", time.Now().UnixNano())
	now := time.Now().UTC()

	// Create roadmap
	roadmap, err := NewRoadmapEntity(roadmapID, vision, successCriteria, now, now)
	if err != nil {
		return fmt.Errorf("failed to create roadmap entity: %w", err)
	}

	// Save to repository
	if err := repo.SaveRoadmap(ctx, roadmap); err != nil {
		return fmt.Errorf("failed to save roadmap: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Roadmap created successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "ID:                %s\n", roadmap.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "Vision:            %s\n", roadmap.Vision)
	fmt.Fprintf(cmdCtx.GetStdout(), "Success Criteria:  %s\n", roadmap.SuccessCriteria)

	return nil
}

// RoadmapShowCommand displays the current roadmap
type RoadmapShowCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *RoadmapShowCommand) GetName() string {
	return "roadmap.show"
}

func (c *RoadmapShowCommand) GetDescription() string {
	return "Display the current roadmap"
}

func (c *RoadmapShowCommand) GetUsage() string {
	return "dw task-manager roadmap show"
}

func (c *RoadmapShowCommand) GetHelp() string {
	return `Displays the details of the current active roadmap.

If no roadmap exists, you can create one using:
  dw task-manager roadmap init --vision <vision> --success-criteria <criteria>

Examples:
  dw task-manager roadmap show

Output:
  ID:                roadmap-1234567890
  Vision:            Build extensible framework
  Success Criteria:  Support 10 plugins
  Created:           2025-10-31T10:00:00Z
  Updated:           2025-10-31T10:00:00Z`
}

func (c *RoadmapShowCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
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

	// Get active roadmap
	roadmap, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "No roadmap found.\n")
			fmt.Fprintf(cmdCtx.GetStdout(), "Run 'dw task-manager roadmap init' to create one.\n")
			return nil
		}
		return fmt.Errorf("failed to get roadmap: %w", err)
	}

	// Display roadmap details
	fmt.Fprintf(cmdCtx.GetStdout(), "Roadmap:\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  ID:                %s\n", roadmap.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Vision:            %s\n", roadmap.Vision)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Success Criteria:  %s\n", roadmap.SuccessCriteria)
	fmt.Fprintf(cmdCtx.GetStdout(), "  Created:           %s\n", roadmap.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(cmdCtx.GetStdout(), "  Updated:           %s\n", roadmap.UpdatedAt.Format(time.RFC3339))

	return nil
}

// RoadmapUpdateCommand updates the current roadmap
type RoadmapUpdateCommand struct {
	Plugin  *TaskManagerPlugin
	project string
}

func (c *RoadmapUpdateCommand) GetName() string {
	return "roadmap.update"
}

func (c *RoadmapUpdateCommand) GetDescription() string {
	return "Update the current roadmap"
}

func (c *RoadmapUpdateCommand) GetUsage() string {
	return "dw task-manager roadmap update [--vision <vision>] [--success-criteria <criteria>]"
}

func (c *RoadmapUpdateCommand) GetHelp() string {
	return `Updates properties of the current active roadmap.

At least one flag must be provided to update.

Flags:
  --vision <vision>              New vision statement
  --success-criteria <criteria>  New success criteria

Examples:
  # Update vision
  dw task-manager roadmap update --vision "Create unified platform"

  # Update both
  dw task-manager roadmap update \
    --vision "New vision" \
    --success-criteria "New criteria"

Notes:
  - At least one flag is required
  - Run 'dw task-manager roadmap show' to see current values
  - Updated_at timestamp is automatically updated`
}

func (c *RoadmapUpdateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	var vision, successCriteria *string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--vision":
			if i+1 < len(args) {
				v := args[i+1]
				vision = &v
				i++
			}
		case "--success-criteria":
			if i+1 < len(args) {
				sc := args[i+1]
				successCriteria = &sc
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

	// At least one flag must be provided
	if vision == nil && successCriteria == nil {
		return fmt.Errorf("at least one flag must be provided (--vision or --success-criteria)")
	}

	// Get active roadmap
	roadmap, err := repo.GetActiveRoadmap(ctx)
	if err != nil {
		if errors.Is(err, pluginsdk.ErrNotFound) {
			fmt.Fprintf(cmdCtx.GetStdout(), "No roadmap found.\n")
			fmt.Fprintf(cmdCtx.GetStdout(), "Run 'dw task-manager roadmap init' to create one.\n")
			return nil
		}
		return fmt.Errorf("failed to get roadmap: %w", err)
	}

	// Update provided fields
	if vision != nil {
		roadmap.Vision = *vision
	}
	if successCriteria != nil {
		roadmap.SuccessCriteria = *successCriteria
	}
	roadmap.UpdatedAt = time.Now().UTC()

	// Save to repository
	if err := repo.UpdateRoadmap(ctx, roadmap); err != nil {
		return fmt.Errorf("failed to update roadmap: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Roadmap updated successfully\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "ID:                %s\n", roadmap.ID)
	fmt.Fprintf(cmdCtx.GetStdout(), "Vision:            %s\n", roadmap.Vision)
	fmt.Fprintf(cmdCtx.GetStdout(), "Success Criteria:  %s\n", roadmap.SuccessCriteria)
	fmt.Fprintf(cmdCtx.GetStdout(), "Updated:           %s\n", roadmap.UpdatedAt.Format(time.RFC3339))

	return nil
}
