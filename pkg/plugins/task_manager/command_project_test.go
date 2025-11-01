package task_manager_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kgatilin/darwinflow-pub/pkg/plugins/task_manager"
)

// mockCommandContext is defined in other test files
// We'll use the existing one

func TestProjectCreateCommand(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create plugin
	logger := &MockLogger{}
	plugin, err := task_manager.NewTaskManagerPlugin(logger, tmpDir, nil)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Create command
	cmd := &task_manager.ProjectCreateCommand{Plugin: plugin}

	// Execute
	ctx := context.Background()
	stdout := bytes.NewBuffer(nil)
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     stdout,
		stdin:      bytes.NewReader(nil),
		logger:     logger,
	}

	err = cmd.Execute(ctx, cmdCtx, []string{"test-project"})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Verify project directory was created
	projectDir := filepath.Join(tmpDir, ".darwinflow", "projects", "test-project")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Fatal("project directory was not created")
	}

	// Verify database was created
	dbPath := filepath.Join(projectDir, "roadmap.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("project database was not created")
	}

	// Verify output
	output := stdout.String()
	if !strings.Contains(output, "Project created successfully: test-project") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestProjectCreateCommand_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	logger := &MockLogger{}
	plugin, _ := task_manager.NewTaskManagerPlugin(logger, tmpDir, nil)

	cmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	ctx := context.Background()
	stdout := bytes.NewBuffer(nil)
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     stdout,
		stdin:      bytes.NewReader(nil),
		logger:     logger,
	}

	// Try to create project with invalid name (spaces)
	err := cmd.Execute(ctx, cmdCtx, []string{"invalid project"})
	if err == nil {
		t.Fatal("expected error for invalid project name")
	}
	if !strings.Contains(err.Error(), "invalid project name") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectCreateCommand_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	logger := &MockLogger{}
	plugin, _ := task_manager.NewTaskManagerPlugin(logger, tmpDir, nil)

	cmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	ctx := context.Background()
	stdout := bytes.NewBuffer(nil)
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     stdout,
		stdin:      bytes.NewReader(nil),
		logger:     logger,
	}

	// Create project first time
	err := cmd.Execute(ctx, cmdCtx, []string{"test"})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Try to create again
	stdout.Reset()
	err = cmd.Execute(ctx, cmdCtx, []string{"test"})
	if err == nil {
		t.Fatal("expected error when creating duplicate project")
	}
	if !strings.Contains(err.Error(), "project already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectListCommand(t *testing.T) {
	tmpDir := t.TempDir()
	logger := &MockLogger{}
	plugin, _ := task_manager.NewTaskManagerPlugin(logger, tmpDir, nil)

	// Create a few projects
	createCmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	ctx := context.Background()

	for _, name := range []string{"project1", "project2", "project3"} {
		stdout := bytes.NewBuffer(nil)
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     stdout,
		stdin:      bytes.NewReader(nil),
		logger:     logger,
	}
		if err := createCmd.Execute(ctx, cmdCtx, []string{name}); err != nil {
			t.Fatalf("failed to create project %s: %v", name, err)
		}
	}

	// List projects
	listCmd := &task_manager.ProjectListCommand{Plugin: plugin}
	stdout := bytes.NewBuffer(nil)
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     stdout,
		stdin:      bytes.NewReader(nil),
		logger:     logger,
	}

	err := listCmd.Execute(ctx, cmdCtx, []string{})
	if err != nil {
		t.Fatalf("failed to list projects: %v", err)
	}

	output := stdout.String()

	// Verify all projects are listed
	for _, name := range []string{"project1", "project2", "project3"} {
		if !strings.Contains(output, name) {
			t.Errorf("project %s not found in output: %s", name, output)
		}
	}

	// Verify total count
	if !strings.Contains(output, "Total: 3 project(s)") {
		t.Errorf("unexpected total count in output: %s", output)
	}
}

func TestProjectSwitchCommand(t *testing.T) {
	tmpDir := t.TempDir()
	logger := &MockLogger{}
	plugin, _ := task_manager.NewTaskManagerPlugin(logger, tmpDir, nil)

	ctx := context.Background()

	// Create a project
	createCmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	createCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}
	if err := createCmd.Execute(ctx, createCtx, []string{"test"}); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Switch to project
	switchCmd := &task_manager.ProjectSwitchCommand{Plugin: plugin}
	switchCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}

	err := switchCmd.Execute(ctx, switchCtx, []string{"test"})
	if err != nil {
		t.Fatalf("failed to switch project: %v", err)
	}

	// Verify output
	output := switchCtx.stdout.String()
	if !strings.Contains(output, "Switched to project: test") {
		t.Errorf("unexpected output: %s", output)
	}

	// Verify active-project.txt was created
	activeFile := filepath.Join(tmpDir, ".darwinflow", "active-project.txt")
	data, err := os.ReadFile(activeFile)
	if err != nil {
		t.Fatalf("failed to read active project file: %v", err)
	}
	if strings.TrimSpace(string(data)) != "test" {
		t.Errorf("unexpected active project: %s", string(data))
	}
}

func TestProjectShowCommand(t *testing.T) {
	tmpDir := t.TempDir()
	logger := &MockLogger{}
	plugin, _ := task_manager.NewTaskManagerPlugin(logger, tmpDir, nil)

	ctx := context.Background()

	// Create and switch to a project
	createCmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	createCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}
	if err := createCmd.Execute(ctx, createCtx, []string{"my-project"}); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	switchCmd := &task_manager.ProjectSwitchCommand{Plugin: plugin}
	switchCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}
	if err := switchCmd.Execute(ctx, switchCtx, []string{"my-project"}); err != nil {
		t.Fatalf("failed to switch project: %v", err)
	}

	// Show active project
	showCmd := &task_manager.ProjectShowCommand{Plugin: plugin}
	showCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}

	err := showCmd.Execute(ctx, showCtx, []string{})
	if err != nil {
		t.Fatalf("failed to show project: %v", err)
	}

	output := showCtx.stdout.String()
	if !strings.Contains(output, "Active project: my-project") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestProjectDeleteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	logger := &MockLogger{}
	plugin, _ := task_manager.NewTaskManagerPlugin(logger, tmpDir, nil)

	ctx := context.Background()

	// Create two projects
	createCmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	for _, name := range []string{"project1", "project2"} {
		stdout := bytes.NewBuffer(nil)
	cmdCtx := &mockCommandContext{
		workingDir: tmpDir,
		stdout:     stdout,
		stdin:      bytes.NewReader(nil),
		logger:     logger,
	}
		if err := createCmd.Execute(ctx, cmdCtx, []string{name}); err != nil {
			t.Fatalf("failed to create project %s: %v", name, err)
		}
	}

	// Switch to project1 (so we can delete project2)
	switchCmd := &task_manager.ProjectSwitchCommand{Plugin: plugin}
	switchCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}
	if err := switchCmd.Execute(ctx, switchCtx, []string{"project1"}); err != nil {
		t.Fatalf("failed to switch project: %v", err)
	}

	// Delete project2 with --force
	deleteCmd := &task_manager.ProjectDeleteCommand{Plugin: plugin}
	deleteCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}

	err := deleteCmd.Execute(ctx, deleteCtx, []string{"project2", "--force"})
	if err != nil {
		t.Fatalf("failed to delete project: %v", err)
	}

	// Verify project directory was deleted
	projectDir := filepath.Join(tmpDir, ".darwinflow", "projects", "project2")
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		t.Fatal("project directory was not deleted")
	}

	// Verify output
	output := deleteCtx.stdout.String()
	if !strings.Contains(output, "Project deleted successfully: project2") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestProjectDeleteCommand_CannotDeleteActive(t *testing.T) {
	tmpDir := t.TempDir()
	logger := &MockLogger{}
	plugin, _ := task_manager.NewTaskManagerPlugin(logger, tmpDir, nil)

	ctx := context.Background()

	// Create project
	createCmd := &task_manager.ProjectCreateCommand{Plugin: plugin}
	createCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}
	if err := createCmd.Execute(ctx, createCtx, []string{"test"}); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}

	// Switch to it
	switchCmd := &task_manager.ProjectSwitchCommand{Plugin: plugin}
	switchCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}
	if err := switchCmd.Execute(ctx, switchCtx, []string{"test"}); err != nil {
		t.Fatalf("failed to switch project: %v", err)
	}

	// Try to delete active project
	deleteCmd := &task_manager.ProjectDeleteCommand{Plugin: plugin}
	deleteCtx := &mockCommandContext{
		stdin:  bytes.NewBuffer(nil),
		stdout: bytes.NewBuffer(nil),
	}

	err := deleteCmd.Execute(ctx, deleteCtx, []string{"test", "--force"})
	if err == nil {
		t.Fatal("expected error when deleting active project")
	}
	if !strings.Contains(err.Error(), "cannot delete active project") {
		t.Errorf("unexpected error: %v", err)
	}
}
