package task_manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kgatilin/darwinflow-pub/pkg/pluginsdk"
)

// Project name validation regex: alphanumeric + hyphens/underscores only
var projectNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ============================================================================
// ProjectCreateCommand creates a new project
// ============================================================================

type ProjectCreateCommand struct {
	Plugin      *TaskManagerPlugin
	projectName string
}

func (c *ProjectCreateCommand) GetName() string {
	return "project.create"
}

func (c *ProjectCreateCommand) GetDescription() string {
	return "Create a new project"
}

func (c *ProjectCreateCommand) GetUsage() string {
	return "dw task-manager project create <project-name>"
}

func (c *ProjectCreateCommand) GetHelp() string {
	return `Creates a new project with its own isolated database.

Each project maintains separate roadmaps, tracks, tasks, and iterations.
Project names must be alphanumeric with hyphens or underscores only.

Arguments:
  <project-name>  Name of the project to create (required)

Examples:
  # Create a test project
  dw task-manager project create test

  # Create a product project
  dw task-manager project create my-product

  # Create a project with underscores
  dw task-manager project create real_work

Notes:
  - Project name must be alphanumeric with hyphens/underscores only
  - Project databases are stored in .darwinflow/projects/<name>/roadmap.db
  - Use 'project switch' to change active project
  - Use '--project <name>' flag to override active project on any command`
}

func (c *ProjectCreateCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("project name is required")
	}
	c.projectName = args[0]

	// Validate project name
	if !projectNameRegex.MatchString(c.projectName) {
		return fmt.Errorf("invalid project name: must be alphanumeric with hyphens or underscores only")
	}

	// Check if project already exists
	projectDir := filepath.Join(c.Plugin.workingDir, ".darwinflow", "projects", c.projectName)
	if _, err := os.Stat(projectDir); err == nil {
		return fmt.Errorf("project already exists: %s", c.projectName)
	}

	// Create project directory
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Initialize database for this project
	db, err := c.Plugin.getProjectDatabase(c.projectName)
	if err != nil {
		return fmt.Errorf("failed to initialize project database: %w", err)
	}
	defer db.Close()

	fmt.Fprintf(cmdCtx.GetStdout(), "Project created successfully: %s\n", c.projectName)
	fmt.Fprintf(cmdCtx.GetStdout(), "Database: %s\n", filepath.Join(projectDir, "roadmap.db"))
	fmt.Fprintf(cmdCtx.GetStdout(), "\nTo switch to this project, run:\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  dw task-manager project switch %s\n", c.projectName)

	return nil
}

// ============================================================================
// ProjectListCommand lists all projects
// ============================================================================

type ProjectListCommand struct {
	Plugin *TaskManagerPlugin
}

func (c *ProjectListCommand) GetName() string {
	return "project.list"
}

func (c *ProjectListCommand) GetDescription() string {
	return "List all projects"
}

func (c *ProjectListCommand) GetUsage() string {
	return "dw task-manager project list"
}

func (c *ProjectListCommand) GetHelp() string {
	return `Lists all available projects.

The active project is marked with an asterisk (*).

Examples:
  dw task-manager project list

Output:
  * default       (active)
    test
    my-product

Notes:
  - Active project is used by default for all commands
  - Use 'project switch' to change active project
  - Use '--project <name>' flag to override active project on any command`
}

func (c *ProjectListCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	// Get projects directory
	projectsDir := filepath.Join(c.Plugin.workingDir, ".darwinflow", "projects")

	// Check if directory exists
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		fmt.Fprintf(cmdCtx.GetStdout(), "No projects found.\n")
		fmt.Fprintf(cmdCtx.GetStdout(), "Run 'dw task-manager project create <name>' to create one.\n")
		return nil
	}

	// Read directory entries
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return fmt.Errorf("failed to read projects directory: %w", err)
	}

	// Get active project
	activeProject, err := c.Plugin.getActiveProject()
	if err != nil {
		// If no active project, use "default"
		activeProject = "default"
	}

	// Collect project names
	projects := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	if len(projects) == 0 {
		fmt.Fprintf(cmdCtx.GetStdout(), "No projects found.\n")
		fmt.Fprintf(cmdCtx.GetStdout(), "Run 'dw task-manager project create <name>' to create one.\n")
		return nil
	}

	// Sort alphabetically
	sort.Strings(projects)

	// Display projects
	fmt.Fprintf(cmdCtx.GetStdout(), "Projects:\n\n")
	for _, project := range projects {
		if project == activeProject {
			fmt.Fprintf(cmdCtx.GetStdout(), "  * %s (active)\n", project)
		} else {
			fmt.Fprintf(cmdCtx.GetStdout(), "    %s\n", project)
		}
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "\nTotal: %d project(s)\n", len(projects))
	fmt.Fprintf(cmdCtx.GetStdout(), "\nTo switch active project:\n")
	fmt.Fprintf(cmdCtx.GetStdout(), "  dw task-manager project switch <name>\n")

	return nil
}

// ============================================================================
// ProjectSwitchCommand switches the active project
// ============================================================================

type ProjectSwitchCommand struct {
	Plugin      *TaskManagerPlugin
	projectName string
}

func (c *ProjectSwitchCommand) GetName() string {
	return "project.switch"
}

func (c *ProjectSwitchCommand) GetDescription() string {
	return "Switch the active project"
}

func (c *ProjectSwitchCommand) GetUsage() string {
	return "dw task-manager project switch <project-name>"
}

func (c *ProjectSwitchCommand) GetHelp() string {
	return `Switches the active project.

The active project is used by default for all commands unless
overridden with the --project flag.

Arguments:
  <project-name>  Name of the project to switch to (required)

Examples:
  # Switch to test project
  dw task-manager project switch test

  # Switch to default project
  dw task-manager project switch default

Notes:
  - Project must exist (create with 'project create')
  - All subsequent commands will use this project by default
  - Use '--project <name>' flag to temporarily override active project`
}

func (c *ProjectSwitchCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("project name is required")
	}
	c.projectName = args[0]

	// Validate project name
	if !projectNameRegex.MatchString(c.projectName) {
		return fmt.Errorf("invalid project name: must be alphanumeric with hyphens or underscores only")
	}

	// Check if project exists
	projectDir := filepath.Join(c.Plugin.workingDir, ".darwinflow", "projects", c.projectName)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("project does not exist: %s", c.projectName)
	}

	// Set as active project
	if err := c.Plugin.setActiveProject(c.projectName); err != nil {
		return fmt.Errorf("failed to set active project: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Switched to project: %s\n", c.projectName)

	return nil
}

// ============================================================================
// ProjectShowCommand shows the current active project
// ============================================================================

type ProjectShowCommand struct {
	Plugin *TaskManagerPlugin
}

func (c *ProjectShowCommand) GetName() string {
	return "project.show"
}

func (c *ProjectShowCommand) GetDescription() string {
	return "Show the current active project"
}

func (c *ProjectShowCommand) GetUsage() string {
	return "dw task-manager project show"
}

func (c *ProjectShowCommand) GetHelp() string {
	return `Displays the current active project name.

Examples:
  dw task-manager project show

Output:
  Active project: default

Notes:
  - The active project is used by default for all commands
  - Use 'project switch' to change active project
  - Use '--project <name>' flag to override active project on any command`
}

func (c *ProjectShowCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	activeProject, err := c.Plugin.getActiveProject()
	if err != nil {
		return fmt.Errorf("failed to get active project: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Active project: %s\n", activeProject)

	return nil
}

// ============================================================================
// ProjectDeleteCommand deletes a project
// ============================================================================

type ProjectDeleteCommand struct {
	Plugin      *TaskManagerPlugin
	projectName string
	force       bool
}

func (c *ProjectDeleteCommand) GetName() string {
	return "project.delete"
}

func (c *ProjectDeleteCommand) GetDescription() string {
	return "Delete a project"
}

func (c *ProjectDeleteCommand) GetUsage() string {
	return "dw task-manager project delete <project-name> [--force]"
}

func (c *ProjectDeleteCommand) GetHelp() string {
	return `Deletes a project and all its data.

This operation is permanent and deletes all roadmaps, tracks, tasks,
and iterations associated with the project.

Arguments:
  <project-name>  Name of the project to delete (required)

Flags:
  --force  Skip confirmation prompt (optional)

Examples:
  # Delete with confirmation prompt
  dw task-manager project delete test

  # Delete without confirmation
  dw task-manager project delete test --force

Notes:
  - This operation is permanent and cannot be undone
  - All data in the project will be deleted
  - You cannot delete the active project (switch to another project first)
  - Database file is permanently removed`
}

func (c *ProjectDeleteCommand) Execute(ctx context.Context, cmdCtx pluginsdk.CommandContext, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("project name is required")
	}
	c.projectName = args[0]

	// Check for --force flag
	for i := 1; i < len(args); i++ {
		if args[i] == "--force" {
			c.force = true
		}
	}

	// Validate project name
	if !projectNameRegex.MatchString(c.projectName) {
		return fmt.Errorf("invalid project name: must be alphanumeric with hyphens or underscores only")
	}

	// Check if project exists
	projectDir := filepath.Join(c.Plugin.workingDir, ".darwinflow", "projects", c.projectName)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("project does not exist: %s", c.projectName)
	}

	// Prevent deleting active project
	activeProject, err := c.Plugin.getActiveProject()
	if err == nil && activeProject == c.projectName {
		return fmt.Errorf("cannot delete active project: %s (switch to another project first)", c.projectName)
	}

	// Prompt for confirmation unless --force
	if !c.force {
		fmt.Fprintf(cmdCtx.GetStdout(), "Are you sure you want to delete project '%s'? This will delete all data. (y/N): ", c.projectName)
		var response string
		if _, err := fmt.Fscanln(cmdCtx.GetStdin(), &response); err != nil {
			fmt.Fprintf(cmdCtx.GetStdout(), "Deletion cancelled.\n")
			return nil
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Fprintf(cmdCtx.GetStdout(), "Deletion cancelled.\n")
			return nil
		}
	}

	// Delete project directory
	if err := os.RemoveAll(projectDir); err != nil {
		return fmt.Errorf("failed to delete project directory: %w", err)
	}

	fmt.Fprintf(cmdCtx.GetStdout(), "Project deleted successfully: %s\n", c.projectName)

	return nil
}
