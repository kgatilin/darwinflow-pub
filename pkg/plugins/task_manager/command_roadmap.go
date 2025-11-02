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
	return "roadmap init"
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
	return "roadmap show"
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
	return "roadmap update"
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

// RoadmapFullCommand displays the complete roadmap overview
type RoadmapFullCommand struct {
	Plugin   *TaskManagerPlugin
	project  string
	verbose  bool
	format   string
	sections []string
}

func (c *RoadmapFullCommand) GetName() string {
	return "roadmap full"
}

func (c *RoadmapFullCommand) GetDescription() string {
	return "Display complete roadmap overview in LLM-optimized format"
}

func (c *RoadmapFullCommand) GetUsage() string {
	return "dw task-manager roadmap full [--verbose] [--format json] [--sections <list>]"
}

func (c *RoadmapFullCommand) GetHelp() string {
	return `Displays the complete roadmap overview in LLM-optimized markdown format.

Shows:
  - Vision and success criteria
  - All tracks with their tasks (titles only by default)
  - All iterations with assigned tasks and progress
  - Backlog (tasks not in any iteration)

Flags:
  --verbose             Include task descriptions and additional details
  --format <format>     Output format (default: markdown)
                        Values: markdown, json
  --sections <list>     Only show specific sections (comma-separated)
                        Values: vision, tracks, iterations, backlog

Examples:
  # Basic overview
  dw task-manager roadmap full

  # Verbose with full details
  dw task-manager roadmap full --verbose

  # Only show tracks and iterations
  dw task-manager roadmap full --sections tracks,iterations

  # Output as JSON
  dw task-manager roadmap full --format json

Notes:
  - Uses status icons: ✅ (complete), ⏸️ (planned/in-progress), ○ (todo)
  - Optimized for LLM consumption
  - JSON format includes all metadata`
}

func (c *RoadmapFullCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Parse flags
	c.format = "markdown"
	c.sections = nil

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 < len(args) {
				c.project = args[i+1]
				i++
			}
		case "--verbose":
			c.verbose = true
		case "--format":
			if i+1 < len(args) {
				c.format = args[i+1]
				i++
			}
		case "--sections":
			if i+1 < len(args) {
				c.sections = splitSections(args[i+1])
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

	// Get all tracks
	tracks, err := repo.ListTracks(ctx, roadmap.ID, TrackFilters{})
	if err != nil {
		return fmt.Errorf("failed to list tracks: %w", err)
	}

	// Get all tasks
	tasks, err := repo.ListTasks(ctx, TaskFilters{})
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Get all iterations
	iterations, err := repo.ListIterations(ctx)
	if err != nil {
		return fmt.Errorf("failed to list iterations: %w", err)
	}

	// Format output based on requested format
	if c.format == "json" {
		return c.outputJSON(cmdCtx, roadmap, tracks, tasks, iterations)
	}

	return c.outputMarkdown(ctx, cmdCtx, roadmap, tracks, tasks, iterations, repo)
}

func (c *RoadmapFullCommand) outputJSON(cmdCtx pluginsdk.CommandContext, roadmap *RoadmapEntity, tracks []*TrackEntity, tasks []*TaskEntity, iterations []*IterationEntity) error {
	output := map[string]interface{}{
		"roadmap":    roadmap,
		"tracks":     tracks,
		"tasks":      tasks,
		"iterations": iterations,
	}

	// Use fmt.Sprintf to format the output
	data := fmt.Sprintf("%v", output)
	fmt.Fprintf(cmdCtx.GetStdout(), "%s\n", data)
	return nil
}

func (c *RoadmapFullCommand) outputMarkdown(ctx context.Context, cmdCtx pluginsdk.CommandContext, roadmap *RoadmapEntity, tracks []*TrackEntity, tasks []*TaskEntity, iterations []*IterationEntity, repo RoadmapRepository) error {
	stdout := cmdCtx.GetStdout()

	// Vision section
	if c.shouldShowSection("vision") {
		fmt.Fprintf(stdout, "# Roadmap: %s\n\n", roadmap.ID)
		fmt.Fprintf(stdout, "## Vision\n%s\n\n", roadmap.Vision)
		fmt.Fprintf(stdout, "## Success Criteria\n%s\n\n", roadmap.SuccessCriteria)
	}

	// Tracks section
	if c.shouldShowSection("tracks") {
		fmt.Fprintf(stdout, "## Tracks\n\n")
		for _, track := range tracks {
			statusIcon := getStatusIcon(track.Status)
			fmt.Fprintf(stdout, "### %s %s\n", statusIcon, track.Title)
			fmt.Fprintf(stdout, "**ID**: %s | **Status**: %s | **Rank**: %d\n", track.ID, track.Status, track.Rank)

			if c.verbose && track.Description != "" {
				fmt.Fprintf(stdout, "**Description**: %s\n", track.Description)
			}

			// Get tasks for this track
			trackTasks := filterTasksByTrack(tasks, track.ID)
			if len(trackTasks) > 0 {
				fmt.Fprintf(stdout, "**Progress**: %d/%d tasks complete\n", countCompleteTasks(trackTasks), len(trackTasks))
				fmt.Fprintf(stdout, "**Tasks**:\n")
				for _, task := range trackTasks {
					taskIcon := getStatusIcon(task.Status)
					fmt.Fprintf(stdout, "- %s %s", taskIcon, task.Title)
					if c.verbose && task.Description != "" {
						fmt.Fprintf(stdout, " - %s", task.Description)
					}
					fmt.Fprintf(stdout, "\n")
				}
			}
			fmt.Fprintf(stdout, "\n")
		}
	}

	// Iterations section
	if c.shouldShowSection("iterations") {
		fmt.Fprintf(stdout, "## Iterations\n\n")
		for _, iter := range iterations {
			statusIcon := getStatusIcon(iter.Status)
			fmt.Fprintf(stdout, "### %s Iteration %d: %s\n", statusIcon, iter.Number, iter.Name)
			fmt.Fprintf(stdout, "**Status**: %s | **Goal**: %s\n", iter.Status, iter.Goal)

			if c.verbose && iter.Deliverable != "" {
				fmt.Fprintf(stdout, "**Deliverable**: %s\n", iter.Deliverable)
			}

			// Get tasks for this iteration
			iterTasks := filterTasksByIDs(tasks, iter.TaskIDs)
			if len(iterTasks) > 0 {
				fmt.Fprintf(stdout, "**Progress**: %d/%d tasks complete (%.0f%%)\n",
					countCompleteTasks(iterTasks), len(iterTasks),
					float64(countCompleteTasks(iterTasks))/float64(len(iterTasks))*100)
				fmt.Fprintf(stdout, "**Tasks**:\n")
				for _, task := range iterTasks {
					taskIcon := getStatusIcon(task.Status)
					fmt.Fprintf(stdout, "- %s %s", taskIcon, task.Title)
					if c.verbose && task.Description != "" {
						fmt.Fprintf(stdout, " - %s", task.Description)
					}
					fmt.Fprintf(stdout, "\n")
				}
			}
			fmt.Fprintf(stdout, "\n")
		}
	}

	// Backlog section
	if c.shouldShowSection("backlog") {
		backlogTasks := getBacklogTasks(tasks, iterations)
		if len(backlogTasks) > 0 {
			fmt.Fprintf(stdout, "## Backlog\n\n")
			fmt.Fprintf(stdout, "%d tasks not assigned to any iteration:\n\n", len(backlogTasks))
			for _, task := range backlogTasks {
				taskIcon := getStatusIcon(task.Status)
				fmt.Fprintf(stdout, "- %s %s", taskIcon, task.Title)
				if c.verbose && task.Description != "" {
					fmt.Fprintf(stdout, " - %s", task.Description)
				}
				fmt.Fprintf(stdout, "\n")
			}
		}
	}

	return nil
}

func (c *RoadmapFullCommand) shouldShowSection(section string) bool {
	if c.sections == nil {
		return true
	}
	for _, s := range c.sections {
		if s == section {
			return true
		}
	}
	return false
}

// Helper functions for roadmap full command

func filterTasksByTrack(tasks []*TaskEntity, trackID string) []*TaskEntity {
	var result []*TaskEntity
	for _, task := range tasks {
		if task.TrackID == trackID {
			result = append(result, task)
		}
	}
	return result
}

func filterTasksByIDs(tasks []*TaskEntity, taskIDs []string) []*TaskEntity {
	var result []*TaskEntity
	taskMap := make(map[string]*TaskEntity)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	for _, id := range taskIDs {
		if task, ok := taskMap[id]; ok {
			result = append(result, task)
		}
	}
	return result
}

func countCompleteTasks(tasks []*TaskEntity) int {
	count := 0
	for _, task := range tasks {
		if task.Status == "done" {
			count++
		}
	}
	return count
}

func getBacklogTasks(tasks []*TaskEntity, iterations []*IterationEntity) []*TaskEntity {
	// Build set of all task IDs in iterations
	inIteration := make(map[string]bool)
	for _, iter := range iterations {
		for _, taskID := range iter.TaskIDs {
			inIteration[taskID] = true
		}
	}

	// Return tasks not in any iteration
	var backlog []*TaskEntity
	for _, task := range tasks {
		if !inIteration[task.ID] {
			backlog = append(backlog, task)
		}
	}
	return backlog
}

func splitSections(sectionStr string) []string {
	var sections []string
	for _, s := range splitString(sectionStr, ",") {
		trimmed := trimSpace(s)
		if trimmed != "" {
			sections = append(sections, trimmed)
		}
	}
	return sections
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}
